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

