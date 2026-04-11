package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeGroupPoolHealthMatchesExistingRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		group Group
		want  GroupPoolHealth
	}{
		{
			name: "healthy when all active accounts are available",
			group: Group{
				ActiveAccountCount:      5,
				RateLimitedAccountCount: 0,
			},
			want: GroupPoolHealth{
				AvailableAccountCount:   5,
				RateLimitedAccountCount: 0,
				HealthPercent:           100,
				HealthState:             "healthy",
			},
		},
		{
			name: "degraded when some active accounts are rate limited",
			group: Group{
				ActiveAccountCount:      5,
				RateLimitedAccountCount: 2,
			},
			want: GroupPoolHealth{
				AvailableAccountCount:   3,
				RateLimitedAccountCount: 2,
				HealthPercent:           60,
				HealthState:             "degraded",
			},
		},
		{
			name: "down when rate limited exceeds active after clamping negatives",
			group: Group{
				ActiveAccountCount:      -2,
				RateLimitedAccountCount: 3,
			},
			want: GroupPoolHealth{
				AvailableAccountCount:   0,
				RateLimitedAccountCount: 3,
				HealthPercent:           0,
				HealthState:             "down",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ComputeGroupPoolHealth(&tt.group)
			require.Equal(t, tt.want, got)
		})
	}
}
