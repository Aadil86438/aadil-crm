package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"core-service-aws/config"
	"core-service-aws/database"
	"core-service-aws/routes"
	"core-service-aws/services"
	"core-service-aws/utils"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()
	utils.SetJWTSecret(cfg.JWT.Secret)

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.DB.Close()

	_ = database.InitRedis(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)

	migrationsPath := filepath.Join(".", "migrations")
	if err := database.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	services.StartDailyHealthReport(cfg)

	handler := routes.Setup(cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[core-service-aws] Starting on %s (env: %s)", addr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[core-service-aws] Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("[core-service-aws] Exited cleanly")
}

