package session

import (
	"context"
	"time"
)

// Manager handles session lifecycle operations
type Manager interface {
	CreateSession() (*Session, error)
	GetSession(id string) (*Session, error)
	UpdateActivity(id string) error
	UpdateCursorChatID(id string, cursorChatID string) error
	AskQuestion(ctx context.Context, id string, question string, workspaceDir string) (answer string, cursorChatID string, err error)
	AddToConversationLog(id string, messages []Message) error
	EndSession(id string) error
	GetAllSessions() []*Session
	CleanupInactiveSessions(timeout time.Duration)
}
