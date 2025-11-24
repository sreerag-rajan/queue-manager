package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv_Success(t *testing.T) {
	t.Setenv("APP_HOST", "127.0.0.1")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("QUEUE_PROVIDER", "RABBITMQ")

	got, err := LoadFromEnv(os.LookupEnv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AppHost != "127.0.0.1" || got.AppPort != "9090" {
		t.Fatalf("unexpected config values: %+v", got)
	}
	if got.Addr() != "127.0.0.1:9090" {
		t.Fatalf("unexpected addr: %s", got.Addr())
	}
}

func TestLoadFromEnv_MissingHost(t *testing.T) {
	t.Setenv("APP_HOST", "")
	t.Setenv("APP_PORT", "8080")
	_, err := LoadFromEnv(os.LookupEnv)
	if err == nil {
		t.Fatalf("expected error for missing APP_HOST")
	}
}

func TestLoadFromEnv_MissingPort(t *testing.T) {
	t.Setenv("APP_HOST", "0.0.0.0")
	t.Setenv("APP_PORT", "")
	_, err := LoadFromEnv(os.LookupEnv)
	if err == nil {
		t.Fatalf("expected error for missing APP_PORT")
	}
}


