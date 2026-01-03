package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPartyHandler_SearchParties(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if exchange != "tmf.party" || routingKey != "query.party.search" {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`[{"id":"1", "name":"Party 1"}]`), nil
		}

		req := httptest.NewRequest("GET", "/api/parties?givenName=John", nil)
		w := httptest.NewRecorder()

		handler.SearchParties(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Party 1") {
			t.Errorf("Expected response to contain 'Party 1', got %s", w.Body.String())
		}
	})

	t.Run("RPC Error", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("RPC failed")
		}

		req := httptest.NewRequest("GET", "/api/parties", nil)
		w := httptest.NewRecorder()

		handler.SearchParties(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestPartyHandler_CreateParty(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if exchange != "tmf.party" || routingKey != "cmd.party.create" {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`{"id":"1"}`), nil
		}

		body := `{"givenName":"John"}`
		req := httptest.NewRequest("POST", "/api/parties", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateParty(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", w.Code)
		}
	})
}

func TestPartyHandler_GetParty(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /parties/{id}", handler.GetParty)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"123"}`), nil
		}

		req := httptest.NewRequest("GET", "/parties/123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})
}
