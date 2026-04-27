//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/userproductsubscription"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestSubscriptionProductAdminService_AssignOrExtendProductSubscription(t *testing.T) {
	client := newSubscriptionProductAdminTestClient(t)
	ctx := context.Background()
	product := mustCreateSubscriptionProductAdminTestProduct(t, ctx, client, 14)

	svc := NewSubscriptionProductAdminService(client)
	sub, reused, err := svc.AssignOrExtendProductSubscription(ctx, &AssignProductSubscriptionInput{
		UserID:       77,
		ProductID:    product.ID,
		ValidityDays: 7,
		Notes:        "first grant",
	})
	require.NoError(t, err)
	require.False(t, reused)
	require.Equal(t, int64(77), sub.UserID)
	require.Equal(t, product.ID, sub.ProductID)
	require.Equal(t, SubscriptionStatusActive, sub.Status)
	require.WithinDuration(t, time.Now().AddDate(0, 0, 7), sub.ExpiresAt, 5*time.Second)

	extended, reused, err := svc.AssignOrExtendProductSubscription(ctx, &AssignProductSubscriptionInput{
		UserID:       77,
		ProductID:    product.ID,
		ValidityDays: 3,
		Notes:        "second grant",
	})
	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, sub.ID, extended.ID)
	require.WithinDuration(t, sub.ExpiresAt.AddDate(0, 0, 3), extended.ExpiresAt, time.Second)
	require.Contains(t, extended.Notes, "first grant")
	require.Contains(t, extended.Notes, "second grant")
}

func TestRedeemService_Redeem_ProductSubscriptionCodeCreatesUserProductSubscription(t *testing.T) {
	client := newSubscriptionProductAdminTestClient(t)
	ctx := context.Background()
	product := mustCreateSubscriptionProductAdminTestProduct(t, ctx, client, 30)
	productID := product.ID

	redeemRepo := &productRedeemCodeRepoStub{
		code: &RedeemCode{
			ID:           1,
			Code:         "PRODUCT-30",
			Type:         RedeemTypeSubscription,
			Status:       StatusUnused,
			SourceType:   RedeemSourceSystemGrant,
			ProductID:    &productID,
			ValidityDays: 30,
		},
	}
	productAdmin := NewSubscriptionProductAdminService(client)
	redeemSvc := NewRedeemService(
		redeemRepo,
		&userRepoStub{user: &User{ID: 42, Email: "product-redeem@test.com", Status: StatusActive}},
		nil,
		nil,
		nil,
		nil,
		client,
		nil,
		productAdmin,
	)

	redeemed, err := redeemSvc.Redeem(ctx, 42, "PRODUCT-30")
	require.NoError(t, err)
	require.NotNil(t, redeemed.ProductID)
	require.Equal(t, product.ID, *redeemed.ProductID)
	require.Nil(t, redeemed.GroupID)

	subs, err := client.UserProductSubscription.Query().
		Where(
			userproductsubscription.UserIDEQ(42),
			userproductsubscription.ProductIDEQ(product.ID),
			userproductsubscription.DeletedAtIsNil(),
		).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	require.Equal(t, SubscriptionStatusActive, subs[0].Status)
	require.WithinDuration(t, time.Now().AddDate(0, 0, 30), subs[0].ExpiresAt, 5*time.Second)
	require.Equal(t, StatusUsed, redeemRepo.code.Status)
}

