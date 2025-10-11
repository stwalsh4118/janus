package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StubManager is a temporary implementation of Manager interface
// This maintains compatibility with existing handlers until task 1-2
// replaces it with the thread-safe MemorySessionManager
type StubManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewStubManager creates a new stub session manager
func NewStubManager() Manager {
	return &StubManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session with a unique ID
func (m *StubManager) CreateSession() (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:              sessionID,
		CreatedAt:       now,
		LastActivity:    now,
		ConversationLog: make([]Message, 0),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
// WARNING: Returns a pointer to internal state. Callers must not mutate returned session.
// This limitation will be fixed in the MemorySessionManager implementation (task 1-2).
func (m *StubManager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	session, exists := m.sessions[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

// UpdateActivity updates the LastActivity timestamp for a session
func (m *StubManager) UpdateActivity(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	session.LastActivity = time.Now()
	return nil
}

// EndSession removes a session from the manager
func (m *StubManager) EndSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	return nil
}

// GetAllSessions returns all active sessions
// WARNING: Returns pointers to internal state. Callers must treat returned sessions as read-only.
// This limitation will be fixed in the MemorySessionManager implementation (task 1-2).
func (m *StubManager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CleanupInactiveSessions removes sessions inactive for longer than timeout
// NOTE: This method is not invoked in the stub implementation
// Proper background cleanup will be implemented in task 1-7
func (m *StubManager) CleanupInactiveSessions(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.Sub(session.LastActivity) > timeout {
			delete(m.sessions, id)
		}
	}
}
