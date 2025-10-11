package session

import (
	"time"
)

// Session represents a cursor-agent chat session
type Session struct {
	ID              string
	CreatedAt       time.Time
	LastActivity    time.Time
	ConversationLog []Message
}

// Message represents a single message in the conversation
type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// Manager manages cursor-agent sessions
// This is a stub implementation for PBI-0
type Manager struct {
	sessions map[string]*Session
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session (stub implementation)
func (m *Manager) CreateSession(sessionID string) *Session {
	session := &Session{
		ID:              sessionID,
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		ConversationLog: []Message{},
	}
	m.sessions[sessionID] = session
	return session
}

// GetSession retrieves a session by ID (stub implementation)
func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	session, exists := m.sessions[sessionID]
	return session, exists
}

// UpdateActivity updates the last activity timestamp
func (m *Manager) UpdateActivity(sessionID string) {
	if session, exists := m.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// EndSession removes a session (stub implementation)
func (m *Manager) EndSession(sessionID string) {
	delete(m.sessions, sessionID)
}

// GetActiveSessions returns the count of active sessions
func (m *Manager) GetActiveSessions() int {
	return len(m.sessions)
}
