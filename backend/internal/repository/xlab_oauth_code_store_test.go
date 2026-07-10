package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestXlabOAuthCodeStoreConsumesCodeOnce(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	store := NewXlabOAuthCodeStore(rdb)
	ctx := context.Background()
	require.NoError(t, store.StoreCode(ctx, "code-id", time.Minute))

	consumed, err := store.ConsumeCode(ctx, "code-id")
	require.NoError(t, err)
	require.True(t, consumed)

	consumed, err = store.ConsumeCode(ctx, "code-id")
	require.NoError(t, err)
	require.False(t, consumed)
}
