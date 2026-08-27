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

// ContactHandler handles contact-related HTTP requests
type ContactHandler struct {
	contactRepo *repositories.ContactRepository
	auditRepo   *repositories.AuditRepository
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(contactRepo *repositories.ContactRepository, auditRepo *repositories.AuditRepository) *ContactHandler {
	return &ContactHandler{contactRepo: contactRepo, auditRepo: auditRepo}
}

// List handles GET /api/contacts
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	sortCol, sortDir := utils.ParseSort(r, []string{"first_name", "last_name", "email", "created_at"})

	params := models.ContactListParams{
		Search:     r.URL.Query().Get("search"),
		AccountID:  r.URL.Query().Get("account_id"),
		OwnerID:    r.URL.Query().Get("owner_id"),
		Page:       page,
		PageSize:   pageSize,
		SortColumn: sortCol,
		SortDir:    sortDir,
	}

	contacts, total, err := h.contactRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch contacts")
		return
	}
	if contacts == nil {
		contacts = []*models.Contact{}
	}
	utils.Paginated(w, contacts, total, page, pageSize)
}

// Get handles GET /api/contacts/:id
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/contacts/")
	contact, err := h.contactRepo.FindByID(id)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch contact")
		return
	}
	if contact == nil {
		utils.NotFound(w, "Contact not found")
		return
	}
	utils.Success(w, contact)
}

// Create handles POST /api/contacts
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		utils.BadRequest(w, "First name and last name are required")
		return
	}

	contact, err := h.contactRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create contact")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "create", Entity: "contact", EntityID: &contact.ID, IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, contact)
}

// Update handles PUT /api/contacts/:id
func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/contacts/")

	existing, err := h.contactRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Contact not found")
		return
	}

	var req models.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	contact, err := h.contactRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update contact")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "update", Entity: "contact", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, contact)
}

// Delete handles DELETE /api/contacts/:id
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/contacts/")

	existing, err := h.contactRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Contact not found")
		return
	}

	if err := h.contactRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete contact")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "delete", Entity: "contact", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, map[string]string{"message": "Contact deleted successfully"})
}