func TestRedeemService_Redeem_CommercialProductSubscriptionAccruesAffiliateRebate(t *testing.T) {
	client := newSubscriptionProductAdminTestClient(t)
	ctx := context.Background()
	product := mustCreateSubscriptionProductAdminTestProductWithLimits(t, ctx, client, 30, 225, 1575, 6750)
	productID := product.ID
	inviterID := int64(7)

	redeemRepo := &productRedeemCodeRepoStub{
		code: &RedeemCode{
			ID:           1,
			Code:         "PRODUCT-COMMERCIAL-30",
			Type:         RedeemTypeSubscription,
			Value:        225,
			Status:       StatusUnused,
			SourceType:   RedeemSourceCommercial,
			ProductID:    &productID,
			ValidityDays: 30,
		},
	}
	affiliateRepo := &affiliateTierRepoStub{
		summaries: map[int64]*AffiliateSummary{
			7:  {UserID: 7, AffCode: "INVITER", CreatedAt: time.Now().Add(-24 * time.Hour)},
			42: {UserID: 42, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-24 * time.Hour)},
		},
		effectiveInvitees: map[int64]int64{7: 3},
	}
	affiliateSettings := NewSettingService(&affiliateTierSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateRebateRate:          "3",
		SettingKeyAffiliateRebateFreezeHours:   "0",
		SettingKeyAffiliateRebateDurationDays:  "0",
		SettingKeyAffiliateRebatePerInviteeCap: "0",
	}}, nil)
	productAdmin := NewSubscriptionProductAdminService(client)
	redeemSvc := NewRedeemService(
		redeemRepo,
		&userRepoStub{user: &User{ID: 42, Email: "product-redeem@test.com", Status: StatusActive}},
		NewAffiliateService(affiliateRepo, affiliateSettings, nil, nil),
		nil,
		nil,
		nil,
		client,
		nil,
		productAdmin,
	)

	redeemed, err := redeemSvc.Redeem(ctx, 42, "PRODUCT-COMMERCIAL-30")

	require.NoError(t, err)
	require.NotNil(t, redeemed.ProductID)
	require.Len(t, affiliateRepo.accrued, 1)
	require.Equal(t, inviterID, affiliateRepo.accrued[0].inviterID)
	require.Equal(t, int64(42), affiliateRepo.accrued[0].inviteeID)
	require.InDelta(t, 162.0, affiliateRepo.accrued[0].amount, 1e-9)
}

func TestResolveSubscriptionAffiliateRebateTerms_UsesTotalQuotaAndSkuFactor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		validityDays int
		daily        float64
		weekly       float64
		monthly      float64
		wantBase     float64
		wantFactor   float64
		wantOK       bool
	}{
		{name: "daily card uses daily total quota and full factor", validityDays: 1, daily: 225, weekly: 1575, monthly: 6750, wantBase: 225, wantFactor: 1, wantOK: true},
		{name: "weekly card uses weekly total quota and 60 percent factor", validityDays: 7, daily: 225, weekly: 1575, monthly: 6750, wantBase: 1575, wantFactor: 0.6, wantOK: true},
		{name: "monthly card uses monthly total quota and 30 percent factor", validityDays: 30, daily: 225, weekly: 1575, monthly: 6750, wantBase: 6750, wantFactor: 0.3, wantOK: true},
		{name: "unknown subscription shape is not rebated", validityDays: 3, daily: 0, weekly: 0, monthly: 0, wantOK: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base, factor, ok := resolveSubscriptionAffiliateRebateTerms(tc.validityDays, tc.daily, tc.weekly, tc.monthly)

			require.Equal(t, tc.wantOK, ok)
			if !tc.wantOK {
				return
			}
			require.InDelta(t, tc.wantBase, base, 1e-9)
			require.InDelta(t, tc.wantFactor, factor, 1e-9)
		})
	}
}

