package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/session"
)

var startTime = time.Now()

// HealthHandler handles health check requests
type HealthHandler struct {
	sessionManager session.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(sessionManager session.Manager) *HealthHandler {
	return &HealthHandler{
		sessionManager: sessionManager,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status         string  `json:"status"`
	Version        string  `json:"version"`
	UptimeSeconds  int64   `json:"uptime_seconds"`
	ActiveSessions int     `json:"active_sessions"`
	MemoryUsageMB  float64 `json:"memory_usage_mb"`
}

// Handle processes health check requests
func (h *HealthHandler) Handle(c *gin.Context) {
	uptime := time.Since(startTime).Seconds()

	// Get active session count
	activeSessions := len(h.sessionManager.GetAllSessions())

	// Get memory usage statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryMB := float64(memStats.Alloc) / 1024 / 1024

	response := HealthResponse{
		Status:         "ok",
		Version:        "1.0.0",
		UptimeSeconds:  int64(uptime),
		ActiveSessions: activeSessions,
		MemoryUsageMB:  memoryMB,
	}

	c.JSON(http.StatusOK, response)
}
