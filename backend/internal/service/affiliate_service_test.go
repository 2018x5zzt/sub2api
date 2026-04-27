package service

import (
	"context"
	"math"
	"testing"
	"time"

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

func TestAffiliateService_AccrueInviteRebate_UsesTierRateFromEffectiveCommercialInvitees(t *testing.T) {
	ctx := context.Background()
	inviterID := int64(7)
	repo := &affiliateTierRepoStub{
		summaries: map[int64]*AffiliateSummary{
			7: {UserID: 7, AffCode: "INVITER", CreatedAt: time.Now().Add(-24 * time.Hour)},
			8: {UserID: 8, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-24 * time.Hour)},
		},
		effectiveInvitees: map[int64]int64{7: 100},
	}
	settings := NewSettingService(&affiliateTierSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateRebateRate:          "3",
		SettingKeyAffiliateRebateFreezeHours:   "0",
		SettingKeyAffiliateRebateDurationDays:  "0",
		SettingKeyAffiliateRebatePerInviteeCap: "0",
		SettingKeyAffiliateRebateTiers:         `[{"min_effective_invitees":0,"rebate_rate":3},{"min_effective_invitees":100,"rebate_rate":8}]`,
	}}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)

	rebate, err := svc.AccrueInviteRebate(ctx, 8, 100)

	require.NoError(t, err)
	require.InDelta(t, 8.0, rebate, 1e-9)
	require.Len(t, repo.accrued, 1)
	require.Equal(t, inviterID, repo.accrued[0].inviterID)
	require.Equal(t, int64(8), repo.accrued[0].inviteeID)
	require.InDelta(t, 8.0, repo.accrued[0].amount, 1e-9)
}

func TestAffiliateService_AccrueInviteRebate_PerUserRateOverridesTierRate(t *testing.T) {
	ctx := context.Background()
	inviterID := int64(7)
	overrideRate := 12.0
	repo := &affiliateTierRepoStub{
		summaries: map[int64]*AffiliateSummary{
			7: {UserID: 7, AffCode: "INVITER", AffRebateRatePercent: &overrideRate, CreatedAt: time.Now().Add(-24 * time.Hour)},
			8: {UserID: 8, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-24 * time.Hour)},
		},
		effectiveInvitees: map[int64]int64{7: 500},
	}
	settings := NewSettingService(&affiliateTierSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateRebateRate:          "3",
		SettingKeyAffiliateRebateFreezeHours:   "0",
		SettingKeyAffiliateRebateDurationDays:  "0",
		SettingKeyAffiliateRebatePerInviteeCap: "0",
		SettingKeyAffiliateRebateTiers:         `[{"min_effective_invitees":100,"rebate_rate":8},{"min_effective_invitees":500,"rebate_rate":15}]`,
	}}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)

	rebate, err := svc.AccrueInviteRebate(ctx, 8, 100)

	require.NoError(t, err)
	require.InDelta(t, 12.0, rebate, 1e-9)
	require.Len(t, repo.accrued, 1)
	require.InDelta(t, 12.0, repo.accrued[0].amount, 1e-9)
}

type affiliateTierRepoStub struct {
	summaries         map[int64]*AffiliateSummary
	accruedByPair     map[[2]int64]float64
	effectiveInvitees map[int64]int64
	accrued           []affiliateTierAccrual
}

type affiliateTierAccrual struct {
	inviterID int64
	inviteeID int64
	amount    float64
}

func (s *affiliateTierRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := s.summaries[userID]; ok {
		clone := *summary
		return &clone, nil
	}
	return nil, ErrAffiliateProfileNotFound
}

func (s *affiliateTierRepoStub) GetAffiliateByCode(context.Context, string) (*AffiliateSummary, error) {
	return nil, ErrAffiliateProfileNotFound
}

func (s *affiliateTierRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (s *affiliateTierRepoStub) AccrueQuota(_ context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int) (bool, error) {
	s.accrued = append(s.accrued, affiliateTierAccrual{inviterID: inviterID, inviteeID: inviteeUserID, amount: amount})
	return true, nil
}

func (s *affiliateTierRepoStub) GetAccruedRebateFromInvitee(_ context.Context, inviterID, inviteeUserID int64) (float64, error) {
	if s.accruedByPair == nil {
		return 0, nil
	}
	return s.accruedByPair[[2]int64{inviterID, inviteeUserID}], nil
}

func (s *affiliateTierRepoStub) CountEffectiveInvitees(_ context.Context, inviterID int64) (int64, error) {
	if s.effectiveInvitees == nil {
		return 0, nil
	}
	return s.effectiveInvitees[inviterID], nil
}

func (s *affiliateTierRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	return 0, nil
}

func (s *affiliateTierRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	return 0, 0, ErrAffiliateQuotaEmpty
}

func (s *affiliateTierRepoStub) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	return nil, nil
}

func (s *affiliateTierRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	return nil
}

func (s *affiliateTierRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	return "", nil
}

func (s *affiliateTierRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	return nil
}

func (s *affiliateTierRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	return nil
}

func (s *affiliateTierRepoStub) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}

type affiliateTierSettingRepoStub struct {
	values map[string]string
}

func (s *affiliateTierSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}

func (s *affiliateTierSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *affiliateTierSettingRepoStub) Set(context.Context, string, string) error {
	return nil
}

func (s *affiliateTierSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		result[key] = s.values[key]
	}
	return result, nil
}

func (s *affiliateTierSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	return nil
}

func (s *affiliateTierSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *affiliateTierSettingRepoStub) Delete(context.Context, string) error {
	return nil
}