func TestRedeemService_Redeem_SystemGrantProductSubscriptionSkipsAffiliateRebate(t *testing.T) {
	client := newSubscriptionProductAdminTestClient(t)
	ctx := context.Background()
	product := mustCreateSubscriptionProductAdminTestProduct(t, ctx, client, 30)
	productID := product.ID
	inviterID := int64(7)

	redeemRepo := &productRedeemCodeRepoStub{
		code: &RedeemCode{
			ID:           1,
			Code:         "PRODUCT-GRANT-30",
			Type:         RedeemTypeSubscription,
			Value:        200,
			Status:       StatusUnused,
			SourceType:   RedeemSourceSystemGrant,
			ProductID:    &productID,
			ValidityDays: 30,
		},
	}
	affiliateRepo := &affiliateTierRepoStub{
		summaries: map[int64]*AffiliateSummary{
			7:  {UserID: 7, AffCode: "INVITER", CreatedAt: time.Now().Add(-24 * time.Hour)},
			42: {UserID: 42, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-24 * time.Hour)},
		},
	}
	affiliateSettings := NewSettingService(&affiliateTierSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateRebateRate:          "10",
		SettingKeyAffiliateRebateFreezeHours:   "0",
		SettingKeyAffiliateRebateDurationDays:  "0",
		SettingKeyAffiliateRebatePerInviteeCap: "0",
	}}, nil)
	productAdmin := NewSubscriptionProductAdminService(client)
	redeemSvc := NewRedeemService(
		redeemRepo,
		&userRepoStub{user: &User{ID: 42, Email: "product-redeem@test.com", Status: StatusActive}},
		NewAffiliateService(affiliateRepo, affiliateSettings, nil, nil),
		nil,
		nil,
		nil,
		client,
		nil,
		productAdmin,
	)

	redeemed, err := redeemSvc.Redeem(ctx, 42, "PRODUCT-GRANT-30")

	require.NoError(t, err)
	require.NotNil(t, redeemed.ProductID)
	require.Empty(t, affiliateRepo.accrued)
}

func newSubscriptionProductAdminTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:subscription_product_admin_test?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func mustCreateSubscriptionProductAdminTestProduct(t *testing.T, ctx context.Context, client *dbent.Client, defaultValidityDays int) *dbent.SubscriptionProduct {
	t.Helper()

	return mustCreateSubscriptionProductAdminTestProductWithLimits(t, ctx, client, defaultValidityDays, 0, 0, 0)
}

func mustCreateSubscriptionProductAdminTestProductWithLimits(t *testing.T, ctx context.Context, client *dbent.Client, defaultValidityDays int, dailyLimit, weeklyLimit, monthlyLimit float64) *dbent.SubscriptionProduct {
	t.Helper()

	product, err := client.SubscriptionProduct.Create().
		SetCode("product-admin-test").
		SetName("Product Admin Test").
		SetStatus(SubscriptionProductStatusActive).
		SetDefaultValidityDays(defaultValidityDays).
		SetDailyLimitUsd(dailyLimit).
		SetWeeklyLimitUsd(weeklyLimit).
		SetMonthlyLimitUsd(monthlyLimit).
		Save(ctx)
	require.NoError(t, err)
	return product
}

type productRedeemCodeRepoStub struct {
	code *RedeemCode
}

func (s *productRedeemCodeRepoStub) Create(context.Context, *RedeemCode) error {
	panic("unexpected Create call")
}

func (s *productRedeemCodeRepoStub) CreateBatch(context.Context, []RedeemCode) error {
	panic("unexpected CreateBatch call")
}

func (s *productRedeemCodeRepoStub) GetByID(_ context.Context, id int64) (*RedeemCode, error) {
	if s.code == nil || s.code.ID != id {
		return nil, ErrRedeemCodeNotFound
	}
	clone := *s.code
	return &clone, nil
}

func (s *productRedeemCodeRepoStub) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	if s.code == nil || s.code.Code != code {
		return nil, ErrRedeemCodeNotFound
	}
	clone := *s.code
	return &clone, nil
}

func (s *productRedeemCodeRepoStub) Update(context.Context, *RedeemCode) error {
	panic("unexpected Update call")
}

func (s *productRedeemCodeRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *productRedeemCodeRepoStub) Use(_ context.Context, id, userID int64) error {
	if s.code == nil || s.code.ID != id || s.code.Status != StatusUnused {
		return ErrRedeemCodeUsed
	}
	now := time.Now()
	s.code.Status = StatusUsed
	s.code.UsedBy = &userID
	s.code.UsedAt = &now
	return nil
}

func (s *productRedeemCodeRepoStub) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *productRedeemCodeRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *productRedeemCodeRepoStub) ListByUser(context.Context, int64, int) ([]RedeemCode, error) {
	panic("unexpected ListByUser call")
}

func (s *productRedeemCodeRepoStub) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserPaginated call")
}

func (s *productRedeemCodeRepoStub) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	panic("unexpected SumPositiveBalanceByUser call")
}
