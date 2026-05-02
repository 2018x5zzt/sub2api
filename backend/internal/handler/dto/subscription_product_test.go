package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAdminSubscriptionProductFromServiceIncludesProductFamily(t *testing.T) {
	product := &service.SubscriptionProduct{
		ID:            88,
		Code:          "gpt_daily_45",
		Name:          "GPT Daily 45",
		ProductFamily: "gpt_shared",
		Status:        service.SubscriptionProductStatusActive,
	}

	out := AdminSubscriptionProductFromService(product)

	require.NotNil(t, out)
	require.Equal(t, "gpt_shared", out.ProductFamily)
}
