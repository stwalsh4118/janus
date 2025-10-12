package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/api/handlers"
	"github.com/sean/janus/internal/api/middleware"
	"github.com/sean/janus/internal/config"
	"github.com/sean/janus/internal/logger"
	"github.com/sean/janus/internal/session"
)

// SetupRouter configures and returns a Gin router
func SetupRouter(cfg *config.Config, sessionManager session.Manager) *gin.Engine {
	// Set Gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() instead of Default() to have full control over middleware
	router := gin.New()

	// Apply middleware in correct order
	router.Use(middleware.Recovery())                                       // 1st - catch panics
	router.Use(middleware.RequestID())                                      // 2nd - add request ID
	router.Use(middleware.Logger())                                         // 3rd - log with ID
	router.Use(middleware.RequestTimeout(middleware.DefaultRequestTimeout)) // 4th - enforce timeout
	router.Use(middleware.CORSConfig(cfg.CORSAllowedOrigins))               // 5th - CORS headers

	// Create handlers
	healthHandler := handlers.NewHealthHandler(sessionManager)
	sessionHandler := handlers.NewSessionHandler(sessionManager, cfg.WorkspaceDir)

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

	// Log registered routes
	logRoutes(router)

	return router
}

// logRoutes logs all registered routes with zerolog
func logRoutes(router *gin.Engine) {
	routes := router.Routes()

	logger.Get().Info().
		Int("total", len(routes)).
		Msg("Routes registered")

	for _, route := range routes {
		logger.Get().Info().
			Str("method", route.Method).
			Str("path", route.Path).
			Str("handler", route.Handler).
			Msgf("%-6s %s", route.Method, route.Path)
	}
}
