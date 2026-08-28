package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"crm/database"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	leadRepo     *repositories.LeadRepository
	contactRepo  *repositories.ContactRepository
	accountRepo  *repositories.AccountRepository
	dealRepo     *repositories.DealRepository
	taskRepo     *repositories.TaskRepository
	activityRepo *repositories.ActivityRepository
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	leadRepo *repositories.LeadRepository,
	contactRepo *repositories.ContactRepository,
	accountRepo *repositories.AccountRepository,
	dealRepo *repositories.DealRepository,
	taskRepo *repositories.TaskRepository,
	activityRepo *repositories.ActivityRepository,
) *DashboardHandler {
	return &DashboardHandler{
		leadRepo:     leadRepo,
		contactRepo:  contactRepo,
		accountRepo:  accountRepo,
		dealRepo:     dealRepo,
		taskRepo:     taskRepo,
		activityRepo: activityRepo,
	}
}

// GetStats handles GET /api/dashboard
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	cacheKey := "dashboard:stats"

	// 1. Check Redis Cache
	if database.RedisClient != nil {
		cachedData, err := database.RedisClient.Get(r.Context(), cacheKey).Result()
		if err == nil && cachedData != "" {
			var respMap map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &respMap); err == nil {
				utils.Success(w, respMap)
				return
			}
		}
	}

	stats := &models.DashboardStats{}

	// Lead stats
	leadsByStatus, _ := h.leadRepo.CountByStatus()
	for status, count := range leadsByStatus {
		stats.TotalLeads += count
		if status == "New" {
			stats.NewLeads = count
		}
		if status == "Qualified" {
			stats.QualifiedLeads = count
		}
	}

	// Contact and account counts
	stats.TotalContacts, _ = h.contactRepo.Count()
	stats.TotalAccounts, _ = h.accountRepo.Count()

	// Deal stats
	pipelineSummary, _ := h.dealRepo.GetPipelineSummary()
	if pipelineSummary != nil {
		stats.OpenDeals = pipelineSummary.TotalDeals - pipelineSummary.WonDeals - pipelineSummary.LostDeals
		stats.WonDeals = pipelineSummary.WonDeals
		stats.LostDeals = pipelineSummary.LostDeals
		stats.TotalRevenue = pipelineSummary.WonRevenue
		stats.PipelineValue = pipelineSummary.TotalValue
	}

	// Activity and task stats
	stats.ActivitiesToday, _ = h.activityRepo.CountDueToday()
	stats.UpcomingTasks, _ = h.taskRepo.CountUpcoming()

	// Chart data
	leadsByStatus2, _ := h.leadRepo.CountByStatus()
	leadsBySource, _ := h.leadRepo.CountBySource()
	dealsByStage, _ := h.dealRepo.CountByStage()
	activitiesByType, _ := h.activityRepo.CountByType()
	monthlyRevenue, _ := h.dealRepo.MonthlyRevenue()

	// Recent data
	recentLeads, _ := h.leadRepo.RecentLeads(5)
	recentDeals, _ := h.dealRepo.RecentDeals(5)

	if recentLeads == nil {
		recentLeads = []*models.Lead{}
	}
	if recentDeals == nil {
		recentDeals = []*models.Deal{}
	}

	responseData := map[string]interface{}{
		"stats":              stats,
		"leads_by_status":    leadsByStatus2,
		"leads_by_source":    leadsBySource,
		"deals_by_stage":     dealsByStage,
		"activities_by_type": activitiesByType,
		"monthly_revenue":    monthlyRevenue,
		"recent_leads":       recentLeads,
		"recent_deals":       recentDeals,
	}

	// Save payload into Redis cache for 60 seconds
	if database.RedisClient != nil {
		if jsonBytes, err := json.Marshal(responseData); err == nil {
			database.RedisClient.Set(r.Context(), cacheKey, jsonBytes, 60*time.Second)
		}
	}

	utils.Success(w, responseData)
}
