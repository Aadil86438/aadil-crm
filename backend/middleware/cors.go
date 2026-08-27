package middleware

import (
	"net/http"
)

// CORS sets appropriate CORS headers for cross-origin requests
func CORS(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Allow specific origins or wildcard in development
			allowedOrigins := []string{frontendURL, "http://localhost:8081", "http://localhost:3000"}
			allowed := false
			for _, o := range allowedOrigins {
				if o == origin {
					allowed = true
					break
				}
			}

			if allowed || origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// In production, only allow configured frontend
				w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
