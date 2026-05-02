//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type redeemCreateRepoStub struct {
	created []RedeemCode
}

func (s *redeemCreateRepoStub) Create(_ context.Context, code *RedeemCode) error {
	if code != nil {
		s.created = append(s.created, *code)
	}
	return nil
}
func (s *redeemCreateRepoStub) CreateBatch(context.Context, []RedeemCode) error { panic("unexpected") }
func (s *redeemCreateRepoStub) GetByID(context.Context, int64) (*RedeemCode, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) GetByCode(context.Context, string) (*RedeemCode, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) Update(context.Context, *RedeemCode) error { panic("unexpected") }
func (s *redeemCreateRepoStub) Delete(context.Context, int64) error       { panic("unexpected") }
func (s *redeemCreateRepoStub) Use(context.Context, int64, int64) error   { panic("unexpected") }
func (s *redeemCreateRepoStub) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) ListByUser(context.Context, int64, int) ([]RedeemCode, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected")
}
func (s *redeemCreateRepoStub) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	panic("unexpected")
}

type redeemProductAwareAssignerStub struct {
	productID int64
	groupID   int64
	products  []SubscriptionProduct
}

func (s redeemProductAwareAssignerStub) AssignOrExtendSubscription(context.Context, *AssignSubscriptionInput) (*UserSubscription, bool, error) {
	panic("unexpected")
}
func (s redeemProductAwareAssignerStub) ResolveActiveProductIDByGroupID(_ context.Context, groupID int64) (int64, bool, error) {
	if groupID == s.groupID {
		return s.productID, true, nil
	}
	return 0, false, nil
}
func (s redeemProductAwareAssignerStub) ListProducts(context.Context) ([]SubscriptionProduct, error) {
	out := make([]SubscriptionProduct, len(s.products))
	copy(out, s.products)
	return out, nil
}

func TestAdminServiceGenerateSubscriptionCardCodesConvertsMappedLegacyGroupToProduct(t *testing.T) {
	t.Parallel()

	groupID := int64(21)
	repo := &redeemCreateRepoStub{}
	svc := &adminServiceImpl{
		redeemCodeRepo:     repo,
		groupRepo:          &subscriptionGroupRepoStub{group: &Group{ID: groupID, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}},
		defaultSubAssigner: redeemProductAwareAssignerStub{groupID: groupID, productID: 88},
	}

	codes, err := svc.GenerateRedeemCodes(context.Background(), &GenerateRedeemCodesInput{
		Count:        1,
		Type:         RedeemTypeSubscription,
		GroupID:      &groupID,
		ValidityDays: 30,
	})

	require.NoError(t, err)
	require.Len(t, codes, 1)
	require.Len(t, repo.created, 1)
	require.Nil(t, repo.created[0].GroupID)
	require.NotNil(t, repo.created[0].ProductID)
	require.Equal(t, int64(88), *repo.created[0].ProductID)
	require.Equal(t, 30, repo.created[0].ValidityDays)
}

func TestAdminServiceGenerateSubscriptionCardCodesRejectsUnknownProduct(t *testing.T) {
	t.Parallel()

	productID := int64(404)
	repo := &redeemCreateRepoStub{}
	svc := &adminServiceImpl{
		redeemCodeRepo: repo,
		defaultSubAssigner: redeemProductAwareAssignerStub{
			groupID:   21,
			productID: 88,
			products: []SubscriptionProduct{
				{ID: 88, Status: SubscriptionProductStatusActive},
			},
		},
	}

	codes, err := svc.GenerateRedeemCodes(context.Background(), &GenerateRedeemCodesInput{
		Count:        1,
		Type:         RedeemTypeSubscription,
		ProductID:    &productID,
		ValidityDays: 30,
	})

	require.Error(t, err)
	require.Nil(t, codes)
	require.Empty(t, repo.created)
}

func TestAdminServiceGenerateSubscriptionCardCodesStoresActiveProductID(t *testing.T) {
	t.Parallel()

	productID := int64(88)
	repo := &redeemCreateRepoStub{}
	svc := &adminServiceImpl{
		redeemCodeRepo: repo,
		defaultSubAssigner: redeemProductAwareAssignerStub{
			products: []SubscriptionProduct{
				{ID: productID, Status: SubscriptionProductStatusActive},
			},
		},
	}

	codes, err := svc.GenerateRedeemCodes(context.Background(), &GenerateRedeemCodesInput{
		Count:        1,
		Type:         RedeemTypeSubscription,
		ProductID:    &productID,
		ValidityDays: 30,
	})

	require.NoError(t, err)
	require.Len(t, codes, 1)
	require.Len(t, repo.created, 1)
	require.Nil(t, repo.created[0].GroupID)
	require.NotNil(t, repo.created[0].ProductID)
	require.Equal(t, productID, *repo.created[0].ProductID)
}
