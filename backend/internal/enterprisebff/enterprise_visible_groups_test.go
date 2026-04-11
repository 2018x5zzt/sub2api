package enterprisebff

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSelectEnterpriseVisibleGroups_IncludesPublicStandardGroupsByDefault(t *testing.T) {
	groups := []service.Group{
		{
			ID:               1,
			Name:             "public-default",
			Platform:         service.PlatformAnthropic,
			Status:           service.StatusActive,
			IsExclusive:      false,
			SubscriptionType: service.SubscriptionTypeStandard,
		},
		{
			ID:               2,
			Name:             "exclusive-private",
			Platform:         service.PlatformOpenAI,
			Status:           service.StatusActive,
			IsExclusive:      true,
			SubscriptionType: service.SubscriptionTypeStandard,
		},
		{
			ID:               3,
			Name:             "subscription-only",
			Platform:         service.PlatformGemini,
			Status:           service.StatusActive,
			IsExclusive:      false,
			SubscriptionType: service.SubscriptionTypeSubscription,
		},
	}

	got := selectEnterpriseVisibleGroups(groups, map[int64]struct{}{}, map[int64]struct{}{})

	require.Equal(t, []EnterpriseVisibleGroup{
		{ID: 1, Name: "public-default", Platform: service.PlatformAnthropic},
	}, got)
}
