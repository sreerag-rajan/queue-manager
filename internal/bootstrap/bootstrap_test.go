package bootstrap

import (
	"os"
	"testing"

	"queue-manager/internal/config"
)

func TestNewProvider_Rabbit(t *testing.T) {
	t.Setenv("APP_HOST", "0.0.0.0")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("QUEUE_PROVIDER", "RABBITMQ")
	t.Setenv("RABBITMQ_AMQP_URI", "amqp://guest:guest@localhost:5672/")

	cfg, err := config.LoadFromEnv(os.LookupEnv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prov, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov == nil {
		t.Fatalf("expected provider instance, got nil")
	}
}

func TestNewProvider_Unsupported(t *testing.T) {
	cfg := config.Config{
		AppHost:       "0.0.0.0",
		AppPort:       "8080",
		QueueProvider: "SOMETHING",
	}
	if _, err := NewProvider(cfg); err == nil {
		t.Fatalf("expected error for unsupported provider")
	}
}

