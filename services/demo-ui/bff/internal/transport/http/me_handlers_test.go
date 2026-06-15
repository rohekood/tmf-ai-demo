package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tmf/services/demo-ui/bff/internal/auth"
)

// callRecorder captures the RPC calls a handler makes and lets a test script the
// reply per routing key.
type callRecorder struct {
	calls   []string
	replies map[string][]byte
}

func (c *callRecorder) handle(_ context.Context, _, routingKey string, _ any, _ map[string]any) ([]byte, error) {
	c.calls = append(c.calls, routingKey)
	if b, ok := c.replies[routingKey]; ok {
		return b, nil
	}
	return []byte(`[]`), nil
}

func newMeHandler(rec *callRecorder) *MeHandler {
	return NewMeHandler(&MockRPCClient{CallRPCFunc: rec.handle})
}

func withEmail(req *http.Request, email string) *http.Request {
	return req.WithContext(auth.ContextWithEmail(req.Context(), email))
}

func TestResolveCustomer_Ready(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[{"id":"party-1","@type":"Individual"}]`),
		queryCustomerSearch: []byte(`[{"id":"cust-1","partyId":"party-1"}]`),
	}}
	h := newMeHandler(rec)

	req := withEmail(httptest.NewRequest("GET", "/api/me/customer", nil), "user@example.com")
	rr := httptest.NewRecorder()
	h.ResolveCustomer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var resp meResolveResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != provisionStatusReady {
		t.Errorf("status = %q, want ready", resp.Status)
	}
	if resp.PartyID != "party-1" {
		t.Errorf("partyId = %q, want party-1", resp.PartyID)
	}
}

func TestResolveCustomer_NeedsParty(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch: []byte(`[]`),
	}}
	h := newMeHandler(rec)

	req := withEmail(httptest.NewRequest("GET", "/api/me/customer", nil), "new@example.com")
	rr := httptest.NewRecorder()
	h.ResolveCustomer(rr, req)

	var resp meResolveResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != provisionStatusNeedsParty {
		t.Errorf("status = %q, want needs_party", resp.Status)
	}
}

func TestResolveCustomer_NeedsCustomer(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[{"id":"party-9"}]`),
		queryCustomerSearch: []byte(`[]`),
	}}
	h := newMeHandler(rec)

	req := withEmail(httptest.NewRequest("GET", "/api/me/customer", nil), "halfway@example.com")
	rr := httptest.NewRecorder()
	h.ResolveCustomer(rr, req)

	var resp meResolveResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != provisionStatusNeedsCustomer || resp.PartyID != "party-9" {
		t.Errorf("got status=%q partyId=%q, want needs_customer/party-9", resp.Status, resp.PartyID)
	}
}

func TestResolveCustomer_NoEmail(t *testing.T) {
	h := newMeHandler(&callRecorder{replies: map[string][]byte{}})
	req := httptest.NewRequest("GET", "/api/me/customer", nil) // no email in context
	rr := httptest.NewRecorder()
	h.ResolveCustomer(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestProvision_CreatesPartyThenCustomer(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[]`),         // no party yet
		queryCustomerSearch: []byte(`[]`),         // no customer yet
		cmdPartyCreate:      []byte(`{"id":"p"}`), // party created
		cmdCustomerOnboard:  []byte(`{"id":"c","partyId":"p"}`),
	}}
	h := newMeHandler(rec)

	body := `{"givenName":"Jane","familyName":"Doe"}`
	req := withEmail(httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body)), "jane@example.com")
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
	if !hasCall(rec.calls, cmdPartyCreate) || !hasCall(rec.calls, cmdCustomerOnboard) {
		t.Errorf("expected both party create and customer onboard, got %v", rec.calls)
	}
}

func TestProvision_IdempotentWhenPartyExistsNoCustomer(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[{"id":"existing-party"}]`), // party already exists
		queryCustomerSearch: []byte(`[]`),                        // but no customer
		cmdCustomerOnboard:  []byte(`{"id":"c","partyId":"existing-party"}`),
	}}
	h := newMeHandler(rec)

	body := `{"givenName":"Jane","familyName":"Doe"}`
	req := withEmail(httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body)), "jane@example.com")
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if hasCall(rec.calls, cmdPartyCreate) {
		t.Error("party create should NOT be called when party already exists")
	}
	if !hasCall(rec.calls, cmdCustomerOnboard) {
		t.Error("customer onboard should be called")
	}
}

func TestProvision_ReadyWhenCustomerExists(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[{"id":"existing-party"}]`),
		queryCustomerSearch: []byte(`[{"id":"existing-cust","partyId":"existing-party"}]`),
	}}
	h := newMeHandler(rec)

	body := `{"givenName":"Jane","familyName":"Doe"}`
	req := withEmail(httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body)), "jane@example.com")
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if hasCall(rec.calls, cmdPartyCreate) || hasCall(rec.calls, cmdCustomerOnboard) {
		t.Errorf("no creates expected when customer exists, got %v", rec.calls)
	}
}

func TestProvision_FallbackToBodyEmail(t *testing.T) {
	rec := &callRecorder{replies: map[string][]byte{
		queryPartySearch:    []byte(`[]`),
		queryCustomerSearch: []byte(`[]`),
		cmdPartyCreate:      []byte(`{"id":"p"}`),
		cmdCustomerOnboard:  []byte(`{"id":"c"}`),
	}}
	h := newMeHandler(rec)

	// No email in context; provided in body (demo fallback).
	body := `{"email":"body@example.com","givenName":"Jane","familyName":"Doe"}`
	req := httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
}

func TestProvision_RaceCreatePartyConflict(t *testing.T) {
	// Simulate a concurrent winner: the first party lookup finds nothing, the
	// party create fails (unique-email constraint), and the re-resolve then finds
	// the party another request created. Provisioning should still succeed.
	partySearchCalls := 0
	client := &MockRPCClient{
		CallRPCFunc: func(_ context.Context, _, routingKey string, _ any, _ map[string]any) ([]byte, error) {
			switch routingKey {
			case queryPartySearch:
				partySearchCalls++
				if partySearchCalls == 1 {
					return []byte(`[]`), nil // not found yet
				}
				return []byte(`[{"id":"race-party"}]`), nil // found on re-resolve
			case cmdPartyCreate:
				return []byte(`{"error":"duplicate key value violates unique constraint"}`), nil
			case queryCustomerSearch:
				return []byte(`[]`), nil
			case cmdCustomerOnboard:
				return []byte(`{"id":"c","partyId":"race-party"}`), nil
			}
			return []byte(`[]`), nil
		},
	}
	h := NewMeHandler(client)

	body := `{"givenName":"Jane","familyName":"Doe"}`
	req := withEmail(httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body)), "jane@example.com")
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
	if partySearchCalls < 2 {
		t.Errorf("expected a re-resolve after create conflict, party searches = %d", partySearchCalls)
	}
}

func TestProvision_MissingNames(t *testing.T) {
	h := newMeHandler(&callRecorder{replies: map[string][]byte{}})
	body := `{}`
	req := withEmail(httptest.NewRequest("POST", "/api/me/provision", strings.NewReader(body)), "jane@example.com")
	rr := httptest.NewRecorder()
	h.Provision(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func hasCall(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
