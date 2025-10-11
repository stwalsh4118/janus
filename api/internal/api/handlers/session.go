package handlers

import (
	"log"
	"net/http"
	"os/exec"

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
	Answer string `json:"answer"`
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

	// Spawn cursor-agent process
	cmd := exec.Command("cursor-agent")
	cmd.Dir = h.workspaceDir

	// Set up stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Failed to create stdin pipe for session %s: %v", sess.ID, err)
		h.sessionManager.EndSession(sess.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create stdin pipe",
			"details": err.Error(),
		})
		return
	}

	// Set up stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe for session %s: %v", sess.ID, err)
		stdin.Close()
		h.sessionManager.EndSession(sess.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create stdout pipe",
			"details": err.Error(),
		})
		return
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start cursor-agent process for session %s: %v", sess.ID, err)
		stdin.Close()
		stdout.Close()
		h.sessionManager.EndSession(sess.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start cursor-agent process",
			"details": err.Error(),
		})
		return
	}

	// Update session with process info
	if err := h.sessionManager.UpdateProcessInfo(sess.ID, cmd, stdin, stdout); err != nil {
		log.Printf("Failed to update process info for session %s: %v", sess.ID, err)
		// Try to cleanup
		stdin.Close()
		stdout.Close()
		cmd.Process.Kill()
		h.sessionManager.EndSession(sess.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update session with process info",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Session %s started successfully with cursor-agent process (PID: %d)", sess.ID, cmd.Process.Pid)

	response := StartSessionResponse{
		SessionID: sess.ID,
		Message:   "Session started successfully",
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
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update activity (non-critical operation)
	if err := h.sessionManager.UpdateActivity(sessionID); err != nil {
		log.Printf("Warning: failed to update activity for session %s: %v", sessionID, err)
	}

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
