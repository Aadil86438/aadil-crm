package handlers

import (
	"net/http"

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

// GetStats calls core-service-aws /internal/stats endpoint and returns the data
func (h *CoreStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.core.GetStats()
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	utils.Success(w, stats)
}

// GetAuditLogs proxies audit logs from core-service-aws
type CoreAuditHandler struct {
	core *coreclient.CoreClient
}

func NewCoreAuditHandler(core *coreclient.CoreClient) *CoreAuditHandler {
	return &CoreAuditHandler{core: core}
}

func (h *CoreAuditHandler) List(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20
	logs, err := h.core.GetAuditLogs(page, pageSize)
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	utils.Success(w, logs)
}

// GetUsers proxies user list from core-service-aws
type CoreUsersHandler struct {
	core *coreclient.CoreClient
}

func NewCoreUsersHandler(core *coreclient.CoreClient) *CoreUsersHandler {
	return &CoreUsersHandler{core: core}
}

func (h *CoreUsersHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.core.GetUsers()
	if err != nil {
		utils.InternalServerError(w, "Could not reach core-service-aws: "+err.Error())
		return
	}
	utils.Success(w, users)
}

