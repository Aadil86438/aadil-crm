package middleware

import (
	"net/http"
	"strings"
)

// CORS sets appropriate CORS headers for cross-origin requests
func CORS(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if frontendURL == "*" || frontendURL == "" {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			} else {
				allowedOrigins := []string{
					frontendURL,
					"http://localhost",
					"http://localhost:80",
					"http://localhost:8080",
					"http://localhost:8081",
					"http://localhost:3000",
					"http://localhost:5173",
				}
				for _, configured := range strings.Split(frontendURL, ",") {
					allowedOrigins = append(allowedOrigins, strings.TrimSpace(configured))
				}

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
					w.Header().Set("Access-Control-Allow-Origin", frontendURL)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Internal-Service-Key")
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


