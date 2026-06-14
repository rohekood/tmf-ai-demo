package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	orderExchange = "ex.domain.market"   // Qualification service exchange
	cartExchange  = "ex.domain.commerce" // Shopping Cart service exchange
	pocvExchange  = "ex.domain.order"    // POCV saga service exchange (UC-03)

	cmdQualEligibilityCheck = "cmd.qual.eligibility.check"
	queryQualSessionGet     = "query.qual.session.get"

	cmdCartItemAdd    = "cmd.cart.item.add"
	cmdCartItemRemove = "cmd.cart.item.remove"
	queryCartGet      = "query.cart.get"

	cmdOrderCheckoutSubmit = "cmd.order.checkout.submit"
	queryPocvSagaGet       = "query.pocv.saga.get"

	qualRPCTimeout = 30 * time.Second
	cartRPCTimeout = 15 * time.Second
	sagaRPCTimeout = 30 * time.Second

	// qualPollInterval is how often CheckQualification re-queries the
	// qualification session while waiting for the async scatter-gather to
	// persist a result.
	qualPollInterval = 300 * time.Millisecond
)

// OrderHandler handles ordering workflow API endpoints
type OrderHandler struct {
	rpcClient RPCClient
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(client RPCClient) *OrderHandler {
	return &OrderHandler{rpcClient: client}
}

// RegisterRoutes registers all ordering routes
func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	// Qualification routes (UC-01)
	mux.HandleFunc("POST /api/qualification/check", h.CheckQualification)
	mux.HandleFunc("GET /api/qualification/session/{sessionId}", h.GetQualificationSession)

	// Cart routes (UC-02)
	mux.HandleFunc("POST /api/cart/items", h.AddCartItem)
	mux.HandleFunc("GET /api/cart/{cartId}", h.GetCart)
	mux.HandleFunc("DELETE /api/cart/{cartId}/items/{itemId}", h.RemoveCartItem)

	// Checkout/Saga routes (UC-03)
	mux.HandleFunc("POST /api/orders/checkout", h.Checkout)
	mux.HandleFunc("GET /api/orders/saga/{sagaId}", h.GetSagaStatus)
}

// --- Qualification Handlers ---

