package handlers

import (
	"net/http"

	"crm/config"
	"crm/services"
	"crm/utils"
)

type HealthReportHandler struct {
	Cfg *config.Config
}

func NewHealthReportHandler(cfg *config.Config) *HealthReportHandler {
	return &HealthReportHandler{Cfg: cfg}
}

// GetHealthStatus returns system telemetry in JSON format
func (h *HealthReportHandler) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	data := services.PerformHealthCheck(h.Cfg)
	utils.Success(w, data)
}

// TriggerHealthReport sends an immediate email report
func (h *HealthReportHandler) TriggerHealthReport(w http.ResponseWriter, r *http.Request) {
	go func() {
		_ = services.SendHealthReportEmail(h.Cfg)
	}()

	utils.Success(w, "Health report dispatch triggered in background goroutine")
}
