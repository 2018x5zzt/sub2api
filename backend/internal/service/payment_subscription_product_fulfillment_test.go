//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

type captureSubscriptionAssigner struct {
	calls []AssignSubscriptionInput
}

func (a *captureSubscriptionAssigner) AssignOrExtendSubscription(_ context.Context, input *AssignSubscriptionInput) (*UserSubscription, bool, error) {
	if input != nil {
		a.calls = append(a.calls, *input)
	}
	return &UserSubscription{
		UserID:  input.UserID,
		GroupID: input.GroupID,
	}, false, nil
}

func TestExecuteSubscriptionFulfillmentPassesExplicitProductID(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("product-plan-buyer@example.com").
		SetPasswordHash("hash").
		SetUsername("product-plan-buyer").
		Save(ctx)
	require.NoError(t, err)

	groupID := int64(45)
	productID := int64(225)
	days := 225
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(225).
		SetPayAmount(225).
		SetFeeRate(0).
		SetRechargeCode("PAY-PRODUCT-225").
		SetOutTradeNo("sub2_product_plan_225").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-product-225").
		SetOrderType(payment.OrderTypeSubscription).
		SetPlanID(1001).
		SetSubscriptionGroupID(groupID).
		SetSubscriptionProductID(productID).
		SetSubscriptionDays(days).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	assigner := &captureSubscriptionAssigner{}
	svc := &PaymentService{
		entClient: client,
		groupRepo: &subscriptionGroupRepoStub{group: &Group{
			ID:               groupID,
			Status:           payment.EntityStatusActive,
			SubscriptionType: SubscriptionTypeSubscription,
		}},
		subscriptionAssigner: assigner,
	}

	require.NoError(t, svc.ExecuteSubscriptionFulfillment(ctx, order.ID))
	require.Len(t, assigner.calls, 1)
	require.Equal(t, user.ID, assigner.calls[0].UserID)
	require.Equal(t, groupID, assigner.calls[0].GroupID)
	require.Equal(t, productID, assigner.calls[0].ProductID)
	require.Equal(t, days, assigner.calls[0].ValidityDays)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
}

func TestCreateSubscriptionOrderSnapshotsPlanProductID(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)

	entUser, err := client.User.Create().
		SetEmail("snapshot-product-buyer@example.com").
		SetPasswordHash("hash").
		SetUsername("snapshot-product-buyer").
		Save(ctx)
	require.NoError(t, err)

	const (
		groupID   = int64(45)
		productID = int64(225)
	)
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetProductID(productID).
		SetName("225 Day Card").
		SetDescription("225 day product subscription").
		SetPrice(225).
		SetValidityDays(225).
		SetValidityUnit("days").
		SetFeatures("").
		SetProductName("225 Day Card").
		SetForSale(true).
		SetSortOrder(20).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	order, err := svc.createOrderInTx(ctx, CreateOrderRequest{
		UserID:      entUser.ID,
		Amount:      plan.Price,
		PaymentType: payment.TypeAlipay,
		ClientIP:    "127.0.0.1",
		SrcHost:     "api.example.com",
		OrderType:   payment.OrderTypeSubscription,
		PlanID:      plan.ID,
	}, &User{
		ID:       entUser.ID,
		Email:    entUser.Email,
		Username: entUser.Username,
	}, plan, &PaymentConfig{
		OrderTimeoutMin:  30,
		MaxPendingOrders: 3,
	}, plan.Price, plan.Price, 0, plan.Price, nil)
	require.NoError(t, err)

	require.NotNil(t, order.SubscriptionGroupID)
	require.Equal(t, groupID, *order.SubscriptionGroupID)
	require.NotNil(t, order.SubscriptionProductID)
	require.Equal(t, productID, *order.SubscriptionProductID)
	require.NotNil(t, order.SubscriptionDays)
	require.Equal(t, 225, *order.SubscriptionDays)
}
