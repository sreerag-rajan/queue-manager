package rabbitmq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		amqpURI  string
		validate func(t *testing.T, p *Provider)
	}{
		{
			name:    "empty URI",
			amqpURI: "",
			validate: func(t *testing.T, p *Provider) {
				assert.Empty(t, p.amqpURI)
				assert.Empty(t, p.httpURI)
				assert.Equal(t, "guest", p.username)
				assert.Equal(t, "guest", p.password)
			},
		},
		{
			name:    "simple URI without credentials",
			amqpURI: "amqp://localhost:5672/",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "amqp://localhost:5672/", p.amqpURI)
				assert.Equal(t, "http://localhost:15672", p.httpURI)
				assert.Equal(t, "guest", p.username)
				assert.Equal(t, "guest", p.password)
			},
		},
		{
			name:    "URI with credentials",
			amqpURI: "amqp://user:pass@localhost:5672/",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "amqp://user:pass@localhost:5672/", p.amqpURI)
				assert.Equal(t, "http://localhost:15672", p.httpURI)
				assert.Equal(t, "user", p.username)
				assert.Equal(t, "pass", p.password)
			},
		},
		{
			name:    "URI with username only",
			amqpURI: "amqp://user@localhost:5672/",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "user", p.username)
				assert.Equal(t, "guest", p.password) // default
			},
		},
		{
			name:    "invalid URI",
			amqpURI: "://invalid",
			validate: func(t *testing.T, p *Provider) {
				// Should still create provider with defaults
				assert.Equal(t, "://invalid", p.amqpURI)
				assert.Equal(t, "guest", p.username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.amqpURI)
			tt.validate(t, p)
		})
	}
}

func TestNewWithHTTP(t *testing.T) {
	tests := []struct {
		name     string
		amqpURI  string
		httpURI  string
		validate func(t *testing.T, p *Provider)
	}{
		{
			name:    "with HTTP URI",
			amqpURI: "amqp://localhost:5672/",
			httpURI: "http://localhost:15672",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "http://localhost:15672", p.httpURI)
			},
		},
		{
			name:    "HTTP URI with trailing slash",
			amqpURI: "amqp://localhost:5672/",
			httpURI: "http://localhost:15672/",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "http://localhost:15672", p.httpURI) // trailing slash removed
			},
		},
		{
			name:    "HTTP URI with credentials",
			amqpURI: "amqp://user1:pass1@localhost:5672/",
			httpURI: "http://user2:pass2@localhost:15672",
			validate: func(t *testing.T, p *Provider) {
				assert.Equal(t, "http://localhost:15672", p.httpURI) // credentials removed
				assert.Equal(t, "user2", p.username)                 // HTTP credentials override
				assert.Equal(t, "pass2", p.password)
			},
		},
		{
			name:    "empty HTTP URI",
			amqpURI: "amqp://localhost:5672/",
			httpURI: "",
			validate: func(t *testing.T, p *Provider) {
				// Should use default from AMQP URI
				assert.Equal(t, "http://localhost:15672", p.httpURI)
			},
		},
		{
			name:    "invalid HTTP URI",
			amqpURI: "amqp://localhost:5672/",
			httpURI: "://invalid",
			validate: func(t *testing.T, p *Provider) {
				// Should use as-is but remove trailing slash
				assert.Equal(t, "://invalid", p.httpURI)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewWithHTTP(tt.amqpURI, tt.httpURI)
			tt.validate(t, p)
		})
	}
}

func TestProvider_Connect(t *testing.T) {
	t.Run("empty AMQP URI", func(t *testing.T) {
		p := New("")
		err := p.Connect()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RABBITMQ_AMQP_URI is required")
	})

	// Note: Testing actual connection requires a running RabbitMQ instance
	// This would be better suited for integration tests
}

func TestProvider_Close(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		p := New("amqp://localhost:5672/")
		err := p.Close()
		assert.NoError(t, err)
	})

	// Note: Testing actual close requires a connection, better for integration tests
}

func TestProvider_Health(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		p := New("amqp://localhost:5672/")
		status := p.Health()
		assert.False(t, status.OK)
		assert.Contains(t, status.Details, "closed")
	})

	// Note: Testing healthy connection requires actual connection, better for integration tests
}

