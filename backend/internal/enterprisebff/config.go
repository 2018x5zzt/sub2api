package enterprisebff

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	coreconfig "github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	defaultHost           = "0.0.0.0"
	defaultPort           = 8090
	defaultCoreBaseURL    = "http://127.0.0.1:8080/api/v1"
	defaultTimeoutSeconds = 30
)

type Config struct {
	ListenAddr                            string
	CoreBaseURL                           *url.URL
	RequestTimeout                        time.Duration
	TrustedProxies                        []string
	EnterpriseAttributeKey                string
	EnterpriseDisplayNameAttributeKey     string
	EnterpriseSupportContactAttribute     string
	EnterpriseVisibleGroupIDsByEnterprise map[string][]int64
}

func LoadConfig() (*Config, *coreconfig.Config, error) {
	host := getenvTrimmed("ENTERPRISE_BFF_HOST", defaultHost)
	port := getenvInt("ENTERPRISE_BFF_PORT", defaultPort)
	rawCoreBaseURL := getenvTrimmed("ENTERPRISE_BFF_CORE_BASE_URL", defaultCoreBaseURL)
	timeoutSeconds := getenvInt("ENTERPRISE_BFF_REQUEST_TIMEOUT_SECONDS", defaultTimeoutSeconds)
	visibleGroupIDsByEnterprise, err := parseEnterpriseVisibleGroupIDsByEnterprise(os.Getenv("ENTERPRISE_BFF_VISIBLE_GROUP_IDS_BY_ENTERPRISE"))
	if err != nil {
		return nil, nil, fmt.Errorf("parse ENTERPRISE_BFF_VISIBLE_GROUP_IDS_BY_ENTERPRISE: %w", err)
	}

	coreBaseURL, err := url.Parse(rawCoreBaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ENTERPRISE_BFF_CORE_BASE_URL: %w", err)
	}
	if !coreBaseURL.IsAbs() || strings.TrimSpace(coreBaseURL.Host) == "" {
		return nil, nil, fmt.Errorf("ENTERPRISE_BFF_CORE_BASE_URL must be an absolute URL")
	}

	sharedCfg, err := coreconfig.LoadForBootstrap()
	if err != nil {
		return nil, nil, fmt.Errorf("load core config: %w", err)
	}

	cfg := &Config{
		ListenAddr:                            fmt.Sprintf("%s:%d", host, port),
		CoreBaseURL:                           coreBaseURL,
		RequestTimeout:                        time.Duration(timeoutSeconds) * time.Second,
		TrustedProxies:                        splitCSV(os.Getenv("ENTERPRISE_BFF_TRUSTED_PROXIES")),
		EnterpriseAttributeKey:                getenvTrimmed("ENTERPRISE_BFF_ENTERPRISE_ATTRIBUTE_KEY", "enterprise_name"),
		EnterpriseDisplayNameAttributeKey:     getenvTrimmed("ENTERPRISE_BFF_ENTERPRISE_DISPLAY_NAME_ATTRIBUTE_KEY", "enterprise_display_name"),
		EnterpriseSupportContactAttribute:     getenvTrimmed("ENTERPRISE_BFF_ENTERPRISE_SUPPORT_CONTACT_ATTRIBUTE_KEY", "enterprise_support_contact"),
		EnterpriseVisibleGroupIDsByEnterprise: visibleGroupIDsByEnterprise,
	}
	return cfg, sharedCfg, nil
}

func getenvTrimmed(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseEnterpriseVisibleGroupIDsByEnterprise(raw string) (map[string][]int64, error) {
	return service.ParseEnterpriseVisibleGroupIDsByEnterprise(raw)
}
