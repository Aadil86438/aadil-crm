package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"crm/config"

	_ "github.com/lib/pq"
)

// DB is the global database connection
var DB *sql.DB

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

// RunMigrations runs all SQL migration files in order
func RunMigrations(migrationsPath string) error {
	// Create migrations tracking table if not exists
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		// Check if already applied
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", file).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			log.Printf("Migration %s already applied, skipping", file)
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(filepath.Join(migrationsPath, file))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		if _, err := DB.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		// Record migration
		if _, err := DB.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		log.Printf("Applied migration: %s", file)
	}

	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck() error {
	return DB.Ping()
}
