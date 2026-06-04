package handler

import (
	"errors"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIsOpenAIImageMissingOutputError(t *testing.T) {
	require.False(t, isOpenAIImageMissingOutputError(nil))
	require.True(t, isOpenAIImageMissingOutputError(errors.New("upstream did not return image output")))
	require.True(t, isOpenAIImageMissingOutputError(errors.New("prefix: Upstream Did Not Return Image Output")))
	require.False(t, isOpenAIImageMissingOutputError(errors.New("other error")))
}

func TestExtractOpenAIImageUpstreamRequestBodyForLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set(service.OpsUpstreamRequestBodyKey, []byte(`{"model":"gpt-image-2"}`))
	require.Equal(t, `{"model":"gpt-image-2"}`, extractOpenAIImageUpstreamRequestBodyForLog(c))

	c2, _ := gin.CreateTestContext(nil)
	c2.Set(service.OpsUpstreamRequestBodyKey, `{"model":"gpt-image-2","prompt":"cat"}`)
	require.Equal(t, `{"model":"gpt-image-2","prompt":"cat"}`, extractOpenAIImageUpstreamRequestBodyForLog(c2))
}

func TestExtractOpenAIImageUpstreamResponseBodyForLog_PrefersDetailAndEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(nil)
	c.Set(service.OpsUpstreamErrorDetailKey, "detail-body")
	require.Equal(t, "detail-body", extractOpenAIImageUpstreamResponseBodyForLog(c, nil))

	c2, _ := gin.CreateTestContext(nil)
	c2.Set(service.OpsUpstreamErrorsKey, []*service.OpsUpstreamErrorEvent{
		{Detail: "first"},
		{UpstreamResponseBody: "second"},
	})
	require.Equal(t, "second", extractOpenAIImageUpstreamResponseBodyForLog(c2, nil))
}

func TestExtractOpenAIImageUpstreamResponseBodyForLog_UsesFailoverErrorFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	err := &service.UpstreamFailoverError{
		StatusCode:   502,
		ResponseBody: []byte(`{"error":"x"}`),
	}
	require.Equal(t, `{"error":"x"}`, extractOpenAIImageUpstreamResponseBodyForLog(c, err))
}

func TestTruncateOpenAIImageForwardLogBody(t *testing.T) {
	body := strings.Repeat("a", openAIImageForwardBodyLogMaxBytes+128)
	trimmed := truncateOpenAIImageForwardLogBody(body)
	require.Len(t, trimmed, openAIImageForwardBodyLogMaxBytes)
}
