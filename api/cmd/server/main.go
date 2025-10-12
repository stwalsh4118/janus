package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sean/janus/internal/api"
	"github.com/sean/janus/internal/config"
	"github.com/sean/janus/internal/logger"
	"github.com/sean/janus/internal/session"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)
	log := logger.Get()

	log.Info().
		Str("port", cfg.Port).
		Msg("Starting Cursor Voice Chat server")
	log.Info().
		Str("log_level", cfg.LogLevel).
		Str("cors_origins", cfg.CORSAllowedOrigins).
		Str("workspace_dir", cfg.WorkspaceDir).
		Msg("Configuration loaded")

	// Create session manager
	sessionManager := session.NewMemorySessionManager()

	// Start cleanup service for inactive sessions
	sessionTimeout := time.Duration(cfg.SessionTimeoutMinutes) * time.Minute
	cleanupService := session.NewCleanupService(
		sessionManager,
		sessionTimeout,
		session.DefaultCleanupInterval,
	)
	cleanupService.Start()

	// Setup router
	router := api.SetupRouter(cfg, sessionManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("address", fmt.Sprintf("http://localhost:%s", cfg.Port)).
			Str("health_check", fmt.Sprintf("http://localhost:%s/api/health", cfg.Port)).
			Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Stop cleanup service
	cleanupService.Stop()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
