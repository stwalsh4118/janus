package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemorySessionManager implements Manager interface with in-memory storage
// and thread-safe operations. Returns deep copies to prevent external mutations.
type MemorySessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewMemorySessionManager creates a new in-memory session manager
func NewMemorySessionManager() Manager {
	return &MemorySessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session with a unique ID
func (m *MemorySessionManager) CreateSession() (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:              sessionID,
		CreatedAt:       now,
		LastActivity:    now,
		ConversationLog: make([]Message, 0),
	}

	m.sessions[sessionID] = session

	// Return a clone to prevent external mutations of internal state
	return session.Clone(), nil
}

// GetSession retrieves a session by ID and returns a deep copy
// to prevent external mutations of internal state
func (m *MemorySessionManager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Return a clone to prevent external mutations
	return session.Clone(), nil
}

// UpdateActivity updates the LastActivity timestamp for a session
func (m *MemorySessionManager) UpdateActivity(id string) error {
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
func (m *MemorySessionManager) EndSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(m.sessions, id)
	return nil
}

// GetAllSessions returns all active sessions as deep copies
// to prevent external mutations of internal state
func (m *MemorySessionManager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		// Clone each session to prevent external mutations
		sessions = append(sessions, session.Clone())
	}

	return sessions
}

// CleanupInactiveSessions removes sessions inactive for longer than timeout
func (m *MemorySessionManager) CleanupInactiveSessions(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.Sub(session.LastActivity) > timeout {
			delete(m.sessions, id)
		}
	}
}
