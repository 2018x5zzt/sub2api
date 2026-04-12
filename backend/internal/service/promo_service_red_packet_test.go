package service

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

var _ io.Reader = zeroReader{}

func TestPromoServiceAllocateBenefitRandomBonus_DeterministicMinimumDraw(t *testing.T) {
	svc := &PromoService{randomReader: zeroReader{}}

	got, err := svc.allocateBenefitRandomBonus(&PromoCode{
		Scene:                 PromoCodeSceneBenefit,
		RandomBonusPoolAmount: 10,
		RandomBonusRemaining:  10,
		MaxUses:               2,
		UsedCount:             0,
	})
	require.NoError(t, err)
	require.Equal(t, 0.01, got)
}

func TestRankBenefitUsages_SortsByRandomBonusDescending(t *testing.T) {
	now := time.Now()
	result := rankBenefitUsages([]PromoCodeUsage{
		{
			UserID:            1,
			BonusAmount:       38,
			FixedBonusAmount:  20,
			RandomBonusAmount: 18,
			UsedAt:            now,
			User:              &User{Username: "alpha"},
		},
		{
			UserID:            2,
			BonusAmount:       52,
			FixedBonusAmount:  20,
			RandomBonusAmount: 32,
			UsedAt:            now.Add(time.Second),
			User:              &User{Username: "bravo"},
		},
	}, 1, 20)

	require.Len(t, result.Entries, 2)
	require.Equal(t, "bravo", result.Entries[0].DisplayName)
	require.Equal(t, 32.0, result.Entries[0].RandomBonusAmount)
	require.Equal(t, "alpha", result.Entries[1].DisplayName)
	require.NotNil(t, result.CurrentUserRank)
	require.Equal(t, 2, *result.CurrentUserRank)
}

func TestRankBenefitUsages_WithNonPositiveLimitDefaultsToTopTwenty(t *testing.T) {
	now := time.Now()
	usages := make([]PromoCodeUsage, 0, 21)
	for i := 1; i <= 21; i++ {
		usages = append(usages, PromoCodeUsage{
			UserID:            int64(i),
			BonusAmount:       float64(22 - i),
			FixedBonusAmount:  1,
			RandomBonusAmount: float64(21 - i),
			UsedAt:            now.Add(time.Duration(i) * time.Second),
			User:              &User{Username: fmt.Sprintf("user_%02d", i)},
		})
	}

	result := rankBenefitUsages(usages, 1, 0)

	require.Len(t, result.Entries, 20)
	require.Equal(t, "user_01", result.Entries[0].DisplayName)
	require.Equal(t, "user_20", result.Entries[19].DisplayName)
	require.NotNil(t, result.CurrentUserRank)
	require.Equal(t, 1, *result.CurrentUserRank)
}
