package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr     string
	CoreAPIBaseURL string
	CoreTimeout    time.Duration
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("XLAB_SERVER_ADDR"))
	if addr == "" {
		addr = ":8090"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("CORE_API_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080/api/v1"
	}

	timeoutSeconds := 10
	if raw := strings.TrimSpace(os.Getenv("XLAB_CORE_TIMEOUT_SECONDS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	return Config{
		ServerAddr:     addr,
		CoreAPIBaseURL: baseURL,
		CoreTimeout:    time.Duration(timeoutSeconds) * time.Second,
	}
}
