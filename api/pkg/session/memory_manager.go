package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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

// UpdateCursorChatID updates the cursor-agent chat session ID for a session
func (m *MemorySessionManager) UpdateCursorChatID(id string, cursorChatID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.CursorChatID = cursorChatID
	return nil
}

// CursorAgentResponse represents the JSON response from cursor-agent --print --output-format json
type CursorAgentResponse struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	IsError   bool   `json:"is_error"`
	Result    string `json:"result"`
	SessionID string `json:"session_id"`
}

// AskQuestion sends a question to cursor-agent and returns the answer
// It runs cursor-agent as a command with --print and --resume flags
// The context is used to cancel the command if the request times out
func (m *MemorySessionManager) AskQuestion(ctx context.Context, id string, question string, workspaceDir string) (string, string, error) {
	m.mu.RLock()
	session, exists := m.sessions[id]
	m.mu.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("session not found: %s", id)
	}

	// Build cursor-agent command
	args := []string{"--print", "--output-format", "json"}

	// If we have a cursor chat ID, resume that conversation
	if session.CursorChatID != "" {
		args = append(args, "--resume", session.CursorChatID)
	}

	args = append(args, question)

	// Use CommandContext to respect timeout/cancellation
	cmd := exec.CommandContext(ctx, "cursor-agent", args...)
	cmd.Dir = workspaceDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command - will be killed if context is cancelled
	if err := cmd.Run(); err != nil {
		// Check if error was due to context cancellation
		if ctx.Err() != nil {
			return "", "", fmt.Errorf("cursor-agent command cancelled: %w", ctx.Err())
		}
		return "", "", fmt.Errorf("cursor-agent command failed: %w, stderr: %s", err, stderr.String())
	}

	// Parse JSON response
	var response CursorAgentResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return "", "", fmt.Errorf("failed to parse cursor-agent response: %w, output: %s", err, stdout.String())
	}

	// Check for errors in response
	if response.IsError {
		return "", "", fmt.Errorf("cursor-agent returned error: %s", response.Result)
	}

	return response.Result, response.SessionID, nil
}

// AddToConversationLog appends messages to the session's conversation log
func (m *MemorySessionManager) AddToConversationLog(id string, messages []Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	session.ConversationLog = append(session.ConversationLog, messages...)
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
