package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldBypassEmbeddedFrontendAPICompatibilityPrefixes(t *testing.T) {
	for _, path := range []string{
		"/api/v1/users",
		"/api-proxy/v1/images/generations",
		"/api-proxy/v1/images/edits",
		"/v1/models",
		"/v1beta/chat",
		"/openai/v1/images/generations",
		"/openai/v1/images/edits",
		"/backend-api/codex/responses",
		"/antigravity/test",
		"/setup/init",
		"/health",
		"/responses",
		"/responses/compact",
		"/images/generations",
	} {
		require.True(t, shouldBypassEmbeddedFrontend(path), "path=%s should bypass embedded frontend", path)
	}
}

func TestShouldBypassEmbeddedFrontendAllowsSPARoutes(t *testing.T) {
	for _, path := range []string{
		"/",
		"/dashboard",
		"/users/123",
		"/settings/profile",
	} {
		require.False(t, shouldBypassEmbeddedFrontend(path), "path=%s should be handled as an SPA route", path)
	}
}
