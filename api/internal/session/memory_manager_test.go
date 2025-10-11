package session

import (
	"sync"
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("creates session with valid UUID", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if session.ID == "" {
			t.Error("expected non-empty session ID")
		}
		if len(session.ID) != 36 { // UUID v4 length
			t.Errorf("expected UUID length 36, got %d", len(session.ID))
		}
	})

	t.Run("initializes timestamps correctly", func(t *testing.T) {
		before := time.Now()
		session, err := manager.CreateSession()
		after := time.Now()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if session.CreatedAt.Before(before) || session.CreatedAt.After(after) {
			t.Error("CreatedAt timestamp not within expected range")
		}
		if session.LastActivity.Before(before) || session.LastActivity.After(after) {
			t.Error("LastActivity timestamp not within expected range")
		}
		if !session.CreatedAt.Equal(session.LastActivity) {
			t.Error("CreatedAt and LastActivity should be equal on creation")
		}
	})

	t.Run("returns unique IDs for multiple sessions", func(t *testing.T) {
		session1, err1 := manager.CreateSession()
		session2, err2 := manager.CreateSession()
		session3, err3 := manager.CreateSession()

		if err1 != nil || err2 != nil || err3 != nil {
			t.Fatal("expected no errors creating sessions")
		}

		if session1.ID == session2.ID || session2.ID == session3.ID || session1.ID == session3.ID {
			t.Error("expected unique IDs for all sessions")
		}
	})

	t.Run("initializes empty conversation log", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if session.ConversationLog == nil {
			t.Error("expected conversation log to be initialized")
		}
		if len(session.ConversationLog) != 0 {
			t.Errorf("expected empty conversation log, got %d messages", len(session.ConversationLog))
		}
	})
}

