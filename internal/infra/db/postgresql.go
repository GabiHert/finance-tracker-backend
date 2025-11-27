// Package db provides database connection and management functionality.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/finance-tracker/backend/config"
)

// Database wraps the GORM database connection.
type Database struct {
	db  *gorm.DB
	cfg *config.DatabaseConfig
}

// NewPostgresConnection creates a new PostgreSQL database connection.
func NewPostgresConnection(cfg *config.DatabaseConfig) (*Database, error) {
	// Configure GORM logger based on environment
	gormLogger := logger.Default.LogMode(logger.Silent)
	
	// Open connection
	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established",
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
	)

	return &Database{
		db:  db,
		cfg: cfg,
	}, nil
}

// DB returns the underlying GORM database instance.
func (d *Database) DB() *gorm.DB {
	return d.db
}

// HealthCheck performs a health check on the database connection.
func (d *Database) HealthCheck() bool {
	sqlDB, err := d.db.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB for health check", "error", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		slog.Error("Database health check failed", "error", err)
		return false
	}

	return true
}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB for closing: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	slog.Info("Database connection closed")
	return nil
}

// AutoMigrate runs GORM auto-migration for the given models.
func (d *Database) AutoMigrate(models ...interface{}) error {
	if err := d.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}
	return nil
}
