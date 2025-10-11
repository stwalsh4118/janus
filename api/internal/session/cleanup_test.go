package session

import (
	"testing"
	"time"
)

func TestNewCleanupService(t *testing.T) {
	manager := NewMemorySessionManager()
	timeout := 10 * time.Minute
	interval := 1 * time.Minute

	service := NewCleanupService(manager, timeout, interval)

	if service == nil {
		t.Fatal("NewCleanupService returned nil")
	}

	if service.manager == nil {
		t.Error("manager not set")
	}

	if service.timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, service.timeout)
	}

	if service.interval != interval {
		t.Errorf("expected interval %v, got %v", interval, service.interval)
	}

	if service.ctx == nil {
		t.Error("context not initialized")
	}

	if service.cancel == nil {
		t.Error("cancel function not initialized")
	}
}

func TestCleanupService_StartStop(t *testing.T) {
	manager := NewMemorySessionManager()
	service := NewCleanupService(manager, 10*time.Minute, 100*time.Millisecond)

	// Start the service
	service.Start()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Give it a moment to stop
	time.Sleep(50 * time.Millisecond)

	// Verify context is done
	select {
	case <-service.ctx.Done():
		// Context properly cancelled
	default:
		t.Error("context not cancelled after Stop()")
	}
}

func TestCleanupService_RemovesInactiveSessions(t *testing.T) {
	manager := NewMemorySessionManager()

	// Create sessions with different activity times
	sess1, _ := manager.CreateSession()
	sess2, _ := manager.CreateSession()
	sess3, _ := manager.CreateSession()

	// Simulate different activity times by manually setting LastActivity
	// We need to access the internal sessions for testing
	// Get the sessions and manipulate them
	allSessions := manager.GetAllSessions()
	if len(allSessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(allSessions))
	}

	// Create a cleanup service with 1 second timeout
	timeout := 1 * time.Second
	service := NewCleanupService(manager, timeout, 100*time.Millisecond)

	// Wait 1.5 seconds so all sessions become inactive
	time.Sleep(1500 * time.Millisecond)

	// Run cleanup manually (not through ticker)
	service.cleanupInactiveSessions()

	// Verify all sessions were removed
	remainingSessions := manager.GetAllSessions()
	if len(remainingSessions) != 0 {
		t.Errorf("expected 0 sessions after cleanup, got %d", len(remainingSessions))
	}

	// Verify specific sessions are gone
	_, err := manager.GetSession(sess1.ID)
	if err == nil {
		t.Error("sess1 should have been removed")
	}

	_, err = manager.GetSession(sess2.ID)
	if err == nil {
		t.Error("sess2 should have been removed")
	}

	_, err = manager.GetSession(sess3.ID)
	if err == nil {
		t.Error("sess3 should have been removed")
	}
}

func TestCleanupService_PreservesActiveSessions(t *testing.T) {
	manager := NewMemorySessionManager()

	// Create sessions
	sess1, _ := manager.CreateSession()
	sess2, _ := manager.CreateSession()

	// Create a cleanup service with 1 second timeout
	timeout := 1 * time.Second
	service := NewCleanupService(manager, timeout, 100*time.Millisecond)

	// Wait a bit but keep sess1 active
	time.Sleep(600 * time.Millisecond)
	manager.UpdateActivity(sess1.ID) // Keep sess1 active

	// Wait more so sess2 becomes inactive
	time.Sleep(600 * time.Millisecond)

	// Run cleanup
	service.cleanupInactiveSessions()

	// Verify sess1 is still there
	_, err := manager.GetSession(sess1.ID)
	if err != nil {
		t.Errorf("sess1 should still exist: %v", err)
	}

	// Verify sess2 was removed
	_, err = manager.GetSession(sess2.ID)
	if err == nil {
		t.Error("sess2 should have been removed")
	}

	remainingSessions := manager.GetAllSessions()
	if len(remainingSessions) != 1 {
		t.Errorf("expected 1 active session, got %d", len(remainingSessions))
	}
}

func TestCleanupService_AutomaticCleanup(t *testing.T) {
	manager := NewMemorySessionManager()

	// Create sessions
	sess1, _ := manager.CreateSession()
	sess2, _ := manager.CreateSession()

	// Create a cleanup service with short timeout and interval for testing
	timeout := 500 * time.Millisecond
	interval := 300 * time.Millisecond
	service := NewCleanupService(manager, timeout, interval)

	// Start the service
	service.Start()
	defer service.Stop()

	// Verify sessions exist initially
	if len(manager.GetAllSessions()) != 2 {
		t.Fatal("expected 2 sessions initially")
	}

	// Wait for sessions to become inactive and cleanup to run
	// We need to wait: timeout + interval + a bit more
	time.Sleep(1 * time.Second)

	// Verify sessions were cleaned up automatically
	remainingSessions := manager.GetAllSessions()
	if len(remainingSessions) != 0 {
		t.Errorf("expected 0 sessions after automatic cleanup, got %d", len(remainingSessions))
	}

	_, err := manager.GetSession(sess1.ID)
	if err == nil {
		t.Error("sess1 should have been automatically removed")
	}

	_, err = manager.GetSession(sess2.ID)
	if err == nil {
		t.Error("sess2 should have been automatically removed")
	}
}

func TestCleanupService_NoCleanupWhenAllActive(t *testing.T) {
	manager := NewMemorySessionManager()

	// Create sessions
	sess1, _ := manager.CreateSession()
	sess2, _ := manager.CreateSession()

	// Create a cleanup service
	timeout := 2 * time.Second
	service := NewCleanupService(manager, timeout, 100*time.Millisecond)

	// Run cleanup immediately (sessions just created, should be active)
	service.cleanupInactiveSessions()

	// Verify no sessions were removed
	remainingSessions := manager.GetAllSessions()
	if len(remainingSessions) != 2 {
		t.Errorf("expected 2 sessions (no cleanup), got %d", len(remainingSessions))
	}

	// Verify specific sessions still exist
	_, err := manager.GetSession(sess1.ID)
	if err != nil {
		t.Errorf("sess1 should still exist: %v", err)
	}

	_, err = manager.GetSession(sess2.ID)
	if err != nil {
		t.Errorf("sess2 should still exist: %v", err)
	}
}
