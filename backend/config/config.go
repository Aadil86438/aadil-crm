package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// SMTPConfig holds email dispatch settings
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// RazorpayConfig holds Razorpay payment gateway settings
type RazorpayConfig struct {
	KeyID     string
	KeySecret string
}

// HealthReportConfig holds configuration for the daily health report
type HealthReportConfig struct {
	Enabled   bool
	Recipient string
	Time      string // Format "15:04" e.g., "21:00"
}

// Config holds all application configuration
type Config struct {
	AppEnv       string
	Port         string
	DB           DBConfig
	JWT          JWTConfig
	Redis        RedisConfig
	FrontendURL  string
	SMTP         SMTPConfig
	HealthReport HealthReportConfig
	Razorpay     RazorpayConfig
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Load reads configuration from environment variables
func Load() *Config {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatal("Invalid DB_PORT:", err)
	}

	jwtExp, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION:", err)
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	reportEnabled, _ := strconv.ParseBool(getEnv("ENABLE_DAILY_HEALTH_REPORT", "true"))

	return &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   getEnv("PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("DB_NAME", "crm"),
			User:     getEnv("DB_USER", "crm_user"),
			Password: getEnv("DB_PASSWORD", "2466"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev_secret_change_in_production"),
			Expiration: jwtExp,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:8081"),
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     smtpPort,
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		Razorpay: RazorpayConfig{
			KeyID:     getEnv("RAZORPAY_KEY_ID", ""),
			KeySecret: getEnv("RAZORPAY_KEY_SECRET", ""),
		},
		HealthReport: HealthReportConfig{
			Enabled:   reportEnabled,
			Recipient: getEnv("HEALTH_REPORT_TO", ""),
			Time:      getEnv("HEALTH_REPORT_TIME", "21:00"),
		},
	}
}

// IsDevelopment returns true in development mode
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// IsProduction returns true in production mode
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
