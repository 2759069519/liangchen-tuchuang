package middleware

import (
	"net/http"
	"os"
)

var allowedOrigin string

func init() {
	allowedOrigin = os.Getenv("CORS_ORIGIN")
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowedOrigin == "" {
			// If no CORS_ORIGIN set, allow all (dev mode)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
