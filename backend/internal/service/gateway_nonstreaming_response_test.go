package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_HandleNonStreamingResponse_InvalidJSONTriggersFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

	svc := &GatewayService{}
	account := &Account{
		ID:       301,
		Name:     "anthropic-invalid-json",
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"rid-invalid-json"},
		},
		Body: io.NopCloser(strings.NewReader("\x1b[31mnot json")),
	}

	usage, err := svc.handleNonStreamingResponse(context.Background(), resp, c, account, "claude-opus-4-6", "claude-opus-4-6")

	require.Nil(t, usage)

	var failoverErr *UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
	require.Equal(t, "rid-invalid-json", failoverErr.ResponseHeaders.Get("x-request-id"))
	require.Contains(t, ExtractUpstreamErrorMessage(failoverErr.ResponseBody), "invalid JSON")
	require.Zero(t, rec.Body.Len(), "service should return failover so handler can retry another account")
}
