package subscriptions

import (
	"context"
	"encoding/json"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type ReadSource string

const (
	ReadSourceCore   ReadSource = "core"
	ReadSourceHybrid ReadSource = "hybrid"
	ReadSourceXlab   ReadSource = "xlab"
)

type Group struct {
	GroupID                  int64   `json:"group_id"`
	GroupName                string  `json:"group_name"`
	GroupPlatform            string  `json:"group_platform"`
	BalanceFallbackGroupID   *int64  `json:"balance_fallback_group_id,omitempty"`
	BalanceFallbackGroupName *string `json:"balance_fallback_group_name,omitempty"`
	DebitMultiplier          float64 `json:"debit_multiplier"`
	Status                   string  `json:"status"`
	SortOrder                int     `json:"sort_order"`
}

type ActiveProduct struct {
	ProductID                  int64      `json:"product_id"`
	SubscriptionID             int64      `json:"subscription_id"`
	Code                       string     `json:"code"`
	Name                       string     `json:"name"`
	Description                string     `json:"description"`
	Status                     string     `json:"status"`
	ExpiresAt                  *time.Time `json:"expires_at,omitempty"`
	DailyUsageUSD              float64    `json:"daily_usage_usd"`
	WeeklyUsageUSD             float64    `json:"weekly_usage_usd"`
	MonthlyUsageUSD            float64    `json:"monthly_usage_usd"`
	DailyLimitUSD              float64    `json:"daily_limit_usd"`
	WeeklyLimitUSD             float64    `json:"weekly_limit_usd"`
	MonthlyLimitUSD            float64    `json:"monthly_limit_usd"`
	DailyCarryoverInUSD        float64    `json:"daily_carryover_in_usd"`
	DailyCarryoverRemainingUSD float64    `json:"daily_carryover_remaining_usd"`
	Groups                     []Group    `json:"groups"`
}

type Summary struct {
	ActiveCount          int             `json:"active_count"`
	TotalMonthlyUsageUSD float64         `json:"total_monthly_usage_usd"`
	TotalMonthlyLimitUSD float64         `json:"total_monthly_limit_usd"`
	Products             []ActiveProduct `json:"products"`
}

func SummaryFromActiveProducts(items []ActiveProduct) Summary {
	summary := Summary{ActiveCount: len(items), Products: items}
	for _, item := range items {
		summary.TotalMonthlyUsageUSD += item.MonthlyUsageUSD
		summary.TotalMonthlyLimitUSD += item.MonthlyLimitUSD
	}
	return summary
}

type SyncState struct {
	SourceName    string
	LastSuccessAt *time.Time
	RowCount      int
}

type CoreProxy interface {
	ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error)
}

type MirrorRepository interface {
	ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error)
	SyncState(ctx context.Context, sourceName string) (*SyncState, error)
}

type ReadService interface {
	Active(ctx context.Context, user *core.User, token string) (any, error)
	Summary(ctx context.Context, user *core.User, token string) (any, error)
	Progress(ctx context.Context, user *core.User, token string) (any, error)
}
