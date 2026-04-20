package enterprisebff

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type settingRepoStub struct {
	value string
	err   error
}

func (s *settingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *settingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	return s.value, s.err
}

func (s *settingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSelectEnterpriseVisibleGroups_ReturnsOnlyConfiguredGroups(t *testing.T) {
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
		{
			ID:               4,
			Name:             "inactive-configured",
			Platform:         service.PlatformOpenAI,
			Status:           service.StatusDisabled,
			IsExclusive:      false,
			SubscriptionType: service.SubscriptionTypeStandard,
		},
	}

	got := selectEnterpriseVisibleGroups(groups, map[int64]struct{}{
		2: {},
		3: {},
		4: {},
	})

	require.Equal(t, []EnterpriseVisibleGroup{
		{ID: 2, Name: "exclusive-private", Platform: service.PlatformOpenAI},
		{ID: 3, Name: "subscription-only", Platform: service.PlatformGemini},
	}, got)
}

func TestSelectEnterpriseVisibleGroups_ReturnsEmptyWhenEnterpriseHasNoConfiguredGroups(t *testing.T) {
	groups := []service.Group{
		{
			ID:               1,
			Name:             "public-default",
			Platform:         service.PlatformAnthropic,
			Status:           service.StatusActive,
			IsExclusive:      false,
			SubscriptionType: service.SubscriptionTypeStandard,
		},
	}

	got := selectEnterpriseVisibleGroups(groups, nil)

	require.Empty(t, got)
}

func TestEntEnterpriseStore_VisibleGroupIDSetForEnterprise_PrefersDBSetting(t *testing.T) {
	store := &entEnterpriseStore{
		settingRepo: &settingRepoStub{
			value: `{"acme":[2,9]}`,
		},
		visibleGroupIDsByEnterprise: normalizeEnterpriseVisibleGroupIDsByEnterprise(map[string][]int64{
			"acme": {11},
		}),
	}

	got := store.visibleGroupIDSetForEnterprise(context.Background(), &EnterpriseProfile{Name: " Acme "})

	require.Equal(t, map[int64]struct{}{
		2: {},
		9: {},
	}, got)
}

func TestEntEnterpriseStore_VisibleGroupIDSetForEnterprise_FallsBackToEnvWhenDBSettingMissing(t *testing.T) {
	store := &entEnterpriseStore{
		settingRepo: &settingRepoStub{
			err: service.ErrSettingNotFound,
		},
		visibleGroupIDsByEnterprise: normalizeEnterpriseVisibleGroupIDsByEnterprise(map[string][]int64{
			"acme": {11, 16},
		}),
	}

	got := store.visibleGroupIDSetForEnterprise(context.Background(), &EnterpriseProfile{Name: " Acme "})

	require.Equal(t, map[int64]struct{}{
		11: {},
		16: {},
	}, got)
}
