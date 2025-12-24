package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	// Customer Management RabbitMQ topics
	customerExchange    = "tmf.customer"
	cmdCustomerOnboard  = "cmd.customer.onboard"
	cmdCustomerUpdate   = "cmd.customer.update"
	cmdCustomerDelete   = "cmd.customer.delete"
	queryCustomerGet    = "query.customer.get"
	queryCustomerSearch = "query.customer.search"

	customerRPCTimeout = 10 * time.Second
)

// Handler handles Customer Management API endpoints
type Handler struct {
	rpcClient    RPCClient
	partyHandler *PartyHandler
	hub          *Hub
}

// NewHandler creates a new Handler
func NewHandler(client RPCClient, hub *Hub) *Handler {
	return &Handler{
		rpcClient:    client,
		partyHandler: NewPartyHandler(client),
		hub:          hub,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	// WebSocket route (outside /api to avoid auth if needed, or included)
	// For demo, we keep it unsecured or basic auth.
	r.Get("/ws/debug", func(w http.ResponseWriter, r *http.Request) {
		h.hub.ServeWs(w, r)
	})

	r.Route("/api", func(r chi.Router) {
		// Customer routes
		r.Route("/customers", func(r chi.Router) {
			r.Get("/", h.SearchCustomers)
			r.Post("/", h.CreateCustomer)
			r.Get("/{id}", h.GetCustomer)
			r.Put("/{id}", h.UpdateCustomer)
			r.Delete("/{id}", h.DeleteCustomer)
		})

		// Party routes (delegated to PartyHandler)
		h.partyHandler.RegisterRoutes(r)
	})
}

func getHeaders(r *http.Request) map[string]interface{} {
	headers := make(map[string]interface{})
	if auth := r.Header.Get("Authorization"); auth != "" {
		headers["Authorization"] = auth
	}
	// Extract user from context if available (set by Auth middleware)
	if user, ok := r.Context().Value("user").(string); ok {
		headers["user"] = user
	}
	return headers
}

// SearchCustomers handles GET /api/customers
// Query params: name, status, partyId
func (h *Handler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	payload := map[string]string{}

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
	w.Write(responseBytes)
}

// GetCustomer handles GET /api/customers/:id
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

// CreateCustomer handles POST /api/customers
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerOnboard, payload, getHeaders(r))
	if err != nil {
		slog.Error("error creating customer", "error", err)
		http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseBytes)
}

// UpdateCustomer handles PUT /api/customers/:id
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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
	w.Write(responseBytes)
}

// DeleteCustomer handles DELETE /api/customers/:id
func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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
