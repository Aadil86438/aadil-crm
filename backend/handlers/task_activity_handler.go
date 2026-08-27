package handlers

import (
	"encoding/json"
	"net/http"

	"crm/middleware"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	taskRepo  *repositories.TaskRepository
	auditRepo *repositories.AuditRepository
}

// NewTaskHandler creates a new TaskHandler
func NewTaskHandler(taskRepo *repositories.TaskRepository, auditRepo *repositories.AuditRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo, auditRepo: auditRepo}
}

// List handles GET /api/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)
	_, sortDir := utils.ParseSort(r, []string{"due_date"})

	params := models.TaskListParams{
		Search:   r.URL.Query().Get("search"),
		Status:   r.URL.Query().Get("status"),
		Priority: r.URL.Query().Get("priority"),
		OwnerID:  r.URL.Query().Get("owner_id"),
		Page:     page,
		PageSize: pageSize,
		SortDir:  sortDir,
	}

	tasks, total, err := h.taskRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch tasks")
		return
	}
	if tasks == nil {
		tasks = []*models.Task{}
	}
	utils.Paginated(w, tasks, total, page, pageSize)
}

// Get handles GET /api/tasks/:id
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/tasks/")
	task, err := h.taskRepo.FindByID(id)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch task")
		return
	}
	if task == nil {
		utils.NotFound(w, "Task not found")
		return
	}
	utils.Success(w, task)
}

// Create handles POST /api/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Subject == "" {
		utils.BadRequest(w, "Subject is required")
		return
	}

	task, err := h.taskRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create task")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "create", Entity: "task", EntityID: &task.ID, IPAddress: r.RemoteAddr,
		})
	}

	utils.Created(w, task)
}

// Update handles PUT /api/tasks/:id
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/tasks/")

	existing, err := h.taskRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Task not found")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	task, err := h.taskRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update task")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil {
		h.auditRepo.Create(&models.CreateAuditLogRequest{
			UserID: &claims.UserID, UserEmail: claims.Email,
			Action: "update", Entity: "task", EntityID: &id, IPAddress: r.RemoteAddr,
		})
	}

	utils.Success(w, task)
}

// Delete handles DELETE /api/tasks/:id
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/tasks/")

	existing, err := h.taskRepo.FindByID(id)
	if err != nil || existing == nil {
		utils.NotFound(w, "Task not found")
		return
	}

	if err := h.taskRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete task")
		return
	}

	utils.Success(w, map[string]string{"message": "Task deleted successfully"})
}

// ActivityHandler handles activity-related HTTP requests
type ActivityHandler struct {
	activityRepo *repositories.ActivityRepository
	auditRepo    *repositories.AuditRepository
}

// NewActivityHandler creates a new ActivityHandler
func NewActivityHandler(activityRepo *repositories.ActivityRepository, auditRepo *repositories.AuditRepository) *ActivityHandler {
	return &ActivityHandler{activityRepo: activityRepo, auditRepo: auditRepo}
}

// List handles GET /api/activities
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := utils.ParsePagination(r)

	params := models.ActivityListParams{
		Type:             r.URL.Query().Get("type"),
		RelatedLeadID:    r.URL.Query().Get("related_lead_id"),
		RelatedContactID: r.URL.Query().Get("related_contact_id"),
		RelatedAccountID: r.URL.Query().Get("related_account_id"),
		RelatedDealID:    r.URL.Query().Get("related_deal_id"),
		OwnerID:          r.URL.Query().Get("owner_id"),
		Page:             page,
		PageSize:         pageSize,
	}

	activities, total, err := h.activityRepo.List(params)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch activities")
		return
	}
	if activities == nil {
		activities = []*models.Activity{}
	}
	utils.Paginated(w, activities, total, page, pageSize)
}

// Create handles POST /api/activities
func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	validTypes := map[string]bool{"call": true, "meeting": true, "task": true, "note": true, "email": true}
	if !validTypes[req.Type] {
		utils.BadRequest(w, "Invalid activity type")
		return
	}
	if req.Subject == "" {
		utils.BadRequest(w, "Subject is required")
		return
	}

	claims := middleware.GetClaims(r)
	if claims != nil && req.OwnerID == nil {
		req.OwnerID = &claims.UserID
	}

	activity, err := h.activityRepo.Create(&req)
	if err != nil {
		utils.InternalServerError(w, "Failed to create activity")
		return
	}

	utils.Created(w, activity)
}

// Delete handles DELETE /api/activities/:id
func (h *ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/activities/")
	if err := h.activityRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete activity")
		return
	}
	utils.Success(w, map[string]string{"message": "Activity deleted"})
}

// Calendar handles GET /api/calendar
func (h *ActivityHandler) Calendar(w http.ResponseWriter, r *http.Request) {
	// Default to current month if no dates provided
	activities, _, err := h.activityRepo.List(models.ActivityListParams{Page: 1, PageSize: 200})
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch calendar events")
		return
	}
	if activities == nil {
		activities = []*models.Activity{}
	}
	utils.Success(w, activities)
}
