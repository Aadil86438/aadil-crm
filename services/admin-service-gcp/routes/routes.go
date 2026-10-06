package routes

import (
	"net/http"

	"admin-service-gcp/coreclient"
	"admin-service-gcp/config"
	"admin-service-gcp/database"
	"admin-service-gcp/handlers"
	"admin-service-gcp/middleware"
)

// Setup configures admin-service-gcp routes
// This service runs on GCP and communicates with core-service-aws on AWS
func Setup(cfg *config.Config) http.Handler {
	frontendURL := cfg.FrontendURL

	// ── Inter-Service Client ──────────────────────────────────────────────────
	// This is the key component: an HTTP client that calls core-service-aws on AWS
	// using the shared INTERNAL_SERVICE_KEY for authentication.
	core := coreclient.New(cfg.CoreServiceURL, cfg.InternalServiceKey)

	// ── Local Admin Handlers (DevOps tools that run locally on GCP) ──────────
	adminPanelH := handlers.NewAdminPanelHandler(nil) // no DB — pure admin ops
	k8sH := handlers.NewK8sHandler()
	logViewerH := handlers.NewLogViewerHandler()

	// ── Proxy Handlers (call core-service-aws on AWS via HTTP) ───────────────
	coreStatsH := handlers.NewCoreStatsHandler(core)
	coreAuditH := handlers.NewCoreAuditHandler(core)
	coreUsersH := handlers.NewCoreUsersHandler(core)

	// ── Redis (connected to the same Redis as AWS, or GCP-local Redis) ────────
	_ = database.RedisClient // already initialized in main.go

	mux := http.NewServeMux()

	// ─── HEALTH CHECK ─────────────────────────────────────────────────────────
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"admin-service-gcp","status":"ok","cloud":"GCP"}`))
	})

	// ─── ADMIN VERIFY (PIN gate — public) ─────────────────────────────────────
	mux.HandleFunc("/api/admin/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		adminPanelH.VerifyCode(w, r)
	})

	// ─── PROTECTED ADMIN ROUTES ───────────────────────────────────────────────
	adminAuth := middleware.Auth

	// Redis Inspector (local GCP Redis or shared AWS Redis)
	mux.Handle("/api/admin/redis", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminPanelH.GetRedisData(w, r)
		case http.MethodDelete:
			adminPanelH.DeleteRedisKey(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	// Kubernetes Visualizer (local GCP cluster tools)
	mux.Handle("/api/admin/k8s/status", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			k8sH.GetStatus(w, r)
		}
	})))
	mux.Handle("/api/admin/k8s/kill-pod", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			k8sH.KillPod(w, r)
		}
	})))

	// Log Viewer (reads GCP VM local log files)
	mux.Handle("/api/admin/logs", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.ListLogs(w, r)
		}
	})))
	mux.Handle("/api/admin/logs/view", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.ViewLog(w, r)
		}
	})))
	mux.Handle("/api/admin/logs/download", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.DownloadLog(w, r)
		}
	})))

	// ─── CROSS-CLOUD PROXY ROUTES ─────────────────────────────────────────────
	// These routes call core-service-aws on AWS via HTTP to retrieve live data.
	// The admin frontend on GCP can show AWS data without any DB access on GCP!
	mux.Handle("/api/admin/core/stats", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			coreStatsH.GetStats(w, r)
		}
	})))
	mux.Handle("/api/admin/core/audit-logs", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			coreAuditH.List(w, r)
		}
	})))
	mux.Handle("/api/admin/core/users", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			coreUsersH.List(w, r)
		}
	})))

	corsMiddleware := middleware.CORS(frontendURL)
	return corsMiddleware(middleware.Logger(mux))
}
