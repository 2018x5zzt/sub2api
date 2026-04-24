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

func TestGatewayService_ForwardAsResponses_BufferedAnthropicStreamPreservesToolCallAndUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5","input":"Call the weather function","stream":false}`)
	upstreamSSE := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_resp_1","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[],"usage":{"input_tokens":6,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"get_weather","input":{"city":"Berlin"}}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"input_tokens":6,"output_tokens":2}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-gw-responses"}},
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

	result, err := svc.ForwardAsResponses(context.Background(), c, account, body, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Equal(t, "gpt-5", gjson.GetBytes(rec.Body.Bytes(), "model").String())
	require.Equal(t, "completed", gjson.GetBytes(rec.Body.Bytes(), "status").String())
	require.Equal(t, "function_call", gjson.GetBytes(rec.Body.Bytes(), "output.0.type").String())
	require.Equal(t, "get_weather", gjson.GetBytes(rec.Body.Bytes(), "output.0.name").String())
	require.Contains(t, gjson.GetBytes(rec.Body.Bytes(), "output.0.arguments").String(), "Berlin")
	require.True(t, strings.HasPrefix(gjson.GetBytes(rec.Body.Bytes(), "output.0.call_id").String(), "fc_"))
	require.Equal(t, int64(6), gjson.GetBytes(rec.Body.Bytes(), "usage.input_tokens").Int())
	require.Equal(t, int64(2), gjson.GetBytes(rec.Body.Bytes(), "usage.output_tokens").Int())
	require.Equal(t, 6, result.Usage.InputTokens)
	require.Equal(t, 2, result.Usage.OutputTokens)
}

func TestGatewayService_ForwardAsResponses_StreamingEmptyAnthropicStreamReturnsFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))

	body := []byte(`{"model":"gpt-5","input":"Hi","stream":true}`)
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
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}, "x-request-id": []string{"rid-gw-responses-empty"}},
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

	result, err := svc.ForwardAsResponses(context.Background(), c, account, body, nil)
	require.Nil(t, result)
	require.Error(t, err)

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.Contains(t, string(failoverErr.ResponseBody), "Upstream returned empty response")
	require.Empty(t, rec.Body.String(), "empty upstream stream must not commit a 200 SSE response")
	require.False(t, c.Writer.Written(), "service should return failover before writing response bytes")
}
