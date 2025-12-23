package http

import (
	"context"
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
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}) ([]byte, error) {
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
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}) ([]byte, error) {
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
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}) ([]byte, error) {
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

// Helper needed because Handler uses chi.URLParam which requires chi context
// I'll import chi in test or just mock context?
// Installing go-chi/chi is required for tests too if I use it.
// I will add the import above.
