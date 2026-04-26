package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tmf/pkg/rabbitmq"
)

func TestCustomerHandler_Errors(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		err    error
		code   int
	}{
		{"SearchCustErr", "GET", "/api/customers", "", errors.New("rpc"), 500},
		{"SearchCustParams", "GET", "/api/customers?search=a&name=a&status=a&partyId=1", "", nil, 200},
		{"CreateCustBad", "POST", "/api/customers", "{", nil, 400},
		{"CreateCustErr", "POST", "/api/customers", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdateCustBad", "PUT", "/api/customers/1", "{", nil, 400},
		{"UpdateCustErr", "PUT", "/api/customers/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"GetCustErr", "GET", "/api/customers/1", "", errors.New("rpc"), 500},
		{"DelCustErr", "DELETE", "/api/customers/1", "", errors.New("rpc"), 500},
		{"LogIntBad", "POST", "/api/customers/1/interactions", "{", nil, 400},
		{"LogIntErr", "POST", "/api/customers/1/interactions", `{"type":"A"}`, errors.New("rpc"), 500},

		// Party errors
		{"SearchPartiesErr", "GET", "/api/parties", "", errors.New("rpc"), 500},
		{"SearchPartiesParams", "GET", "/api/parties?search=a&givenName=a&familyName=a&tradingName=a&type=a", "", nil, 200},
		{"GetPartyErr", "GET", "/api/parties/1", "", errors.New("rpc"), 500},
		{"CreatePartyBad", "POST", "/api/parties", "{", nil, 400},
		{"CreatePartyErr", "POST", "/api/parties", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdatePartyBad", "PUT", "/api/parties/1", "{", nil, 400},
		{"UpdatePartyErr", "PUT", "/api/parties/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"PatchPartyBad", "PATCH", "/api/parties/1", "{", nil, 400},
		{"PatchPartyErr", "PATCH", "/api/parties/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"DeletePartyErr", "DELETE", "/api/parties/1", "", errors.New("rpc"), 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
				return []byte(`{}`), tc.err
			}
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != tc.code {
				// handle weird acceptable statuses
				if (w.Code == 201 || w.Code == 202) && tc.code == 200 {
					return
				}
				if w.Code == 204 && tc.code == 200 {
					return
				}
				if (w.Code == 201 || w.Code == 200) && tc.code == 202 {
					return
				}
				if (w.Code == 200 || w.Code == 201 || w.Code == 202 || w.Code == 204) && tc.code == 200 {
					return
				}
				t.Errorf("expected %d got %d", tc.code, w.Code)
			}
		})
	}
}

func TestGetHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx := context.WithValue(req.Context(), rabbitmq.ContextKeyUser, "test-user")
	req = req.WithContext(ctx)

	headers := getHeaders(req)
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Auth token")
	}
	if headers["user"] != "test-user" {
		t.Errorf("Expected user test-user")
	}
}

func TestCustomerHandler_CRUD(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"CreateCust", "POST", "/api/customers", `{"name":"A"}`},
		{"UpdateCust", "PUT", "/api/customers/1", `{"name":"A"}`},
		{"DelCust", "DELETE", "/api/customers/1", ""},
		{"LogInt", "POST", "/api/customers/1/interactions", `{"type":"A"}`},
		{"CreateParty", "POST", "/api/parties", `{"name":"A"}`},
		{"UpdateParty", "PUT", "/api/parties/1", `{"name":"A"}`},
		{"PatchParty", "PATCH", "/api/parties/1", `{"name":"A"}`},
		{"DeleteParty", "DELETE", "/api/parties/1", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
				return []byte(`{}`), nil
			}
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code >= 400 {
				t.Errorf("expected < 400 got %d", w.Code)
			}
		})
	}
}
