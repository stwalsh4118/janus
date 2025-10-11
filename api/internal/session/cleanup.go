package session

import (
	"context"
	"log"
	"sync"
	"time"
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
	log.Printf("Starting cleanup service (checking every %v, timeout %v)", s.interval, s.timeout)
	go s.run()
}

// Stop gracefully stops the cleanup goroutine
func (s *CleanupService) Stop() {
	log.Println("Stopping cleanup service...")
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
			log.Println("Cleanup service stopped")
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
		log.Printf("Cleanup: removed %d inactive session(s), %d active session(s) remaining",
			removed, sessionsAfter)
	}
}
