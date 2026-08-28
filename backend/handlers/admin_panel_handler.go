package handlers

import (
	"encoding/json"
	"net/http"

	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// AdminPanelHandler handles the admin approval panel (code-gated, not JWT-based)
type AdminPanelHandler struct {
	regRepo  *repositories.RegistrationRepository
	userRepo *repositories.UserRepository
}

// NewAdminPanelHandler creates a new AdminPanelHandler
func NewAdminPanelHandler(regRepo *repositories.RegistrationRepository, userRepo *repositories.UserRepository) *AdminPanelHandler {
	return &AdminPanelHandler{regRepo: regRepo, userRepo: userRepo}
}

// adminCode is the hardcoded admin access code
const adminCode = "1101"

// VerifyCode handles POST /api/admin/verify — validates the admin code
func (h *AdminPanelHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req models.AdminVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Code != adminCode {
		utils.Forbidden(w, "Invalid admin code")
		return
	}

	// Generate a short-lived token for admin operations
	token, err := utils.GenerateToken("admin-panel", "admin@propertier.app", "admin", "Admin Panel", 2*60*60*1e9) // 2 hours
	if err != nil {
		utils.InternalServerError(w, "Failed to generate session")
		return
	}

	utils.Success(w, map[string]string{
		"token":   token,
		"message": "Admin access granted",
	})
}

// ListPending handles GET /api/admin/pending
func (h *AdminPanelHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	regs, err := h.regRepo.ListPending()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch pending requests")
		return
	}
	if regs == nil {
		regs = []*models.RegistrationRequest{}
	}
	utils.Success(w, regs)
}

// ListAll handles GET /api/admin/all
func (h *AdminPanelHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	regs, err := h.regRepo.ListAll()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch registration requests")
		return
	}
	if regs == nil {
		regs = []*models.RegistrationRequest{}
	}
	utils.Success(w, regs)
}

// Approve handles POST /api/admin/approve/:id
func (h *AdminPanelHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/admin/approve/")
	if id == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	// Get registration request
	reg, err := h.regRepo.FindByID(id)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration request not found")
		return
	}

	if reg.ApprovalStatus != "pending" {
		utils.BadRequest(w, "This request has already been processed")
		return
	}

	// Create user from registration data
	createReq := &models.CreateUserRequest{
		Name:  reg.Name,
		Email: reg.Email,
		Role:  "sales_user",
	}

	_, err = h.userRepo.Create(createReq, reg.PasswordHash)
	if err != nil {
		utils.InternalServerError(w, "Failed to create user account")
		return
	}

	// Mark registration as approved
	h.regRepo.Approve(id)

	utils.Success(w, map[string]string{
		"message": "User approved successfully. They can now sign in.",
	})
}

// Reject handles POST /api/admin/reject/:id
func (h *AdminPanelHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/admin/reject/")
	if id == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	reg, err := h.regRepo.FindByID(id)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration request not found")
		return
	}

	if reg.ApprovalStatus != "pending" {
		utils.BadRequest(w, "This request has already been processed")
		return
	}

	h.regRepo.Reject(id)

	utils.Success(w, map[string]string{
		"message": "Registration request rejected.",
	})
}
