package repository

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newOpenAIImageJobStoreTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb, mr
}

func TestOpenAIImageJobRedisStorePersistsAcrossInstances(t *testing.T) {
	rdb, _ := newOpenAIImageJobStoreTestRedis(t)
	ctx := context.Background()
	createdAt := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	owner := service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 7}

	storeA := NewOpenAIImageJobStore(rdb)
	storeB := NewOpenAIImageJobStore(rdb)
	require.NoError(t, storeA.Create(ctx, &service.OpenAIImageJobRecord{
		ID:        "imgjob_cross_instance",
		Owner:     owner,
		Status:    service.OpenAIImageJobStatusQueued,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}, time.Hour))

	got, ok, err := storeB.Get(ctx, "imgjob_cross_instance", owner)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, service.OpenAIImageJobStatusQueued, got.Status)
	require.Equal(t, createdAt, got.CreatedAt)

	_, ok, err = storeB.Get(ctx, "imgjob_cross_instance", service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 8})
	require.NoError(t, err)
	require.False(t, ok)
}

func TestOpenAIImageJobRedisStoreCompleteSurvivesNewInstance(t *testing.T) {
	rdb, _ := newOpenAIImageJobStoreTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	owner := service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 7}

	storeA := NewOpenAIImageJobStore(rdb)
	require.NoError(t, storeA.Create(ctx, &service.OpenAIImageJobRecord{
		ID:        "imgjob_completed",
		Owner:     owner,
		Status:    service.OpenAIImageJobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}, time.Hour))
	started, err := storeA.SetRunning(ctx, "imgjob_completed", now.Add(time.Second), time.Hour)
	require.NoError(t, err)
	require.True(t, started)
	require.NoError(t, storeA.Complete(ctx, "imgjob_completed", service.OpenAIImageJobStatusSucceeded, http.StatusOK,
		http.Header{"Content-Type": []string{"application/json"}},
		[]byte(`{"data":[{"url":"https://example.test/a.png"}]}`),
		now.Add(2*time.Second), time.Hour))

	storeB := NewOpenAIImageJobStore(rdb)
	got, ok, err := storeB.Get(ctx, "imgjob_completed", owner)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, service.OpenAIImageJobStatusSucceeded, got.Status)
	require.Equal(t, http.StatusOK, got.StatusCode)
	require.JSONEq(t, `{"data":[{"url":"https://example.test/a.png"}]}`, string(got.Body))
	require.Equal(t, now.Add(2*time.Second), got.CompletedAt)
}

func TestOpenAIImageJobRedisStoreTTLExpiresRecord(t *testing.T) {
	rdb, mr := newOpenAIImageJobStoreTestRedis(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	owner := service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 7}
	store := NewOpenAIImageJobStore(rdb)

	require.NoError(t, store.Create(ctx, &service.OpenAIImageJobRecord{
		ID:        "imgjob_ttl",
		Owner:     owner,
		Status:    service.OpenAIImageJobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}, time.Second))
	mr.FastForward(2 * time.Second)

	got, ok, err := store.Get(ctx, "imgjob_ttl", owner)
	require.NoError(t, err)
	require.False(t, ok)
	require.Nil(t, got)
}

func TestOpenAIImageJobRedisStoreMarksOnlyStaleActiveJobsTimeout(t *testing.T) {
	rdb, _ := newOpenAIImageJobStoreTestRedis(t)
	ctx := context.Background()
	base := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	owner := service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 7}
	store := NewOpenAIImageJobStore(rdb)

	for _, rec := range []*service.OpenAIImageJobRecord{
		{ID: "imgjob_stale_queued", Owner: owner, Status: service.OpenAIImageJobStatusQueued, CreatedAt: base, UpdatedAt: base},
		{ID: "imgjob_stale_running", Owner: owner, Status: service.OpenAIImageJobStatusRunning, CreatedAt: base, UpdatedAt: base},
		{ID: "imgjob_fresh_running", Owner: owner, Status: service.OpenAIImageJobStatusRunning, CreatedAt: base, UpdatedAt: base.Add(9 * time.Minute)},
		{ID: "imgjob_terminal", Owner: owner, Status: service.OpenAIImageJobStatusSucceeded, StatusCode: http.StatusOK, CreatedAt: base, UpdatedAt: base, CompletedAt: base},
	} {
		require.NoError(t, store.Create(ctx, rec, time.Hour))
	}

	changed, err := store.MarkStaleTimeouts(ctx, base.Add(10*time.Minute), 5*time.Minute, 100, time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(2), changed)

	for _, id := range []string{"imgjob_stale_queued", "imgjob_stale_running"} {
		got, ok, err := store.Get(ctx, id, owner)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, service.OpenAIImageJobStatusTimeout, got.Status)
		require.Equal(t, http.StatusGatewayTimeout, got.StatusCode)
		require.Equal(t, base.Add(10*time.Minute), got.UpdatedAt)
		require.Equal(t, base.Add(10*time.Minute), got.CompletedAt)
		require.JSONEq(t, `{"error":{"type":"api_error","message":"Image job timed out"}}`, string(got.Body))
	}

	fresh, ok, err := store.Get(ctx, "imgjob_fresh_running", owner)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, service.OpenAIImageJobStatusRunning, fresh.Status)

	terminal, ok, err := store.Get(ctx, "imgjob_terminal", owner)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, service.OpenAIImageJobStatusSucceeded, terminal.Status)
}

func TestOpenAIImageJobRedisStoreLateCompleteDoesNotOverwriteTimeout(t *testing.T) {
	rdb, _ := newOpenAIImageJobStoreTestRedis(t)
	ctx := context.Background()
	base := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	owner := service.OpenAIImageJobOwner{UserID: 42, APIKeyID: 7}
	store := NewOpenAIImageJobStore(rdb)

	require.NoError(t, store.Create(ctx, &service.OpenAIImageJobRecord{
		ID:        "imgjob_late_complete",
		Owner:     owner,
		Status:    service.OpenAIImageJobStatusRunning,
		CreatedAt: base,
		UpdatedAt: base,
	}, time.Hour))
	changed, err := store.MarkStaleTimeouts(ctx, base.Add(10*time.Minute), 5*time.Minute, 100, time.Hour)
	require.NoError(t, err)
	require.Equal(t, int64(1), changed)

	require.NoError(t, store.Complete(ctx, "imgjob_late_complete", service.OpenAIImageJobStatusSucceeded, http.StatusOK,
		http.Header{"Content-Type": []string{"application/json"}},
		[]byte(`{"data":[{"url":"https://example.test/late.png"}]}`),
		base.Add(11*time.Minute), time.Hour))

	got, ok, err := store.Get(ctx, "imgjob_late_complete", owner)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, service.OpenAIImageJobStatusTimeout, got.Status)
	require.Equal(t, http.StatusGatewayTimeout, got.StatusCode)
	require.JSONEq(t, `{"error":{"type":"api_error","message":"Image job timed out"}}`, string(got.Body))
}
