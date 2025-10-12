package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error     string `json:"error"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Error codes
const (
	ErrSessionNotFound      = "SESSION_NOT_FOUND"
	ErrInvalidSessionID     = "INVALID_SESSION_ID"
	ErrInvalidRequest       = "INVALID_REQUEST"
	ErrProcessSpawnFailed   = "PROCESS_SPAWN_FAILED"
	ErrProcessCommunication = "PROCESS_COMMUNICATION_FAILED"
	ErrTimeout              = "REQUEST_TIMEOUT"
	ErrInternalServer       = "INTERNAL_SERVER_ERROR"
)

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, status int, errorCode string, details string) {
	requestID := ""
	if id, exists := c.Get("request_id"); exists {
		if idStr, ok := id.(string); ok {
			requestID = idStr
		}
	}

	c.JSON(status, ErrorResponse{
		Error:     errorCode,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
