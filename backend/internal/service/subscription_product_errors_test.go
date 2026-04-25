package service

import (
	"net/http"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionProductError_ProductLimitExceededIncludesMetadata(t *testing.T) {
	binding := &SubscriptionProductBinding{
		ProductID:       101,
		ProductName:     "GPT Monthly",
		GroupID:         202,
		GroupName:       "pro",
		DebitMultiplier: 1.5,
	}

	err := NewProductLimitExceededError(binding, 3.25)

	appErr := infraerrors.FromError(err)
	require.Equal(t, int32(http.StatusForbidden), appErr.Code)
	require.Equal(t, "PRODUCT_LIMIT_EXCEEDED", appErr.Reason)
	require.Equal(t, "101", appErr.Metadata["product_id"])
	require.Equal(t, "GPT Monthly", appErr.Metadata["product_name"])
	require.Equal(t, "202", appErr.Metadata["group_id"])
	require.Equal(t, "pro", appErr.Metadata["group_name"])
	require.Equal(t, "1.5", appErr.Metadata["debit_multiplier"])
	require.Equal(t, "3.250000", appErr.Metadata["remaining_budget"])
}

func TestSubscriptionProductError_ProductBindingInactiveIncludesMetadata(t *testing.T) {
	binding := &SubscriptionProductBinding{
		ProductID:       101,
		ProductName:     "GPT Monthly",
		GroupID:         202,
		GroupName:       "plus",
		DebitMultiplier: 1,
	}

	err := NewProductBindingInactiveError(binding)

	appErr := infraerrors.FromError(err)
	require.Equal(t, int32(http.StatusForbidden), appErr.Code)
	require.Equal(t, "PRODUCT_GROUP_BINDING_INACTIVE", appErr.Reason)
	require.Equal(t, "101", appErr.Metadata["product_id"])
	require.Equal(t, "GPT Monthly", appErr.Metadata["product_name"])
	require.Equal(t, "202", appErr.Metadata["group_id"])
	require.Equal(t, "plus", appErr.Metadata["group_name"])
	require.Equal(t, "1", appErr.Metadata["debit_multiplier"])
}
