// Package config provides application configuration management.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	Email    EmailConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Environment  string
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret              string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
}

// MinIOConfig holds MinIO/S3 configuration.
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// EmailConfig holds email service configuration.
type EmailConfig struct {
	ResendAPIKey  string
	FromName      string
	FromEmail     string
	AppBaseURL    string
	WorkerEnabled bool
	PollInterval  time.Duration
	BatchSize     int
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			Environment:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://app_user:app_password@localhost:5433/finance_tracker?sslmode=disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvAsDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
			Bucket:    getEnv("MINIO_BUCKET", "finance-tracker-uploads"),
			UseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
		},
		Email: EmailConfig{
			ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
			FromName:      getEnv("RESEND_FROM_NAME", "Finance Tracker"),
			FromEmail:     getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
			AppBaseURL:    getEnv("APP_BASE_URL", "http://localhost:5173"),
			WorkerEnabled: getEnvAsBool("EMAIL_WORKER_ENABLED", true),
			PollInterval:  getEnvAsDuration("EMAIL_WORKER_POLL_INTERVAL", 5*time.Second),
			BatchSize:     getEnvAsInt("EMAIL_WORKER_BATCH_SIZE", 10),
		},
	}
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
