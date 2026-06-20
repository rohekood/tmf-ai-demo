package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"tmf/services/demo-ui/bff/internal/auth"
)

// normalizeEmail lower-cases and trims an email so it matches the
// case-insensitive uniqueness constraint and is resolved consistently.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Provisioning status discriminators returned by GET /api/me/customer.
const (
	provisionStatusReady         = "ready"          // customer already exists
	provisionStatusNeedsParty    = "needs_party"    // neither party nor customer exists
	provisionStatusNeedsCustomer = "needs_customer" // party exists, customer does not
)

// MeHandler serves the "current user" provisioning endpoints. It resolves and
// creates the Customer (and underlying Party) that backs the logged-in user,
// linked by email.
type MeHandler struct {
	rpcClient RPCClient
}

// NewMeHandler creates a MeHandler.
func NewMeHandler(client RPCClient) *MeHandler {
	return &MeHandler{rpcClient: client}
}

// resolveEmail returns the caller's verified email, taken only from the signed
// access-token claim (`https://tmf-demo/email`). It never reads a client-supplied
// value. Returns false when no verified email is present, in which case the
// caller fails closed with 401.
func (h *MeHandler) resolveEmail(r *http.Request) (string, bool) {
	if email, ok := auth.EmailFromContext(r.Context()); ok {
		return normalizeEmail(email), true
	}
	return "", false
}

// RegisterRoutes registers the provisioning routes. These are NOT in the public
// allowlist, so the auth middleware requires a valid JWT for them.
func (h *MeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/me/customer", h.ResolveCustomer)
	mux.HandleFunc("POST /api/me/provision", h.Provision)
}

type meResolveResponse struct {
	Status   string          `json:"status"`
	PartyID  string          `json:"partyId,omitempty"`
	Customer json.RawMessage `json:"customer,omitempty"`
}

// provisionRequest carries only profile fields. Identity (email) is taken from
// the verified token, never from the body.
type provisionRequest struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Phone      string `json:"phone,omitempty"`
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	Postcode   string `json:"postcode,omitempty"`
	Country    string `json:"country,omitempty"`
}

