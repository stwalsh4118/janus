package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sean/janus/internal/session"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	sessionManager *session.Manager
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager *session.Manager) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
	}
}

// StartSessionResponse represents the response for starting a session
type StartSessionResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// AskRequest represents a question request
type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

// AskResponse represents a response to a question
type AskResponse struct {
	Answer string `json:"answer"`
}

// GenericResponse represents a generic success response
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Start handles session start requests (stub implementation)
func (h *SessionHandler) Start(c *gin.Context) {
	// Generate a new session ID
	sessionID := uuid.New().String()

	// Create session in manager
	h.sessionManager.CreateSession(sessionID)

	response := StartSessionResponse{
		SessionID: sessionID,
		Message:   "Session started successfully (stub implementation)",
	}

	c.JSON(http.StatusOK, response)
}

// Ask handles question requests (stub implementation)
func (h *SessionHandler) Ask(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Check if session exists
	_, exists := h.sessionManager.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update activity
	h.sessionManager.UpdateActivity(sessionID)

	// Return stub response
	response := AskResponse{
		Answer: "This is a stub response. Cursor-agent integration will be implemented in PBI-2. Your question was: " + req.Question,
	}

	c.JSON(http.StatusOK, response)
}

// Heartbeat handles heartbeat requests (stub implementation)
func (h *SessionHandler) Heartbeat(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Check if session exists
	_, exists := h.sessionManager.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Update activity
	h.sessionManager.UpdateActivity(sessionID)

	response := GenericResponse{
		Success: true,
		Message: "Heartbeat received",
	}

	c.JSON(http.StatusOK, response)
}

// End handles session end requests (stub implementation)
func (h *SessionHandler) End(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Check if session exists
	_, exists := h.sessionManager.GetSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// End session
	h.sessionManager.EndSession(sessionID)

	response := GenericResponse{
		Success: true,
		Message: "Session ended successfully",
	}

	c.JSON(http.StatusOK, response)
}
