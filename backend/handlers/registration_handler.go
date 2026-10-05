package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// RegistrationHandler handles self-service registration endpoints
type RegistrationHandler struct {
	userRepo *repositories.UserRepository
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(userRepo *repositories.UserRepository) *RegistrationHandler {
	return &RegistrationHandler{userRepo: userRepo}
}

// Register handles POST /api/auth/register (Direct Signup)
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Full name is required")
		return
	}
	if !utils.IsValidEmail(req.Email) {
		utils.BadRequest(w, "Valid email address is required")
		return
	}
	if len(req.Password) < 8 {
		utils.BadRequest(w, "Password must be at least 8 characters")
		return
	}
	if strings.TrimSpace(req.CompanyName) == "" {
		utils.BadRequest(w, "Company/Organization name is required")
		return
	}

	// Check if email already exists in users table
	existingUser, _ := h.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		utils.Conflict(w, "An account with this email already exists. Please sign in.")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerError(w, "Failed to process password")
		return
	}

	// Directly create user as active
	user, err := h.userRepo.Create(&models.CreateUserRequest{
		Name:  req.Name,
		Email: req.Email,
		Role:  "admin", // First user registering company is admin
	}, hashedPassword)
	if err != nil {
		utils.InternalServerError(w, "Failed to create user account")
		return
	}

	// Generate JWT Token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, user.Name, 24*time.Hour)
	if err != nil {
		utils.InternalServerError(w, "Failed to generate authentication token")
		return
	}

	utils.Created(w, models.LoginResponse{
		Token: token,
		User:  user,
	})
}
