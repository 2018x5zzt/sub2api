package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type rejectingStartImageJobPersistence struct {
	created chan struct{}
	mu      sync.Mutex
	jobs    map[string]*service.OpenAIImageJobRecord
}

func newRejectingStartImageJobPersistence() *rejectingStartImageJobPersistence {
	return &rejectingStartImageJobPersistence{
		created: make(chan struct{}),
		jobs:    make(map[string]*service.OpenAIImageJobRecord),
	}
}

func (p *rejectingStartImageJobPersistence) Create(_ context.Context, record *service.OpenAIImageJobRecord, _ time.Duration) error {
	p.mu.Lock()
	p.jobs[record.ID] = cloneOpenAIImageJobRecord(record)
	p.mu.Unlock()
	close(p.created)
	return nil
}

func (p *rejectingStartImageJobPersistence) Get(_ context.Context, id string, owner service.OpenAIImageJobOwner) (*service.OpenAIImageJobRecord, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	job, ok := p.jobs[id]
	if !ok || job.Owner.UserID != owner.UserID || job.Owner.APIKeyID != owner.APIKeyID {
		return nil, false, nil
	}
	return cloneOpenAIImageJobRecord(job), true, nil
}

func (p *rejectingStartImageJobPersistence) SetRunning(context.Context, string, time.Time, time.Duration) (bool, error) {
	return false, nil
}

func (p *rejectingStartImageJobPersistence) Complete(context.Context, string, service.OpenAIImageJobStatus, int, map[string][]string, []byte, time.Time, time.Duration) error {
	return nil
}

func (p *rejectingStartImageJobPersistence) MarkStaleTimeouts(context.Context, time.Time, time.Duration, int, time.Duration) (int64, error) {
	return 0, nil
}

func TestOpenAIImageJobStoreSuccessIsVisibleOnlyToOwner(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
		MaxAttempts: 1,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	started := make(chan struct{})

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		close(started)
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"data":[{"url":"https://example.test/a.png"}]}`),
		}
	})
	require.NoError(t, err)

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

func TestOpenAIImageJobStoreSkipsRunnerWhenStartTransitionFails(t *testing.T) {
	persistence := newRejectingStartImageJobPersistence()
	store := newOpenAIImageJobStoreWithPersistence(persistence, openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
		MaxAttempts: 1,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	var ran atomic.Bool

	_, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		ran.Store(true)
		return openAIImageJobResult{StatusCode: http.StatusOK, Body: []byte(`{"data":[]}`)}
	})
	require.NoError(t, err)

	select {
	case <-persistence.created:
	case <-time.After(time.Second):
		t.Fatal("job was not created")
	}
	require.Never(t, ran.Load, 100*time.Millisecond, 10*time.Millisecond)
}

func TestOpenAIImageJobStoreFailureIsNotSuccessful(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
		MaxAttempts: 1,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{
			StatusCode: http.StatusBadGateway,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"error":{"type":"api_error","message":"upstream timeout"}}`),
		}
	})
	require.NoError(t, err)

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

func TestOpenAIImageJobStoreRetriesRetryableFailureUntilSuccess(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency:  1,
		Timeout:      time.Second,
		TTL:          time.Hour,
		MaxAttempts:  3,
		RetryBackoff: []time.Duration{0, 0},
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	var attempts atomic.Int32

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		attempt := attempts.Add(1)
		if attempt == 1 {
			return openAIImageJobResult{
				StatusCode: http.StatusBadGateway,
				Headers:    http.Header{"Content-Type": []string{"application/json"}},
				Body:       []byte(`{"error":{"type":"upstream_error","message":"temporary"}}`),
			}
		}
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"data":[{"url":"https://example.test/retry.png"}]}`),
		}
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusSucceeded
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusSucceeded, got.Status)
	require.Equal(t, http.StatusOK, got.StatusCode)
	require.Equal(t, int32(2), attempts.Load())
	require.JSONEq(t, `{"data":[{"url":"https://example.test/retry.png"}]}`, string(got.Body))
}

func TestOpenAIImageJobStoreFailsAfterRetryableAttemptsExhausted(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency:  1,
		Timeout:      time.Second,
		TTL:          time.Hour,
		MaxAttempts:  3,
		RetryBackoff: []time.Duration{0, 0},
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	var attempts atomic.Int32

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		attempts.Add(1)
		return openAIImageJobResult{
			StatusCode: http.StatusBadGateway,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"error":{"type":"upstream_error","message":"temporary"}}`),
		}
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusFailed
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusFailed, got.Status)
	require.Equal(t, http.StatusBadGateway, got.StatusCode)
	require.Equal(t, int32(3), attempts.Load())
	require.JSONEq(t, `{"error":{"type":"upstream_error","message":"Image job failed after 3 attempts","last_status":502,"attempts":3}}`, string(got.Body))
}

