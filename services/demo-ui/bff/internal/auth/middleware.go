package auth

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"

	"tmf/pkg/rabbitmq"
)

// Public (anonymous-accessible) API route patterns. Only the Check Availability
// (qualification) flow is reachable without authentication; every other /api
// route requires a valid JWT.
const (
	qualCheckPath   = "/api/qualification/check"
	qualSessionPath = "/api/qualification/session/"
)

// EmailClaimKey is the namespaced custom claim that carries the user's email.
// It must match the claim emitted by the Auth0 Login Action that enriches the
// access token (see docs/plans/08_login_customer_provisioning.md, task A1).
const EmailClaimKey = "https://tmf-demo/email"

// emailContextKey is the context key under which the authenticated user's email
// is stored. It is BFF-local: email is read by request handlers (provisioning),
// not forwarded as an RPC transport header.
type emailContextKey struct{}

// CustomClaims captures the namespaced email claim from the access token. It
// implements validator.CustomClaims so the Auth0 validator unmarshals it.
type CustomClaims struct {
	Email string `json:"https://tmf-demo/email"`
}

// Validate satisfies validator.CustomClaims. No extra validation is required.
func (c *CustomClaims) Validate(_ context.Context) error { return nil }

// EmailFromContext returns the authenticated user's email if it was present on
// the validated token.
func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailContextKey{}).(string)
	return email, ok && email != ""
}

// ContextWithEmail returns a copy of ctx carrying the given authenticated email.
// Exported for composition and testing.
func ContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailContextKey{}, email)
}

// withIdentity injects the user id (sub) and, when present, the email from the
// validated claims into the context.
func withIdentity(ctx context.Context, vc *validator.ValidatedClaims) context.Context {
	if vc == nil {
		return ctx
	}
	if sub := vc.RegisteredClaims.Subject; sub != "" {
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyUser, sub)
	}
	if cc, ok := vc.CustomClaims.(*CustomClaims); ok && cc.Email != "" {
		ctx = context.WithValue(ctx, emailContextKey{}, cc.Email)
	}
	return ctx
}

// IsPublicRoute reports whether the given HTTP method and path may be called
// without authentication. The match is method-aware and prefix-safe so that
// look-alike paths (e.g. "/api/qualification/checkfoo") are NOT treated as
// public and protected routes never leak through.
func IsPublicRoute(method, path string) bool {
	switch {
	case method == http.MethodPost && path == qualCheckPath:
		return true
	case method == http.MethodGet && strings.HasPrefix(path, qualSessionPath):
		// Require exactly one non-empty trailing segment:
		// /api/qualification/session/{sessionId}
		rest := strings.TrimPrefix(path, qualSessionPath)
		return rest != "" && !strings.Contains(rest, "/")
	default:
		return false
	}
}

// OptionalToken validates a bearer token if one is present and injects the user
// ID into the request context, but never rejects the request when the token is
// missing or invalid. It is intended for public routes so that an authenticated
// caller still gets their identity (used for customer-specific pricing) while
// anonymous callers proceed unauthenticated.
func OptionalToken(tokenValidator TokenValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const bearerPrefix = "Bearer "
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, bearerPrefix) {
				token := strings.TrimPrefix(authHeader, bearerPrefix)
				if claims, err := tokenValidator.ValidateToken(r.Context(), token); err == nil {
					if vc, ok := claims.(*validator.ValidatedClaims); ok && vc.RegisteredClaims.Subject != "" {
						r = r.WithContext(withIdentity(r.Context(), vc))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TokenValidator defines the interface for validating JWT tokens.
type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (any, error)
}

// NewAuth0Validator creates a new TokenValidator using Auth0.
func NewAuth0Validator(domain, audience string) (TokenValidator, error) {
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		return nil, err
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{audience},
		validator.WithAllowedClockSkew(time.Minute),
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
	)
	if err != nil {
		return nil, err
	}

	return jwtValidator, nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken(tokenValidator TokenValidator, domain, audience string) func(next http.Handler) http.Handler {
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)
		log.Printf("DEBUG: Expected Domain: %s, Audience: %s", domain, audience)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		tokenValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims and inject into context if needed for application logic
			claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			if ok {
				// Inject the User ID (sub) — relied upon by customer_handlers.go —
				// and the email (when the access token carries the namespaced
				// claim) into the context.
				next.ServeHTTP(w, r.WithContext(withIdentity(r.Context(), claims)))
			} else {
				// Should not happen if CheckJWT passes, but safe fallback
				next.ServeHTTP(w, r)
			}
		}))
	}
}
