package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHandler_SearchCustomers(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if exchange != "tmf.customer" || routingKey != "query.customer.search" {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`[{"id":"1", "name":"Customer 1"}]`), nil
		}

		req := httptest.NewRequest("GET", "/api/customers?name=test", nil)
		w := httptest.NewRecorder()

		handler.SearchCustomers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Customer 1") {
			t.Errorf("Expected response to contain 'Customer 1', got %s", w.Body.String())
		}
	})

	t.Run("RPC Error", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("RPC failed")
		}

		req := httptest.NewRequest("GET", "/api/customers", nil)
		w := httptest.NewRecorder()

		handler.SearchCustomers(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestHandler_GetCustomer(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	r := chi.NewRouter()
	r.Get("/api/customers/{id}", handler.GetCustomer)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"1", "name":"Customer 1"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/customers/123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})
}

func TestHandler_CreateCustomer_DerivesName(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "query.party.get" {
			return []byte(`{"id":"p1", "@type":"Individual", "givenName":"John", "familyName":"Doe"}`), nil
		}
		if routingKey == "cmd.customer.onboard" {
			p := payload.(map[string]interface{})
			if p["name"] != "John Doe" {
				t.Errorf("Expected name to be derived as 'John Doe', got %v", p["name"])
			}
			return []byte(`{"id":"c1", "name":"John Doe"}`), nil
		}
		return nil, errors.New("unexpected call")
	}

	body := `{"partyId":"p1"}` // No name provided
	req := httptest.NewRequest("POST", "/api/customers", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateCustomer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", w.Code)
	}
}

func TestHandler_CreateCustomer_RespectsProvidedName(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "cmd.customer.onboard" {
			p := payload.(map[string]interface{})
			if p["name"] != "Custom Name" {
				t.Errorf("Expected name to be 'Custom Name', got %v", p["name"])
			}
			return []byte(`{"id":"c1", "name":"Custom Name"}`), nil
		}
		return nil, errors.New("unexpected call")
	}

	body := `{"partyId":"p1", "name":"Custom Name"}`
	req := httptest.NewRequest("POST", "/api/customers", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateCustomer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", w.Code)
	}
}

func TestHandler_GetCustomer_EnrichesPartyDetails(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "query.customer.get" {
			return []byte(`{"id":"c1", "name":"Cust", "partyId":"p1"}`), nil
		}
		if routingKey == "query.party.get" {
			return []byte(`{"id":"p1", "@type":"Individual", "givenName":"John", "familyName":"Doe"}`), nil
		}
		return nil, errors.New("unexpected call: " + routingKey)
	}

	req := httptest.NewRequest("GET", "/api/customers/c1", nil)
	r := chi.NewRouter()
	r.Get("/api/customers/{id}", handler.GetCustomer)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["partyName"] != "John Doe" {
		t.Errorf("Expected partyName 'John Doe', got %v", response["partyName"])
	}
	if response["partyType"] != "Individual" {
		t.Errorf("Expected partyType 'Individual', got %v", response["partyType"])
	}
}

// Helper needed because Handler uses chi.URLParam which requires chi context
// I'll import chi in test or just mock context?
// Installing go-chi/chi is required for tests too if I use it.
// I will add the import above.
