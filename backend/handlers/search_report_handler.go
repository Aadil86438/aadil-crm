package handlers

import (
	"net/http"

	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// SearchHandler handles global search
type SearchHandler struct {
	leadRepo    *repositories.LeadRepository
	contactRepo *repositories.ContactRepository
	accountRepo *repositories.AccountRepository
	dealRepo    *repositories.DealRepository
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(
	leadRepo *repositories.LeadRepository,
	contactRepo *repositories.ContactRepository,
	accountRepo *repositories.AccountRepository,
	dealRepo *repositories.DealRepository,
) *SearchHandler {
	return &SearchHandler{
		leadRepo:    leadRepo,
		contactRepo: contactRepo,
		accountRepo: accountRepo,
		dealRepo:    dealRepo,
	}
}

// Search handles GET /api/search?q=term
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")
	if len(term) < 2 {
		utils.BadRequest(w, "Search term must be at least 2 characters")
		return
	}

	result := &models.SearchResult{
		Leads:    []models.SearchItem{},
		Contacts: []models.SearchItem{},
		Accounts: []models.SearchItem{},
		Deals:    []models.SearchItem{},
		Tasks:    []models.SearchItem{},
	}

	// Search leads
	leads, _ := h.leadRepo.Search(term, 5)
	for _, l := range leads {
		result.Leads = append(result.Leads, models.SearchItem{
			ID:       l.ID,
			Type:     "lead",
			Title:    l.FirstName + " " + l.LastName,
			Subtitle: l.Company,
		})
	}

	// Search contacts
	contacts, _ := h.contactRepo.Search(term, 5)
	for _, c := range contacts {
		result.Contacts = append(result.Contacts, models.SearchItem{
			ID:       c.ID,
			Type:     "contact",
			Title:    c.FirstName + " " + c.LastName,
			Subtitle: c.Email,
		})
	}

	// Search accounts
	accounts, _ := h.accountRepo.Search(term, 5)
	for _, a := range accounts {
		result.Accounts = append(result.Accounts, models.SearchItem{
			ID:       a.ID,
			Type:     "account",
			Title:    a.Name,
			Subtitle: a.Industry,
		})
	}

	// Search deals
	deals, _ := h.dealRepo.Search(term, 5)
	for _, d := range deals {
		subtitle := ""
		if d.Amount != nil {
			subtitle = d.Stage
		}
		result.Deals = append(result.Deals, models.SearchItem{
			ID:       d.ID,
			Type:     "deal",
			Title:    d.Name,
			Subtitle: subtitle,
		})
	}

	utils.Success(w, result)
}

// ReportHandler handles report endpoints
type ReportHandler struct {
	leadRepo    *repositories.LeadRepository
	dealRepo    *repositories.DealRepository
	activityRepo *repositories.ActivityRepository
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(
	leadRepo *repositories.LeadRepository,
	dealRepo *repositories.DealRepository,
	activityRepo *repositories.ActivityRepository,
) *ReportHandler {
	return &ReportHandler{
		leadRepo:     leadRepo,
		dealRepo:     dealRepo,
		activityRepo: activityRepo,
	}
}

// SalesReport handles GET /api/reports/sales
func (h *ReportHandler) SalesReport(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dealRepo.GetPipelineSummary()
	if err != nil {
		utils.InternalServerError(w, "Failed to generate sales report")
		return
	}

	dealsByStage, _ := h.dealRepo.CountByStage()
	monthlyRevenue, _ := h.dealRepo.MonthlyRevenue()

	utils.Success(w, map[string]interface{}{
		"summary":         summary,
		"deals_by_stage":  dealsByStage,
		"monthly_revenue": monthlyRevenue,
	})
}

// LeadReport handles GET /api/reports/leads
func (h *ReportHandler) LeadReport(w http.ResponseWriter, r *http.Request) {
	byStatus, _ := h.leadRepo.CountByStatus()
	bySource, _ := h.leadRepo.CountBySource()

	utils.Success(w, map[string]interface{}{
		"leads_by_status": byStatus,
		"leads_by_source": bySource,
	})
}

// ActivityReport handles GET /api/reports/activities
func (h *ReportHandler) ActivityReport(w http.ResponseWriter, r *http.Request) {
	byType, _ := h.activityRepo.CountByType()

	utils.Success(w, map[string]interface{}{
		"activities_by_type": byType,
	})
}