func TestOpenAIImageJobStoreDoesNotRetryPermanentClientError(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency:  1,
		Timeout:      time.Second,
		TTL:          time.Hour,
		MaxAttempts:  3,
		RetryBackoff: []time.Duration{0, 0},
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	var attempts atomic.Int32

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		attempts.Add(1)
		return openAIImageJobResult{
			StatusCode: http.StatusBadRequest,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"error":{"type":"invalid_request_error","message":"bad prompt"}}`),
		}
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusFailed
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, http.StatusBadRequest, got.StatusCode)
	require.Equal(t, int32(1), attempts.Load())
	require.JSONEq(t, `{"error":{"type":"invalid_request_error","message":"bad prompt"}}`, string(got.Body))
}

func TestOpenAIImageJobStoreStopsRetryingWhenContextTimesOut(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency:  1,
		Timeout:      20 * time.Millisecond,
		TTL:          time.Hour,
		MaxAttempts:  3,
		RetryBackoff: []time.Duration{time.Second, time.Second},
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	var attempts atomic.Int32

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		attempts.Add(1)
		return openAIImageJobResult{
			StatusCode: http.StatusBadGateway,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"error":{"type":"upstream_error","message":"temporary"}}`),
		}
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusTimeout
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusTimeout, got.Status)
	require.Equal(t, http.StatusGatewayTimeout, got.StatusCode)
	require.Equal(t, int32(1), attempts.Load())
	require.JSONEq(t, `{"error":{"type":"api_error","message":"Image job timed out"}}`, string(got.Body))
}

func TestOpenAIImageJobStoreMarksTimeoutWhenRunnerDoesNotReturn(t *testing.T) {
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     20 * time.Millisecond,
		TTL:         time.Hour,
		MaxAttempts: 1,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	started := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		close(started)
		<-release
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"data":[{"url":"https://example.test/late.png"}]}`),
		}
	})
	require.NoError(t, err)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job runner did not start")
	}

	require.Eventually(t, func() bool {
		got, ok := store.get(job.ID, owner)
		return ok && got.Status == openAIImageJobStatusTimeout
	}, time.Second, 10*time.Millisecond)

	got, ok := store.get(job.ID, owner)
	require.True(t, ok)
	require.Equal(t, openAIImageJobStatusTimeout, got.Status)
	require.Equal(t, http.StatusGatewayTimeout, got.StatusCode)
	require.JSONEq(t, `{"error":{"type":"api_error","message":"Image job timed out"}}`, string(got.Body))
	require.False(t, got.CompletedAt.IsZero())
	require.True(t, got.UpdatedAt.After(got.CreatedAt))
}

func TestOpenAIGatewayHandlerImageJobStatusReturnsSucceededResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newOpenAIImageJobStore(openAIImageJobStoreOptions{
		Concurrency: 1,
		Timeout:     time.Second,
		TTL:         time.Hour,
	})
	owner := openAIImageJobOwner{UserID: 42, APIKeyID: 7}
	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Body:       []byte(`{"created":123,"data":[{"url":"https://example.test/a.png"}]}`),
		}
	})
	require.NoError(t, err)
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
	require.Contains(t, rec.Header().Get("Cache-Control"), "no-store")
	require.Equal(t, "no-cache", rec.Header().Get("Pragma"))
	require.Equal(t, "0", rec.Header().Get("Expires"))
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
	job, err := store.submit(owner, openAIImageJobRequest{Endpoint: EndpointImagesGenerations}, func(ctx context.Context, req openAIImageJobRequest) openAIImageJobResult {
		return openAIImageJobResult{StatusCode: http.StatusOK, Body: []byte(`{"data":[]}`)}
	})
	require.NoError(t, err)
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
