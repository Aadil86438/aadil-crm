package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// DBConfig holds database configuration (optional for GCP admin service)
type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// Config holds admin-service-gcp configuration
type Config struct {
	AppEnv              string
	Port                string
	DB                  DBConfig
	JWT                 JWTConfig
	Redis               RedisConfig
	FrontendURL         string
	CoreServiceURL      string
	InternalServiceKey  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	jwtExp, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION:", err)
	}

	port := getEnv("PORT", "8081")

	return &Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   port,
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
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		CoreServiceURL:     getEnv("CORE_SERVICE_URL", "http://localhost:8080"),
		InternalServiceKey: getEnv("INTERNAL_SERVICE_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