func TestGetSession(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("returns correct session by ID", func(t *testing.T) {
		created, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		retrieved, err := manager.GetSession(created.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("expected ID %s, got %s", created.ID, retrieved.ID)
		}
		if !retrieved.CreatedAt.Equal(created.CreatedAt) {
			t.Error("CreatedAt timestamps don't match")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := manager.GetSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session")
		}
		expectedMsg := "session not found: non-existent-id"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("returns deep copy preventing external mutations", func(t *testing.T) {
		created, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		retrieved1, _ := manager.GetSession(created.ID)
		retrieved2, _ := manager.GetSession(created.ID)

		// Mutate retrieved1's conversation log
		retrieved1.ConversationLog = append(retrieved1.ConversationLog, Message{
			Role:      "user",
			Content:   "test message",
			Timestamp: time.Now(),
		})

		// retrieved2 should not be affected
		if len(retrieved2.ConversationLog) != 0 {
			t.Error("external mutation affected another retrieved copy")
		}

		// Verify internal state is not affected
		retrieved3, _ := manager.GetSession(created.ID)
		if len(retrieved3.ConversationLog) != 0 {
			t.Error("external mutation affected internal session state")
		}
	})
}

func TestUpdateActivity(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("updates LastActivity timestamp", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		originalActivity := session.LastActivity
		time.Sleep(10 * time.Millisecond)

		err = manager.UpdateActivity(session.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated, _ := manager.GetSession(session.ID)
		if !updated.LastActivity.After(originalActivity) {
			t.Error("LastActivity should be updated to a later time")
		}
	})

	t.Run("returns error for invalid session ID", func(t *testing.T) {
		err := manager.UpdateActivity("invalid-id")
		if err == nil {
			t.Error("expected error for invalid session ID")
		}
		expectedMsg := "session not found: invalid-id"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestUpdateCursorChatID(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("updates cursor chat ID successfully", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		cursorChatID := "test-chat-id-123"
		err = manager.UpdateCursorChatID(session.ID, cursorChatID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the cursor chat ID was updated
		retrieved, _ := manager.GetSession(session.ID)
		if retrieved.CursorChatID != cursorChatID {
			t.Errorf("expected cursor chat ID %q, got %q", cursorChatID, retrieved.CursorChatID)
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		err := manager.UpdateCursorChatID("non-existent-id", "chat-id")
		if err == nil {
			t.Error("expected error for non-existent session")
		}
		expectedMsg := "session not found: non-existent-id"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestAskQuestion(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("returns error for non-existent session", func(t *testing.T) {
		_, _, err := manager.AskQuestion("non-existent-id", "test question", "/tmp")
		if err == nil {
			t.Error("expected error for non-existent session")
		}
	})

	// Note: Full integration test with actual cursor-agent would require cursor-agent to be installed
	// and would be slow, so we skip it in unit tests. The method is tested via integration tests.
}

func TestAddToConversationLog(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("adds messages successfully", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		messages := []Message{
			{
				Role:      "user",
				Content:   "test question",
				Timestamp: time.Now(),
			},
			{
				Role:      "assistant",
				Content:   "test answer",
				Timestamp: time.Now(),
			},
		}

		err = manager.AddToConversationLog(session.ID, messages)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify messages were added
		retrieved, _ := manager.GetSession(session.ID)
		if len(retrieved.ConversationLog) != 2 {
			t.Errorf("expected 2 messages, got %d", len(retrieved.ConversationLog))
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		messages := []Message{{Role: "user", Content: "test"}}
		err := manager.AddToConversationLog("non-existent-id", messages)
		if err == nil {
			t.Error("expected error for non-existent session")
		}
		expectedMsg := "session not found: non-existent-id"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestEndSession(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("removes session from map", func(t *testing.T) {
		session, err := manager.CreateSession()
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		err = manager.EndSession(session.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify session is removed
		_, err = manager.GetSession(session.ID)
		if err == nil {
			t.Error("expected error when getting ended session")
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		err := manager.EndSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session")
		}
		expectedMsg := "session not found: non-existent-id"
		if err.Error() != expectedMsg {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})
}

func TestGetAllSessions(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("returns all active sessions", func(t *testing.T) {
		session1, _ := manager.CreateSession()
		session2, _ := manager.CreateSession()
		session3, _ := manager.CreateSession()

		sessions := manager.GetAllSessions()
		if len(sessions) != 3 {
			t.Errorf("expected 3 sessions, got %d", len(sessions))
		}

		// Verify all created sessions are in the result
		ids := make(map[string]bool)
		for _, s := range sessions {
			ids[s.ID] = true
		}
		if !ids[session1.ID] || !ids[session2.ID] || !ids[session3.ID] {
			t.Error("not all created sessions were returned")
		}
	})

	t.Run("returns empty slice when no sessions", func(t *testing.T) {
		emptyManager := NewMemorySessionManager()
		sessions := emptyManager.GetAllSessions()
		if sessions == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions, got %d", len(sessions))
		}
	})

	t.Run("returns deep copies preventing external mutations", func(t *testing.T) {
		manager.CreateSession()

		sessions1 := manager.GetAllSessions()
		sessions2 := manager.GetAllSessions()

		// Mutate first result
		if len(sessions1) > 0 {
			sessions1[0].ConversationLog = append(sessions1[0].ConversationLog, Message{
				Role:    "user",
				Content: "test",
			})
		}

		// Second result should not be affected
		if len(sessions2) > 0 && len(sessions2[0].ConversationLog) != 0 {
			t.Error("external mutation affected another retrieved copy")
		}
	})
}

func TestCleanupInactiveSessions(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("removes sessions older than timeout", func(t *testing.T) {
		// Create an old session by manipulating internal state
		oldSession, _ := manager.CreateSession()
		time.Sleep(10 * time.Millisecond)

		// Create a new session
		newSession, _ := manager.CreateSession()

		// Cleanup sessions older than 5ms (should only affect oldSession)
		manager.CleanupInactiveSessions(5 * time.Millisecond)

		// Old session should be removed
		_, err := manager.GetSession(oldSession.ID)
		if err == nil {
			t.Error("expected old session to be removed")
		}

		// New session should still exist
		_, err = manager.GetSession(newSession.ID)
		if err != nil {
			t.Error("expected new session to still exist")
		}
	})

	t.Run("keeps active sessions", func(t *testing.T) {
		session, _ := manager.CreateSession()

		// Update activity
		manager.UpdateActivity(session.ID)

		// Cleanup with a short timeout
		manager.CleanupInactiveSessions(1 * time.Millisecond)

		// Session should still exist
		_, err := manager.GetSession(session.ID)
		if err != nil {
			t.Error("expected active session to be kept")
		}
	})

	t.Run("handles empty session map", func(t *testing.T) {
		emptyManager := NewMemorySessionManager()
		// Should not panic
		emptyManager.CleanupInactiveSessions(1 * time.Minute)
	})
}

func TestThreadSafety(t *testing.T) {
	manager := NewMemorySessionManager()

	t.Run("concurrent CreateSession calls", func(t *testing.T) {
		const numGoroutines = 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		sessions := make([]*Session, numGoroutines)
		errors := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				sessions[index], errors[index] = manager.CreateSession()
			}(i)
		}

		wg.Wait()

		// Verify all sessions were created
		for i, err := range errors {
			if err != nil {
				t.Errorf("goroutine %d failed to create session: %v", i, err)
			}
		}

		// Verify all IDs are unique
		ids := make(map[string]bool)
		for i, session := range sessions {
			if session == nil {
				t.Errorf("goroutine %d created nil session", i)
				continue
			}
			if ids[session.ID] {
				t.Errorf("duplicate session ID: %s", session.ID)
			}
			ids[session.ID] = true
		}

		if len(ids) != numGoroutines {
			t.Errorf("expected %d unique IDs, got %d", numGoroutines, len(ids))
		}
	})

	t.Run("concurrent read/write operations", func(t *testing.T) {
		// Create some sessions
		const numSessions = 10
		sessionIDs := make([]string, numSessions)
		for i := 0; i < numSessions; i++ {
			session, _ := manager.CreateSession()
			sessionIDs[i] = session.ID
		}

		const numOperations = 100
		var wg sync.WaitGroup
		wg.Add(numOperations * 4) // 4 types of operations

		// Concurrent GetSession operations
		for i := 0; i < numOperations; i++ {
			go func(index int) {
				defer wg.Done()
				sessionID := sessionIDs[index%numSessions]
				_, _ = manager.GetSession(sessionID)
			}(i)
		}

		// Concurrent UpdateActivity operations
		for i := 0; i < numOperations; i++ {
			go func(index int) {
				defer wg.Done()
				sessionID := sessionIDs[index%numSessions]
				_ = manager.UpdateActivity(sessionID)
			}(i)
		}

		// Concurrent GetAllSessions operations
		for i := 0; i < numOperations; i++ {
			go func() {
				defer wg.Done()
				_ = manager.GetAllSessions()
			}()
		}

		// Concurrent CleanupInactiveSessions operations
		for i := 0; i < numOperations; i++ {
			go func() {
				defer wg.Done()
				manager.CleanupInactiveSessions(1 * time.Hour)
			}()
		}

		wg.Wait()
		// If we reach here without deadlock or panic, thread safety is good
	})

	t.Run("concurrent create and cleanup", func(t *testing.T) {
		const duration = 100 * time.Millisecond
		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: Create sessions
		go func() {
			defer wg.Done()
			end := time.Now().Add(duration)
			for time.Now().Before(end) {
				manager.CreateSession()
				time.Sleep(1 * time.Millisecond)
			}
		}()

		// Goroutine 2: Cleanup sessions
		go func() {
			defer wg.Done()
			end := time.Now().Add(duration)
			for time.Now().Before(end) {
				manager.CleanupInactiveSessions(10 * time.Millisecond)
				time.Sleep(5 * time.Millisecond)
			}
		}()

		wg.Wait()
		// If we reach here without deadlock or panic, thread safety is good
	})
}
