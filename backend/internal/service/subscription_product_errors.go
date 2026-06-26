package service

import infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

var ErrProductSubscriptionAssignerUnavailable = infraerrors.ServiceUnavailable(
	"PRODUCT_SUBSCRIPTION_ASSIGNER_UNAVAILABLE",
	"product subscription assigner is not configured",
)
