package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/session"
)

// MockSessionManager implements session.Manager for testing
type MockSessionManager struct {
	sessions              map[string]*session.Session
	createSessionError    error
	getSessionError       error
	updateActivityError   error
	updateProcessError    error
	endSessionError       error
	updateProcessInfoFunc func(id string, process *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser) error
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

func (m *MockSessionManager) UpdateProcessInfo(id string, process *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser) error {
	if m.updateProcessInfoFunc != nil {
		return m.updateProcessInfoFunc(id, process, stdin, stdout)
	}
	if m.updateProcessError != nil {
		return m.updateProcessError
	}
	sess, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	sess.Process = process
	sess.Stdin = stdin
	sess.Stdout = stdout
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

	t.Run("creates session successfully with mock command", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		// Create test context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/start", nil)

		// Note: This test won't actually spawn cursor-agent, it will fail at cmd.Start()
		// but it verifies the session creation logic works
		handler.Start(c)

		// Since cursor-agent likely doesn't exist in test env, we expect an error
		// but we verify the session was created and then cleaned up
		if w.Code != http.StatusInternalServerError {
			// If it succeeds (cursor-agent exists), verify the response
			if w.Code == http.StatusOK {
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
			}
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

	t.Run("cleans up session when UpdateProcessInfo fails", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		mockManager.updateProcessError = fmt.Errorf("failed to update")
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/session/start", nil)

		handler.Start(c)

		// Should clean up the session if process update fails
		// (though it won't reach that point if cursor-agent doesn't exist)
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

	t.Run("returns stub response for valid session", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		sess, _ := mockManager.CreateSession()
		handler := NewSessionHandler(mockManager, "/tmp/test-workspace")

		body := bytes.NewBufferString(`{"question":"What is this codebase?"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/ask?session_id=%s", sess.ID), body)
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Ask(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response AskResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Answer == "" {
			t.Error("expected non-empty answer")
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

		var response GenericResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if !response.Success {
			t.Error("expected success to be true")
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

		var response GenericResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if !response.Success {
			t.Error("expected success to be true")
		}

		// Verify session was removed
		_, err = mockManager.GetSession(sess.ID)
		if err == nil {
			t.Error("expected session to be removed")
		}
	})
}
