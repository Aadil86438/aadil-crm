package handlers

import (
	"encoding/json"
	"net/http"

	"crm/database"
	"crm/middleware"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo *repositories.UserRepository
	auditRepo *repositories.AuditRepository
	jwtExpiration interface{ Hours() float64 }
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *repositories.UserRepository, auditRepo *repositories.AuditRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, auditRepo: auditRepo}
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "Email and password are required")
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		utils.InternalServerError(w, "Authentication failed")
		return
	}
	if user == nil || !utils.CheckPassword(user.PasswordHash, req.Password) {
		// Log failed login attempt
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserEmail: req.Email,
			Action:    "login_failed",
			Entity:    "auth",
			IPAddress: r.RemoteAddr,
		})
		utils.Unauthorized(w, "Invalid email or password")
		return
	}

	if user.Status != "active" {
		utils.Forbidden(w, "Account is disabled")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, user.Name, 24*60*60*1e9)
	if err != nil {
		utils.InternalServerError(w, "Failed to generate token")
		return
	}

	// Log successful login
	h.auditRepo.Create(&models.CreateAuditLogRequest{
		UserID:    &user.ID,
		UserEmail: user.Email,
		Action:    "login",
		Entity:    "auth",
		IPAddress: r.RemoteAddr,
	})

	utils.Success(w, models.LoginResponse{Token: token, User: user})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID:    &claims.UserID,
			UserEmail: claims.Email,
			Action:    "logout",
			Entity:    "auth",
			IPAddress: r.RemoteAddr,
		})
	}
	utils.Success(w, map[string]string{"message": "Logged out successfully"})
}

// Me handles GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		utils.Unauthorized(w, "Not authenticated")
		return
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil || user == nil {
		utils.NotFound(w, "User not found")
		return
	}

	utils.Success(w, user)
}

// Health handles GET /health
func Health(w http.ResponseWriter, r *http.Request) {
	dbErr := database.HealthCheck()
	if dbErr != nil {
		utils.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded", "db": dbErr.Error(),
		})
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
