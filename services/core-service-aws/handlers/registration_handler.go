package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"core-service-aws/models"
	"core-service-aws/repositories"
	"core-service-aws/utils"
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

	// Create user as pending approval (inactive until admin approves)
	user, err := h.userRepo.Create(&models.CreateUserRequest{
		Name:           req.Name,
		Email:          req.Email,
		Role:           "sales_user",
		CompanyName:    req.CompanyName,
		ApprovalStatus: "pending",
	}, hashedPassword)
	if err != nil {
		utils.InternalServerError(w, "Failed to submit registration request")
		return
	}

	utils.Created(w, map[string]interface{}{
		"message":           "Registration submitted successfully. Your account is pending admin approval. You will receive your CRM Access Code via email once approved.",
		"requires_approval": true,
		"user": map[string]interface{}{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"company_name":    user.CompanyName,
			"approval_status": user.ApprovalStatus,
		},
	})
}

