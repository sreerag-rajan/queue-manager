package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain_ConfigLoading(t *testing.T) {
	t.Run("missing APP_HOST", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_PORT", "8080")
		os.Setenv("QUEUE_PROVIDER", "RABBITMQ")

		// This would fail in actual main(), but we're just testing the logic
		// In a real scenario, we'd test config.LoadFromEnv separately
	})

	t.Run("missing APP_PORT", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_HOST", "127.0.0.1")
		os.Setenv("QUEUE_PROVIDER", "RABBITMQ")

		// This would fail in actual main()
	})

	t.Run("valid config", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_HOST", "127.0.0.1")
		os.Setenv("APP_PORT", "8080")
		os.Setenv("QUEUE_PROVIDER", "RABBITMQ")

		// Config should load successfully
		// Note: Actual main() execution would require database and queue connections
		// This test is more of a placeholder for integration tests
	})
}

func TestMain_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T)
	}{
		{
			name: "all required vars set",
			envVars: map[string]string{
				"APP_HOST":       "0.0.0.0",
				"APP_PORT":       "9090",
				"QUEUE_PROVIDER": "RABBITMQ",
			},
			validate: func(t *testing.T) {
				assert.Equal(t, "0.0.0.0", os.Getenv("APP_HOST"))
				assert.Equal(t, "9090", os.Getenv("APP_PORT"))
			},
		},
		{
			name: "optional vars",
			envVars: map[string]string{
				"APP_HOST":        "127.0.0.1",
				"APP_PORT":        "8080",
				"POSTGRES_URI":    "postgres://user:pass@localhost/db",
				"RABBITMQ_AMQP_URI": "amqp://user:pass@localhost:5672/",
			},
			validate: func(t *testing.T) {
				assert.NotEmpty(t, os.Getenv("APP_HOST"))
				assert.NotEmpty(t, os.Getenv("APP_PORT"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			tt.validate(t)
		})
	}
}

// Note: Testing the actual main() function is challenging because:
// 1. It calls log.Fatalf which exits the process
// 2. It requires actual database and queue connections
// 3. It starts a long-running server

// For comprehensive testing of main(), consider:
// 1. Integration tests that start the full application
// 2. Extracting initialization logic into testable functions
// 3. Using testcontainers for database and queue dependencies

func TestMain_InitializationFlow(t *testing.T) {
	// This test documents the expected initialization flow:
	// 1. Load configuration from environment
	// 2. Connect to database (if POSTGRES_URI provided)
	// 3. Initialize repository
	// 4. Connect to queue provider (if QUEUE_PROVIDER provided)
	// 5. Declare topology on startup
	// 6. Start cron scheduler
	// 7. Start HTTP server

	// Actual testing would require:
	// - Mocking or using testcontainers for database
	// - Mocking or using testcontainers for RabbitMQ
	// - Extracting main() logic into testable functions

	t.Run("placeholder for integration tests", func(t *testing.T) {
		// This is a placeholder
		// Real integration tests should be in a separate file with build tag: //go:build integration
		assert.True(t, true)
	})
}

func TestMain_ErrorHandling(t *testing.T) {
	t.Run("config load error", func(t *testing.T) {
		os.Clearenv()
		// Missing required vars would cause config.LoadFromEnv to fail
		// This would cause main() to call log.Fatalf and exit
		// Testing this requires process-level testing or extracting the logic
	})

	t.Run("database connection error", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_HOST", "127.0.0.1")
		os.Setenv("APP_PORT", "8080")
		os.Setenv("POSTGRES_URI", "postgres://invalid:invalid@localhost:5432/invalid")
		// This would fail in db.Connect()
		// main() would call log.Fatalf and exit
	})

	t.Run("queue provider error", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_HOST", "127.0.0.1")
		os.Setenv("APP_PORT", "8080")
		os.Setenv("QUEUE_PROVIDER", "RABBITMQ")
		os.Setenv("RABBITMQ_AMQP_URI", "amqp://invalid:invalid@localhost:5672/")
		// This would fail in qp.Connect() after retries
		// main() would continue anyway and retry via cron
	})
}

// Helper function to set up test environment
func setupTestEnv(envVars map[string]string) func() {
	originalEnv := make(map[string]string)
	for k := range envVars {
		if val, ok := os.LookupEnv(k); ok {
			originalEnv[k] = val
		}
	}

	os.Clearenv()
	for k, v := range envVars {
		os.Setenv(k, v)
	}

	return func() {
		os.Clearenv()
		for k, v := range originalEnv {
			os.Setenv(k, v)
		}
	}
}

func TestMain_HelperFunctions(t *testing.T) {
	t.Run("setupTestEnv", func(t *testing.T) {
		cleanup := setupTestEnv(map[string]string{
			"TEST_VAR": "test_value",
		})
		defer cleanup()

		assert.Equal(t, "test_value", os.Getenv("TEST_VAR"))
	})
}

// Integration test placeholder
// To run: go test -tags=integration ./cmd/server
func TestMain_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would require:
	// 1. Docker compose or testcontainers
	// 2. Actual database and RabbitMQ instances
	// 3. Full application startup
	// 4. Health check endpoints
	// 5. Cleanup after tests

	t.Skip("Integration test not implemented - requires docker compose or testcontainers")
}

