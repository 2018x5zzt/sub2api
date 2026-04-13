//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type gatewayForwardingSettingRepoStub struct {
	values map[string]string
}

func (s *gatewayForwardingSettingRepoStub) Get(_ context.Context, key string) (*Setting, error) {
	if val, ok := s.values[key]; ok {
		return &Setting{Key: key, Value: val}, nil
	}
	return nil, ErrSettingNotFound
}

func (s *gatewayForwardingSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if val, ok := s.values[key]; ok {
		return val, nil
	}
	return "", ErrSettingNotFound
}

func (s *gatewayForwardingSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *gatewayForwardingSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if val, ok := s.values[key]; ok {
			result[key] = val
		}
	}
	return result, nil
}

func (s *gatewayForwardingSettingRepoStub) SetMultiple(_ context.Context, settings map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range settings {
		s.values[key] = value
	}
	return nil
}

func (s *gatewayForwardingSettingRepoStub) GetAll(_ context.Context) (map[string]string, error) {
	result := make(map[string]string, len(s.values))
	for key, value := range s.values {
		result[key] = value
	}
	return result, nil
}

func (s *gatewayForwardingSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

type gatewayForwardingIdentityCacheStub struct {
	fingerprint     *Fingerprint
	maskedSessionID string
}

func (s *gatewayForwardingIdentityCacheStub) GetFingerprint(_ context.Context, _ int64) (*Fingerprint, error) {
	if s.fingerprint == nil {
		return nil, nil
	}
	cp := *s.fingerprint
	return &cp, nil
}

func (s *gatewayForwardingIdentityCacheStub) SetFingerprint(_ context.Context, _ int64, fp *Fingerprint) error {
	if fp == nil {
		s.fingerprint = nil
		return nil
	}
	cp := *fp
	s.fingerprint = &cp
	return nil
}

func (s *gatewayForwardingIdentityCacheStub) GetMaskedSessionID(_ context.Context, _ int64) (string, error) {
	return s.maskedSessionID, nil
}

func (s *gatewayForwardingIdentityCacheStub) SetMaskedSessionID(_ context.Context, _ int64, sessionID string) error {
	s.maskedSessionID = sessionID
	return nil
}

func resetGatewayForwardingCacheForTest() {
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{expiresAt: 0})
}

func newGatewayForwardingTestService(settings map[string]string, fingerprint *Fingerprint) *GatewayService {
	resetGatewayForwardingCacheForTest()

	cfg := &config.Config{RunMode: config.RunModeStandard}
	return &GatewayService{
		cfg:            cfg,
		settingService: NewSettingService(&gatewayForwardingSettingRepoStub{values: settings}, cfg),
		identityService: NewIdentityService(&gatewayForwardingIdentityCacheStub{
			fingerprint: fingerprint,
		}),
	}
}

func newGatewayForwardingRequestContext(t *testing.T, body []byte, userAgent, stainlessLang string) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", userAgent)
	c.Request.Header.Set("X-Stainless-Lang", stainlessLang)
	c.Request.Header.Set("X-Stainless-Package-Version", "0.70.0")
	c.Request.Header.Set("X-Stainless-OS", "Linux")
	c.Request.Header.Set("X-Stainless-Arch", "arm64")
	c.Request.Header.Set("X-Stainless-Runtime", "node")
	c.Request.Header.Set("X-Stainless-Runtime-Version", "v24.13.0")
	return c
}

func newGatewayForwardingOAuthAccount() *Account {
	return &Account{
		ID:       42,
		Type:     AccountTypeOAuth,
		Platform: PlatformAnthropic,
		Extra: map[string]any{
			"account_uuid": "11111111-2222-4333-8444-555555555555",
		},
	}
}

func readRequestBody(t *testing.T, req *http.Request) []byte {
	t.Helper()

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	return body
}

func TestGatewayService_BuildUpstreamRequest_GatewayForwardingSettingsPreserveClientIdentity(t *testing.T) {
	originalUserID := FormatMetadataUserID(
		strings.Repeat("a", 64),
		"",
		"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
		"2.1.22",
	)
	body := []byte(`{"metadata":{"user_id":"` + originalUserID + `"}}`)
	clientUserAgent := "claude-cli/2.1.80 (external, cli)"
	clientStainlessLang := "js-client"
	fingerprintUserAgent := "claude-cli/2.1.99 (external, cli)"

	svc := newGatewayForwardingTestService(
		map[string]string{
			SettingKeyEnableFingerprintUnification: "false",
			SettingKeyEnableMetadataPassthrough:    "true",
		},
		&Fingerprint{
			ClientID:                strings.Repeat("b", 64),
			UserAgent:               fingerprintUserAgent,
			StainlessLang:           "go-fingerprint",
			StainlessPackageVersion: "9.9.9",
			StainlessOS:             "Darwin",
			StainlessArch:           "amd64",
			StainlessRuntime:        "go",
			StainlessRuntimeVersion: "go1.99.0",
			UpdatedAt:               time.Now().Unix(),
		},
	)

	c := newGatewayForwardingRequestContext(t, body, clientUserAgent, clientStainlessLang)
	req, err := svc.buildUpstreamRequest(context.Background(), c, newGatewayForwardingOAuthAccount(), body, "token", "oauth", "claude-3-5-sonnet-20241022", false, false)
	require.NoError(t, err)

	gotBody := readRequestBody(t, req)
	require.Equal(t, originalUserID, gjson.GetBytes(gotBody, "metadata.user_id").String())
	require.Equal(t, clientUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, clientStainlessLang, req.Header.Get("X-Stainless-Lang"))
}

func TestGatewayService_BuildCountTokensRequest_GatewayForwardingSettingsPreserveClientIdentity(t *testing.T) {
	originalUserID := FormatMetadataUserID(
		strings.Repeat("c", 64),
		"",
		"ffffffff-1111-4222-8333-444444444444",
		"2.1.22",
	)
	body := []byte(`{"metadata":{"user_id":"` + originalUserID + `"}}`)
	clientUserAgent := "claude-cli/2.1.80 (external, cli)"
	clientStainlessLang := "ts-client"
	fingerprintUserAgent := "claude-cli/2.1.99 (external, cli)"

	svc := newGatewayForwardingTestService(
		map[string]string{
			SettingKeyEnableFingerprintUnification: "false",
			SettingKeyEnableMetadataPassthrough:    "true",
		},
		&Fingerprint{
			ClientID:                strings.Repeat("d", 64),
			UserAgent:               fingerprintUserAgent,
			StainlessLang:           "python-fingerprint",
			StainlessPackageVersion: "8.8.8",
			StainlessOS:             "Windows",
			StainlessArch:           "x86_64",
			StainlessRuntime:        "python",
			StainlessRuntimeVersion: "3.14.0",
			UpdatedAt:               time.Now().Unix(),
		},
	)

	c := newGatewayForwardingRequestContext(t, body, clientUserAgent, clientStainlessLang)
	req, err := svc.buildCountTokensRequest(context.Background(), c, newGatewayForwardingOAuthAccount(), body, "token", "oauth", "claude-3-5-sonnet-20241022", false)
	require.NoError(t, err)

	gotBody := readRequestBody(t, req)
	require.Equal(t, originalUserID, gjson.GetBytes(gotBody, "metadata.user_id").String())
	require.Equal(t, clientUserAgent, req.Header.Get("User-Agent"))
	require.Equal(t, clientStainlessLang, req.Header.Get("X-Stainless-Lang"))
}
