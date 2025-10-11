package session

import (
	"io"
	"os/exec"
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
	Process         *exec.Cmd
	Stdin           io.WriteCloser
	Stdout          io.ReadCloser
	CreatedAt       time.Time
	LastActivity    time.Time
	ConversationLog []Message
}

// Clone creates a deep copy of the Session
// Note: Process, Stdin, and Stdout are not cloned as they represent
// unique system resources that should not be duplicated
func (s *Session) Clone() *Session {
	if s == nil {
		return nil
	}

	// Deep copy the conversation log
	conversationCopy := make([]Message, len(s.ConversationLog))
	copy(conversationCopy, s.ConversationLog)

	return &Session{
		ID:              s.ID,
		Process:         s.Process,
		Stdin:           s.Stdin,
		Stdout:          s.Stdout,
		CreatedAt:       s.CreatedAt,
		LastActivity:    s.LastActivity,
		ConversationLog: conversationCopy,
	}
}
