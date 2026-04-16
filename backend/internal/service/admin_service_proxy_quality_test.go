package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinalizeProxyQualityResult_ScoreAndGrade(t *testing.T) {
	result := &ProxyQualityCheckResult{
		PassedCount:    2,
		WarnCount:      1,
		FailedCount:    1,
		ChallengeCount: 1,
	}

	finalizeProxyQualityResult(result)

	require.Equal(t, 38, result.Score)
	require.Equal(t, "F", result.Grade)
	require.Contains(t, result.Summary, "通过 2 项")
	require.Contains(t, result.Summary, "告警 1 项")
	require.Contains(t, result.Summary, "失败 1 项")
	require.Contains(t, result.Summary, "挑战 1 项")
}

func TestRunProxyQualityTarget_SoraChallenge(t *testing.T) {
	target := proxyQualityTarget{
		Target: "sora",
		URL:    "http://proxy-quality.test/sora",
		Method: http.MethodGet,
		AllowedStatuses: map[int]struct{}{
			http.StatusUnauthorized: {},
		},
	}

	client := newProxyQualityTestClient(http.StatusForbidden, map[string]string{
		"Content-Type": "text/html",
		"cf-ray":       "test-ray-123",
	}, "<!DOCTYPE html><title>Just a moment...</title><script>window._cf_chl_opt={};</script>")

	item := runProxyQualityTarget(context.Background(), client, target)
	require.Equal(t, "challenge", item.Status)
	require.Equal(t, http.StatusForbidden, item.HTTPStatus)
	require.Equal(t, "test-ray-123", item.CFRay)
}

func TestRunProxyQualityTarget_AllowedStatusPass(t *testing.T) {
	target := proxyQualityTarget{
		Target: "gemini",
		URL:    "http://proxy-quality.test/gemini",
		Method: http.MethodGet,
		AllowedStatuses: map[int]struct{}{
			http.StatusOK: {},
		},
	}

	client := newProxyQualityTestClient(http.StatusOK, nil, `{"models":[]}`)

	item := runProxyQualityTarget(context.Background(), client, target)
	require.Equal(t, "pass", item.Status)
	require.Equal(t, http.StatusOK, item.HTTPStatus)
}

func TestRunProxyQualityTarget_AllowedStatusWarnForUnauthorized(t *testing.T) {
	target := proxyQualityTarget{
		Target: "openai",
		URL:    "http://proxy-quality.test/openai",
		Method: http.MethodGet,
		AllowedStatuses: map[int]struct{}{
			http.StatusUnauthorized: {},
		},
	}

	client := newProxyQualityTestClient(http.StatusUnauthorized, nil, `{"error":"unauthorized"}`)

	item := runProxyQualityTarget(context.Background(), client, target)
	require.Equal(t, "warn", item.Status)
	require.Equal(t, http.StatusUnauthorized, item.HTTPStatus)
	require.Contains(t, item.Message, "目标可达")
}

type proxyQualityRoundTripFunc func(*http.Request) (*http.Response, error)

func (f proxyQualityRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newProxyQualityTestClient(statusCode int, headers map[string]string, body string) *http.Client {
	return &http.Client{
		Transport: proxyQualityRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			header := make(http.Header, len(headers))
			for k, v := range headers {
				header.Set(k, v)
			}
			return &http.Response{
				StatusCode: statusCode,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}
}
