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
	leadRepo := repositories.NewLeadRepository(database.DB)
	contactRepo := repositories.NewContactRepository(database.DB)
	accountRepo := repositories.NewAccountRepository(database.DB)
	dealRepo := repositories.NewDealRepository(database.DB)
	taskRepo := repositories.NewTaskRepository(database.DB)
	activityRepo := repositories.NewActivityRepository(database.DB)
	noteRepo := repositories.NewNoteRepository(database.DB)
	auditRepo := repositories.NewAuditRepository(database.DB)
	regRepo := repositories.NewRegistrationRepository(database.DB)

	// Initialize handlers
	authH := handlers.NewAuthHandler(userRepo, auditRepo)
	registrationH := handlers.NewRegistrationHandler(regRepo, userRepo)
	adminPanelH := handlers.NewAdminPanelHandler(regRepo, userRepo)
	leadH := handlers.NewLeadHandler(leadRepo, contactRepo, accountRepo, dealRepo, auditRepo)
	contactH := handlers.NewContactHandler(contactRepo, auditRepo)
	accountH := handlers.NewAccountHandler(accountRepo, auditRepo)
	dealH := handlers.NewDealHandler(dealRepo, auditRepo)
	taskH := handlers.NewTaskHandler(taskRepo, auditRepo)
	activityH := handlers.NewActivityHandler(activityRepo, auditRepo)
	noteH := handlers.NewNoteHandler(noteRepo, auditRepo)
	dashboardH := handlers.NewDashboardHandler(leadRepo, contactRepo, accountRepo, dealRepo, taskRepo, activityRepo)
	searchH := handlers.NewSearchHandler(leadRepo, contactRepo, accountRepo, dealRepo)
	reportH := handlers.NewReportHandler(leadRepo, dealRepo, activityRepo)
	userH := handlers.NewUserHandler(userRepo, accountRepo, contactRepo, auditRepo)
	auditH := handlers.NewAuditHandler(auditRepo)
	healthReportH := handlers.NewHealthReportHandler(cfg)
	paymentH := handlers.NewPaymentHandler(cfg, regRepo)
	k8sH := handlers.NewK8sHandler()

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
	mux.HandleFunc("/api/auth/submit-payment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		registrationH.SubmitPayment(w, r)
	})
	mux.HandleFunc("/api/auth/registration-status/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			registrationH.CheckStatus(w, r)
		}
	})

	// ─── PUBLIC PAYMENT ROUTES (Razorpay) ─────────────────────────
	mux.HandleFunc("/api/payment/create-order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		paymentH.CreateOrder(w, r)
	})
	mux.HandleFunc("/api/payment/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		paymentH.VerifyPayment(w, r)
	})
	mux.HandleFunc("/api/payment/key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paymentH.GetRazorpayKey(w, r)
		}
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
	mux.Handle("/api/admin/pending", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminPanelH.ListPending(w, r)
		}
	})))
	mux.Handle("/api/admin/all", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminPanelH.ListAll(w, r)
		}
	})))
	mux.Handle("/api/admin/approve/", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			adminPanelH.Approve(w, r)
		}
	})))
	mux.Handle("/api/admin/reject/", adminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			adminPanelH.Reject(w, r)
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

	// Search
	mux.Handle("/api/search", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		searchH.Search(w, r)
	})))

	// Leads
	mux.Handle("/api/leads", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			leadH.List(w, r)
		case http.MethodPost:
			leadH.Create(w, r)
		default:
			http.NotFound(w, r)
		}
	})))
	mux.Handle("/api/leads/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if hasPathSuffix(path, "/convert") {
			if r.Method == http.MethodPost {
				leadH.Convert(w, r)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			leadH.Get(w, r)
		case http.MethodPut:
			leadH.Update(w, r)
		case http.MethodDelete:
			leadH.Delete(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	// Contacts
	mux.Handle("/api/contacts", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contactH.List(w, r)
		case http.MethodPost:
			contactH.Create(w, r)
		}
	})))
	mux.Handle("/api/contacts/simple", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contactH.ListSimple(w, r)
	})))
	mux.Handle("/api/contacts/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contactH.Get(w, r)
		case http.MethodPut:
			contactH.Update(w, r)
		case http.MethodDelete:
			contactH.Delete(w, r)
		}
	})))

	// Accounts
	mux.Handle("/api/accounts", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountH.List(w, r)
		case http.MethodPost:
			accountH.Create(w, r)
		}
	})))
	mux.Handle("/api/accounts/simple", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountH.ListSimple(w, r)
	})))
	mux.Handle("/api/accounts/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountH.Get(w, r)
		case http.MethodPut:
			accountH.Update(w, r)
		case http.MethodDelete:
			accountH.Delete(w, r)
		}
	})))

	// Deals
	mux.Handle("/api/deals/pipeline", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dealH.Pipeline(w, r)
	})))
	mux.Handle("/api/deals", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			dealH.List(w, r)
		case http.MethodPost:
			dealH.Create(w, r)
		}
	})))
	mux.Handle("/api/deals/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if hasPathSuffix(path, "/stage") {
			if r.Method == http.MethodPatch {
				dealH.UpdateStage(w, r)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			dealH.Get(w, r)
		case http.MethodPut:
			dealH.Update(w, r)
		case http.MethodDelete:
			dealH.Delete(w, r)
		}
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

	// Reports
	mux.Handle("/api/reports/sales", auth(http.HandlerFunc(reportH.SalesReport)))
	mux.Handle("/api/reports/leads", auth(http.HandlerFunc(reportH.LeadReport)))
	mux.Handle("/api/reports/activities", auth(http.HandlerFunc(reportH.ActivityReport)))

	// Apply CORS and logger middleware to all routes
	corsMiddleware := middleware.CORS(frontendURL)
	return corsMiddleware(middleware.Logger(mux))
}

// hasPathSuffix checks if a path ends with the given suffix
func hasPathSuffix(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}
