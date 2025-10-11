package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/session"
)

// MockSessionManager implements session.Manager for testing
type MockSessionManager struct {
	sessions                map[string]*session.Session
	createSessionError      error
	getSessionError         error
	updateActivityError     error
	updateCursorChatIDError error
	askQuestionFunc         func(id string, question string, workspaceDir string) (string, string, error)
	addToLogError           error
	endSessionError         error
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		sessions: make(map[string]*session.Session),
	}
}

func (m *MockSessionManager) CreateSession() (*session.Session, error) {
	if m.createSessionError != nil {
		return nil, m.createSessionError
	}
	sess := &session.Session{
		ID:              fmt.Sprintf("test-session-%d", len(m.sessions)+1),
		CreatedAt:       time.Now(),
		LastActivity:    time.Now(),
		ConversationLog: make([]session.Message, 0),
	}
	m.sessions[sess.ID] = sess
	return sess, nil
}

func (m *MockSessionManager) GetSession(id string) (*session.Session, error) {
	if m.getSessionError != nil {
		return nil, m.getSessionError
	}
	sess, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess, nil
}

func (m *MockSessionManager) UpdateActivity(id string) error {
	if m.updateActivityError != nil {
		return m.updateActivityError
	}
	sess, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	sess.LastActivity = time.Now()
	return nil
}

func (m *MockSessionManager) UpdateCursorChatID(id string, cursorChatID string) error {
	if m.updateCursorChatIDError != nil {
		return m.updateCursorChatIDError
	}
	sess, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	sess.CursorChatID = cursorChatID
	return nil
}

func (m *MockSessionManager) AskQuestion(id string, question string, workspaceDir string) (string, string, error) {
	if m.askQuestionFunc != nil {
		return m.askQuestionFunc(id, question, workspaceDir)
	}
	sess, exists := m.sessions[id]
	if !exists {
		return "", "", fmt.Errorf("session not found: %s", id)
	}
	// Default mock answer - use existing cursor chat ID or generate one
	cursorChatID := sess.CursorChatID
	if cursorChatID == "" {
		cursorChatID = "mock-cursor-chat-" + id
	}
	return "Mock cursor-agent response to: " + question, cursorChatID, nil
}

func (m *MockSessionManager) AddToConversationLog(id string, messages []session.Message) error {
	if m.addToLogError != nil {
		return m.addToLogError
	}
	sess, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	sess.ConversationLog = append(sess.ConversationLog, messages...)
	return nil
}

func (m *MockSessionManager) EndSession(id string) error {
	if m.endSessionError != nil {
		return m.endSessionError
	}
	if _, exists := m.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	return nil
}

func (m *MockSessionManager) GetAllSessions() []*session.Session {
	sessions := make([]*session.Session, 0, len(m.sessions))
	for _, sess := range m.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

func (m *MockSessionManager) CleanupInactiveSessions(timeout time.Duration) {
	now := time.Now()
	for id, sess := range m.sessions {
		if now.Sub(sess.LastActivity) > timeout {
			delete(m.sessions, id)
		}
	}
}

func TestStartSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates session successfully", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/start", nil)

		handler.Start(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response StartSessionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.SessionID == "" {
			t.Error("expected non-empty session_id")
		}
		if response.Message != "Session started successfully" {
			t.Errorf("unexpected message: %s", response.Message)
		}
	})

	t.Run("returns error when session creation fails", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		mockManager.createSessionError = fmt.Errorf("database connection failed")
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/start", nil)

		handler.Start(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse error response: %v", err)
		}

		if response["error"] != "Failed to create session" {
			t.Errorf("unexpected error message: %v", response["error"])
		}
	})
}

