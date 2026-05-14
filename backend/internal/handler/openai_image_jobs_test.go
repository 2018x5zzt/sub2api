package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIImageJobStoreSuccessIsVisibleOnlyToOwner(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	started := make(chan struct{})

	job := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		close(started)
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"data":[{"url":"https://example.test/a.png"}]}`),
		}
	})

	require.NotEmpty(t, job.ID)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job runner did not start")
	}

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusSucceeded
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusSucceeded, got.Status)
	require.Equal(t, http.StatusOK, got.StatusCode)
	require.JSONEq(t, `{"data":[{"url":"https://example.test/a.png"}]}`, string(got.Body))

	_, ok = store.get(job.ID, openAIImageJobOwner{UserID: 42, APIKeyID: 8})
	require.False(t, ok)
	_, ok = store.get(job.ID, openAIImageJobOwner{UserID: 99, APIKeyID: 7})
	require.False(t, ok)
}

func TestOpenAIImageJobStoreFailureIsNotSuccessful(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}

	job := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{
			StatusCode: http.StatusBadGateway,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"error":{"type":"api_error","message":"upstream timeout"}}`),
		}
	})

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusFailed
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusFailed, got.Status)
	require.Equal(t, http.StatusBadGateway, got.StatusCode)
	require.JSONEq(t, `{"error":{"type":"api_error","message":"upstream timeout"}}`, string(got.Body))
}

func TestOpenAIGatewayHandlerImageJobStatusReturnsSucceededResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	job := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"created":123,"data":[{"url":"https://example.test/a.png"}]}`),
		}
	})
	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusSucceeded
	}, time.Second, 10*time.Millisecond)

	h := &OpenAIGatewayHandler{imageJobStore: store}
	router := gin.New()
	router.GET("/v1/images/jobs/:id", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{ID: owner.APIKeyID})
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: owner.UserID})
		h.ImageJobStatus(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/images/jobs/"+job.ID, nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Equal(t, job.ID, payload["job_id"])
	require.Equal(t, string(openAIImageJobStatusSucceeded), payload["status"])
	response, ok := payload["response"].(map[string]any)
	require.True(t, ok)
	data, ok := response["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 1)
	require.Equal(t, "https://example.test/a.png", data[0].(map[string]any)["url"])
}

func TestOpenAIGatewayHandlerImageJobStatusHidesOtherOwnersJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	job := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{StatusCode: http.StatusOK, Body: []byte(`{"data":[]}`)}
	})
	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusSucceeded
	}, time.Second, 10*time.Millisecond)

	h := &OpenAIGatewayHandler{imageJobStore: store}
	router := gin.New()
	router.GET("/v1/images/jobs/:id", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{ID: 99})
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: owner.UserID})
		h.ImageJobStatus(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/images/jobs/"+job.ID, nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
