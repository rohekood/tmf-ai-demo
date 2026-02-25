package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"tmf/pkg/rabbitmq"
)

type mockValidator struct {
	valid bool
}

func (m *mockValidator) ValidateToken(ctx context.Context, tokenString string) (interface{}, error) {
	if m.valid {
		return &validator.ValidatedClaims{
			RegisteredClaims: validator.RegisteredClaims{
				Subject: "test-user-id",
			},
		}, nil
	}
	return nil, nil // error handling in middleware expects error or false validation
}

func TestEnsureValidToken(t *testing.T) {
	validatorMock := &mockValidator{valid: true}
	middleware := EnsureValidToken(validatorMock, "test-domain", "test-audience")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(rabbitmq.ContextKeyUser)
		if user != "test-user-id" {
			t.Errorf("expected user test-user-id, got %v", user)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer dummy-token")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	// Test failing auth
	validatorFailMock := &mockValidator{valid: false}
	middlewareFail := EnsureValidToken(validatorFailMock, "test-domain", "test-audience")
	handlerFail := middlewareFail(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	reqFail := httptest.NewRequest("GET", "/", nil)
	rrFail := httptest.NewRecorder()
	handlerFail.ServeHTTP(rrFail, reqFail) // without header it should fail

	if rrFail.Code != http.StatusUnauthorized {
		t.Errorf("expected status Unauthorized, got %v", rrFail.Code)
	}
}

func TestNewAuth0Validator(t *testing.T) {
	_, err := NewAuth0Validator("test.auth0.com", "test-audience")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Wait briefly to allow cache provider initialization
	time.Sleep(10 * time.Millisecond)
	
	// Test error case (invalid URL - need something that fails url.Parse or caching provider)
	// URL parsing rarely fails unless there's a control character
	_, err = NewAuth0Validator(string([]byte{0x7f}) + "invalid", "test-audience")
	if err == nil {
		t.Errorf("expected error for invalid url")
	}
}
