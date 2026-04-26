package service

import "context"

type ProductSubscriptionRepository interface {
	GetActiveProductBindingByGroupID(ctx context.Context, groupID int64) (*SubscriptionProductBinding, error)
	GetActiveUserProductSubscription(ctx context.Context, userID, productID int64) (*UserProductSubscription, error)
	ListVisibleGroupsByUserID(ctx context.Context, userID int64) ([]Group, error)
	ListActiveProductsByUserID(ctx context.Context, userID int64) ([]ActiveSubscriptionProduct, error)
}

type ProductSubscriptionByUserGroupRepository interface {
	GetActiveProductSubscriptionByUserAndGroupID(ctx context.Context, userID, groupID int64) (*SubscriptionProductBinding, *UserProductSubscription, error)
}
