//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_Forward_ModelUnsupported400ReturnsFailoverError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	body := []byte(`{"model":"claude-opus-4-7","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`)
	parsed := &ParsedRequest{
		Body:   body,
		Model:  "claude-opus-4-7",
		Stream: false,
	}

	upstream := &anthropicHTTPUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"X-Request-Id": []string{"rid-unsupported-model"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"暂不支持","type":"invalid_request_error"},"type":"error"}`)),
		},
	}

	svc := &GatewayService{
		cfg:              &config.Config{},
		httpUpstream:     upstream,
		rateLimitService: &RateLimitService{},
	}

	account := &Account{
		ID:          501,
		Name:        "mixed-relay",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "upstream-key",
			"base_url": "https://api.anthropic.com",
			"model_mapping": map[string]any{
				"claude-opus-4-7": "claude-opus-4-7",
			},
		},
		Status:      StatusActive,
		Schedulable: true,
	}

	result, err := svc.Forward(context.Background(), c, account, parsed)
	require.Nil(t, result)

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadRequest, failoverErr.StatusCode)
	require.Contains(t, string(failoverErr.ResponseBody), "暂不支持")
	require.Zero(t, rec.Body.Len(), "service should return failover so handler can retry another account")
}
