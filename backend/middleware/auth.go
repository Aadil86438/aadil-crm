package middleware

import (
	"context"
	"net/http"

	"crm/utils"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Auth validates the JWT token and injects claims into the request context
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := utils.ExtractTokenFromRequest(r)
		if tokenString == "" {
			utils.Unauthorized(w, "Authorization token required")
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.Unauthorized(w, "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaims extracts the claims from the request context
func GetClaims(r *http.Request) *utils.Claims {
	if claims, ok := r.Context().Value(ClaimsKey).(*utils.Claims); ok {
		return claims
	}
	return nil
}
