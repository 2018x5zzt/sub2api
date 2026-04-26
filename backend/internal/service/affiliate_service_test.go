package service

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateService_XlabapiDefaults(t *testing.T) {
	t.Parallel()

	svc := &AffiliateService{}

	require.True(t, svc.IsEnabled(context.Background()))
	require.Equal(t, AffiliateEnabledDefault, svc.IsEnabled(context.Background()))
	require.InDelta(t, 3.0, AffiliateRebateRateDefault, 1e-9)
	require.InDelta(t, 3.0, svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{}), 1e-9)
}

func TestAffiliateService_ResolveRebateRatePercent_PerUserOverride(t *testing.T) {
	t.Parallel()

	svc := &AffiliateService{}

	rate := 50.0
	require.InDelta(t, 50.0, svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &rate}), 1e-9)

	zero := 0.0
	require.InDelta(t, 0.0, svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &zero}), 1e-9)

	tooHigh := 250.0
	require.InDelta(t, AffiliateRebateRateMax, svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooHigh}), 1e-9)

	tooLow := -5.0
	require.InDelta(t, AffiliateRebateRateMin, svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooLow}), 1e-9)
}

func TestAffiliateService_ValidateExclusiveRate(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateExclusiveRate(nil))
	for _, v := range []float64{0, 0.01, 50, 99.99, 100} {
		v := v
		require.NoError(t, validateExclusiveRate(&v), "value %v should be valid", v)
	}
	for _, v := range []float64{-0.01, 100.01, -100, 200} {
		v := v
		require.Error(t, validateExclusiveRate(&v), "value %v should be rejected", v)
	}

	nan := math.NaN()
	require.Error(t, validateExclusiveRate(&nan))
	posInf := math.Inf(1)
	require.Error(t, validateExclusiveRate(&posInf))
	negInf := math.Inf(-1)
	require.Error(t, validateExclusiveRate(&negInf))
}

func TestAffiliateService_IsValidAffiliateCodeFormat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid canonical 12-char", "ABCDEFGHJKLM", true},
		{"valid all digits", "012345678901", true},
		{"valid admin custom short", "VIP1", true},
		{"valid admin custom with hyphen", "NEW-USER", true},
		{"valid admin custom with underscore", "VIP_2026", true},
		{"valid 32-char max", "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345", true},
		{"too short", "ABC", false},
		{"too long", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456", false},
		{"lowercase rejected before caller normalization", "abcdefghjklm", false},
		{"empty", "", false},
		{"non-ascii", "ÄÄÄÄÄÄ", false},
		{"punctuation", "ABCDEFGHJK.M", false},
		{"whitespace", "ABCDEFGHJK M", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, isValidAffiliateCodeFormat(tc.in))
		})
	}
}
