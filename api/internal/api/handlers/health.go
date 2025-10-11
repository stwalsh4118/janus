package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/session"
)

var startTime = time.Now()

// HealthHandler handles health check requests
type HealthHandler struct {
	sessionManager *session.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(sessionManager *session.Manager) *HealthHandler {
	return &HealthHandler{
		sessionManager: sessionManager,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status         string `json:"status"`
	Version        string `json:"version"`
	UptimeSeconds  int64  `json:"uptime_seconds"`
	ActiveSessions int    `json:"active_sessions"`
}

// Handle processes health check requests
func (h *HealthHandler) Handle(c *gin.Context) {
	uptime := time.Since(startTime).Seconds()

	response := HealthResponse{
		Status:         "ok",
		Version:        "1.0.0",
		UptimeSeconds:  int64(uptime),
		ActiveSessions: h.sessionManager.GetActiveSessions(),
	}

	c.JSON(http.StatusOK, response)
}
