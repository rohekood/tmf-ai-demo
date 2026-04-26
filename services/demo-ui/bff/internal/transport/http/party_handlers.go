package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	// Party Management RabbitMQ topics
	partyExchange    = "tmf.party"
	cmdPartyCreate   = "cmd.party.create"
	cmdPartyUpdate   = "cmd.party.update"
	cmdPartyPatch    = "cmd.party.patch"
	cmdPartyDelete   = "cmd.party.delete"
	queryPartyGet    = "query.party.get"
	queryPartySearch = "query.party.search"

	defaultRPCTimeout = 10 * time.Second
)

// PartyHandler handles Party Management API endpoints
type PartyHandler struct {
	rpcClient RPCClient
}

// NewPartyHandler creates a new PartyHandler
func NewPartyHandler(client RPCClient) *PartyHandler {
	return &PartyHandler{rpcClient: client}
}

// RegisterRoutes registers all party routes
func (h *PartyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/parties", h.SearchParties)
	mux.HandleFunc("POST /api/parties", h.CreateParty)
	mux.HandleFunc("GET /api/parties/{id}", h.GetParty)
	mux.HandleFunc("PUT /api/parties/{id}", h.UpdateParty)
	mux.HandleFunc("PATCH /api/parties/{id}", h.PatchParty)
	mux.HandleFunc("DELETE /api/parties/{id}", h.DeleteParty)
}

// SearchParties handles GET /api/parties
// Query params: givenName, familyName, tradingName, type
func (h *PartyHandler) SearchParties(w http.ResponseWriter, r *http.Request) {
	payload := map[string]*string{}

	if v := r.URL.Query().Get("search"); v != "" {
		payload["search"] = &v
	}
	if v := r.URL.Query().Get("givenName"); v != "" {
		payload["givenName"] = &v
	}
	if v := r.URL.Query().Get("name"); v != "" {
		payload["name"] = &v
	}
	if v := r.URL.Query().Get("familyName"); v != "" {
		payload["familyName"] = &v
	}
	if v := r.URL.Query().Get("tradingName"); v != "" {
		payload["tradingName"] = &v
	}
	if v := r.URL.Query().Get("type"); v != "" {
		payload["type"] = &v
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, queryPartySearch, payload, getHeaders(r))
	if err != nil {
		slog.Error("error searching parties", "error", err)
		http.Error(w, "Failed to search parties: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// GetParty handles GET /api/parties/:id
func (h *PartyHandler) GetParty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Party ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, queryPartyGet, payload, getHeaders(r))
	if err != nil {
		slog.Error("error getting party", "error", err)
		http.Error(w, "Failed to get party: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// CreateParty handles POST /api/parties
// Body should contain the party data with @type field indicating Individual or Organization
func (h *PartyHandler) CreateParty(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, cmdPartyCreate, payload, getHeaders(r))
	if err != nil {
		slog.Error("error creating party", "error", err)
		http.Error(w, "Failed to create party: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(responseBytes)
}

// UpdateParty handles PUT /api/parties/:id
// Full replacement of party data
func (h *PartyHandler) UpdateParty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Party ID is required", http.StatusBadRequest)
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

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Ensure ID is set in payload
	payload["id"] = id

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, cmdPartyUpdate, payload, getHeaders(r))
	if err != nil {
		slog.Error("error updating party", "error", err)
		http.Error(w, "Failed to update party: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// PatchParty handles PATCH /api/parties/:id
// Partial update of party data
func (h *PartyHandler) PatchParty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Party ID is required", http.StatusBadRequest)
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

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Ensure ID is set in payload
	payload["id"] = id

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	responseBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, cmdPartyPatch, payload, getHeaders(r))
	if err != nil {
		slog.Error("error patching party", "error", err)
		http.Error(w, "Failed to patch party: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBytes)
}

// DeleteParty handles DELETE /api/parties/:id
func (h *PartyHandler) DeleteParty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Party ID is required", http.StatusBadRequest)
		return
	}

	payload := map[string]string{"id": id}

	ctx, cancel := context.WithTimeout(r.Context(), defaultRPCTimeout)
	defer cancel()

	_, err := h.rpcClient.CallRPC(ctx, partyExchange, cmdPartyDelete, payload, getHeaders(r))
	if err != nil {
		slog.Error("error deleting party", "error", err)
		http.Error(w, "Failed to delete party: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
