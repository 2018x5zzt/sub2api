//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type productSubscriptionRepoStub struct {
	binding         *SubscriptionProductBinding
	sub             *UserProductSubscription
	products        []ActiveSubscriptionProduct
	groups          []Group
	resolvedProduct *SubscriptionProduct
	requestedFamily *string
	err             error
	assigned        *UserProductSubscription
	assignInputs    []AssignProductSubscriptionInput
	reused          bool
}

func (r *productSubscriptionRepoStub) GetActiveProductSubscriptionByUserAndGroupID(_ context.Context, _ int64, _ int64, productFamily *string) (*SubscriptionProductBinding, *UserProductSubscription, error) {
	r.requestedFamily = productFamily
	if r.err != nil {
		return nil, nil, r.err
	}
	if r.binding == nil || r.sub == nil {
		return nil, nil, ErrSubscriptionNotFound
	}
	return r.binding, r.sub, nil
}

func (r *productSubscriptionRepoStub) ListActiveProductsByUserID(context.Context, int64) ([]ActiveSubscriptionProduct, error) {
	out := make([]ActiveSubscriptionProduct, len(r.products))
	copy(out, r.products)
	return out, nil
}

func (r *productSubscriptionRepoStub) ListVisibleGroupsByUserID(context.Context, int64) ([]Group, error) {
	out := make([]Group, len(r.groups))
	copy(out, r.groups)
	return out, nil
}

func (r *productSubscriptionRepoStub) ListProducts(context.Context) ([]SubscriptionProduct, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) ResolveActiveProductByGroupID(context.Context, int64) (*SubscriptionProduct, error) {
	if r.resolvedProduct == nil {
		return nil, ErrSubscriptionNotFound
	}
	return r.resolvedProduct, nil
}

