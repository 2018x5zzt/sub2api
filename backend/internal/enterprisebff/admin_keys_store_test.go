package enterprisebff

import (
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestNewAdminKeyStore_ConfiguresAPIKeyCacheFromRedisClient(t *testing.T) {
	t.Run("without redis client", func(t *testing.T) {
		store, ok := NewAdminKeyStore(nil, nil, nil, nil).(*adminKeyStore)
		require.True(t, ok)
		require.NotNil(t, store.apiKeyService)

		cacheField := reflect.ValueOf(store.apiKeyService).Elem().FieldByName("cache")
		require.True(t, cacheField.IsValid())
		require.True(t, cacheField.IsNil())
	})

	t.Run("with redis client", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
		t.Cleanup(func() { _ = rdb.Close() })

		store, ok := NewAdminKeyStore(nil, nil, rdb, nil).(*adminKeyStore)
		require.True(t, ok)
		require.NotNil(t, store.apiKeyService)

		cacheField := reflect.ValueOf(store.apiKeyService).Elem().FieldByName("cache")
		require.True(t, cacheField.IsValid())
		require.False(t, cacheField.IsNil())
	})
}
