package middleware

import (
	"net/http"
	"strings"
)

// Auth returns middleware that checks for a valid Bearer token.
// If token is empty, auth is disabled (open access).
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for GET requests (public image access)
			if r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// If no token configured, allow all
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization header: "Bearer <token>"
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != token {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