func TestAsk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 400 when session_id is missing", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/ask", nil)

		handler.Ask(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 400 when request body is invalid", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		body := bytes.NewBufferString(`{"invalid":"json"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/ask?session_id=%s", sess.ID), body)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Ask(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 404 when session not found", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		body := bytes.NewBufferString(`{"question":"test"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/ask?session_id=non-existent", body)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Ask(c)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("successfully processes question and returns answer", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()

		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		body := bytes.NewBufferString(`{"question":"What is this codebase?"}`)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/ask?session_id=%s", sess.ID), body)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Ask(c)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response AskResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Answer == "" {
			t.Error("expected non-empty answer")
		}
		if response.SessionID != sess.ID {
			t.Errorf("expected session_id %s, got %s", sess.ID, response.SessionID)
		}
	})

	t.Run("handles cursor-agent error", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()

		// Mock cursor-agent failure
		mockManager.askQuestionFunc = func(id string, question string, workspaceDir string) (string, string, error) {
			return "", "", fmt.Errorf("cursor-agent command failed")
		}

		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		body := bytes.NewBufferString(`{"question":"test"}`)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/ask?session_id=%s", sess.ID), body)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Ask(c)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHeartbeat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 400 when session_id is missing", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/heartbeat", nil)

		handler.Heartbeat(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 404 when session not found", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/heartbeat?session_id=non-existent", nil)

		handler.Heartbeat(c)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("updates activity for valid session", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/heartbeat?session_id=%s", sess.ID), nil)

		handler.Heartbeat(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response HeartbeatResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Message != "Heartbeat received" {
			t.Errorf("expected message 'Heartbeat received', got '%s'", response.Message)
		}

		if response.SessionID != sess.ID {
			t.Errorf("expected session_id '%s', got '%s'", sess.ID, response.SessionID)
		}

		if response.LastActivity.IsZero() {
			t.Error("expected last_activity to be set")
		}
	})

	t.Run("timestamp updates on subsequent heartbeats", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		// First heartbeat
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/heartbeat?session_id=%s", sess.ID), nil)
		handler.Heartbeat(c1)

		var response1 HeartbeatResponse
		json.Unmarshal(w1.Body.Bytes(), &response1)
		firstTimestamp := response1.LastActivity

		// Wait a small amount to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Second heartbeat
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/heartbeat?session_id=%s", sess.ID), nil)
		handler.Heartbeat(c2)

		var response2 HeartbeatResponse
		json.Unmarshal(w2.Body.Bytes(), &response2)
		secondTimestamp := response2.LastActivity

		if !secondTimestamp.After(firstTimestamp) {
			t.Errorf("expected second timestamp %v to be after first %v", secondTimestamp, firstTimestamp)
		}
	})
}

func TestEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 400 when session_id is missing", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/end", nil)

		handler.End(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns 404 when session not found", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/end?session_id=non-existent", nil)

		handler.End(c)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("ends session successfully", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/session/end?session_id=%s", sess.ID), nil)

		handler.End(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response EndSessionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Message != "Session ended successfully" {
			t.Errorf("expected message 'Session ended successfully', got '%s'", response.Message)
		}

		if response.SessionID != sess.ID {
			t.Errorf("expected session_id '%s', got '%s'", sess.ID, response.SessionID)
		}

		// Verify session was removed
		_, err = mockManager.GetSession(sess.ID)
		if err == nil {
			t.Error("expected session to be removed")
		}
	})

	t.Run("ending session twice returns 404 second time", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		// End session first time
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/session/end?session_id=%s", sess.ID), nil)
		handler.End(c1)

		if w1.Code != http.StatusOK {
			t.Errorf("first end: expected status 200, got %d", w1.Code)
		}

		// Try to end same session again
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/session/end?session_id=%s", sess.ID), nil)
		handler.End(c2)

		if w2.Code != http.StatusNotFound {
			t.Errorf("second end: expected status 404, got %d", w2.Code)
		}

		// Verify error message
		var errorResponse map[string]interface{}
		err := json.Unmarshal(w2.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("failed to parse error response: %v", err)
		}

		if errorResponse["error"] != "Session not found" {
			t.Errorf("expected error 'Session not found', got '%v'", errorResponse["error"])
		}
	})
}
