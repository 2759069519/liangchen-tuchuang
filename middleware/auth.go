package middleware

import (
	"context"
	"net/http"
	"strings"

	"imgbed/auth"
)

type contextKey string

const AuthValidKey contextKey = "auth_valid"
const AuthUserKey contextKey = "auth_user"

func isPublicPath(method, path string) bool {
	if method == http.MethodGet && path == "/" {
		return true
	}
	if method == http.MethodGet && strings.HasPrefix(path, "/img/") {
		return true
	}
	if method == http.MethodPost && path == "/api/login" {
		return true
	}
	return false
}

func isUploadPath(method, path string) bool {
	return method == http.MethodPost && path == "/api/upload"
}

func isListPath(method, path string) bool {
	return method == http.MethodGet && path == "/api/images"
}

func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			if isUploadPath(r.Method, r.URL.Path) || isListPath(r.Method, r.URL.Path) {
				valid, username := checkAuth(r, token)
				ctx := context.WithValue(r.Context(), AuthValidKey, valid)
				if username != "" {
					ctx = context.WithValue(ctx, AuthUserKey, username)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			valid, username := checkAuth(r, token)
			if !valid {
				http.Error(w, `{"error":"invalid or missing token"}`, http.StatusUnauthorized)
				return
			}
			ctx := r.Context()
			if username != "" {
				ctx = context.WithValue(ctx, AuthUserKey, username)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func checkAuth(r *http.Request, staticToken string) (bool, string) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false, ""
	}
	token := parts[1]
	if staticToken != "" && token == staticToken {
		return true, "admin"
	}
	return auth.ValidateToken(token)
}

func IsAuthValid(r *http.Request) bool {
	v, ok := r.Context().Value(AuthValidKey).(bool)
	return ok && v
}

func GetAuthUser(r *http.Request) string {
	v, _ := r.Context().Value(AuthUserKey).(string)
	return v
}

// GetClientIP extracts the real client IP from X-Forwarded-For, X-Real-IP, or RemoteAddr.
func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
