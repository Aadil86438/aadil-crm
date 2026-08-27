package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"crm/config"
	"crm/database"
	"crm/models"
	"crm/repositories"
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

	// --- Users (2 only) ---
	fmt.Println("Creating users...")
	adminHash, _ := utils.HashPassword("Admin@123")
	salesHash, _ := utils.HashPassword("Sales@123")

	adminUser, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Mohammed Aadil", Email: "admin@propertier.app", Role: "admin",
	}, adminHash)
	if err != nil {
		adminUser, _ = userRepo.FindByEmail("admin@propertier.app")
		if adminUser != nil {
			userRepo.UpdatePassword(adminUser.ID, adminHash)
		}
	}

	salesUser, err := userRepo.Create(&models.CreateUserRequest{
		Name: "Sales Rep", Email: "sales@propertier.app", Role: "sales_user",
	}, salesHash)
	if err != nil {
		salesUser, _ = userRepo.FindByEmail("sales@propertier.app")
	}

	ownerID1 := adminUser.ID
	ownerID2 := salesUser.ID

	// --- Accounts (2 only) ---
	fmt.Println("Creating accounts...")
	accounts := []models.CreateAccountRequest{
		{Name: "Tata Consultancy Services", Website: "https://tcs.com", Industry: "Technology", Phone: "+91-22-1234567", Email: "info@tcs.com", AccountType: "Customer", OwnerID: &ownerID1, Country: "India"},
		{Name: "Infosys Limited", Website: "https://infosys.com", Industry: "Technology", Phone: "+91-80-2345678", Email: "contact@infosys.com", AccountType: "Prospect", OwnerID: &ownerID2, Country: "India"},
	}
	createdAccounts := make([]*models.Account, 0)
	for _, a := range accounts {
		acc, err := accountRepo.Create(&a)
		if err == nil {
			createdAccounts = append(createdAccounts, acc)
		}
	}

	// --- Contacts (2 only) ---
	fmt.Println("Creating contacts...")
	contacts := []models.CreateContactRequest{
		{FirstName: "Rahul", LastName: "Kumar", Email: "rahul.kumar@tcs.com", Phone: "+91-9876543210", JobTitle: "IT Director", Department: "Technology", OwnerID: &ownerID1},
		{FirstName: "Priya", LastName: "Sharma", Email: "priya.sharma@infosys.com", Phone: "+91-9876543211", JobTitle: "VP Sales", Department: "Sales", OwnerID: &ownerID2},
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

	// --- Leads (2 only) ---
	fmt.Println("Creating leads...")
	leads := []models.CreateLeadRequest{
		{FirstName: "Arjun", LastName: "Nair", Company: "Reliance Industries", Email: "arjun@reliance.com", Phone: "+91-22-9876543", LeadSource: "Website", LeadStatus: "New", Industry: "Energy", JobTitle: "Manager IT", OwnerID: &ownerID1, Country: "India"},
		{FirstName: "Deepa", LastName: "Krishnan", Company: "Mahindra Group", Email: "deepa@mahindra.com", Phone: "+91-22-8765432", LeadSource: "Referral", LeadStatus: "Contacted", Industry: "Manufacturing", JobTitle: "Director", OwnerID: &ownerID2, Country: "India"},
	}
	for _, l := range leads {
		leadRepo.Create(&l)
	}

	// --- Deals (2 only) ---
	fmt.Println("Creating deals...")
	closeDate1 := time.Now().AddDate(0, 1, 0)
	closeDate2 := time.Now().AddDate(0, 2, 0)
	amount1 := 500000.0
	amount2 := 1200000.0

	deals := []models.CreateDealRequest{
		{Name: "TCS - Enterprise License", Amount: &amount1, Stage: "Proposal", Probability: 40, ExpectedCloseDate: &closeDate1, LeadSource: "Referral", OwnerID: &ownerID1},
		{Name: "Infosys - Cloud Migration", Amount: &amount2, Stage: "Negotiation", Probability: 60, ExpectedCloseDate: &closeDate2, LeadSource: "Website", OwnerID: &ownerID2},
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

	// --- Tasks (2 only) ---
	fmt.Println("Creating tasks...")
	dueToday := time.Now()
	dueTomorrow := time.Now().AddDate(0, 0, 1)

	tasks := []models.CreateTaskRequest{
		{Subject: "Follow up with TCS lead", Description: "Call to discuss proposal", DueDate: &dueToday, Priority: "High", Status: "Not Started", OwnerID: &ownerID1},
		{Subject: "Prepare demo for Infosys", Description: "Cloud migration demo", DueDate: &dueTomorrow, Priority: "High", Status: "In Progress", OwnerID: &ownerID2},
	}
	for _, t := range tasks {
		taskRepo.Create(&t)
	}

	// --- Activities (2 only) ---
	fmt.Println("Creating activities...")
	past1 := time.Now().AddDate(0, 0, -3)
	past2 := time.Now().AddDate(0, 0, -1)

	activities := []models.CreateActivityRequest{
		{Type: "call", Subject: "Initial call with TCS", Description: "Discussed requirements", DueDate: &past1, OwnerID: &ownerID1},
		{Type: "meeting", Subject: "Demo for Infosys", Description: "Cloud platform demo", DueDate: &past2, OwnerID: &ownerID2},
	}
	for _, a := range activities {
		activityRepo.Create(&a)
	}

	fmt.Println("\n✅ Seed data created successfully!")
	fmt.Println("\nCREDENTIALS:")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Admin: admin@propertier.app / Admin@123")
	fmt.Println("Sales: sales@propertier.app / Sales@123")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Admin Panel Code: 1101")
}
