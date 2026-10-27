package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"admin-service-gcp/coreclient"
	"admin-service-gcp/utils"
)

// CoreStatsHandler proxies system stats from core-service-aws to the admin frontend
type CoreStatsHandler struct {
	core *coreclient.CoreClient
}

func NewCoreStatsHandler(core *coreclient.CoreClient) *CoreStatsHandler {
	return &CoreStatsHandler{core: core}
}

// GetStats calls core-service-aws /internal/stats endpoint and streams the response
func (h *CoreStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	data, err := h.core.DoRequest("/internal/stats")
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CoreAuditHandler proxies audit logs from core-service-aws
type CoreAuditHandler struct {
	core *coreclient.CoreClient
}

func NewCoreAuditHandler(core *coreclient.CoreClient) *CoreAuditHandler {
	return &CoreAuditHandler{core: core}
}

// List forwards GET /internal/audit-logs with query params
func (h *CoreAuditHandler) List(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageSize := r.URL.Query().Get("page_size")
	if pageSize == "" {
		pageSize = "20"
	}

	path := fmt.Sprintf("/internal/audit-logs?page=%s&page_size=%s", page, pageSize)
	data, err := h.core.DoRequest(path)
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CoreUsersHandler proxies user list from core-service-aws
type CoreUsersHandler struct {
	core *coreclient.CoreClient
}

func NewCoreUsersHandler(core *coreclient.CoreClient) *CoreUsersHandler {
	return &CoreUsersHandler{core: core}
}

// List forwards GET /internal/users with query params
func (h *CoreUsersHandler) List(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageSize := r.URL.Query().Get("page_size")
	if pageSize == "" {
		pageSize = "100"
	}

	path := fmt.Sprintf("/internal/users?page=%s&page_size=%s", page, pageSize)
	data, err := h.core.DoRequest(path)
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Approve forwards POST /api/admin/users/:id/approve to core-service-aws
func (h *CoreUsersHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := extractActionID(r.URL.Path, "approve")
	if id == "" {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	path := fmt.Sprintf("/internal/users/%s/approve", id)
	data, err := h.core.DoPostRequest(path, r.Body)
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Reject forwards POST /api/admin/users/:id/reject to core-service-aws
func (h *CoreUsersHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := extractActionID(r.URL.Path, "reject")
	if id == "" {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	path := fmt.Sprintf("/internal/users/%s/reject", id)
	data, err := h.core.DoPostRequest(path, r.Body)
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func extractActionID(path, action string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == action && i > 0 {
			return parts[i-1]
		}
	}
	return ""
}
