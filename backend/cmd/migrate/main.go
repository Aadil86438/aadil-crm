package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"crm/config"
	"crm/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(filepath.Join("..", "..", ".env")); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.DB.Close()

	migrationsPath := filepath.Join("..", "..", "migrations")
	if len(os.Args) > 1 {
		migrationsPath = os.Args[1]
	}

	if err := database.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations completed successfully!")
}
