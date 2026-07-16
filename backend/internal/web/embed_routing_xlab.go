//go:build embed || unit

package web

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
)

const modelHubPath = "/models"

func shouldServeModelHub(request *http.Request) bool {
	if request == nil || request.URL == nil {
		return false
	}
	if strings.TrimSpace(request.URL.Path) != modelHubPath {
		return false
	}
	if request.Method != http.MethodGet {
		return false
	}
	for _, name := range []string{"Authorization", "X-API-Key", "X-Goog-API-Key"} {
		if _, present := headerValues(request.Header, name); present {
			return false
		}
	}
	query := request.URL.Query()
	for _, name := range []string{"client_version", "key", "api_key"} {
		if query.Has(name) {
			return false
		}
	}
	if destinations, present := headerValues(request.Header, "Sec-Fetch-Dest"); present {
		if len(destinations) != 1 || !strings.EqualFold(strings.TrimSpace(destinations[0]), "document") {
			return false
		}
	}
	acceptValues, _ := headerValues(request.Header, "Accept")
	return acceptsHTML(acceptValues)
}

func headerValues(header http.Header, name string) ([]string, bool) {
	for key, values := range header {
		if strings.EqualFold(key, name) {
			return values, true
		}
	}
	return nil, false
}

func acceptsHTML(values []string) bool {
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(part))
			if err != nil || !strings.EqualFold(mediaType, "text/html") {
				continue
			}
			if rawQuality, ok := params["q"]; ok {
				quality, err := strconv.ParseFloat(rawQuality, 64)
				if err != nil || quality <= 0 {
					continue
				}
			}
			return true
		}
	}
	return false
}

func addModelHubVaryHeader(header http.Header, path string) {
	if strings.TrimSpace(path) == modelHubPath {
		header.Add("Vary", "Accept, Sec-Fetch-Dest, Authorization, X-API-Key, X-Goog-API-Key")
	}
}
