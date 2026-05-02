//go:build unit

package service

import (
	"context"
	"testing"
	"time"
)

type redeemProductAssignerStub struct {
	inputs []AssignProductSubscriptionInput
}

func (s *redeemProductAssignerStub) AssignOrExtendProductSubscription(_ context.Context, input *AssignProductSubscriptionInput) (*UserProductSubscription, bool, error) {
	if input != nil {
		s.inputs = append(s.inputs, *input)
	}
	return &UserProductSubscription{
		ID:        99,
		UserID:    input.UserID,
		ProductID: input.ProductID,
		Status:    SubscriptionStatusActive,
		ExpiresAt: time.Now().AddDate(0, 0, input.ValidityDays),
	}, false, nil
}

func TestRedeemServiceAssignProductSubscriptionFromRedeem(t *testing.T) {
	productID := int64(88)
	assigner := &redeemProductAssignerStub{}
	svc := &RedeemService{productSubAssigner: assigner}

	err := svc.assignProductSubscriptionFromRedeem(context.Background(), 42, &RedeemCode{
		Code:         "PRODUCT-30",
		ProductID:    &productID,
		ValidityDays: 30,
	})
	if err != nil {
		t.Fatalf("assignProductSubscriptionFromRedeem returned error: %v", err)
	}
	if len(assigner.inputs) != 1 {
		t.Fatalf("assign calls = %d, want 1", len(assigner.inputs))
	}
	got := assigner.inputs[0]
	if got.UserID != 42 || got.ProductID != productID || got.ValidityDays != 30 {
		t.Fatalf("assignment input = %+v, want user/product/validity", got)
	}
	if got.Notes == "" {
		t.Fatal("assignment notes is empty")
	}
}

func TestValidateSubscriptionRedeemCodeAllowsHistoricalProductCodeWithLegacyGroupID(t *testing.T) {
	productID := int64(88)
	groupID := int64(21)

	err := validateSubscriptionRedeemCodeShape(&RedeemCode{
		Type:      RedeemTypeSubscription,
		ProductID: &productID,
		GroupID:   &groupID,
	})

	if err != nil {
		t.Fatalf("validateSubscriptionRedeemCodeShape returned error: %v", err)
	}
}

func TestRedeemServiceCreateSubscriptionCardCodeConvertsMappedLegacyGroupToProduct(t *testing.T) {
	groupID := int64(21)
	repo := &redeemCreateRepoStub{}
	svc := &RedeemService{
		redeemRepo:           repo,
		subscriptionAssigner: redeemProductAwareAssignerStub{groupID: groupID, productID: 88},
	}

	err := svc.CreateCode(context.Background(), &RedeemCode{
		Code:         "GROUP-COMPAT",
		Type:         RedeemTypeSubscription,
		Value:        1,
		Status:       StatusUnused,
		GroupID:      &groupID,
		ValidityDays: 30,
	})
	if err != nil {
		t.Fatalf("CreateCode returned error: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created codes = %d, want 1", len(repo.created))
	}
	got := repo.created[0]
	if got.GroupID != nil {
		t.Fatalf("GroupID = %v, want nil after product conversion", *got.GroupID)
	}
	if got.ProductID == nil || *got.ProductID != 88 {
		t.Fatalf("ProductID = %v, want 88", got.ProductID)
	}
}