func TestIsSystemExchange(t *testing.T) {
	tests := []struct {
		name     string
		exchange string
		expected bool
	}{
		{"empty string", "", true},
		{"amq. prefix", "amq.direct", true},
		{"amq. prefix with more", "amq.topic.test", true},
		{"regular exchange", "my-exchange", false},
		{"user exchange", "user-exchange", false},
		{"exchange with dots", "my.exchange", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSystemExchange(tt.exchange)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProvider_ListExchanges(t *testing.T) {
	t.Run("HTTP URI not configured", func(t *testing.T) {
		p := New("")
		p.httpURI = ""
		_, err := p.ListExchanges()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP URI not configured")
	})

	t.Run("successful list", func(t *testing.T) {
		// Mock HTTP server
		exchanges := []map[string]interface{}{
			{"name": "exchange1", "type": "topic"},
			{"name": "exchange2", "type": "direct"},
			{"name": "amq.direct", "type": "direct"}, // system exchange, should be filtered
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The path should contain /api/exchanges
			assert.Contains(t, r.URL.Path, "/api/exchanges")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(exchanges)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		result, err := p.ListExchanges()
		require.NoError(t, err)
		assert.Len(t, result, 2) // system exchange filtered
		assert.Contains(t, result, "exchange1")
		assert.Contains(t, result, "exchange2")
		assert.NotContains(t, result, "amq.direct")
	})

	t.Run("fallback to /exchanges endpoint", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call to /exchanges/%2F returns 404
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// Second call to /exchanges succeeds
			exchanges := []map[string]interface{}{
				{"name": "exchange1", "type": "topic"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(exchanges)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		result, err := p.ListExchanges()
		require.NoError(t, err)
		assert.Len(t, result, 1)
		// The fallback logic may make multiple calls - at least 2 (initial + fallback)
		assert.GreaterOrEqual(t, callCount, 2)
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		_, err := p.ListExchanges()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})
}

func TestProvider_ListQueues(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		queues := []map[string]interface{}{
			{"name": "queue1"},
			{"name": "queue2"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/queues", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(queues)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		result, err := p.ListQueues()
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "queue1")
		assert.Contains(t, result, "queue2")
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		_, err := p.ListQueues()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 401")
	})
}

func TestProvider_ListBindings(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		bindings := []map[string]interface{}{
			{"source": "exchange1", "routing_key": "key1"},
			{"source": "exchange2", "routing_key": "key2"},
			{"source": "", "routing_key": "key3"}, // empty source should be filtered
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "/api/queues")
			assert.Contains(t, r.URL.Path, "test-queue")
			assert.Contains(t, r.URL.Path, "bindings")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bindings)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		result, err := p.ListBindings("test-queue")
		require.NoError(t, err)
		assert.Len(t, result, 2) // empty source filtered
		assert.Equal(t, [3]string{"test-queue", "exchange1", "key1"}, result[0])
		assert.Equal(t, [3]string{"test-queue", "exchange2", "key2"}, result[1])
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		_, err := p.ListBindings("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 404")
	})
}

func TestProvider_DeleteQueue(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Contains(t, r.URL.Path, "/api/queues")
			assert.Contains(t, r.URL.Path, "test-queue")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		err := p.DeleteQueue("test-queue")
		assert.NoError(t, err)
	})

	t.Run("queue not found (idempotent)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		err := p.DeleteQueue("nonexistent")
		assert.NoError(t, err) // Should be idempotent
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		err := p.DeleteQueue("test-queue")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})
}

func TestProvider_DeleteExchange(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Contains(t, r.URL.Path, "/api/exchanges")
			assert.Contains(t, r.URL.Path, "test-exchange")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		err := p.DeleteExchange("test-exchange")
		assert.NoError(t, err)
	})

	t.Run("system exchange error", func(t *testing.T) {
		p := New("amqp://localhost:5672/")
		err := p.DeleteExchange("amq.direct")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system exchange")
	})

	t.Run("empty exchange error", func(t *testing.T) {
		p := New("amqp://localhost:5672/")
		err := p.DeleteExchange("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system exchange")
	})

	t.Run("exchange not found (idempotent)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		p := New("amqp://localhost:5672/")
		// Extract host:port from test server URL for httpURI
		// makeHTTPRequest adds /api, so we need just the base URL
		serverURL := server.URL
		if strings.HasPrefix(serverURL, "http://") {
			p.httpURI = serverURL
		} else {
			p.httpURI = "http://" + serverURL
		}

		err := p.DeleteExchange("nonexistent")
		assert.NoError(t, err) // Should be idempotent
	})
}

func TestProvider_makeHTTPRequest(t *testing.T) {
	t.Run("authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "testuser", user)
			assert.Equal(t, "testpass", pass)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		p := New("amqp://testuser:testpass@localhost:5672/")
		p.httpURI = strings.TrimPrefix(server.URL, "http://")
		p.httpURI = "http://" + p.httpURI

		resp, err := p.makeHTTPRequest("GET", "/test")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("HTTP URI not configured", func(t *testing.T) {
		p := New("amqp://localhost:5672/")
		p.httpURI = ""
		_, err := p.makeHTTPRequest("GET", "/test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP URI not configured")
	})
}