func (r *productSubscriptionRepoStub) CreateProduct(context.Context, *CreateSubscriptionProductInput) (*SubscriptionProduct, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) UpdateProduct(context.Context, int64, *UpdateSubscriptionProductInput) (*SubscriptionProduct, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) SyncProductBindings(context.Context, int64, []SubscriptionProductBindingInput) ([]SubscriptionProductBindingDetail, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) ListProductBindings(context.Context, int64) ([]SubscriptionProductBindingDetail, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) ListProductSubscriptions(context.Context, int64) ([]UserProductSubscription, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) ListUserProductSubscriptionsForAdmin(context.Context, AdminProductSubscriptionListParams) ([]AdminProductSubscriptionListItem, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (r *productSubscriptionRepoStub) AdjustProductSubscription(context.Context, int64, *AdjustProductSubscriptionInput) (*UserProductSubscription, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) ResetProductSubscriptionQuota(context.Context, int64, *ResetProductSubscriptionQuotaInput) (*UserProductSubscription, error) {
	return nil, nil
}

func (r *productSubscriptionRepoStub) RevokeProductSubscription(context.Context, int64) error {
	return nil
}

func (r *productSubscriptionRepoStub) AssignOrExtendProductSubscription(_ context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if r.assignInputs == nil {
		r.assignInputs = []AssignProductSubscriptionInput{}
	}
	if input != nil {
		r.assignInputs = append(r.assignInputs, *input)
	}
	return r.assigned, r.reused, nil
}

func TestSubscriptionProductServiceListActiveUserProductsReturnsSharedProductGroups(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(24 * time.Hour)
	svc := NewSubscriptionProductService(&productSubscriptionRepoStub{
		products: []ActiveSubscriptionProduct{
			{
				Product: SubscriptionProduct{
					ID:              88,
					Code:            "gpt_monthly",
					Name:            "GPT Monthly",
					Description:     "shared GPT pool",
					Status:          SubscriptionProductStatusActive,
					DailyLimitUSD:   45,
					WeeklyLimitUSD:  315,
					MonthlyLimitUSD: 1350,
				},
				Subscription: UserProductSubscription{
					ID:                         99,
					UserID:                     7,
					ProductID:                  88,
					ExpiresAt:                  expiresAt,
					Status:                     SubscriptionStatusActive,
					DailyUsageUSD:              6,
					DailyCarryoverInUSD:        38,
					DailyCarryoverRemainingUSD: 32,
				},
				Groups: []SubscriptionProductGroupSummary{
					{GroupID: 21, GroupName: "plus-team", DebitMultiplier: 1, Status: SubscriptionProductBindingStatusActive, SortOrder: 10},
					{GroupID: 36, GroupName: "pro", DebitMultiplier: 1.5, Status: SubscriptionProductBindingStatusActive, SortOrder: 20},
				},
			},
		},
	})

	products, err := svc.ListActiveUserProducts(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListActiveUserProducts returned error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("len(products) = %d, want 1", len(products))
	}
	got := products[0]
	if got.Product.Code != "gpt_monthly" {
		t.Fatalf("Product.Code = %q, want gpt_monthly", got.Product.Code)
	}
	if got.Subscription.ID != 99 {
		t.Fatalf("Subscription.ID = %d, want 99", got.Subscription.ID)
	}
	if len(got.Groups) != 2 {
		t.Fatalf("len(Groups) = %d, want 2", len(got.Groups))
	}
	if got.Groups[1].GroupName != "pro" || got.Groups[1].DebitMultiplier != 1.5 {
		t.Fatalf("second group = %+v, want pro at 1.5x", got.Groups[1])
	}
}

func TestSubscriptionProductServicePassesExplicitProductFamilyToRepository(t *testing.T) {
	t.Parallel()

	repo := &productSubscriptionRepoStub{
		binding: &SubscriptionProductBinding{
			ProductStatus: SubscriptionProductStatusActive,
			BindingStatus: SubscriptionProductBindingStatusActive,
		},
		sub: &UserProductSubscription{
			Status:    SubscriptionStatusActive,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
	svc := NewSubscriptionProductService(repo)
	family := "image"

	_, err := svc.GetActiveProductSubscriptionForFamily(context.Background(), 7, 30, &family)

	if err != nil {
		t.Fatalf("GetActiveProductSubscriptionForFamily returned error: %v", err)
	}
	if repo.requestedFamily == nil || *repo.requestedFamily != family {
		t.Fatalf("requested family = %v, want %q", repo.requestedFamily, family)
	}
}

func TestSubscriptionProductServiceReturnsFamilyRequired(t *testing.T) {
	t.Parallel()

	svc := NewSubscriptionProductService(&productSubscriptionRepoStub{err: ErrProductFamilyRequired})

	_, err := svc.GetActiveProductSubscriptionForFamily(context.Background(), 7, 30, nil)

	if !errors.Is(err, ErrProductFamilyRequired) {
		t.Fatalf("err = %v, want ErrProductFamilyRequired", err)
	}
}

func TestNormalizeExpiredProductSubscriptionWindowKeepsMonthlyOnCalendarMonthBoundary(t *testing.T) {
	t.Parallel()

	location := time.UTC
	previousMonthWindow := time.Date(2026, 4, 30, 0, 0, 0, 0, location)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, location)
	sub := &UserProductSubscription{
		Status:             SubscriptionStatusActive,
		ExpiresAt:          now.Add(24 * time.Hour),
		MonthlyWindowStart: &previousMonthWindow,
		MonthlyUsageUSD:    149,
	}
	product := &SubscriptionProduct{MonthlyLimitUSD: 150}

	NormalizeExpiredProductSubscriptionWindowForRepository(sub, product, now)

	if sub.MonthlyWindowStart == nil {
		t.Fatal("MonthlyWindowStart is nil, want original rolling window start")
	}
	if !sub.MonthlyWindowStart.Equal(previousMonthWindow) {
		t.Fatalf("MonthlyWindowStart = %s, want %s", sub.MonthlyWindowStart.Format(time.RFC3339), previousMonthWindow.Format(time.RFC3339))
	}
	if sub.MonthlyUsageUSD != 149 {
		t.Fatalf("MonthlyUsageUSD = %v, want 149", sub.MonthlyUsageUSD)
	}
}

func TestNormalizeExpiredProductSubscriptionWindowResetsMonthlyAfterRollingThirtyDays(t *testing.T) {
	t.Parallel()

	location := time.UTC
	previousWindow := time.Date(2026, 4, 1, 9, 30, 0, 0, location)
	now := previousWindow.Add(31 * 24 * time.Hour)
	sub := &UserProductSubscription{
		Status:             SubscriptionStatusActive,
		ExpiresAt:          now.Add(24 * time.Hour),
		MonthlyWindowStart: &previousWindow,
		MonthlyUsageUSD:    149,
	}
	product := &SubscriptionProduct{MonthlyLimitUSD: 150}

	NormalizeExpiredProductSubscriptionWindowForRepository(sub, product, now)

	wantWindow := previousWindow.Add(30 * 24 * time.Hour)
	if sub.MonthlyWindowStart == nil {
		t.Fatal("MonthlyWindowStart is nil, want next rolling 30 day start")
	}
	if !sub.MonthlyWindowStart.Equal(wantWindow) {
		t.Fatalf("MonthlyWindowStart = %s, want %s", sub.MonthlyWindowStart.Format(time.RFC3339), wantWindow.Format(time.RFC3339))
	}
	if sub.MonthlyUsageUSD != 0 {
		t.Fatalf("MonthlyUsageUSD = %v, want 0", sub.MonthlyUsageUSD)
	}
}

func TestNormalizeExpiredProductSubscriptionWindowResetsDailyAtBeijingMidnight(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 4, 16, 30, 0, 0, time.UTC)
	sub := &UserProductSubscription{
		Status:             SubscriptionStatusActive,
		ExpiresAt:          now.Add(24 * time.Hour),
		DailyWindowStart:   &windowStart,
		DailyUsageUSD:      9,
		WeeklyWindowStart:  &windowStart,
		WeeklyUsageUSD:     9,
		MonthlyWindowStart: &windowStart,
		MonthlyUsageUSD:    9,
	}
	product := &SubscriptionProduct{DailyLimitUSD: 10, WeeklyLimitUSD: 100, MonthlyLimitUSD: 300}

	NormalizeExpiredProductSubscriptionWindowForRepository(sub, product, now)

	wantDaily := time.Date(2026, 5, 4, 16, 0, 0, 0, time.UTC)
	if sub.DailyWindowStart == nil {
		t.Fatal("DailyWindowStart is nil, want Beijing midnight in UTC")
	}
	if !sub.DailyWindowStart.Equal(wantDaily) {
		t.Fatalf("DailyWindowStart = %s, want %s", sub.DailyWindowStart.Format(time.RFC3339), wantDaily.Format(time.RFC3339))
	}
	if sub.DailyUsageUSD != 0 {
		t.Fatalf("DailyUsageUSD = %v, want 0", sub.DailyUsageUSD)
	}
}

func TestNormalizeExpiredProductSubscriptionWindowRollsWeeklyFromCurrentStart(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)
	now := windowStart.Add(8 * 24 * time.Hour)
	sub := &UserProductSubscription{
		Status:            SubscriptionStatusActive,
		ExpiresAt:         now.Add(24 * time.Hour),
		WeeklyWindowStart: &windowStart,
		WeeklyUsageUSD:    77,
	}
	product := &SubscriptionProduct{WeeklyLimitUSD: 100}

	NormalizeExpiredProductSubscriptionWindowForRepository(sub, product, now)

	wantWindow := windowStart.Add(7 * 24 * time.Hour)
	if sub.WeeklyWindowStart == nil {
		t.Fatal("WeeklyWindowStart is nil, want next rolling 7 day start")
	}
	if !sub.WeeklyWindowStart.Equal(wantWindow) {
		t.Fatalf("WeeklyWindowStart = %s, want %s", sub.WeeklyWindowStart.Format(time.RFC3339), wantWindow.Format(time.RFC3339))
	}
	if sub.WeeklyUsageUSD != 0 {
		t.Fatalf("WeeklyUsageUSD = %v, want 0", sub.WeeklyUsageUSD)
	}
}

