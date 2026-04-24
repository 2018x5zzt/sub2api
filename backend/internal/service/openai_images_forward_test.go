package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayService_ForwardImages_APIKeyHTMLSuccessResponseTriggersFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"model":"gpt-image-2","prompt":"draw a cat"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "curl/8.0")

	svc := &OpenAIGatewayService{
		cfg: &config.Config{
			Gateway: config.GatewayConfig{ForceCodexCLI: false},
		},
		httpUpstream: &httpUpstreamRecorder{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/html; charset=utf-8"},
					"x-request-id": []string{"rid-image-html"},
				},
				Body: io.NopCloser(strings.NewReader(`<!DOCTYPE html><html><head><title>Bad gateway</title></head><body>temporarily unavailable</body></html>`)),
			},
		},
	}

	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)

	account := &Account{
		ID:             456,
		Name:           "apikey-acc",
		Platform:       PlatformOpenAI,
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Credentials:    map[string]any{"api_key": "sk-api-key"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
	require.Nil(t, result)
	require.Error(t, err)
	require.False(t, c.Writer.Written(), "service should not stream HTML upstream bodies to the client")

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.JSONEq(t, `{"error":{"type":"upstream_error","message":"Upstream returned an HTML response","code":"html_response"}}`, string(failoverErr.ResponseBody))
}

func TestNewOpenAIImageStatusError_NormalizesHTMLBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Just a moment...</title></head><body></body></html>`))
	}))
	defer server.Close()

	resp, err := req.C().R().DisableAutoReadResponse().Get(server.URL)
	require.NoError(t, err)

	wrapped := newOpenAIImageStatusError(resp, "chat-requirements failed")

	var statusErr *openAIImageStatusError
	require.ErrorAs(t, wrapped, &statusErr)
	require.Equal(t, http.StatusForbidden, statusErr.StatusCode)
	require.Equal(t, "Upstream returned an HTML error page", statusErr.Message)
	require.JSONEq(t, `{"error":{"type":"upstream_error","message":"Upstream returned an HTML error page","code":"html_error_response"}}`, string(statusErr.ResponseBody))
}
