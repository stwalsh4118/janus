package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns health response with all fields", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		// Verify all fields present
		if response.Status == "" {
			t.Error("status field is empty")
		}

		if response.Version == "" {
			t.Error("version field is empty")
		}

		if response.UptimeSeconds < 0 {
			t.Error("uptime_seconds should be non-negative")
		}

		if response.ActiveSessions < 0 {
			t.Error("active_sessions should be non-negative")
		}

		if response.MemoryUsageMB <= 0 {
			t.Error("memory_usage_mb should be positive")
		}
	})

	t.Run("returns correct status", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Status != "ok" {
			t.Errorf("expected status 'ok', got '%s'", response.Status)
		}
	})

	t.Run("returns version", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Version == "" {
			t.Error("version should not be empty")
		}

		// Current version should be "1.0.0"
		if response.Version != "1.0.0" {
			t.Errorf("expected version '1.0.0', got '%s'", response.Version)
		}
	})

	t.Run("returns zero active sessions when none exist", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ActiveSessions != 0 {
			t.Errorf("expected 0 active sessions, got %d", response.ActiveSessions)
		}
	})

	t.Run("returns correct active session count", func(t *testing.T) {
		mockManager := NewMockSessionManager()

		// Create 3 sessions
		mockManager.CreateSession()
		mockManager.CreateSession()
		mockManager.CreateSession()

		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ActiveSessions != 3 {
			t.Errorf("expected 3 active sessions, got %d", response.ActiveSessions)
		}
	})

	t.Run("active session count updates dynamically", func(t *testing.T) {
		mockManager := NewMockSessionManager()

		// Create 2 sessions
		sess1, _ := mockManager.CreateSession()
		sess2, _ := mockManager.CreateSession()

		handler := NewHealthHandler(mockManager)

		// First call - should have 2 sessions
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest("GET", "/api/health", nil)
		handler.Handle(c1)

		var response1 HealthResponse
		json.Unmarshal(w1.Body.Bytes(), &response1)

		if response1.ActiveSessions != 2 {
			t.Errorf("first call: expected 2 active sessions, got %d", response1.ActiveSessions)
		}

		// End one session
		mockManager.EndSession(sess1.ID)

		// Second call - should have 1 session
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("GET", "/api/health", nil)
		handler.Handle(c2)

		var response2 HealthResponse
		json.Unmarshal(w2.Body.Bytes(), &response2)

		if response2.ActiveSessions != 1 {
			t.Errorf("second call: expected 1 active session, got %d", response2.ActiveSessions)
		}

		// End remaining session
		mockManager.EndSession(sess2.ID)

		// Third call - should have 0 sessions
		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		c3.Request = httptest.NewRequest("GET", "/api/health", nil)
		handler.Handle(c3)

		var response3 HealthResponse
		json.Unmarshal(w3.Body.Bytes(), &response3)

		if response3.ActiveSessions != 0 {
			t.Errorf("third call: expected 0 active sessions, got %d", response3.ActiveSessions)
		}
	})

	t.Run("uptime increases over time", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		// First call
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Request = httptest.NewRequest("GET", "/api/health", nil)
		handler.Handle(c1)

		var response1 HealthResponse
		json.Unmarshal(w1.Body.Bytes(), &response1)
		firstUptime := response1.UptimeSeconds

		// Wait 2 seconds to ensure integer uptime changes
		time.Sleep(2 * time.Second)

		// Second call
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("GET", "/api/health", nil)
		handler.Handle(c2)

		var response2 HealthResponse
		json.Unmarshal(w2.Body.Bytes(), &response2)
		secondUptime := response2.UptimeSeconds

		if secondUptime <= firstUptime {
			t.Errorf("expected uptime to increase: first=%d, second=%d", firstUptime, secondUptime)
		}

		// Should have increased by at least 1 second
		if secondUptime-firstUptime < 1 {
			t.Errorf("expected uptime to increase by at least 1 second, got %d", secondUptime-firstUptime)
		}
	})

	t.Run("memory usage is reasonable", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		// Memory usage should be positive and reasonable (< 1GB for tests)
		if response.MemoryUsageMB <= 0 {
			t.Error("memory usage should be positive")
		}

		if response.MemoryUsageMB > 1024 {
			t.Errorf("memory usage seems too high: %.2f MB", response.MemoryUsageMB)
		}
	})

	t.Run("response format is consistent", func(t *testing.T) {
		mockManager := NewMockSessionManager()
		handler := NewHealthHandler(mockManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/health", nil)

		handler.Handle(c)

		// Verify it's valid JSON
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("response is not valid JSON: %v", err)
		}

		// Verify all expected fields exist
		requiredFields := []string{"status", "version", "uptime_seconds", "active_sessions", "memory_usage_mb"}
		for _, field := range requiredFields {
			if _, exists := response[field]; !exists {
				t.Errorf("missing required field: %s", field)
			}
		}
	})
}
