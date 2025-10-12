package session

import (
	"context"
	"sync"
	"time"

	"github.com/sean/janus/internal/logger"
)

const (
	// DefaultCleanupInterval is how often to check for stale sessions
	DefaultCleanupInterval = 1 * time.Minute
)

// CleanupService manages automatic cleanup of inactive sessions
type CleanupService struct {
	manager  Manager
	timeout  time.Duration
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(manager Manager, timeout time.Duration, interval time.Duration) *CleanupService {
	ctx, cancel := context.WithCancel(context.Background())
	return &CleanupService{
		manager:  manager,
		timeout:  timeout,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the cleanup goroutine
func (s *CleanupService) Start() {
	logger.Get().Info().
		Dur("interval", s.interval).
		Dur("timeout", s.timeout).
		Msg("Starting cleanup service")
	go s.run()
}

// Stop gracefully stops the cleanup goroutine
func (s *CleanupService) Stop() {
	logger.Get().Info().Msg("Stopping cleanup service")
	s.stopOnce.Do(func() {
		s.cancel()
	})
}

// run is the main cleanup loop
func (s *CleanupService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			logger.Get().Info().Msg("Cleanup service stopped")
			return
		case <-ticker.C:
			s.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions uses the manager's cleanup method to remove stale sessions
func (s *CleanupService) cleanupInactiveSessions() {
	// Get count before cleanup for logging
	sessionsBefore := len(s.manager.GetAllSessions())

	// Call the manager's cleanup method
	s.manager.CleanupInactiveSessions(s.timeout)

	// Get count after cleanup
	sessionsAfter := len(s.manager.GetAllSessions())

	if sessionsBefore != sessionsAfter {
		removed := sessionsBefore - sessionsAfter
		logger.Get().Info().
			Int("removed", removed).
			Int("active", sessionsAfter).
			Msg("Cleaned up inactive sessions")
	}
}
