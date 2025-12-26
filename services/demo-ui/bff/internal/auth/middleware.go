package auth

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// TokenValidator defines the interface for validating JWT tokens.
type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (interface{}, error)
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
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
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
				// Inject the User ID (sub) into the context as "user"
				// This is relied upon by customer_handlers.go
				userID := claims.RegisteredClaims.Subject
				ctx := context.WithValue(r.Context(), "user", userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				// Should not happen if CheckJWT passes, but safe fallback
				next.ServeHTTP(w, r)
			}
		}))
	}
}
