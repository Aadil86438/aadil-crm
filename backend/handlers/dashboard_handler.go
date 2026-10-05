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
	taskRepo     *repositories.TaskRepository
	activityRepo *repositories.ActivityRepository
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	taskRepo *repositories.TaskRepository,
	activityRepo *repositories.ActivityRepository,
) *DashboardHandler {
	return &DashboardHandler{
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

	// Activity and task stats
	stats.ActivitiesToday, _ = h.activityRepo.CountDueToday()
	stats.UpcomingTasks, _ = h.taskRepo.CountUpcoming()

	activitiesByType, _ := h.activityRepo.CountByType()

	responseData := map[string]interface{}{
		"stats":              stats,
		"activities_by_type": activitiesByType,
	}

	// Save payload into Redis cache for 60 seconds
	if database.RedisClient != nil {
		if jsonBytes, err := json.Marshal(responseData); err == nil {
			database.RedisClient.Set(r.Context(), cacheKey, jsonBytes, 60*time.Second)
		}
	}

	utils.Success(w, responseData)
}
