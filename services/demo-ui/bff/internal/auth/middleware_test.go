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
	email string
}

func (m *mockValidator) ValidateToken(ctx context.Context, tokenString string) (any, error) {
	if m.valid {
		return &validator.ValidatedClaims{
			RegisteredClaims: validator.RegisteredClaims{
				Subject: "test-user-id",
			},
			CustomClaims: &CustomClaims{Email: m.email},
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
	_, err = NewAuth0Validator(string([]byte{0x7f})+"invalid", "test-audience")
	if err == nil {
		t.Errorf("expected error for invalid url")
	}
}

func TestIsPublicRoute(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{"qualification check", http.MethodPost, "/api/qualification/check", true},
		{"qualification session", http.MethodGet, "/api/qualification/session/abc-123", true},
		{"check wrong method", http.MethodGet, "/api/qualification/check", false},
		{"check lookalike not allowed", http.MethodPost, "/api/qualification/checkfoo", false},
		{"session wrong method", http.MethodPost, "/api/qualification/session/abc", false},
		{"session missing id", http.MethodGet, "/api/qualification/session/", false},
		{"session extra segment", http.MethodGet, "/api/qualification/session/abc/extra", false},
		{"cart blocked", http.MethodPost, "/api/cart/items", false},
		{"checkout blocked", http.MethodPost, "/api/orders/checkout", false},
		{"catalog blocked", http.MethodGet, "/api/offerings", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPublicRoute(tc.method, tc.path); got != tc.want {
				t.Errorf("IsPublicRoute(%q, %q) = %v, want %v", tc.method, tc.path, got, tc.want)
			}
		})
	}
}

func TestOptionalToken_InjectsUserWhenValid(t *testing.T) {
	mw := OptionalToken(&mockValidator{valid: true})

	var gotUser any
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Context().Value(rabbitmq.ContextKeyUser)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/qualification/check", nil)
	req.Header.Set("Authorization", "Bearer dummy-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %v", rr.Code)
	}
	if gotUser != "test-user-id" {
		t.Errorf("expected user test-user-id injected, got %v", gotUser)
	}
}

func TestOptionalToken_InjectsEmailFromClaim(t *testing.T) {
	mw := OptionalToken(&mockValidator{valid: true, email: "user@example.com"})

	var gotEmail string
	var emailOK bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEmail, emailOK = EmailFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/qualification/check", nil)
	req.Header.Set("Authorization", "Bearer dummy-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !emailOK || gotEmail != "user@example.com" {
		t.Errorf("expected email user@example.com injected, got %q (ok=%v)", gotEmail, emailOK)
	}
}

func TestEmailFromContext_AbsentWhenNoClaim(t *testing.T) {
	mw := OptionalToken(&mockValidator{valid: true}) // no email claim

	var emailOK bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, emailOK = EmailFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/qualification/check", nil)
	req.Header.Set("Authorization", "Bearer dummy-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if emailOK {
		t.Error("expected no email in context when claim is absent")
	}
}

func TestOptionalToken_AllowsAnonymous(t *testing.T) {
	mw := OptionalToken(&mockValidator{valid: true})

	var gotUser any
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotUser = r.Context().Value(rabbitmq.ContextKeyUser)
		w.WriteHeader(http.StatusOK)
	}))

	// No Authorization header at all.
	req := httptest.NewRequest("POST", "/api/qualification/check", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected next handler to be called for anonymous request")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK for anonymous, got %v", rr.Code)
	}
	if gotUser != nil {
		t.Errorf("expected no user injected for anonymous, got %v", gotUser)
	}
}

func TestOptionalToken_IgnoresInvalidToken(t *testing.T) {
	// valid:false makes ValidateToken return (nil, nil); the claims type
	// assertion fails, so no user is injected and the request still proceeds.
	mw := OptionalToken(&mockValidator{valid: false})

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/qualification/check", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called || rr.Code != http.StatusOK {
		t.Errorf("expected request to proceed despite invalid token, called=%v code=%v", called, rr.Code)
	}
}
