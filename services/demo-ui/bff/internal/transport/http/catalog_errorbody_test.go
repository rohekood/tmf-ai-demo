package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCatalogHelpers_EmptyID tests the guard branches in the helper methods.
// These are reachable only by calling helpers directly (not via mux routes,
// since {id} patterns always yield a non-empty value).
func TestCatalogHelpers_EmptyID(t *testing.T) {
	mockClient := &MockRPCClient{}
	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
		t.Error("RPC should not be called when ID is empty")
		return nil, nil
	}
	handler := NewCatalogHandler(mockClient)

	t.Run("handleCommandWithID_EmptyID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/", strings.NewReader(`{"name":"x"}`))
		w := httptest.NewRecorder()
		handler.handleCommandWithID(w, req, "cmd.test", "testing")
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("handleQueryByID_EmptyID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.handleQueryByID(w, req, "query.test", "testing")
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("handleDelete_EmptyID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/", nil)
		w := httptest.NewRecorder()
		handler.handleDelete(w, req, "cmd.test.delete", "testing")
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

// Tests for the new error-body detection paths in handleCommand and handleCommandWithID.
func TestCatalogHandler_ErrorBodyReturns422(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewCatalogHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	errorBody := `{"error":"validation failed: name is required"}`

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"CreateCatalog_ErrorBody", "POST", "/api/catalogs", `{"name":"A"}`},
		{"CreateCategory_ErrorBody", "POST", "/api/categories", `{"name":"A"}`},
		{"CreateSpec_ErrorBody", "POST", "/api/specifications", `{"name":"A"}`},
		{"CreateOffering_ErrorBody", "POST", "/api/offerings", `{"name":"A"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
				return []byte(errorBody), nil
			}
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusUnprocessableEntity {
				t.Errorf("%s: expected 422, got %d (body: %s)", tc.name, w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "validation failed") {
				t.Errorf("%s: expected error text in body, got %q", tc.name, w.Body.String())
			}
		})
	}
}

func TestCatalogHandler_UpdateErrorBodyReturns422(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewCatalogHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	errorBody := `{"error":"catalog not found"}`

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"UpdateCatalog_ErrorBody", "PUT", "/api/catalogs/1", `{"name":"A"}`},
		{"UpdateCategory_ErrorBody", "PUT", "/api/categories/1", `{"name":"A"}`},
		{"UpdateSpec_ErrorBody", "PUT", "/api/specifications/1", `{"name":"A"}`},
		{"UpdateOffering_ErrorBody", "PUT", "/api/offerings/1", `{"name":"A"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
				return []byte(errorBody), nil
			}
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusUnprocessableEntity {
				t.Errorf("%s: expected 422, got %d (body: %s)", tc.name, w.Code, w.Body.String())
			}
		})
	}
}
