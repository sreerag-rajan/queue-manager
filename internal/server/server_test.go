package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"queue-manager/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Config{AppHost: "127.0.0.1", AppPort: "0"}
	s := New(cfg, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}


