package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/api/response"
	"github.com/sean/janus/internal/logger"
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

// EndSessionResponse represents the response for ending a session
type EndSessionResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// HeartbeatResponse represents the response for a heartbeat request
type HeartbeatResponse struct {
	Message      string    `json:"message"`
	SessionID    string    `json:"session_id"`
	LastActivity time.Time `json:"last_activity"`
}

// Start handles session start requests
func (h *SessionHandler) Start(c *gin.Context) {
	// Create session in manager
	sess, err := h.sessionManager.CreateSession()
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create session")
		response.RespondWithError(c, http.StatusInternalServerError, response.ErrInternalServer, "Failed to create session")
		return
	}

	logger.Get().Info().
		Str("session_id", sess.ID).
		Msg("Session created successfully")

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
		response.RespondWithError(c, http.StatusBadRequest, response.ErrInvalidRequest, "session_id query parameter is required")
		return
	}

	// Parse request body
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(c, http.StatusBadRequest, response.ErrInvalidRequest, "Invalid request body: missing or malformed question field")
		return
	}

	// Verify session exists
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		response.RespondWithError(c, http.StatusNotFound, response.ErrSessionNotFound, "The specified session does not exist or has expired")
		return
	}

	// Ask question using cursor-agent command (with context for timeout)
	answer, cursorChatID, err := h.sessionManager.AskQuestion(c.Request.Context(), sessionID, req.Question, h.workspaceDir)
	if err != nil {
		// Check if the error was due to context timeout
		if c.Request.Context().Err() != nil {
			logger.Get().Warn().
				Str("session_id", sessionID).
				Err(err).
				Msg("Request timed out")
			response.RespondWithError(c, http.StatusRequestTimeout, response.ErrTimeout, "Request to cursor-agent timed out")
			return
		}
		logger.Get().Error().
			Str("session_id", sessionID).
			Err(err).
			Msg("Failed to ask question")
		response.RespondWithError(c, http.StatusInternalServerError, response.ErrProcessCommunication, "Failed to get response from cursor-agent")
		return
	}

	// Update cursor chat ID if this was the first question
	if err := h.sessionManager.UpdateCursorChatID(sessionID, cursorChatID); err != nil {
		logger.Get().Warn().
			Str("session_id", sessionID).
			Str("cursor_chat_id", cursorChatID).
			Err(err).
			Msg("Failed to update cursor chat ID")
	}

	// Update activity timestamp
	if err := h.sessionManager.UpdateActivity(sessionID); err != nil {
		logger.Get().Warn().
			Str("session_id", sessionID).
			Err(err).
			Msg("Failed to update activity")
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
		logger.Get().Warn().
			Str("session_id", sessionID).
			Err(err).
			Msg("Failed to add to conversation log")
		// Don't fail the request, just log the warning
	}

	logger.Get().Info().
		Str("session_id", sessionID).
		Str("cursor_chat_id", cursorChatID).
		Msg("Question processed successfully")

	response := AskResponse{
		Answer:    answer,
		SessionID: sessionID,
	}

	c.JSON(http.StatusOK, response)
}

// Heartbeat handles heartbeat requests
func (h *SessionHandler) Heartbeat(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.RespondWithError(c, http.StatusBadRequest, response.ErrInvalidRequest, "session_id query parameter is required")
		return
	}

	// Verify session exists
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		response.RespondWithError(c, http.StatusNotFound, response.ErrSessionNotFound, "The specified session does not exist or has expired")
		return
	}

	// Update activity timestamp
	if err := h.sessionManager.UpdateActivity(sessionID); err != nil {
		response.RespondWithError(c, http.StatusInternalServerError, response.ErrInternalServer, "Failed to update session activity")
		return
	}

	// Get updated session to return new timestamp
	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		// Unlikely since we just updated it, but handle anyway
		response.RespondWithError(c, http.StatusInternalServerError, response.ErrInternalServer, "Failed to retrieve updated session")
		return
	}

	logger.Get().Debug().
		Str("session_id", sessionID).
		Msg("Heartbeat received")

	response := HeartbeatResponse{
		Message:      "Heartbeat received",
		SessionID:    sessionID,
		LastActivity: sess.LastActivity,
	}

	c.JSON(http.StatusOK, response)
}

// End handles session end requests
func (h *SessionHandler) End(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.RespondWithError(c, http.StatusBadRequest, response.ErrInvalidRequest, "session_id query parameter is required")
		return
	}

	// Verify session exists
	_, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		response.RespondWithError(c, http.StatusNotFound, response.ErrSessionNotFound, "The specified session does not exist or has expired")
		return
	}

	// Remove session from manager
	if err := h.sessionManager.EndSession(sessionID); err != nil {
		response.RespondWithError(c, http.StatusInternalServerError, response.ErrInternalServer, "Failed to end session")
		return
	}

	logger.Get().Info().
		Str("session_id", sessionID).
		Msg("Session ended successfully")

	response := EndSessionResponse{
		Message:   "Session ended successfully",
		SessionID: sessionID,
	}

	c.JSON(http.StatusOK, response)
}
