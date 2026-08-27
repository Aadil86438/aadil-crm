package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"crm/config"
	"crm/database"
	"crm/repositories"
	"crm/models"
	"crm/utils"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(filepath.Join("..", "..", ".env")); err != nil {
		log.Println("No .env file found, using env vars")
	}

	cfg := config.Load()
	utils.SetJWTSecret(cfg.JWT.Secret)

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.DB.Close()

	migrationsPath := filepath.Join("..", "..", "migrations")
	if err := database.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	userRepo := repositories.NewUserRepository(database.DB)
	leadRepo := repositories.NewLeadRepository(database.DB)
	contactRepo := repositories.NewContactRepository(database.DB)
	accountRepo := repositories.NewAccountRepository(database.DB)
	dealRepo := repositories.NewDealRepository(database.DB)
	taskRepo := repositories.NewTaskRepository(database.DB)
	activityRepo := repositories.NewActivityRepository(database.DB)

	// --- Users ---
	fmt.Println("Creating users...")
	adminHash, _ := utils.HashPassword("Admin@123")
	managerHash, _ := utils.HashPassword("Manager@123")
	salesHash, _ := utils.HashPassword("Sales@123")

	adminUser, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Admin User", Email: "admin@crm.local", Role: "admin",
	}, adminHash)
	if err != nil {
		adminUser, _ = userRepo.FindByEmail("admin@crm.local")
		if adminUser != nil {
			userRepo.UpdatePassword(adminUser.ID, adminHash)
		}
	}

	managerUser, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Sales Manager", Email: "manager@crm.local", Role: "manager",
	}, managerHash)
	if err != nil {
		managerUser, _ = userRepo.FindByEmail("manager@crm.local")
	}

	salesUser, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Sales Rep 1", Email: "sales@crm.local", Role: "sales_user",
	}, salesHash)
	if err != nil {
		salesUser, _ = userRepo.FindByEmail("sales@crm.local")
	}

	sales2User, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Sales Rep 2", Email: "sales2@crm.local", Role: "sales_user",
	}, salesHash)
	if err != nil {
		sales2User, _ = userRepo.FindByEmail("sales2@crm.local")
	}

	ownerID1 := adminUser.ID
	ownerID2 := managerUser.ID
	_ = salesUser
	_ = sales2User

	// --- Accounts ---
	fmt.Println("Creating accounts...")
	accounts := []models.CreateAccountRequest{
		{Name: "Tata Consultancy Services", Website: "https://tcs.com", Industry: "Technology", Phone: "+91-22-1234567", Email: "info@tcs.com", AccountType: "Customer", OwnerID: &ownerID1, Country: "India"},
		{Name: "Infosys Limited", Website: "https://infosys.com", Industry: "Technology", Phone: "+91-80-2345678", Email: "contact@infosys.com", AccountType: "Prospect", OwnerID: &ownerID2, Country: "India"},
		{Name: "Wipro Technologies", Website: "https://wipro.com", Industry: "Technology", Phone: "+91-80-3456789", Email: "info@wipro.com", AccountType: "Partner", OwnerID: &ownerID1, Country: "India"},
		{Name: "HCL Technologies", Website: "https://hcl.com", Industry: "Technology", Phone: "+91-120-4567890", Email: "contact@hcl.com", AccountType: "Customer", OwnerID: &ownerID2, Country: "India"},
		{Name: "Tech Mahindra", Website: "https://techmahindra.com", Industry: "Technology", Phone: "+91-20-5678901", Email: "info@techmahindra.com", AccountType: "Prospect", OwnerID: &ownerID1, Country: "India"},
	}
	createdAccounts := make([]*models.Account, 0)
	for _, a := range accounts {
		acc, err := accountRepo.Create(&a)
		if err == nil {
			createdAccounts = append(createdAccounts, acc)
		}
	}

	// --- Contacts ---
	fmt.Println("Creating contacts...")
	contacts := []models.CreateContactRequest{
		{FirstName: "Rahul", LastName: "Kumar", Email: "rahul.kumar@tcs.com", Phone: "+91-9876543210", JobTitle: "IT Director", Department: "Technology", OwnerID: &ownerID1},
		{FirstName: "Priya", LastName: "Sharma", Email: "priya.sharma@infosys.com", Phone: "+91-9876543211", JobTitle: "VP Sales", Department: "Sales", OwnerID: &ownerID2},
		{FirstName: "Amit", LastName: "Patel", Email: "amit.patel@wipro.com", Phone: "+91-9876543212", JobTitle: "CTO", Department: "Technology", OwnerID: &ownerID1},
		{FirstName: "Sunita", LastName: "Singh", Email: "sunita.singh@hcl.com", Phone: "+91-9876543213", JobTitle: "CEO", Department: "Executive", OwnerID: &ownerID2},
		{FirstName: "Vikram", LastName: "Mehta", Email: "vikram.mehta@techmahindra.com", Phone: "+91-9876543214", JobTitle: "CFO", Department: "Finance", OwnerID: &ownerID1},
	}
	createdContacts := make([]*models.Contact, 0)
	for i, c := range contacts {
		if i < len(createdAccounts) {
			c.AccountID = &createdAccounts[i].ID
		}
		contact, err := contactRepo.Create(&c)
		if err == nil {
			createdContacts = append(createdContacts, contact)
		}
	}

	// --- Leads ---
	fmt.Println("Creating leads...")
	leads := []models.CreateLeadRequest{
		{FirstName: "Arjun", LastName: "Nair", Company: "Reliance Industries", Email: "arjun@reliance.com", Phone: "+91-22-9876543", LeadSource: "Website", LeadStatus: "New", Industry: "Energy", JobTitle: "Manager IT", OwnerID: &ownerID1, Country: "India"},
		{FirstName: "Deepa", LastName: "Krishnan", Company: "Mahindra Group", Email: "deepa@mahindra.com", Phone: "+91-22-8765432", LeadSource: "Referral", LeadStatus: "Contacted", Industry: "Manufacturing", JobTitle: "Director", OwnerID: &ownerID2, Country: "India"},
		{FirstName: "Suresh", LastName: "Babu", Company: "HDFC Bank", Email: "suresh@hdfc.com", Phone: "+91-22-7654321", LeadSource: "Cold Call", LeadStatus: "Qualified", Industry: "Banking", JobTitle: "VP Technology", OwnerID: &ownerID1, Country: "India"},
		{FirstName: "Anjali", LastName: "Desai", Company: "Larsen & Toubro", Email: "anjali@lt.com", Phone: "+91-22-6543210", LeadSource: "Advertisement", LeadStatus: "New", Industry: "Construction", JobTitle: "Head IT", OwnerID: &ownerID2, Country: "India"},
		{FirstName: "Rajesh", LastName: "Joshi", Company: "Bajaj Auto", Email: "rajesh@bajaj.com", Phone: "+91-20-5432109", LeadSource: "Email", LeadStatus: "Unqualified", Industry: "Automotive", JobTitle: "GM Technology", OwnerID: &ownerID1, Country: "India"},
		{FirstName: "Meena", LastName: "Iyer", Company: "ICICI Bank", Email: "meena@icici.com", Phone: "+91-22-4321098", LeadSource: "Social Media", LeadStatus: "New", Industry: "Banking", JobTitle: "CTO", OwnerID: &ownerID2, Country: "India"},
		{FirstName: "Kiran", LastName: "Rao", Company: "Hindustan Unilever", Email: "kiran@hul.com", Phone: "+91-22-3210987", LeadSource: "Campaign", LeadStatus: "Contacted", Industry: "FMCG", JobTitle: "IT Manager", OwnerID: &ownerID1, Country: "India"},
		{FirstName: "Pooja", LastName: "Gupta", Company: "Axis Bank", Email: "pooja@axis.com", Phone: "+91-22-2109876", LeadSource: "Website", LeadStatus: "Qualified", Industry: "Banking", JobTitle: "Director Technology", OwnerID: &ownerID2, Country: "India"},
	}
	for _, l := range leads {
		leadRepo.Create(&l)
	}

	// --- Deals ---
	fmt.Println("Creating deals...")
	closeDate1 := time.Now().AddDate(0, 1, 0)
	closeDate2 := time.Now().AddDate(0, 2, 0)
	closeDate3 := time.Now().AddDate(0, -1, 0)
	amount1 := 500000.0
	amount2 := 1200000.0
	amount3 := 350000.0
	amount4 := 800000.0
	amount5 := 250000.0

	deals := []models.CreateDealRequest{
		{Name: "TCS - Enterprise License", Amount: &amount1, Stage: "Proposal", Probability: 40, ExpectedCloseDate: &closeDate1, LeadSource: "Referral", OwnerID: &ownerID1},
		{Name: "Infosys - Cloud Migration", Amount: &amount2, Stage: "Negotiation", Probability: 60, ExpectedCloseDate: &closeDate2, LeadSource: "Website", OwnerID: &ownerID2},
		{Name: "Wipro - Security Suite", Amount: &amount3, Stage: "Closed Won", Probability: 100, ExpectedCloseDate: &closeDate3, LeadSource: "Cold Call", OwnerID: &ownerID1},
		{Name: "HCL - DevOps Platform", Amount: &amount4, Stage: "Qualification", Probability: 10, ExpectedCloseDate: &closeDate1, LeadSource: "Advertisement", OwnerID: &ownerID2},
		{Name: "Tech Mahindra - Analytics", Amount: &amount5, Stage: "Closed Lost", Probability: 0, ExpectedCloseDate: &closeDate3, LeadSource: "Email", OwnerID: &ownerID1},
		{Name: "TCS - Support Contract", Amount: &amount3, Stage: "Needs Analysis", Probability: 20, ExpectedCloseDate: &closeDate2, LeadSource: "Referral", OwnerID: &ownerID2},
	}
	for i, d := range deals {
		if i < len(createdAccounts) {
			d.AccountID = &createdAccounts[i].ID
		}
		if i < len(createdContacts) {
			d.ContactID = &createdContacts[i].ID
		}
		dealRepo.Create(&d)
	}

	// --- Tasks ---
	fmt.Println("Creating tasks...")
	dueToday := time.Now()
	dueTomorrow := time.Now().AddDate(0, 0, 1)
	dueNextWeek := time.Now().AddDate(0, 0, 7)

	tasks := []models.CreateTaskRequest{
		{Subject: "Follow up with TCS lead", Description: "Call to discuss proposal", DueDate: &dueToday, Priority: "High", Status: "Not Started", OwnerID: &ownerID1},
		{Subject: "Prepare demo for Infosys", Description: "Cloud migration demo", DueDate: &dueTomorrow, Priority: "High", Status: "In Progress", OwnerID: &ownerID2},
		{Subject: "Send proposal to HCL", Description: "DevOps platform proposal", DueDate: &dueNextWeek, Priority: "Medium", Status: "Not Started", OwnerID: &ownerID1},
		{Subject: "Follow up Wipro contract", Description: "Contract renewal discussion", DueDate: &dueTomorrow, Priority: "Low", Status: "Not Started", OwnerID: &ownerID2},
		{Subject: "Quarterly review meeting", Description: "Team performance review", DueDate: &dueNextWeek, Priority: "Medium", Status: "Not Started", OwnerID: &ownerID1},
	}
	for _, t := range tasks {
		taskRepo.Create(&t)
	}

	// --- Activities ---
	fmt.Println("Creating activities...")
	past1 := time.Now().AddDate(0, 0, -3)
	past2 := time.Now().AddDate(0, 0, -1)
	future1 := time.Now().AddDate(0, 0, 2)

	activities := []models.CreateActivityRequest{
		{Type: "call", Subject: "Initial call with TCS", Description: "Discussed requirements", DueDate: &past1, OwnerID: &ownerID1},
		{Type: "meeting", Subject: "Demo for Infosys", Description: "Cloud platform demo", DueDate: &past2, OwnerID: &ownerID2},
		{Type: "email", Subject: "Sent proposal to HCL", Description: "DevOps platform proposal email", DueDate: &past1, OwnerID: &ownerID1},
		{Type: "call", Subject: "Follow-up call with Wipro", Description: "Discussed security requirements", DueDate: &future1, OwnerID: &ownerID2},
		{Type: "meeting", Subject: "Contract signing TCS", Description: "Enterprise license signing", DueDate: &future1, OwnerID: &ownerID1},
		{Type: "note", Subject: "TCS feedback notes", Description: "Client happy with proposal terms", DueDate: &past2, OwnerID: &ownerID2},
	}
	for _, a := range activities {
		activityRepo.Create(&a)
	}

	fmt.Println("\n✅ Seed data created successfully!")
	fmt.Println("\nDEVELOPMENT CREDENTIALS:")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Admin:   admin@crm.local   / Admin@123")
	fmt.Println("Manager: manager@crm.local / Manager@123")
	fmt.Println("Sales:   sales@crm.local   / Sales@123")
	fmt.Println("─────────────────────────────────────────")
}
