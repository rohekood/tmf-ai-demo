package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCatalogHandler_Success(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewCatalogHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		code   int
	}{
		{"CreateCatalog_Success", "POST", "/api/catalogs", `{"name":"A"}`, 201},
		{"GetCatalog_Success", "GET", "/api/catalogs/1", "", 200},
		{"UpdateCatalog_Success", "PUT", "/api/catalogs/1", `{"name":"A"}`, 200},
		{"ListCategories_Success", "GET", "/api/categories", "", 200},
		{"CreateCategory_Success", "POST", "/api/categories", `{"name":"A"}`, 201},
		{"GetCategory_Success", "GET", "/api/categories/1", "", 200},
		{"DeleteCategory_Success", "DELETE", "/api/categories/1", "", 204},
		{"ListSpecifications_Success", "GET", "/api/specifications", "", 200},
		{"GetSpecification_Success", "GET", "/api/specifications/1", "", 200},
		{"UpdateSpecification_Success", "PUT", "/api/specifications/1", `{"name":"A"}`, 200},
		{"DeleteSpecification_Success", "DELETE", "/api/specifications/1", "", 204},
		{"ListOfferings_Success", "GET", "/api/offerings", "", 200},
		{"CreateOffering_Success", "POST", "/api/offerings", `{"name":"A"}`, 201},
		{"UpdateOffering_Success", "PUT", "/api/offerings/1", `{"name":"A"}`, 200},
		{"DeleteOffering_Success", "DELETE", "/api/offerings/1", "", 204},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
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
			if w.Code != tc.code {
				t.Errorf("expected %d got %d", tc.code, w.Code)
			}
		})
	}
}
