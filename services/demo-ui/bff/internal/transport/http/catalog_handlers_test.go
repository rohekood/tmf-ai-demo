package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCatalogHandler(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewCatalogHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("ListCatalogs_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if exchange != catalogExchange || routingKey != queryCatalogList {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`[{"id":"cat1", "name":"Catalog 1"}]`), nil
		}

		req := httptest.NewRequest("GET", "/api/catalogs", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Catalog 1") {
			t.Errorf("Expected response to contain 'Catalog 1', got %s", w.Body.String())
		}
	})

	t.Run("CreateSpecification_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if exchange != catalogExchange || routingKey != cmdSpecCreate {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`{"id":"spec1", "name":"Spec 1"}`), nil
		}

		body := `{"name":"Spec 1"}`
		req := httptest.NewRequest("POST", "/api/specifications", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "spec1") {
			t.Errorf("Expected response to contain 'spec1', got %s", w.Body.String())
		}
	})

	t.Run("GetOffering_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != queryOfferingGet {
				t.Errorf("Unexpected routing key: %s", routingKey)
			}
			p := payload.(map[string]string)
			if p["id"] != "off1" {
				t.Errorf("Expected ID 'off1', got %v", p["id"])
			}
			return []byte(`{"id":"off1", "name":"Offering 1"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/offerings/off1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("UpdateCategory_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != cmdCategoryUpdate {
				t.Errorf("Unexpected routing key: %s", routingKey)
			}
			p := payload.(map[string]any)
			if p["id"] != "cat-id-1" {
				t.Errorf("Expected ID 'cat-id-1', got %v", p["id"])
			}
			return []byte(`{"id":"cat-id-1", "name":"Updated Category"}`), nil
		}

		body := `{"name":"Updated Category"}`
		req := httptest.NewRequest("PUT", "/api/categories/cat-id-1", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("DeleteCatalog_Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != cmdCatalogDelete {
				t.Errorf("Unexpected routing key: %s", routingKey)
			}
			return nil, nil
		}

		req := httptest.NewRequest("DELETE", "/api/catalogs/cat1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status NoContent, got %v", w.Code)
		}
	})
}
