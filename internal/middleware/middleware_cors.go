package middleware

import (
	"net/http"
)

// CORSMiddleware CORS middleware to enable Cross-Origin Resource Sharing
type CORSMiddleware struct {
	allowedOrigins []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

// Handler adds CORS headers to the response
func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin == "" || m.isAllowedOrigin(origin) {
			if len(m.allowedOrigins) == 1 && m.allowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CORSMiddleware) isAllowedOrigin(origin string) bool {
	for _, allowed := range m.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
