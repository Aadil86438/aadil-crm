package routes

import (
	"net/http"

	"core-service-aws/config"
	"core-service-aws/database"
	"core-service-aws/handlers"
	"core-service-aws/middleware"
	"core-service-aws/repositories"
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
	taskH := handlers.NewTaskHandler(taskRepo, auditRepo)
	activityH := handlers.NewActivityHandler(activityRepo, auditRepo)
	noteH := handlers.NewNoteHandler(noteRepo, auditRepo)
	dashboardH := handlers.NewDashboardHandler(taskRepo, activityRepo)
	userH := handlers.NewUserHandler(userRepo, auditRepo)
	auditH := handlers.NewAuditHandler(auditRepo)

	mux := http.NewServeMux()

	// ─── HEALTH CHECK (public) ─────────────────────────────────────
	mux.HandleFunc("/health", handlers.Health)

	// ─── INTERNAL API (called by admin-service-gcp) ───────────────
	// Protected by X-Internal-Service-Key header (service-to-service auth)
	mux.Handle("/internal/stats", middleware.InternalServiceAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			dashboardH.GetStats(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/internal/audit-logs", middleware.InternalServiceAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditH.List(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/internal/users", middleware.InternalServiceAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userH.List(w, r)
		} else {
			http.NotFound(w, r)
		}
	})))

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

	// ─── PROTECTED ROUTES ─────────────────────────────────────────
	auth := middleware.Auth

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

	// Users
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

	corsMiddleware := middleware.CORS(frontendURL)
	return corsMiddleware(middleware.Logger(mux))
}

func hasPathSuffix(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}
