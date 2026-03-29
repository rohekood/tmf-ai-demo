package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"tmf/pkg/rabbitmq"
)

const (
	// Customer Management RabbitMQ topics
	customerExchange          = "tmf.customer"
	cmdCustomerOnboard        = "cmd.customer.onboard"
	cmdCustomerUpdate         = "cmd.customer.update"
	cmdCustomerDelete         = "cmd.customer.delete"
	queryCustomerGet          = "query.customer.get"
	queryCustomerSearch       = "query.customer.search"
	cmdCustomerLogInteraction = "cmd.customer.interaction.log"

	customerRPCTimeout = 10 * time.Second
)

// Handler handles Customer Management API endpoints
type Handler struct {
	rpcClient      RPCClient
	partyHandler   *PartyHandler
	catalogHandler *CatalogHandler
	hub            *Hub
}

// NewHandler creates a new Handler
func NewHandler(client RPCClient, hub *Hub) *Handler {
	return &Handler{
		rpcClient:      client,
		partyHandler:   NewPartyHandler(client),
		catalogHandler: NewCatalogHandler(client),
		hub:            hub,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Customer routes
	mux.HandleFunc("GET /api/customers", h.SearchCustomers)
	mux.HandleFunc("POST /api/customers", h.CreateCustomer)
	mux.HandleFunc("GET /api/customers/{id}", h.GetCustomer)
	mux.HandleFunc("PUT /api/customers/{id}", h.UpdateCustomer)
	mux.HandleFunc("DELETE /api/customers/{id}", h.DeleteCustomer)
	mux.HandleFunc("POST /api/customers/{id}/interactions", h.LogInteraction)

	// Party routes
	h.partyHandler.RegisterRoutes(mux)

	// Catalog routes
	h.catalogHandler.RegisterRoutes(mux)
}

func getHeaders(r *http.Request) map[string]interface{} {
	headers := make(map[string]interface{})
	if auth := r.Header.Get("Authorization"); auth != "" {
		headers["Authorization"] = auth
	}
	// Extract user from context if available (set by Auth middleware)
	if user, ok := r.Context().Value(rabbitmq.ContextKeyUser).(string); ok {
		headers["user"] = user
	} else if user, ok := r.Context().Value(rabbitmq.Key(rabbitmq.HeaderUser)).(string); ok {
		headers["user"] = user
	}
	return headers
}

// SearchCustomers handles GET /api/customers
// Query params: name, status, partyId
func (h *Handler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	payload := map[string]string{}

	if v := r.URL.Query().Get("search"); v != "" {
		payload["search"] = v
	}
	if v := r.URL.Query().Get("name"); v != "" {
		payload["name"] = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		payload["status"] = v
	}
	if v := r.URL.Query().Get("partyId"); v != "" {
		payload["partyId"] = v
	}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, queryCustomerSearch, payload, getHeaders(r))
	if err != nil {
		slog.Error("error searching customers", "error", err)
		http.Error(w, "Failed to search customers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// GetCustomer handles GET /api/customers/:id
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, queryCustomerGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting customer", "error", err)
		http.Error(w, "Failed to get customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich with Party Name if possible
	var customer map[string]interface{}
	if err := json.Unmarshal(responseBytes, &customer); err == nil {
		if partyID, ok := customer["partyId"].(string); ok && partyID != "" {
			partyPayload := map[string]string{"id": partyID}
			partyBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, queryPartyGet, partyPayload, getHeaders(r))
			if err == nil {
				var party map[string]interface{}
				if err := json.Unmarshal(partyBytes, &party); err == nil {
					pType, _ := party["@type"].(string)
					customer["partyType"] = pType

					derivedName := ""
					switch pType {
					case "Individual":
						givenName, _ := party["givenName"].(string)
						familyName, _ := party["familyName"].(string)
						derivedName = givenName + " " + familyName
					case "Organization":
						derivedName, _ = party["tradingName"].(string)
					}
					customer["partyName"] = derivedName

					// Re-marshal
					if enrichedBytes, err := json.Marshal(customer); err == nil {
						responseBytes = enrichedBytes
					}
				}
			} else {
				slog.Warn("failed to fetch party for enrichment", "party_id", partyID, "error", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// CreateCustomer handles POST /api/customers
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	// Name Derivation Logic:
	// If partyId is present but name is missing, fetch party and derive name.
	partyID, _ := payload["partyId"].(string)
	name, _ := payload["name"].(string)

	if partyID != "" && name == "" {
		partyPayload := map[string]string{"id": partyID}
		partyBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, queryPartyGet, partyPayload, getHeaders(r))
		if err != nil {
			// If we fail to get party, we log but maybe we should fail?
			// TMF says party must exist. If we can't check it, we can't derive name.
			// Let's assume we proceed and let backend validate party existence,
			// but we can't set name.
			slog.Warn("failed to fetch party for name derivation", "party_id", partyID, "error", err)
		} else {
			var party map[string]interface{}
			if err := json.Unmarshal(partyBytes, &party); err == nil {
				derivedName := ""
				pType, _ := party["@type"].(string)
				switch pType {
				case "Individual":
					givenName, _ := party["givenName"].(string)
					familyName, _ := party["familyName"].(string)
					derivedName = givenName + " " + familyName
				case "Organization":
					derivedName, _ = party["tradingName"].(string)
				}

				if derivedName != "" {
					payload["name"] = derivedName
					slog.Info("derived customer name from party", "party_id", partyID, "derived_name", derivedName)
				}
			}
		}
	}

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerOnboard, payload, getHeaders(r))
	if err != nil {
		slog.Error("error creating customer", "error", err)
		http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(responseBytes)
}

// UpdateCustomer handles PUT /api/customers/:id
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Ensure ID is set in payload
	payload["id"] = id

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerUpdate, payload, getHeaders(r))
	if err != nil {
		slog.Error("error updating customer", "error", err)
		http.Error(w, "Failed to update customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// DeleteCustomer handles DELETE /api/customers/:id
func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	_, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerDelete, payload, getHeaders(r))
	if err != nil {
		slog.Error("error deleting customer", "error", err)
		http.Error(w, "Failed to delete customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LogInteraction handles POST /api/customers/:id/interactions
func (h *Handler) LogInteraction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Ensure Customer ID is set
	payload["customerId"] = id
	// Ensure ID is generated if not present, though backend might handle it.
	// But let's leave it to backend or payload content.

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerLogInteraction, payload, getHeaders(r))
	if err != nil {
		slog.Error("error logging interaction", "error", err)
		http.Error(w, "Failed to log interaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(responseBytes)
}
