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
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGatewayService_ForwardAsChatCompletions_BufferedAnthropicStreamPreservesReasoningAndUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"Hi"}],"stream":false}`)
	upstreamSSE := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_cc_1","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[],"usage":{"input_tokens":9,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"plan first"}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":"final answer"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":9,"output_tokens":4}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-gw-cc"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}
	upstream := &httpUpstreamRecorder{resp: resp}

	cfg := &config.Config{Gateway: config.GatewayConfig{}}
	svc := &GatewayService{
		cfg:                  cfg,
		httpUpstream:         upstream,
		responseHeaderFilter: compileResponseHeaderFilter(cfg),
	}

	account := &Account{
		ID:             123,
		Name:           "acc",
		Platform:       PlatformAnthropic,
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Credentials:    map[string]any{"api_key": "test-api-key"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Equal(t, "plan first", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.reasoning_content").String())
	require.Equal(t, "final answer", gjson.GetBytes(rec.Body.Bytes(), "choices.0.message.content").String())
	require.Equal(t, int64(9), gjson.GetBytes(rec.Body.Bytes(), "usage.prompt_tokens").Int())
	require.Equal(t, int64(4), gjson.GetBytes(rec.Body.Bytes(), "usage.completion_tokens").Int())
	require.Equal(t, 9, result.Usage.InputTokens)
	require.Equal(t, 4, result.Usage.OutputTokens)
}

func TestGatewayService_ForwardAsChatCompletions_StreamingEmptyAnthropicStreamReturnsFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"Hi"}],"stream":true}`)
	upstreamSSE := strings.Join([]string{
		`event: ping`,
		`data: not-json`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-gw-cc-empty"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}
	upstream := &httpUpstreamRecorder{resp: resp}

	svc := &GatewayService{
		cfg:          &config.Config{Gateway: config.GatewayConfig{}},
		httpUpstream: upstream,
	}

	account := &Account{
		ID:             123,
		Name:           "acc",
		Platform:       PlatformAnthropic,
		Type:           AccountTypeAPIKey,
		Concurrency:    1,
		Credentials:    map[string]any{"api_key": "test-api-key"},
		Status:         StatusActive,
		Schedulable:    true,
		RateMultiplier: f64p(1),
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, account, body, nil)
	require.Nil(t, result)
	require.Error(t, err)

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.Contains(t, string(failoverErr.ResponseBody), "Upstream returned empty response")
	require.Empty(t, rec.Body.String(), "empty upstream stream must not commit a successful chat completions stream")
	require.False(t, c.Writer.Written(), "service should return failover before writing response bytes")
}
