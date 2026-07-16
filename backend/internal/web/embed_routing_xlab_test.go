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
		name         string
		method       string
		target       string
		headers      http.Header
		wantFrontend bool
	}{
		{name: "html_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html,application/xhtml+xml"}}, wantFrontend: true},
		{name: "weighted_html_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"application/json,text/html;q=0.5"}}, wantFrontend: true},
		{name: "html_explicitly_rejected", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html;q=0"}}},
		{name: "json_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"application/json"}}},
		{name: "wildcard_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"*/*"}}},
		{name: "missing_accept_api", method: http.MethodGet, target: "/models"},
		{name: "document_navigation", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {"document"}}, wantFrontend: true},
		{name: "script_fetch_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {"script"}}},
		{name: "empty_fetch_destination_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Sec-Fetch-Dest": {""}}},
		{name: "authorization_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "Authorization": {""}}},
		{name: "x_api_key_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "X-API-Key": {""}}},
		{name: "x_goog_api_key_header_api", method: http.MethodGet, target: "/models", headers: http.Header{"Accept": {"text/html"}, "X-Goog-API-Key": {""}}},
		{name: "client_version_api", method: http.MethodGet, target: "/models?client_version=1", headers: http.Header{"Accept": {"text/html"}}},
		{name: "key_query_api", method: http.MethodGet, target: "/models?key=", headers: http.Header{"Accept": {"text/html"}}},
		{name: "api_key_query_api", method: http.MethodGet, target: "/models?api_key=", headers: http.Header{"Accept": {"text/html"}}},
		{name: "post_api", method: http.MethodPost, target: "/models", headers: http.Header{"Accept": {"text/html"}}},
		{name: "versioned_models_api", method: http.MethodGet, target: "/v1/models", headers: http.Header{"Accept": {"text/html"}}},
		{name: "other_spa_route", method: http.MethodGet, target: "/dashboard", headers: http.Header{"Accept": {"text/html"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, nil)
			request.Header = test.headers.Clone()
			require.Equal(t, test.wantFrontend, shouldServeModelHub(request))
		})
	}
}