type legacyDefaultSubscriptionAssignerStub struct {
	inputs []AssignSubscriptionInput
}

func (s *legacyDefaultSubscriptionAssignerStub) AssignOrExtendSubscription(_ context.Context, input *AssignSubscriptionInput) (*UserSubscription, bool, error) {
	if input != nil {
		s.inputs = append(s.inputs, *input)
	}
	return &UserSubscription{ID: 7, UserID: input.UserID, GroupID: input.GroupID, Status: SubscriptionStatusActive}, false, nil
}

func TestProductAwareSubscriptionAssignerAssignsMappedLegacyGroupAsProduct(t *testing.T) {
	t.Parallel()

	productRepo := &productSubscriptionRepoStub{
		resolvedProduct: &SubscriptionProduct{ID: 88, Status: SubscriptionProductStatusActive},
		assigned: &UserProductSubscription{
			ID:        99,
			UserID:    42,
			ProductID: 88,
			Status:    SubscriptionStatusActive,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
	legacy := &legacyDefaultSubscriptionAssignerStub{}
	assigner := NewProductAwareSubscriptionAssigner(legacy, NewSubscriptionProductService(productRepo))

	sub, reused, err := assigner.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       42,
		GroupID:      21,
		ValidityDays: 30,
		AssignedBy:   5,
		Notes:        "default grant",
	})
	if err != nil {
		t.Fatalf("AssignOrExtendSubscription returned error: %v", err)
	}
	if reused {
		t.Fatal("reused = true, want false")
	}
	if sub.UserID != 42 || sub.GroupID != 21 || sub.ID != 99 {
		t.Fatalf("returned subscription = %+v, want synthetic product-backed user subscription", sub)
	}
	if len(productRepo.assignInputs) != 1 {
		t.Fatalf("product assign calls = %d, want 1", len(productRepo.assignInputs))
	}
	got := productRepo.assignInputs[0]
	if got.UserID != 42 || got.ProductID != 88 || got.ValidityDays != 30 || got.AssignedBy != 5 || got.Notes != "default grant" {
		t.Fatalf("product assignment input = %+v, want mapped product assignment", got)
	}
	if len(legacy.inputs) != 0 {
		t.Fatalf("legacy assign calls = %d, want 0", len(legacy.inputs))
	}
}

func TestProductAwareSubscriptionAssignerFallsBackForUnmappedLegacyGroup(t *testing.T) {
	t.Parallel()

	productRepo := &productSubscriptionRepoStub{}
	legacy := &legacyDefaultSubscriptionAssignerStub{}
	assigner := NewProductAwareSubscriptionAssigner(legacy, NewSubscriptionProductService(productRepo))

	sub, _, err := assigner.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       42,
		GroupID:      22,
		ValidityDays: 7,
	})
	if err != nil {
		t.Fatalf("AssignOrExtendSubscription returned error: %v", err)
	}
	if sub.GroupID != 22 {
		t.Fatalf("returned group id = %d, want legacy group", sub.GroupID)
	}
	if len(productRepo.assignInputs) != 0 {
		t.Fatalf("product assign calls = %d, want 0", len(productRepo.assignInputs))
	}
	if len(legacy.inputs) != 1 {
		t.Fatalf("legacy assign calls = %d, want 1", len(legacy.inputs))
	}
}
