package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"errors"
)

func TestCatalogHandler_Errors(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewCatalogHandler(mockClient)
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
		{"ListCatalogs", "GET", "/api/catalogs", "", errors.New("rpc"), 500},
		{"ListCategories", "GET", "/api/categories", "", errors.New("rpc"), 500},
		{"ListSpecifications", "GET", "/api/specifications", "", errors.New("rpc"), 500},
		{"ListOfferings", "GET", "/api/offerings", "", errors.New("rpc"), 500},
		{"CreateCatBad", "POST", "/api/catalogs", "{", nil, 400},
		{"CreateCatErr", "POST", "/api/catalogs", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdateCatBad", "PUT", "/api/catalogs/1", "{", nil, 400},
		{"UpdateCatErr", "PUT", "/api/catalogs/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"GetCatErr", "GET", "/api/catalogs/1", "", errors.New("rpc"), 500},
		{"DelCatErr", "DELETE", "/api/catalogs/1", "", errors.New("rpc"), 500},
		
		{"CreateCategoryBad", "POST", "/api/categories", "{", nil, 400},
		{"CreateCategoryErr", "POST", "/api/categories", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdateCategoryBad", "PUT", "/api/categories/1", "{", nil, 400},
		{"UpdateCategoryErr", "PUT", "/api/categories/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"GetCategoryErr", "GET", "/api/categories/1", "", errors.New("rpc"), 500},
		{"DelCategoryErr", "DELETE", "/api/categories/1", "", errors.New("rpc"), 500},

		{"CreateSpecBad", "POST", "/api/specifications", "{", nil, 400},
		{"CreateSpecErr", "POST", "/api/specifications", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdateSpecBad", "PUT", "/api/specifications/1", "{", nil, 400},
		{"UpdateSpecErr", "PUT", "/api/specifications/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"GetSpecErr", "GET", "/api/specifications/1", "", errors.New("rpc"), 500},
		{"DelSpecErr", "DELETE", "/api/specifications/1", "", errors.New("rpc"), 500},

		{"CreateOffBad", "POST", "/api/offerings", "{", nil, 400},
		{"CreateOffErr", "POST", "/api/offerings", `{"name":"A"}`, errors.New("rpc"), 500},
		{"UpdateOffBad", "PUT", "/api/offerings/1", "{", nil, 400},
		{"UpdateOffErr", "PUT", "/api/offerings/1", `{"name":"A"}`, errors.New("rpc"), 500},
		{"GetOffErr", "GET", "/api/offerings/1", "", errors.New("rpc"), 500},
		{"DelOffErr", "DELETE", "/api/offerings/1", "", errors.New("rpc"), 500},
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
