package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", "")
	t.Setenv("CORE_API_BASE_URL", "")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "")

	cfg := Load()
	if cfg.ServerAddr != ":8090" {
		t.Fatalf("ServerAddr = %q, want :8090", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "http://127.0.0.1:8080/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 10*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("XLAB_SERVER_ADDR", ":19090")
	t.Setenv("CORE_API_BASE_URL", "https://core.example.com/api/v1/")
	t.Setenv("XLAB_CORE_TIMEOUT_SECONDS", "3")

	cfg := Load()
	if cfg.ServerAddr != ":19090" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr)
	}
	if cfg.CoreAPIBaseURL != "https://core.example.com/api/v1" {
		t.Fatalf("CoreAPIBaseURL = %q", cfg.CoreAPIBaseURL)
	}
	if cfg.CoreTimeout != 3*time.Second {
		t.Fatalf("CoreTimeout = %s", cfg.CoreTimeout)
	}
}
