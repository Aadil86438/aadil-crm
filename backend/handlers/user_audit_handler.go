package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"crm/middleware"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userRepo    *repositories.UserRepository
	accountRepo *repositories.AccountRepository
	contactRepo *repositories.ContactRepository
	auditRepo   *repositories.AuditRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	userRepo *repositories.UserRepository,
	accountRepo *repositories.AccountRepository,
	contactRepo *repositories.ContactRepository,
	auditRepo *repositories.AuditRepository,
) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		contactRepo: contactRepo,
		auditRepo:   auditRepo,
	}
}

// List handles GET /api/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	users, total, err := h.userRepo.ListUsers(page, pageSize)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch users")
		return
	}
	if users == nil {
		users = []*models.User{}
	}
	utils.Paginated(w, users, total, page, pageSize)
}

// ListSimple handles GET /api/users/simple (for dropdowns)
func (h *UserHandler) ListSimple(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListUsersSimple()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch users")
		return
	}
	if users == nil {
		users = []*models.User{}
	}
	utils.Success(w, users)
}

// Create handles POST /api/users (Admin only)
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Name is required")
		return
	}
	if !utils.IsValidEmail(req.Email) {
		utils.BadRequest(w, "Valid email is required")
		return
	}
	if len(req.Password) < 8 {
		utils.BadRequest(w, "Password must be at least 8 characters")
		return
	}
	if !utils.IsValidRole(req.Role) {
		utils.BadRequest(w, "Invalid role. Must be: admin, manager, or sales_user")
		return
	}

	// Check email uniqueness
	exists, err := h.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		utils.InternalServerError(w, "Failed to validate email")
		return
	}
	if exists {
		utils.Conflict(w, "A user with this email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerError(w, "Failed to process password")
		return
	}

	user, err := h.userRepo.Create(&req, hashedPassword)
	if err != nil {
		utils.InternalServerError(w, "Failed to create user")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "create_user", Entity: "user", EntityID: &user.ID, IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, user)
}

// Update handles PUT /api/users/:id (Admin only)
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/users/")
	// Strip off any suffix like /password
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Name is required")
		return
	}

	user, err := h.userRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update user")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "update_user", Entity: "user", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, user)
}

// ResetPassword handles PUT /api/users/:id/password (Admin only)
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id := extractIDWithSuffix(r.URL.Path, "/api/users/", "/password")

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Password) < 8 {
		utils.BadRequest(w, "Password must be at least 8 characters")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerError(w, "Failed to process password")
		return
	}

	if err := h.userRepo.UpdatePassword(id, hashedPassword); err != nil {
		utils.InternalServerError(w, "Failed to update password")
		return
	}

	utils.Success(w, map[string]string{"message": "Password updated successfully"})
}

// Delete handles DELETE /api/users/:id (Admin only)
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/users/")

	// Don't allow self-deletion
	claims := middleware.GetClaims(r)
	if claims != nil && claims.UserID == id {
		utils.BadRequest(w, "Cannot delete your own account")
		return
	}

	if err := h.userRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete user")
		return
	}

	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "delete_user", Entity: "user", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, map[string]string{"message": "User deleted successfully"})
}

// AuditHandler handles audit log HTTP requests
type AuditHandler struct {
	auditRepo *repositories.AuditRepository
}

// NewAuditHandler creates a new AuditHandler
func NewAuditHandler(auditRepo *repositories.AuditRepository) *AuditHandler {
	return &AuditHandler{auditRepo: auditRepo}
}

// List handles GET /api/audit-logs
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	logs, total, err := h.auditRepo.List(page, pageSize)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch audit logs")
		return
	}
	if logs == nil {
		logs = []*models.AuditLog{}
	}
	utils.Paginated(w, logs, total, page, pageSize)
}

// ListAccountsSimple handles GET /api/accounts/simple
func (h *AccountHandler) ListSimple(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.accountRepo.ListSimple()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch accounts")
		return
	}
	if accounts == nil {
		accounts = []*models.Account{}
	}
	utils.Success(w, accounts)
}

// ListContactsSimple handles GET /api/contacts/simple
func (h *ContactHandler) ListSimple(w http.ResponseWriter, r *http.Request) {
	contacts, err := h.contactRepo.ListSimple()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch contacts")
		return
	}
	if contacts == nil {
		contacts = []*models.Contact{}
	}
	utils.Success(w, contacts)
}
