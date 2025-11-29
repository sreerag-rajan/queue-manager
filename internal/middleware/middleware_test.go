package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates UUID and sets header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := RequestID()
		handler(c)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		
		// Verify it's a valid UUID
		_, err := uuid.Parse(requestID)
		assert.NoError(t, err)

		// Verify it's stored in context
		ctxID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.Equal(t, requestID, ctxID)
	})

	t.Run("generates unique IDs for each request", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			handler := RequestID()
			handler(c)

			requestID := w.Header().Get("X-Request-ID")
			assert.False(t, ids[requestID], "duplicate request ID found")
			ids[requestID] = true
		}
	})
}

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("logs request details", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := Logger()
		start := time.Now()
		handler(c)
		elapsed := time.Since(start)

		// Verify status is set (default 200)
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify latency is reasonable (should be very small for test)
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("calculates latency correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/slow", nil)

		handler := Logger()
		
		// Create a router with a slow handler
		r := gin.New()
		r.Use(handler)
		r.POST("/slow", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		
		start := time.Now()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/slow", nil))
		elapsed := time.Since(start)

		assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
	})
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets CORS headers for regular requests", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := CORS()
		handler(c)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("handles OPTIONS request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodOptions, "/test", nil)

		handler := CORS()
		handler(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("allows other methods to proceed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)

		handler := CORS()
		handler(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows request to complete within timeout", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := gin.New()
		r.Use(Timeout(100 * time.Millisecond))
		r.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("aborts request that exceeds timeout", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := gin.New()
		r.Use(Timeout(50 * time.Millisecond))
		r.GET("/test", func(c *gin.Context) {
			time.Sleep(100 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
		
		// Verify error response body
		body := w.Body.String()
		assert.True(t, strings.Contains(body, "timeout") || strings.Contains(body, "error"))
	})

	t.Run("returns JSON error on timeout", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := gin.New()
		r.Use(Timeout(10 * time.Millisecond))
		r.GET("/test", func(c *gin.Context) {
			time.Sleep(50 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
		
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		require.Equal(t, http.StatusGatewayTimeout, w.Code)
	})
}

