package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	orderExchange = "ex.domain.market" // Using default market exchange

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

	ctx, cancel := context.WithTimeout(r.Context(), qualRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, cmdQualEligibilityCheck, payload, getHeaders(r))
	if err != nil {
		slog.Error("error checking qualification", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Qualification check timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "Failed to check qualification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseBytes)
}

func (h *OrderHandler) GetQualificationSession(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("sessionId")
	if sessionId == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": sessionId} // Some systems expect {"sessionId": "..."}

	ctx, cancel := context.WithTimeout(r.Context(), qualRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, queryQualSessionGet, payload, getHeaders(r))
	if err != nil {
		// Could be 422 expired session
		slog.Error("error getting qualification session", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error": "SESSION_EXPIRED"}`))
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

	ctx, cancel := context.WithTimeout(r.Context(), cartRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, cmdCartItemAdd, payload, getHeaders(r))
	if err != nil {
		slog.Error("error adding cart item", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cartId := r.PathValue("cartId")
	if cartId == "" {
		http.Error(w, "Cart ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": cartId}

	ctx, cancel := context.WithTimeout(r.Context(), cartRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, queryCartGet, payload, getHeaders(r))
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

	_, err := h.rpcClient.CallRPC(ctx, orderExchange, cmdCartItemRemove, payload, getHeaders(r))
	if err != nil {
		slog.Error("error removing cart item", "error", err)
		http.Error(w, "Failed to remove cart item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Checkout Handlers ---

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

	ctx, cancel := context.WithTimeout(r.Context(), sagaRPCTimeout)
	defer cancel()

	// Checkout is fire-and-forget in analysis but UI might expect sagaId in response
	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, cmdOrderCheckoutSubmit, payload, getHeaders(r))
	if err != nil {
		slog.Error("error checking out", "error", err)
		http.Error(w, "Failed to initiate checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write(responseBytes)
}

func (h *OrderHandler) GetSagaStatus(w http.ResponseWriter, r *http.Request) {
	sagaId := r.PathValue("sagaId")
	if sagaId == "" {
		http.Error(w, "Saga ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": sagaId, "sagaId": sagaId} // Pass both to be safe depending on backend mapping

	ctx, cancel := context.WithTimeout(r.Context(), sagaRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, orderExchange, queryPocvSagaGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting saga status", "error", err)
		http.Error(w, "Failed to get saga status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}
