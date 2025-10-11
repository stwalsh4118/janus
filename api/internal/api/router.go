package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/api/handlers"
	"github.com/sean/janus/internal/api/middleware"
	"github.com/sean/janus/internal/config"
	"github.com/sean/janus/internal/session"
)

// SetupRouter configures and returns a Gin router
func SetupRouter(cfg *config.Config, sessionManager *session.Manager) *gin.Engine {
	// Set Gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORSConfig(cfg.CORSAllowedOrigins))

	// Create handlers
	healthHandler := handlers.NewHealthHandler(sessionManager)
	sessionHandler := handlers.NewSessionHandler(sessionManager)

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", healthHandler.Handle)

		// Session management
		api.POST("/session/start", sessionHandler.Start)
		api.POST("/ask", sessionHandler.Ask)
		api.POST("/heartbeat", sessionHandler.Heartbeat)
		api.POST("/session/end", sessionHandler.End)
	}

	return router
}
