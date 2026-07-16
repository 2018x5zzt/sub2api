package handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestShouldStopOpenAIOAuthFailoverRunsBeforeSilent429Retry(t *testing.T) {
	gateway := &service.OpenAIGatewayService{}
	account := &service.Account{
		ID:       44,
		Platform: service.PlatformGrok,
		Type:     service.AccountTypeOAuth,
	}
	var rateLimitState openAI429SilentFailoverState
	var oauthState service.OpenAIOAuth429FailoverState
	first429 := &service.UpstreamFailoverError{StatusCode: http.StatusTooManyRequests}

	require.False(t, shouldStopOpenAIOAuthFailover(gateway, account, first429, 0, &rateLimitState, &oauthState))
	require.True(t, rateLimitState.noteSwitch(first429, time.Now()))

	followupFailure := &service.UpstreamFailoverError{StatusCode: http.StatusBadGateway}
	require.True(t, shouldStopOpenAIOAuthFailover(gateway, account, followupFailure, 0, &rateLimitState, &oauthState))
}
