// Package main is the entry point for the Finance Tracker API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/finance-tracker/backend/config"
	"github.com/finance-tracker/backend/internal/infra/db"
	"github.com/finance-tracker/backend/internal/infra/server/router"
	"github.com/finance-tracker/backend/internal/integration/entrypoint/controller"
)

func main() {
	// Load .env file if it exists (development only)
	_ = godotenv.Load()

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	slog.Info("Starting Finance Tracker API",
		"environment", cfg.Server.Environment,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	// Initialize database connection
	var database *db.Database
	var dbHealthChecker func() bool

	database, err := db.NewPostgresConnection(&cfg.Database)
	if err != nil {
		slog.Warn("Database connection failed, running without database",
			"error", err,
		)
		dbHealthChecker = func() bool { return false }
	} else {
		dbHealthChecker = database.HealthCheck
		defer func() {
			if err := database.Close(); err != nil {
				slog.Error("Failed to close database connection", "error", err)
			}
		}()
	}

	// Create health controller with database health checker
	healthController := controller.NewHealthController(dbHealthChecker)

	// Setup router
	r := router.NewRouter(healthController)
	engine := r.Setup(cfg.Server.Environment)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited properly")
}
