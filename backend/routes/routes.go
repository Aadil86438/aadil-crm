package routes

import (
	"net/http"

	"crm/config"
	"crm/database"
	"crm/handlers"
	"crm/middleware"
	"crm/repositories"
)

// Setup configures all application routes and returns the main http.Handler
func Setup(cfg *config.Config) http.Handler {
	frontendURL := cfg.FrontendURL
	// Initialize repositories
	userRepo := repositories.NewUserRepository(database.DB)
	taskRepo := repositories.NewTaskRepository(database.DB)
	activityRepo := repositories.NewActivityRepository(database.DB)
	noteRepo := repositories.NewNoteRepository(database.DB)
	auditRepo := repositories.NewAuditRepository(database.DB)

	// Initialize handlers
	authH := handlers.NewAuthHandler(userRepo, auditRepo)
	registrationH := handlers.NewRegistrationHandler(userRepo)
	adminPanelH := handlers.NewAdminPanelHandler(userRepo)
	taskH := handlers.NewTaskHandler(taskRepo, auditRepo)
	activityH := handlers.NewActivityHandler(activityRepo, auditRepo)
	noteH := handlers.NewNoteHandler(noteRepo, auditRepo)
	dashboardH := handlers.NewDashboardHandler(taskRepo, activityRepo)
	userH := handlers.NewUserHandler(userRepo, auditRepo)
	auditH := handlers.NewAuditHandler(auditRepo)
	healthReportH := handlers.NewHealthReportHandler(cfg)
	k8sH := handlers.NewK8sHandler()
	logViewerH := handlers.NewLogViewerHandler()

	mux := http.NewServeMux()

	// Health check (public)
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/health-report/test-email", func(w http.ResponseWriter, r *http.Request) {
		healthReportH.TriggerHealthReport(w, r)
	})

	// ─── PUBLIC AUTH ROUTES ───────────────────────────────────────
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		authH.Login(w, r)
	})
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		registrationH.Register(w, r)
	})

	// ─── ADMIN PANEL ROUTES (code-gated, not JWT) ─────────────────
	mux.HandleFunc("/api/admin/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		adminPanelH.VerifyCode(w, r)
	})
	// Admin-panel protected routes use JWT token from VerifyCode
	adminAuth := middleware.Auth
	mux.Handle("/api/admin/all", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminPanelH.ListAll(w, r)
		}
	})))
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
	mux.Handle("/api/admin/health-report", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			healthReportH.GetHealthStatus(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/admin/health-report/trigger", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			healthReportH.TriggerHealthReport(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/admin/k8s/status", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			k8sH.GetStatus(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/admin/k8s/kill-pod", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			k8sH.KillPod(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	// ─── LOG VIEWER ROUTES (admin-protected) ──────────────────────
	mux.Handle("/api/admin/logs", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.ListLogs(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/admin/logs/view", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.ViewLog(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/admin/logs/download", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			logViewerH.DownloadLog(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))

	// ─── PROTECTED CRM ROUTES ─────────────────────────────────────
	auth := middleware.Auth

	// Auth - protected
	mux.Handle("/api/auth/logout", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authH.Logout(w, r)
		}
	})))
	mux.Handle("/api/auth/me", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authH.Me(w, r)
		}
	})))

	// Dashboard
	mux.Handle("/api/dashboard", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dashboardH.GetStats(w, r)
	})))

	// Tasks
	mux.Handle("/api/tasks", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskH.List(w, r)
		case http.MethodPost:
			taskH.Create(w, r)
		}
	})))
	mux.Handle("/api/tasks/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskH.Get(w, r)
		case http.MethodPut:
			taskH.Update(w, r)
		case http.MethodDelete:
			taskH.Delete(w, r)
		}
	})))

	// Activities
	mux.Handle("/api/activities", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			activityH.List(w, r)
		case http.MethodPost:
			activityH.Create(w, r)
		}
	})))
	mux.Handle("/api/activities/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			activityH.Delete(w, r)
		}
	})))

	// Calendar
	mux.Handle("/api/calendar", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activityH.Calendar(w, r)
	})))

	// Notes
	mux.Handle("/api/notes", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			noteH.ListByEntity(w, r)
		case http.MethodPost:
			noteH.Create(w, r)
		}
	})))
	mux.Handle("/api/notes/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			noteH.Update(w, r)
		case http.MethodDelete:
			noteH.Delete(w, r)
		}
	})))

	// Users - all authenticated users can list, but only admin can create/edit/delete
	mux.Handle("/api/users/simple", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userH.ListSimple(w, r)
	})))
	mux.Handle("/api/users", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userH.List(w, r)
		case http.MethodPost:
			middleware.RequireAdmin(http.HandlerFunc(userH.Create)).ServeHTTP(w, r)
		}
	})))
	mux.Handle("/api/users/", auth(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if hasPathSuffix(path, "/password") {
			if r.Method == http.MethodPut {
				userH.ResetPassword(w, r)
			}
			return
		}
		switch r.Method {
		case http.MethodPut:
			userH.Update(w, r)
		case http.MethodDelete:
			userH.Delete(w, r)
		}
	}))))

	// Audit logs - admin only
	mux.Handle("/api/audit-logs", auth(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auditH.List(w, r)
	}))))

	// Apply CORS and logger middleware to all routes
	corsMiddleware := middleware.CORS(frontendURL)
	return corsMiddleware(middleware.Logger(mux))
}

// hasPathSuffix checks if a path ends with the given suffix
func hasPathSuffix(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}
