package auth

import (
	"context"
	"net/http"
)

// Middleware is a placeholder for Okta Auth.
// In a real implementation, this would validate the session/JWT.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock Auth: For demo, allowing all requests or assuming a test user
		// In production:
		// 1. Check for Session Cookie
		// 2. Validate Token with Okta
		// 3. Reject if invalid

		// Inject mock user context
		ctx := context.WithValue(r.Context(), "user", "demo-user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
