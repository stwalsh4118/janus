package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sean/janus/internal/api/response"
	"github.com/sean/janus/internal/logger"
)

const (
	// DefaultRequestTimeout is the maximum time for a request
	DefaultRequestTimeout = 60 * time.Second
)

// RequestID middleware adds a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RequestTimeout middleware enforces request timeout by setting a context deadline.
// Handlers should check c.Request.Context().Done() and return early on timeout.
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Recovery middleware recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := ""
				if id, exists := c.Get("request_id"); exists {
					if idStr, ok := id.(string); ok {
						requestID = idStr
					}
				}

				logger.Get().Error().
					Str("request_id", requestID).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Interface("panic", err).
					Msg("Panic recovered")

				response.RespondWithError(c, 500, response.ErrInternalServer, "An unexpected error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Logger middleware logs all requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		requestID := ""
		if id, exists := c.Get("request_id"); exists {
			if idStr, ok := id.(string); ok {
				requestID = idStr
			}
		}

		// Use different log levels based on status code
		event := logger.Get().Info()
		if status >= 500 {
			event = logger.Get().Error()
		} else if status >= 400 {
			event = logger.Get().Warn()
		}

		event.
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("duration", duration).
			Msg("Request completed")
	}
}
