package session

import "time"

// Manager handles session lifecycle operations
type Manager interface {
	CreateSession() (*Session, error)
	GetSession(id string) (*Session, error)
	UpdateActivity(id string) error
	EndSession(id string) error
	GetAllSessions() []*Session
	CleanupInactiveSessions(timeout time.Duration)
}