// ResolveCustomer resolves the caller's customer by email:
//
//	party-by-email -> customer-by-party_id
//
// and reports whether provisioning is still needed.
func (h *MeHandler) ResolveCustomer(w http.ResponseWriter, r *http.Request) {
	email, ok := h.resolveEmail(r)
	if !ok {
		http.Error(w, "Verified email required", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()
	headers := getHeaders(r)

	partyID, found, err := h.findPartyByEmail(ctx, email, headers)
	if err != nil {
		slog.Error("resolve customer: party lookup failed", "error", err)
		http.Error(w, "Failed to resolve party: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, meResolveResponse{Status: provisionStatusNeedsParty})
		return
	}

	customer, found, err := h.findCustomerByPartyID(ctx, partyID, headers)
	if err != nil {
		slog.Error("resolve customer: customer lookup failed", "error", err)
		http.Error(w, "Failed to resolve customer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		writeJSON(w, http.StatusOK, meResolveResponse{Status: provisionStatusNeedsCustomer, PartyID: partyID})
		return
	}

	writeJSON(w, http.StatusOK, meResolveResponse{Status: provisionStatusReady, PartyID: partyID, Customer: customer})
}

// Provision creates the Party (if missing) and Customer (if missing) for the
// caller, linked by email. It is idempotent: each step re-resolves by
// email/party_id, so a retry after a partial failure completes the remaining
// step without duplicating.
func (h *MeHandler) Provision(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req provisionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Identity comes only from a verified source (signed token claim, else a
	// trusted /userinfo lookup) — never from the request body. Fail closed if it
	// cannot be established.
	email, ok := h.resolveEmail(r)
	if !ok {
		http.Error(w, "Verified email required", http.StatusUnauthorized)
		return
	}

	if req.GivenName == "" || req.FamilyName == "" {
		http.Error(w, "givenName and familyName are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), customerRPCTimeout)
	defer cancel()
	headers := getHeaders(r)

	// 1. Find-or-create party (keyed by email).
	partyID, found, err := h.findPartyByEmail(ctx, email, headers)
	if err != nil {
		slog.Error("provision: party lookup failed", "error", err)
		http.Error(w, "Failed to look up party: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		partyID = uuid.New().String()
		if err := h.createParty(ctx, partyID, email, req, headers); err != nil {
			// A concurrent provisioning request may have created the party first,
			// tripping the unique-email constraint. Re-resolve by email and
			// continue if it now exists; otherwise surface the error.
			slog.Warn("provision: party create failed, re-resolving by email", "error", err)
			existingID, exists, lookupErr := h.findPartyByEmail(ctx, email, headers)
			if lookupErr != nil || !exists {
				http.Error(w, "Failed to create party: "+err.Error(), http.StatusInternalServerError)
				return
			}
			partyID = existingID
		}
	}

	// 2. Find-or-create customer (keyed by party_id).
	customer, found, err := h.findCustomerByPartyID(ctx, partyID, headers)
	if err != nil {
		slog.Error("provision: customer lookup failed", "error", err)
		http.Error(w, "Failed to look up customer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		customer, err = h.createCustomer(ctx, partyID, email, req, headers)
		if err != nil {
			slog.Error("provision: customer create failed", "error", err)
			http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusCreated, meResolveResponse{Status: provisionStatusReady, PartyID: partyID, Customer: customer})
}

// findPartyByEmail returns the id of the party owning the given email, if any.
func (h *MeHandler) findPartyByEmail(ctx context.Context, email string, headers map[string]any) (string, bool, error) {
	respBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, queryPartySearch, map[string]any{"email": email}, headers)
	if err != nil {
		return "", false, err
	}
	first, ok, err := firstResult(respBytes)
	if err != nil || !ok {
		return "", false, err
	}
	id, _ := first["id"].(string)
	return id, id != "", nil
}

// findCustomerByPartyID returns the customer for the given party id, if any.
func (h *MeHandler) findCustomerByPartyID(ctx context.Context, partyID string, headers map[string]any) (json.RawMessage, bool, error) {
	respBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, queryCustomerSearch, map[string]any{"partyId": partyID}, headers)
	if err != nil {
		return nil, false, err
	}
	var customers []json.RawMessage
	if err := json.Unmarshal(respBytes, &customers); err != nil {
		return nil, false, err
	}
	if len(customers) == 0 {
		return nil, false, nil
	}
	return customers[0], true, nil
}

// createParty issues cmd.party.create for an Individual with the email (and any
// supplied phone/address) as contact mediums.
func (h *MeHandler) createParty(ctx context.Context, partyID, email string, req provisionRequest, headers map[string]any) error {
	contactMediums := []map[string]any{
		{"mediumType": "email", "value": email, "preferred": true},
	}
	if req.Phone != "" {
		contactMediums = append(contactMediums, map[string]any{"mediumType": "phone", "value": req.Phone})
	}
	if req.Street != "" || req.City != "" || req.Postcode != "" || req.Country != "" {
		contactMediums = append(contactMediums, map[string]any{
			"mediumType": "address",
			"street1":    req.Street,
			"city":       req.City,
			"postcode":   req.Postcode,
			"country":    req.Country,
		})
	}

	payload := map[string]any{
		"@type":          "Individual",
		"id":             partyID,
		"givenName":      req.GivenName,
		"familyName":     req.FamilyName,
		"contactMediums": contactMediums,
	}

	respBytes, err := h.rpcClient.CallRPC(ctx, partyExchange, cmdPartyCreate, payload, headers)
	if err != nil {
		return err
	}
	return replyError(respBytes)
}

// createCustomer issues cmd.customer.onboard referencing the party.
func (h *MeHandler) createCustomer(ctx context.Context, partyID, email string, req provisionRequest, headers map[string]any) (json.RawMessage, error) {
	payload := map[string]any{
		"id":        uuid.New().String(),
		"name":      req.GivenName + " " + req.FamilyName,
		"partyId":   partyID,
		"partyType": "Individual",
		"contactMediums": []map[string]any{
			{"mediumType": "email", "value": email, "preferred": true},
		},
	}

	respBytes, err := h.rpcClient.CallRPC(ctx, customerExchange, cmdCustomerOnboard, payload, headers)
	if err != nil {
		return nil, err
	}
	if err := replyError(respBytes); err != nil {
		return nil, err
	}
	return respBytes, nil
}

// firstResult unmarshals an RPC search reply (a JSON array) and returns the
// first element as a generic map.
func firstResult(respBytes []byte) (map[string]any, bool, error) {
	var arr []map[string]any
	if err := json.Unmarshal(respBytes, &arr); err != nil {
		return nil, false, err
	}
	if len(arr) == 0 {
		return nil, false, nil
	}
	return arr[0], true, nil
}

// replyError returns a non-nil error when an RPC reply is an {"error": "..."}
// envelope rather than a successful entity.
func replyError(respBytes []byte) error {
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &env); err == nil && env.Error != "" {
		return &rpcReplyError{msg: env.Error}
	}
	return nil
}

type rpcReplyError struct{ msg string }

func (e *rpcReplyError) Error() string { return e.msg }

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
