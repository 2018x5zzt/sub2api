//go:build embed || unit

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedFrontendModelsRouteDispatch(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		headers    http.Header
		wantBypass bool
	}{
		{name: "html_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html,application/xhtml+xml"}}, wantBypass: false},
		{name: "weighted_html_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"application/json,text/html;q=0.5"}}, wantBypass: false},
		{name: "html_explicitly_rejected", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html;q=0"}}, wantBypass: true},
		{name: "json_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"application/json"}}, wantBypass: true},
		{name: "wildcard_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"*/*"}}, wantBypass: true},
		{name: "missing_accept_api", method: http.MethodGet, target: "/models", wantBypass: true},
		{name: "document_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {"document"}}, wantBypass: false},
		{name: "script_fetch_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {"script"}}, wantBypass: true},
		{name: "empty_fetch_destination_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {""}}, wantBypass: true},
		{name: "authorization_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Authorization": {""}}, wantBypass: true},
		{name: "x_api_key_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "X-API-Key": {""}}, wantBypass: true},
		{name: "x_goog_api_key_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "X-Goog-API-Key": {""}}, wantBypass: true},
		{name: "client_version_api", method: http.MethodGet, target: "/models?client_version=1", headers: http.Header{"Accept": {"text/html"}}, wantBypass: true},
		{name: "key_query_api", method: http.MethodGet, target: "/models?key=", headers: http.Header{"Accept": {"text/html"}}, wantBypass: true},
		{name: "api_key_query_api", method: http.MethodGet, target: "/models?api_key=", headers: http.Header{"Accept": {"text/html"}}, wantBypass: true},
		{name: "post_api", method: http.MethodPost, target: "/models", headers: http.Header{"Accept": {"text/html"}}, wantBypass: true},
		{name: "versioned_models_api", method: http.MethodGet, target: "/v1/models", headers: http.Header{"Accept": {"text/html"}}, wantBypass: true},
		{name: "other_spa_route", method: http.MethodGet, target: "/dashboard", headers: http.Header{"Accept": {"text/html"}}, wantBypass: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			request.Header = test.headers.Clone()
			require.Equal(t, test.wantBypass, shouldBypassEmbeddedFrontendRequest(request))
		})
	}
}
