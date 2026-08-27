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

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	accountRepo *repositories.AccountRepository
	auditRepo   *repositories.AuditRepository
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(accountRepo *repositories.AccountRepository, auditRepo *repositories.AuditRepository) *AccountHandler {
	return &AccountHandler{accountRepo: accountRepo, auditRepo: auditRepo}
}

// List handles GET /api/accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	sortCol, sortDir := utils.ParseSort(r, []string{"name", "account_type", "created_at"})

	params := models.AccountListParams{
		Search:      r.URL.Query().Get("search"),
		AccountType: r.URL.Query().Get("account_type"),
		OwnerID:     r.URL.Query().Get("owner_id"),
		Page:        page,
		PageSize:    pageSize,
		SortColumn:  sortCol,
		SortDir:     sortDir,
	}

	accounts, total, err := h.accountRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch accounts")
		return
	}
	if accounts == nil {
		accounts = []*models.Account{}
	}
	utils.Paginated(w, accounts, total, page, pageSize)
}

// Get handles GET /api/accounts/:id
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/accounts/")
	account, err := h.accountRepo.FindByID(id)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch account")
		return
	}
	if account == nil {
		utils.NotFound(w, "Account not found")
		return
	}
	utils.Success(w, account)
}

// Create handles POST /api/accounts
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Account name is required")
		return
	}

	account, err := h.accountRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create account")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "create", Entity: "account", EntityID: &account.ID, IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, account)
}

// Update handles PUT /api/accounts/:id
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/accounts/")

	existing, err := h.accountRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Account not found")
		return
	}

	var req models.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Account name is required")
		return
	}

	account, err := h.accountRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update account")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "update", Entity: "account", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, account)
}

// Delete handles DELETE /api/accounts/:id
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/accounts/")

	existing, err := h.accountRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Account not found")
		return
	}

	if err := h.accountRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete account")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "delete", Entity: "account", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, map[string]string{"message": "Account deleted successfully"})
}
