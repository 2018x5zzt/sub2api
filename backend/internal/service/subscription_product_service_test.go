//go:build unit

package service

import (
	"context"
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
	assigned        *UserProductSubscription
	assignInputs    []AssignProductSubscriptionInput
	reused          bool
}

func (r *productSubscriptionRepoStub) GetActiveProductSubscriptionByUserAndGroupID(context.Context, int64, int64) (*SubscriptionProductBinding, *UserProductSubscription, error) {
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
