package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// RegistrationHandler handles self-service registration endpoints
type RegistrationHandler struct {
	regRepo  *repositories.RegistrationRepository
	userRepo *repositories.UserRepository
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(regRepo *repositories.RegistrationRepository, userRepo *repositories.UserRepository) *RegistrationHandler {
	return &RegistrationHandler{regRepo: regRepo, userRepo: userRepo}
}

// Register handles POST /api/auth/register
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

	// Check if email already has a pending registration
	existingReg, _ := h.regRepo.FindByEmail(req.Email)
	if existingReg != nil {
		if existingReg.ApprovalStatus == "rejected" {
			// Allow re-registration after rejection — delete old request
			h.regRepo.Delete(existingReg.ID)
		} else {
			utils.Conflict(w, "A registration is already pending for this email")
			return
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerError(w, "Failed to process password")
		return
	}

	// Create registration request
	reg, err := h.regRepo.Create(&req, hashedPassword)
	if err != nil {
		utils.InternalServerError(w, "Failed to create registration")
		return
	}

	utils.Created(w, reg)
}

// SubmitPayment handles POST /api/auth/submit-payment
func (h *RegistrationHandler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.RegistrationID) == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}
	if strings.TrimSpace(req.TransactionID) == "" {
		utils.BadRequest(w, "Transaction ID is required")
		return
	}

	// Verify registration exists
	reg, err := h.regRepo.FindByID(req.RegistrationID)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration not found")
		return
	}

	if reg.PaymentStatus == "submitted" {
		utils.BadRequest(w, "Payment has already been submitted")
		return
	}

	// Update payment info
	if err := h.regRepo.SubmitPayment(req.RegistrationID, req.TransactionID); err != nil {
		utils.InternalServerError(w, "Failed to record payment")
		return
	}

	utils.Success(w, map[string]string{
		"message": "Payment submitted successfully. Awaiting admin approval.",
		"status":  "submitted",
	})
}

// CheckStatus handles GET /api/auth/registration-status/:id
func (h *RegistrationHandler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/auth/registration-status/")
	if id == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	reg, err := h.regRepo.FindByID(id)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration not found")
		return
	}

	utils.Success(w, map[string]string{
		"id":              reg.ID,
		"payment_status":  reg.PaymentStatus,
		"approval_status": reg.ApprovalStatus,
	})
}
