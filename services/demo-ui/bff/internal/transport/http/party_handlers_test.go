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

func TestPartyHandler_CRUD(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("CreateParty_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"p1", "name":"Party 1"}`), nil
		}
		req := httptest.NewRequest("POST", "/api/parties", strings.NewReader(`{"name":"Party 1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusCreated && w.Code != http.StatusOK {
			t.Errorf("Expected status Created or OK, got %v", w.Code)
		}
	})

	t.Run("UpdateParty_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"p1", "name":"Updated"}`), nil
		}
		req := httptest.NewRequest("PUT", "/api/parties/p1", strings.NewReader(`{"name":"Updated"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
			t.Errorf("Expected status OK or Accepted, got %v", w.Code)
		}
	})

	t.Run("PatchParty_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"p1", "name":"Patched"}`), nil
		}
		req := httptest.NewRequest("PATCH", "/api/parties/p1", strings.NewReader(`{"name":"Patched"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
			t.Errorf("Expected status OK or Accepted, got %v", w.Code)
		}
	})

	t.Run("DeleteParty_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, nil
		}
		req := httptest.NewRequest("DELETE", "/api/parties/p1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Errorf("Expected status NoContent, got %v", w.Code)
		}
	})

	t.Run("Errors_SearchParties_RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("GET", "/api/parties", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %v", w.Code)
		}
	})

	t.Run("Errors_CreateParty_InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/parties", strings.NewReader(`{invalid}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %v", w.Code)
		}
	})

	t.Run("Errors_UpdateParty_InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/parties/1", strings.NewReader(`{invalid}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %v", w.Code)
		}
	})

	t.Run("Errors_UpdateParty_RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("PUT", "/api/parties/1", strings.NewReader(`{"name":"P"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %v", w.Code)
		}
	})

	t.Run("Errors_PatchParty_InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/api/parties/1", strings.NewReader(`{invalid}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %v", w.Code)
		}
	})

	t.Run("Errors_PatchParty_RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("PATCH", "/api/parties/1", strings.NewReader(`{"name":"P"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %v", w.Code)
		}
	})

	t.Run("Errors_DeleteParty_RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("DELETE", "/api/parties/1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %v", w.Code)
		}
	})
}

func TestPartyHandler_AdditionalErrors(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)
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
		{"SearchPartiesBad", "GET", "/api/parties?some=1", "", errors.New("rpc"), 500},
		{"GetPartyBad", "GET", "/api/parties/1", "", errors.New("rpc"), 500},
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
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
				return nil, tc.err
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
				t.Errorf("expected %d got %d", tc.code, w.Code)
			}
		})
	}
}
