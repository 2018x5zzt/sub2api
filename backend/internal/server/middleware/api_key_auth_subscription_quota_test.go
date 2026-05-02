package middleware

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestHasSubscriptionEntitlement(t *testing.T) {
	require.False(t, hasSubscriptionEntitlement(nil, nil))
	require.True(t, hasSubscriptionEntitlement(&service.ProductSettlementContext{}, nil))
	require.True(t, hasSubscriptionEntitlement(nil, &service.UserSubscription{}))
}
