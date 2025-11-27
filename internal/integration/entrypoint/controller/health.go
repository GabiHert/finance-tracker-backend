// Package controller implements HTTP handlers for the API endpoints.
package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthController handles health check endpoints.
type HealthController struct {
	dbHealthChecker func() bool
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

// NewHealthController creates a new health controller instance.
func NewHealthController(dbHealthChecker func() bool) *HealthController {
	return &HealthController{
		dbHealthChecker: dbHealthChecker,
	}
}

// Check handles GET /health requests.
// It returns the current health status of the API and its dependencies.
func (h *HealthController) Check(c *gin.Context) {
	dbStatus := "disconnected"
	if h.dbHealthChecker != nil && h.dbHealthChecker() {
		dbStatus = "connected"
	}

	response := HealthResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
