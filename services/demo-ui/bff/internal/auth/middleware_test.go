package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type MockValidator struct {
	ValidateTokenFunc func(ctx context.Context, tokenString string) (interface{}, error)
}

func (m *MockValidator) ValidateToken(ctx context.Context, tokenString string) (interface{}, error) {
	return m.ValidateTokenFunc(ctx, tokenString)
}

func TestMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mockValidator := &MockValidator{
		ValidateTokenFunc: func(ctx context.Context, tokenString string) (interface{}, error) {
			// Return a mock validated claim
			return &validator.ValidatedClaims{
				RegisteredClaims: validator.RegisteredClaims{
					Subject: "demo-user",
				},
			}, nil
		},
	}

	middleware := EnsureValidToken(mockValidator, "test-domain", "test-audience")
	handler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}
}
