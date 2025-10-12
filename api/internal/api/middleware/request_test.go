package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestRequestID verifies RequestID middleware generates unique IDs
func TestRequestID(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())

	var capturedID1, capturedID2 string

	router.GET("/test", func(c *gin.Context) {
		id, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, id)
		c.String(http.StatusOK, "ok")
	})

	// First request
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	capturedID1 = w1.Header().Get("X-Request-ID")
	assert.NotEmpty(t, capturedID1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	capturedID2 = w2.Header().Get("X-Request-ID")
	assert.NotEmpty(t, capturedID2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// IDs should be different
	assert.NotEqual(t, capturedID1, capturedID2)
}

// TestRequestTimeout_CompletesWithinTimeout verifies requests complete normally within timeout
func TestRequestTimeout_CompletesWithinTimeout(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestTimeout(2 * time.Second))

	router.GET("/fast", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.String(http.StatusOK, "completed")
	})

	req := httptest.NewRequest("GET", "/fast", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "completed")
}

// TestRequestTimeout_ExceedsTimeout verifies context deadline is enforced
func TestRequestTimeout_ExceedsTimeout(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestTimeout(100 * time.Millisecond))

	handlerCalled := false
	contextCancelled := false

	router.GET("/slow", func(c *gin.Context) {
		handlerCalled = true
		select {
		case <-c.Request.Context().Done():
			// Context was cancelled due to timeout - handler should detect this
			contextCancelled = true
			c.String(http.StatusRequestTimeout, "timeout detected")
			return
		case <-time.After(2 * time.Second):
			c.String(http.StatusOK, "completed")
		}
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled, "handler should be called")
	assert.True(t, contextCancelled, "context should be cancelled due to timeout")
	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

// TestRecovery_CatchesPanic verifies panic recovery
func TestRecovery_CatchesPanic(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(Recovery())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Panic is logged, test that server doesn't crash
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_SERVER_ERROR")
	assert.Contains(t, w.Body.String(), "request_id")
}

// TestRecovery_DoesNotAffectNormalRequests verifies recovery doesn't affect normal flow
func TestRecovery_DoesNotAffectNormalRequests(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())

	router.GET("/normal", func(c *gin.Context) {
		c.String(http.StatusOK, "all good")
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "all good", w.Body.String())
}

// TestLogger_LogsRequests verifies logger middleware logs correctly
func TestLogger_LogsRequests(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.Use(Logger())

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "logged")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Logger outputs via zerolog, test that request completes successfully
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "logged", w.Body.String())
}

// TestMiddlewareChain verifies all middleware work together
func TestMiddlewareChain(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.Use(RequestID())
	router.Use(Logger())
	router.Use(RequestTimeout(1 * time.Second))

	handlerCalled := false
	router.GET("/chain", func(c *gin.Context) {
		handlerCalled = true

		// Verify request_id is available
		id, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, id)

		// Verify context has timeout
		deadline, ok := c.Request.Context().Deadline()
		assert.True(t, ok)
		assert.True(t, time.Until(deadline) > 0)

		c.String(http.StatusOK, "chain works")
	})

	req := httptest.NewRequest("GET", "/chain", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "chain works", w.Body.String())
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

// TestRequestTimeout_ContextPropagation verifies context is properly propagated
func TestRequestTimeout_ContextPropagation(t *testing.T) {
	router := gin.New()
	router.Use(RequestTimeout(500 * time.Millisecond))

	router.GET("/context", func(c *gin.Context) {
		ctx := c.Request.Context()

		// Verify context has deadline
		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.True(t, time.Until(deadline) > 0)
		assert.True(t, time.Until(deadline) <= 500*time.Millisecond)

		c.String(http.StatusOK, "context ok")
	})

	req := httptest.NewRequest("GET", "/context", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequestID_HeaderSet verifies X-Request-ID header is set
func TestRequestID_HeaderSet(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())

	router.GET("/header", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/header", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	header := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, header)
	assert.Len(t, header, 36) // UUID format length
}
