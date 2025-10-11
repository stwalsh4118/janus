package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/session"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	sessionManager session.Manager
	workspaceDir   string
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager session.Manager, workspaceDir string) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
		workspaceDir:   workspaceDir,
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
	Answer    string `json:"answer"`
	SessionID string `json:"session_id"`
}

// GenericResponse represents a generic success response
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Start handles session start requests
func (h *SessionHandler) Start(c *gin.Context) {
	// Create session in manager
	sess, err := h.sessionManager.CreateSession()
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create session",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Session %s created successfully (cursor chat will be created on first question)", sess.ID)

	response := StartSessionResponse{
		SessionID: sess.ID,
		Message:   "Session started successfully",
	}

	c.JSON(http.StatusOK, response)
}

// Ask handles question requests
func (h *SessionHandler) Ask(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id query parameter is required",
		})
		return
	}

	// Parse request body
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Verify session exists
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Session not found",
			"details": err.Error(),
		})
		return
	}

	// Ask question using cursor-agent command
	answer, cursorChatID, err := h.sessionManager.AskQuestion(sessionID, req.Question, h.workspaceDir)
	if err != nil {
		log.Printf("Failed to ask question for session %s: %v", sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get response from cursor-agent",
			"details": err.Error(),
		})
		return
	}

	// Update cursor chat ID if this was the first question
	if err := h.sessionManager.UpdateCursorChatID(sessionID, cursorChatID); err != nil {
		log.Printf("Warning: failed to update cursor chat ID for session %s: %v", sessionID, err)
	}

	// Update activity timestamp
	if err := h.sessionManager.UpdateActivity(sessionID); err != nil {
		log.Printf("Warning: failed to update activity for session %s: %v", sessionID, err)
	}

	// Add to conversation log
	now := time.Now()
	messages := []session.Message{
		{
			Role:      "user",
			Content:   req.Question,
			Timestamp: now,
		},
		{
			Role:      "assistant",
			Content:   answer,
			Timestamp: time.Now(),
		},
	}

	if err := h.sessionManager.AddToConversationLog(sessionID, messages); err != nil {
		log.Printf("Warning: failed to add to conversation log for session %s: %v", sessionID, err)
		// Don't fail the request, just log the warning
	}

	log.Printf("Session %s: Question processed successfully (cursor chat: %s)", sessionID, cursorChatID)

	response := AskResponse{
		Answer:    answer,
		SessionID: sessionID,
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
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Update activity (non-critical operation)
	if err := h.sessionManager.UpdateActivity(sessionID); err != nil {
		log.Printf("Warning: failed to update activity for session %s: %v", sessionID, err)
	}

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
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// End session
	if err := h.sessionManager.EndSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to end session"})
		return
	}

	response := GenericResponse{
		Success: true,
		Message: "Session ended successfully",
	}

	c.JSON(http.StatusOK, response)
}
