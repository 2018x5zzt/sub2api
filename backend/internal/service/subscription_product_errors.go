package service

import (
	"net/http"
	"strconv"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func NewSubscriptionProductNotFoundError(groupID int64) *infraerrors.ApplicationError {
	err := infraerrors.New(http.StatusNotFound, "SUBSCRIPTION_PRODUCT_NOT_FOUND", "subscription product not found")
	if groupID > 0 {
		return err.WithMetadata(map[string]string{
			"group_id": strconv.FormatInt(groupID, 10),
		})
	}
	return err
}

func NewSubscriptionProductInactiveError(binding *SubscriptionProductBinding) *infraerrors.ApplicationError {
	return infraerrors.New(http.StatusForbidden, "SUBSCRIPTION_PRODUCT_INACTIVE", "subscription product is inactive").
		WithMetadata(productBindingMetadata(binding))
}

func NewProductSubscriptionInvalidError(binding *SubscriptionProductBinding, sub *UserProductSubscription) *infraerrors.ApplicationError {
	md := productBindingMetadata(binding)
	if sub != nil {
		md["product_subscription_id"] = strconv.FormatInt(sub.ID, 10)
		md["user_id"] = strconv.FormatInt(sub.UserID, 10)
	}
	return infraerrors.New(http.StatusForbidden, "PRODUCT_SUBSCRIPTION_INVALID", "product subscription is invalid").
		WithMetadata(md)
}

func NewProductLimitExceededError(binding *SubscriptionProductBinding, remainingBudget float64) *infraerrors.ApplicationError {
	md := productBindingMetadata(binding)
	md["remaining_budget"] = strconv.FormatFloat(remainingBudget, 'f', 6, 64)
	return infraerrors.New(http.StatusForbidden, "PRODUCT_LIMIT_EXCEEDED", "product subscription limit exceeded").
		WithMetadata(md)
}

func NewProductBindingInactiveError(binding *SubscriptionProductBinding) *infraerrors.ApplicationError {
	return infraerrors.New(http.StatusForbidden, "PRODUCT_GROUP_BINDING_INACTIVE", "product group binding is inactive").
		WithMetadata(productBindingMetadata(binding))
}

func NewProductMigrationConflictError(metadata map[string]string) *infraerrors.ApplicationError {
	return infraerrors.New(http.StatusConflict, "PRODUCT_MIGRATION_CONFLICT", "product subscription migration conflict").
		WithMetadata(metadata)
}

func productBindingMetadata(binding *SubscriptionProductBinding) map[string]string {
	md := map[string]string{}
	if binding == nil {
		return md
	}
	if binding.ProductID > 0 {
		md["product_id"] = strconv.FormatInt(binding.ProductID, 10)
	}
	if binding.ProductName != "" {
		md["product_name"] = binding.ProductName
	}
	if binding.GroupID > 0 {
		md["group_id"] = strconv.FormatInt(binding.GroupID, 10)
	}
	if binding.GroupName != "" {
		md["group_name"] = binding.GroupName
	}
	if binding.DebitMultiplier > 0 {
		md["debit_multiplier"] = strconv.FormatFloat(binding.DebitMultiplier, 'f', -1, 64)
	}
	return md
}
