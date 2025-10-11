package session

import (
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents an active cursor-agent chat session
type Session struct {
	ID              string
	CursorChatID    string // Cursor-agent's internal chat session ID for --resume
	CreatedAt       time.Time
	LastActivity    time.Time
	ConversationLog []Message
}

// Clone creates a deep copy of the Session
func (s *Session) Clone() *Session {
	if s == nil {
		return nil
	}

	// Deep copy the conversation log
	conversationCopy := make([]Message, len(s.ConversationLog))
	copy(conversationCopy, s.ConversationLog)

	return &Session{
		ID:              s.ID,
		CursorChatID:    s.CursorChatID,
		CreatedAt:       s.CreatedAt,
		LastActivity:    s.LastActivity,
		ConversationLog: conversationCopy,
	}
}
