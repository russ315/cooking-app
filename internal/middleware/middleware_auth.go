package middleware

import (
	"context"
	"net/http"
	"strings"

	"cooking-app/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware ...
type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// extractAndValidateToken is shared logic for both required + optional auth
func (m *AuthMiddleware) extractAndValidateToken(r *http.Request) (int, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, false
	}

	tokenStr := parts[1]
	claims, err := m.authService.ValidateToken(tokenStr)
	if err != nil {
		return 0, false
	}

	userIDAny, ok := claims["user_id"]
	if !ok {
		return 0, false
	}

	// Safest way: try different possible numeric types
	switch v := userIDAny.(type) {
	case float64:
		if v < 1 || v > 1<<31-1 { // reasonable guard for int32 user IDs
			return 0, false
		}
		return int(v), true
	case int:
		if v < 1 {
			return 0, false
		}
		return v, true
	case int64:
		if v < 1 || v > 1<<31-1 {
			return 0, false
		}
		return int(v), true
	default:
		return 0, false
	}
}

// Authenticate — requires valid token
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := m.extractAndValidateToken(r)
		if !ok {
			http.Error(w, "Unauthorized - invalid or missing token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth — attaches user_id only if token is valid, otherwise continues
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := m.extractAndValidateToken(r)
		if ok {
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)
		}
		// always continue — even without auth
		next.ServeHTTP(w, r)
	})
}

// GetUserID returns the authenticated user ID (if present)
func GetUserID(r *http.Request) (int, bool) {
	v := r.Context().Value(UserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// MustGetUserID panics or returns error variant if no user (for handlers that require auth)
func MustGetUserID(r *http.Request) int {
	id, ok := GetUserID(r)
	if !ok {
		panic("MustGetUserID called without authenticated user")
		// or return error / use custom error type in real apps
	}
	return id
}
