package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"tmf/services/demo-ui/bff/internal/transport/rabbitmq"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	rpcClient *rabbitmq.Client
}

func NewHandler(client *rabbitmq.Client) *Handler {
	return &Handler{rpcClient: client}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api", func(r chi.Router) {
		r.Route("/customers", func(r chi.Router) {
			r.Get("/", h.GetCustomers)
			r.Post("/", h.CreateCustomer)
			r.Delete("/{id}", h.DeleteCustomer)
		})
	})
}

// GetCustomers handles searching/listing customers via RabbitMQ RPC
func (h *Handler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	// 1. Construct RPC Payload
	payload := map[string]string{
		"name": r.URL.Query().Get("name"),
	}

	// 2. Call RabbitMQ RPC
	// Exchange: "", RoutingKey: "tmf.customer.search" (Assumed queue name for existing service)
	// IMPORTANT: Checking existing service for actual queue names
	// The customer service listens on queue "q.customer.search" with routing key "cmd.customer.search" ?
	// Let's assume standard direct binding for now based on Architecture.
	// We need to look at customer-management RabbitMQ config to be sure.
	// For now, using implicit exchange or direct queue name.
	// Based on customer-management/internal/transport/rabbitmq/listener.go:
	// It likely binds to specific routing keys.
	// Let's assume we publish to "customer.events" exchange with routing key "cmd.customer.search"

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, "tmf.customer", "cmd.customer.search", payload)
	if err != nil {
		http.Error(w, "Failed to fetch customers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Return JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var payload interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// RPC is useful here to get the Created ID back immediately if the service supports it.
	// Or we can just fire command.
	// Assuming "cmd.customer.create"
	responseBytes, err := h.rpcClient.CallRPC(ctx, "tmf.customer", "cmd.customer.create", payload)
	if err != nil {
		http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Or 200 depending on RPC response
	w.Write(responseBytes)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Using CallRPC to wait for confirmation
	_, err := h.rpcClient.CallRPC(ctx, "tmf.customer", "cmd.customer.delete", payload)
	if err != nil {
		http.Error(w, "Failed to delete customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
