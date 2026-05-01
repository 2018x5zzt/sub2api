package enterprisebff

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type upstreamResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (s *Server) proxy(c *gin.Context, upstreamPath string, transformer responseTransformer) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Failed to read request body",
		})
		return
	}

	resp, err := s.doUpstreamRequest(c.Request.Context(), c.Request.Method, upstreamPath, c.Request.URL.RawQuery, c.Request.Header, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    http.StatusBadGateway,
			"message": "Failed to reach upstream core",
		})
		return
	}

	if transformer != nil && isJSONResponse(resp.Header) && len(resp.Body) > 0 {
		transformed, transformErr := transformer(resp.Body)
		if transformErr == nil {
			resp.Body = transformed
		}
	}

	writeUpstreamResponse(c, resp)
}

func (s *Server) doUpstreamRequest(
	ctx context.Context,
	method string,
	upstreamPath string,
	rawQuery string,
	headers http.Header,
	body []byte,
) (*upstreamResponse, error) {
	targetURL, err := s.resolveCoreURL(upstreamPath, rawQuery)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	copyRequestHeaders(req.Header, headers)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &upstreamResponse{
		StatusCode: resp.StatusCode,
		Header:     cloneHeader(resp.Header),
		Body:       respBody,
	}, nil
}

func (s *Server) resolveCoreURL(path, rawQuery string) (string, error) {
	base := *s.cfg.CoreBaseURL
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = rawQuery
	return base.String(), nil
}

func cloneHeader(header http.Header) http.Header {
	out := make(http.Header, len(header))
	for key, values := range header {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func copyRequestHeaders(dst, src http.Header) {
	for key, values := range src {
		if strings.EqualFold(key, "Host") {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func writeUpstreamResponse(c *gin.Context, resp *upstreamResponse) {
	for key, values := range resp.Header {
		if shouldSkipResponseHeader(key) {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Status(resp.StatusCode)
	if len(resp.Body) > 0 {
		_, _ = c.Writer.Write(resp.Body)
	}
}

func shouldSkipResponseHeader(key string) bool {
	switch strings.ToLower(key) {
	case "content-length", "transfer-encoding", "connection":
		return true
	default:
		return false
	}
}

func isJSONResponse(header http.Header) bool {
	contentType := strings.ToLower(strings.TrimSpace(header.Get("Content-Type")))
	return strings.Contains(contentType, "application/json")
}

func buildPathf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func addQueryParam(rawURL, key, value string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
