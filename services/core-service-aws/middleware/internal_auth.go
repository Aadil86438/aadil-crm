package middleware

import (
	"net/http"
	"os"
)

// InternalServiceAuth validates requests coming from admin-service-gcp
// using a shared secret key set in both services via INTERNAL_SERVICE_KEY env var.
// This is how microservices authenticate each other WITHOUT user JWT tokens.
func InternalServiceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedKey := os.Getenv("INTERNAL_SERVICE_KEY")
		if expectedKey == "" {
			// If no key is configured, block all internal calls as a safety measure
			http.Error(w, `{"error":"Internal service authentication not configured"}`, http.StatusServiceUnavailable)
			return
		}

		providedKey := r.Header.Get("X-Internal-Service-Key")
		if providedKey == "" || providedKey != expectedKey {
			http.Error(w, `{"error":"Unauthorized: invalid or missing service key"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
