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

// DealHandler handles deal-related HTTP requests
type DealHandler struct {
	dealRepo  *repositories.DealRepository
	auditRepo *repositories.AuditRepository
}

// NewDealHandler creates a new DealHandler
func NewDealHandler(dealRepo *repositories.DealRepository, auditRepo *repositories.AuditRepository) *DealHandler {
	return &DealHandler{dealRepo: dealRepo, auditRepo: auditRepo}
}

// List handles GET /api/deals
func (h *DealHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	sortCol, sortDir := utils.ParseSort(r, []string{"name", "amount", "stage", "expected_close_date", "created_at"})

	params := models.DealListParams{
		Search:     r.URL.Query().Get("search"),
		Stage:      r.URL.Query().Get("stage"),
		AccountID:  r.URL.Query().Get("account_id"),
		OwnerID:    r.URL.Query().Get("owner_id"),
		Page:       page,
		PageSize:   pageSize,
		SortColumn: sortCol,
		SortDir:    sortDir,
	}

	deals, total, err := h.dealRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch deals")
		return
	}
	if deals == nil {
		deals = []*models.Deal{}
	}
	utils.Paginated(w, deals, total, page, pageSize)
}

// Get handles GET /api/deals/:id
func (h *DealHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/deals/")
	deal, err := h.dealRepo.FindByID(id)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch deal")
		return
	}
	if deal == nil {
		utils.NotFound(w, "Deal not found")
		return
	}
	utils.Success(w, deal)
}

// Create handles POST /api/deals
func (h *DealHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		utils.BadRequest(w, "Deal name is required")
		return
	}
	if req.Stage == "" {
		req.Stage = "Qualification"
	}

	deal, err := h.dealRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create deal")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "create", Entity: "deal", EntityID: &deal.ID, IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, deal)
}

// Update handles PUT /api/deals/:id
func (h *DealHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/deals/")

	existing, err := h.dealRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Deal not found")
		return
	}

	var req models.UpdateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	deal, err := h.dealRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update deal")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		if existing.Stage != deal.Stage {
			h.auditRepo.Create(&models.CreateAuditLogRequest{
				UserID: &claims.UserID, UserEmail: claims.Email,
				Action: "stage_change", Entity: "deal", EntityID: &id, IPAddress: r.RemoteAddr,
			})
		} else {
			h.auditRepo.Create(&models.CreateAuditLogRequest{
				UserID: &claims.UserID, UserEmail: claims.Email,
				Action: "update", Entity: "deal", EntityID: &id, IPAddress: r.RemoteAddr,
			})
		}
	}

	utils.Success(w, deal)
}

// UpdateStage handles PATCH /api/deals/:id/stage
func (h *DealHandler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id := extractIDWithSuffix(r.URL.Path, "/api/deals/", "/stage")

	var req models.UpdateDealStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	validStages := map[string]bool{
		"Qualification": true, "Needs Analysis": true, "Proposal": true,
		"Negotiation": true, "Closed Won": true, "Closed Lost": true,
	}
	if !validStages[req.Stage] {
		utils.BadRequest(w, "Invalid stage")
		return
	}

	deal, err := h.dealRepo.UpdateStage(id, req.Stage)
	if err != nil {
		utils.InternalServerError(w, "Failed to update deal stage")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "stage_change", Entity: "deal", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, deal)
}

// Delete handles DELETE /api/deals/:id
func (h *DealHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/deals/")

	existing, err := h.dealRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Deal not found")
		return
	}

	if err := h.dealRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete deal")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "delete", Entity: "deal", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, map[string]string{"message": "Deal deleted successfully"})
}

// Pipeline handles GET /api/deals/pipeline (returns all deals grouped by stage)
func (h *DealHandler) Pipeline(w http.ResponseWriter, r *http.Request) {
	stages := []string{"Qualification", "Needs Analysis", "Proposal", "Negotiation", "Closed Won", "Closed Lost"}
	pipeline := map[string][]*models.Deal{}
	for _, s := range stages {
		pipeline[s] = []*models.Deal{}
	}

	params := models.DealListParams{
		Page:     1,
		PageSize: 500,
	}
	deals, _, err := h.dealRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch pipeline")
		return
	}
	for _, d := range deals {
		pipeline[d.Stage] = append(pipeline[d.Stage], d)
	}

	summary, _ := h.dealRepo.GetPipelineSummary()
	utils.Success(w, map[string]interface{}{
		"pipeline": pipeline,
		"summary":  summary,
	})
}
