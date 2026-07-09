//go:build unit

package service

import (
	"errors"
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type productAwareProductRepoStub struct {
	getActiveProductSubscriptionByUserAndGroupID func(ctx context.Context, userID, groupID int64, productFamily *string) (*SubscriptionProductBinding, *UserProductSubscription, error)
	listVisibleGroupsByUserID                    func(ctx context.Context, userID int64) ([]Group, error)
	resolveActiveProductByGroupID                func(ctx context.Context, groupID int64) (*SubscriptionProduct, error)
	assignProductSubscription                    func(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error)
}

func (r *productAwareProductRepoStub) GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64, productFamily *string) (*SubscriptionProductBinding, *UserProductSubscription, error) {
	if r.getActiveProductSubscriptionByUserAndGroupID != nil {
		return r.getActiveProductSubscriptionByUserAndGroupID(ctx, userID, groupID, productFamily)
	}
	return nil, nil, ErrSubscriptionNotFound
}

func (r *productAwareProductRepoStub) ListActiveProductsByUserID(context.Context, int64) ([]ActiveSubscriptionProduct, error) {
	return nil, nil
}

func (r *productAwareProductRepoStub) ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]Group, error) {
	if r.listVisibleGroupsByUserID != nil {
		return r.listVisibleGroupsByUserID(ctx, userID)
	}
	return nil, nil
}

func (r *productAwareProductRepoStub) ListProducts(context.Context) ([]SubscriptionProduct, error) {
	return nil, nil
}

func (r *productAwareProductRepoStub) ResolveActiveProductByGroupID(ctx context.Context, groupID int64) (*SubscriptionProduct, error) {
	if r.resolveActiveProductByGroupID != nil {
		return r.resolveActiveProductByGroupID(ctx, groupID)
	}
	return nil, ErrSubscriptionNotFound
}
func (r *productAwareProductRepoStub) GetProductByID(context.Context, int64) (*SubscriptionProduct, error) {
	return nil, errors.New("not implemented")
}


func (r *productAwareProductRepoStub) CreateProduct(context.Context, *CreateSubscriptionProductInput) (*SubscriptionProduct, error) {
	return nil, ErrProductSubscriptionAssignerUnavailable
}

func (r *productAwareProductRepoStub) UpdateProduct(context.Context, int64, *UpdateSubscriptionProductInput) (*SubscriptionProduct, error) {
	return nil, ErrProductSubscriptionAssignerUnavailable
}

func (r *productAwareProductRepoStub) ListProductBindings(context.Context, int64) ([]SubscriptionProductBindingDetail, error) {
	return nil, nil
}

func (r *productAwareProductRepoStub) SyncProductBindings(context.Context, int64, []SubscriptionProductBindingInput) ([]SubscriptionProductBindingDetail, error) {
	return nil, nil
}

func (r *productAwareProductRepoStub) ListProductSubscriptions(context.Context, int64) ([]UserProductSubscription, error) {
	return nil, nil
}

func (r *productAwareProductRepoStub) ListUserProductSubscriptionsForAdmin(context.Context, AdminProductSubscriptionListParams) ([]AdminProductSubscriptionListItem, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{Page: 1, PageSize: 20, Pages: 1}, nil
}

func (r *productAwareProductRepoStub) AssignOrExtendProductSubscription(ctx context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if r.assignProductSubscription != nil {
		return r.assignProductSubscription(ctx, input)
	}
	return nil, false, ErrProductSubscriptionAssignerUnavailable
}

func (r *productAwareProductRepoStub) AdjustProductSubscription(context.Context, int64, *AdjustProductSubscriptionInput) (*UserProductSubscription, error) {
	return nil, ErrProductSubscriptionAssignerUnavailable
}

func (r *productAwareProductRepoStub) ResetProductSubscriptionQuota(context.Context, int64, *ResetProductSubscriptionQuotaInput) (*UserProductSubscription, error) {
	return nil, ErrProductSubscriptionAssignerUnavailable
}

func (r *productAwareProductRepoStub) RevokeProductSubscription(context.Context, int64) error {
	return ErrProductSubscriptionAssignerUnavailable
}

func TestProductAwareSubscriptionAssignerRejectsLegacyFallbackWhenGroupHasNoProduct(t *testing.T) {
	assigner := NewProductAwareSubscriptionAssigner(NewSubscriptionProductService(&productAwareProductRepoStub{
		resolveActiveProductByGroupID: func(context.Context, int64) (*SubscriptionProduct, error) {
			return nil, ErrSubscriptionNotFound
		},
	}))

	_, _, err := assigner.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       1001,
		GroupID:      7,
		ValidityDays: 30,
		Notes:        "payment order 42",
	})

	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}