func (h *OrderHandler) CheckQualification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Generate the session ID up front. The qualification service stores its
	// result under this ID. We publish the (fire-and-forget) check command and
	// then poll the session via RPC until the async scatter-gather has
	// persisted a result, presenting a synchronous response to the UI.
	sessionID := uuid.New().String()
	payload["sessionId"] = sessionID

	ctx, cancel := context.WithTimeout(r.Context(), qualRPCTimeout)
	defer cancel()

	if err := h.rpcClient.PublishCommand(ctx, orderExchange, cmdQualEligibilityCheck, payload); err != nil {
		slog.Error("error publishing qualification command", "error", err)
		http.Error(w, "Failed to start qualification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.awaitQualificationResult(ctx, sessionID, getHeaders(r))
	if err != nil {
		slog.Error("error awaiting qualification result", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Qualification timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "Failed to check qualification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// qualificationResult mirrors the QualificationResult shape the UI expects from
// POST /api/qualification/check.
type qualificationResult struct {
	SessionID            string          `json:"sessionId"`
	Status               string          `json:"status"`
	QualifiedOffers      json.RawMessage `json:"qualifiedOffers"`
	UnavailabilityReason string          `json:"unavailabilityReason,omitempty"`
}

// awaitQualificationResult polls query.qual.session.get until the qualification
// service has persisted a session for sessionID, then maps it to the response
// the UI expects. The session lookup replies with {"error": ...} until the
// async scatter-gather completes, so an empty status means "not ready yet".
func (h *OrderHandler) awaitQualificationResult(ctx context.Context, sessionID string, headers map[string]any) (*qualificationResult, error) {
	ticker := time.NewTicker(qualPollInterval)
	defer ticker.Stop()

	for {
		respBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, queryQualSessionGet, map[string]string{"sessionId": sessionID}, headers)
		if err != nil {
			return nil, err
		}

		var session struct {
			ID              string          `json:"id"`
			Status          string          `json:"status"`
			QualifiedOffers json.RawMessage `json:"qualifiedOffers"`
		}
		if err := json.Unmarshal(respBytes, &session); err != nil {
			return nil, err
		}

		if session.Status != "" {
			// Normalise a missing/null offers field to an empty array so the UI
			// can safely read qualifiedOffers.length.
			offers := session.QualifiedOffers
			if len(offers) == 0 || string(offers) == "null" {
				offers = json.RawMessage("[]")
			}
			return &qualificationResult{
				SessionID:       session.ID,
				Status:          session.Status,
				QualifiedOffers: offers,
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (h *OrderHandler) GetQualificationSession(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("sessionId")
	if sessionId == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"sessionId": sessionId}

	ctx, cancel := context.WithTimeout(r.Context(), qualRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, queryQualSessionGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting qualification session", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Session lookup timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "Failed to get qualification session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// --- Cart Handlers ---

func (h *OrderHandler) AddCartItem(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Ensure cartId is always a non-empty string before forwarding to the cart
	// service. An empty cartId would cause the service to save a cart with ID ""
	// creating a collision shared across all sessions. A missing or non-string
	// cartId value is treated as absent and a new UUID is generated.
	cartID, ok := payload["cartId"].(string)
	if !ok || cartID == "" {
		cartID = uuid.New().String()
		payload["cartId"] = cartID
	}

	ctx, cancel := context.WithTimeout(r.Context(), cartRPCTimeout)
	defer cancel()

	// cmd.cart.item.add is a fire-and-forget command; the cart service processes
	// it asynchronously and does not publish an RPC reply.
	if err := h.rpcClient.PublishCommand(ctx, cartExchange, cmdCartItemAdd, payload); err != nil {
		slog.Error("error publishing add cart item command", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return the cartId so the UI can persist it in localStorage and use it for
	// subsequent cart queries and the checkout flow.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"cartId": cartID})
}

func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cartId := r.PathValue("cartId")
	if cartId == "" {
		http.Error(w, "Cart ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"cartId": cartId}

	ctx, cancel := context.WithTimeout(r.Context(), cartRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, cartExchange, queryCartGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting cart", "error", err)
		http.Error(w, "Failed to get cart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *OrderHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	cartId := r.PathValue("cartId")
	itemId := r.PathValue("itemId")

	if cartId == "" || itemId == "" {
		http.Error(w, "Cart ID and Item ID are required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{
		"cartId": cartId,
		"itemId": itemId,
	}

	ctx, cancel := context.WithTimeout(r.Context(), cartRPCTimeout)
	defer cancel()

	_, err := h.rpcClient.CallRPC(ctx, cartExchange, cmdCartItemRemove, payload, getHeaders(r))
	if err != nil {
		slog.Error("error removing cart item", "error", err)
		http.Error(w, "Failed to remove cart item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Checkout Handlers ---

// checkoutResponse is the 202 response body for POST /api/orders/checkout.
type checkoutResponse struct {
	SagaID string `json:"sagaId"`
	Status string `json:"status"`
	CartID string `json:"cartId"`
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	cartID, _ := payload["cartId"].(string)
	if cartID == "" {
		http.Error(w, "cartId is required", http.StatusBadRequest)
		return
	}

	// Generate sagaId on the BFF side so it can be returned immediately and
	// included in the command payload for POCV to use as its saga identifier.
	sagaID := uuid.New().String()
	payload["sagaId"] = sagaID

	ctx, cancel := context.WithTimeout(r.Context(), sagaRPCTimeout)
	defer cancel()

	// Fire-and-forget: publish the command without waiting for a reply.
	if err := h.rpcClient.PublishCommand(ctx, pocvExchange, cmdOrderCheckoutSubmit, payload); err != nil {
		slog.Error("error publishing checkout command", "error", err)
		http.Error(w, "Failed to initiate checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := checkoutResponse{
		SagaID: sagaID,
		Status: "PENDING",
		CartID: cartID,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		slog.Error("error marshaling checkout response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write(respBytes)
}

func (h *OrderHandler) GetSagaStatus(w http.ResponseWriter, r *http.Request) {
	sagaId := r.PathValue("sagaId")
	if sagaId == "" {
		http.Error(w, "Saga ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": sagaId}

	ctx, cancel := context.WithTimeout(r.Context(), sagaRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, pocvExchange, queryPocvSagaGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting saga status", "error", err)
		http.Error(w, "Failed to get saga status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}
