package middleware

import (
	"net/http"

	"crm/utils"
)

// RequireRole returns middleware that enforces role-based access
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.Forbidden(w, "Insufficient permissions")
		})
	}
}

// RequireAdmin enforces admin-only access
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole("admin")(next)
}

// RequireManagerOrAdmin enforces manager or admin access
func RequireManagerOrAdmin(next http.Handler) http.Handler {
	return RequireRole("admin", "manager")(next)
}
