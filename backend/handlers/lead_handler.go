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

// LeadHandler handles lead-related HTTP requests
type LeadHandler struct {
	leadRepo    *repositories.LeadRepository
	contactRepo *repositories.ContactRepository
	accountRepo *repositories.AccountRepository
	dealRepo    *repositories.DealRepository
	auditRepo   *repositories.AuditRepository
	db          interface{}
}

// NewLeadHandler creates a new LeadHandler
func NewLeadHandler(
	leadRepo *repositories.LeadRepository,
	contactRepo *repositories.ContactRepository,
	accountRepo *repositories.AccountRepository,
	dealRepo *repositories.DealRepository,
	auditRepo *repositories.AuditRepository,
) *LeadHandler {
	return &LeadHandler{
		leadRepo:    leadRepo,
		contactRepo: contactRepo,
		accountRepo: accountRepo,
		dealRepo:    dealRepo,
		auditRepo:   auditRepo,
	}
}

// List handles GET /api/leads
func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	sortCol, sortDir := utils.ParseSort(r, []string{"first_name", "last_name", "company", "email", "lead_status", "created_at"})

	params := models.LeadListParams{
		Search:     r.URL.Query().Get("search"),
		Status:     r.URL.Query().Get("status"),
		Source:     r.URL.Query().Get("source"),
		OwnerID:    r.URL.Query().Get("owner_id"),
		Page:       page,
		PageSize:   pageSize,
		SortColumn: sortCol,
		SortDir:    sortDir,
	}

	leads, total, err := h.leadRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch leads")
		return
	}
	if leads == nil {
		leads = []*models.Lead{}
	}
	utils.Paginated(w, leads, total, page, pageSize)
}

// Get handles GET /api/leads/:id
func (h *LeadHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/leads/")
	lead, err := h.leadRepo.FindByID(id)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch lead")
		return
	}
	if lead == nil {
		utils.NotFound(w, "Lead not found")
		return
	}
	utils.Success(w, lead)
}

// Create handles POST /api/leads
func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		utils.BadRequest(w, "First name and last name are required")
		return
	}

	if req.LeadStatus == "" {
		req.LeadStatus = "New"
	}

	lead, err := h.leadRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create lead")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID:    &claims.UserID,
			UserEmail: claims.Email,
			Action:    "create",
			Entity:    "lead",
			EntityID:  &lead.ID,
			IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, lead)
}

// Update handles PUT /api/leads/:id
func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/leads/")

	existing, err := h.leadRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Lead not found")
		return
	}

	var req models.UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		utils.BadRequest(w, "First name and last name are required")
		return
	}

	lead, err := h.leadRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update lead")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID:    &claims.UserID,
			UserEmail: claims.Email,
			Action:    "update",
			Entity:    "lead",
			EntityID:  &id,
			IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, lead)
}

// Delete handles DELETE /api/leads/:id
func (h *LeadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/leads/")

	existing, err := h.leadRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Lead not found")
		return
	}

	if err := h.leadRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete lead")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID:    &claims.UserID,
			UserEmail: claims.Email,
			Action:    "delete",
			Entity:    "lead",
			EntityID:  &id,
			IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, map[string]string{"message": "Lead deleted successfully"})
}

// Convert handles POST /api/leads/:id/convert
func (h *LeadHandler) Convert(w http.ResponseWriter, r *http.Request) {
	id := extractIDWithSuffix(r.URL.Path, "/api/leads/", "/convert")

	lead, err := h.leadRepo.FindByID(id)
	if err != nil || lead == nil {
		utils.NotFound(w, "Lead not found")
		return
	}

	if lead.IsConverted {
		utils.Conflict(w, "Lead has already been converted")
		return
	}

	var req models.ConvertLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	claims := middleware.GetClaims(r)
	ownerID := ""
	if claims != nil {
		ownerID = claims.UserID
	}

	result := &models.LeadConversionResult{LeadID: id}

	// Create Account if requested
	if req.CreateAccount {
		accountReq := &models.CreateAccountRequest{
			Name:    lead.Company,
			Phone:   lead.Phone,
			Email:   lead.Email,
			Website: lead.Website,
			Industry: lead.Industry,
			OwnerID: &ownerID,
		}
		account, err := h.accountRepo.Create(accountReq)
		if err != nil {
			utils.InternalServerError(w, "Failed to create account during conversion")
			return
		}
		result.AccountID = &account.ID
		req.AccountID = &account.ID
	}

	// Create Contact if requested
	if req.CreateContact {
		contactReq := &models.CreateContactRequest{
			FirstName: lead.FirstName,
			LastName:  lead.LastName,
			Email:     lead.Email,
			Phone:     lead.Phone,
			Mobile:    lead.Mobile,
			JobTitle:  lead.JobTitle,
			AccountID: req.AccountID,
			OwnerID:   &ownerID,
		}
		contact, err := h.contactRepo.Create(contactReq)
		if err != nil {
			utils.InternalServerError(w, "Failed to create contact during conversion")
			return
		}
		result.ContactID = &contact.ID
	}

	// Create Deal if requested
	if req.CreateDeal && req.DealName != "" {
		dealReq := &models.CreateDealRequest{
			Name:      req.DealName,
			Amount:    req.DealAmount,
			AccountID: req.AccountID,
			Stage:     "Qualification",
			OwnerID:   &ownerID,
		}
		if result.ContactID != nil {
			dealReq.ContactID = result.ContactID
		}
		deal, err := h.dealRepo.Create(dealReq)
		if err != nil {
			utils.InternalServerError(w, "Failed to create deal during conversion")
			return
		}
		result.DealID = &deal.ID
	}

	// Mark lead as converted
	if err := h.leadRepo.MarkConverted(id); err != nil {
		utils.InternalServerError(w, "Failed to mark lead as converted")
		return
	}

	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID:    &claims.UserID,
			UserEmail: claims.Email,
			Action:    "convert",
			Entity:    "lead",
			EntityID:  &id,
			IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, result)
}

// extractID extracts the ID from the URL path by removing the base prefix
func extractID(path, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	return strings.Split(id, "/")[0]
}

// extractIDWithSuffix extracts ID from a path with a known suffix
func extractIDWithSuffix(path, prefix, suffix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	return strings.TrimSuffix(trimmed, suffix)
}
