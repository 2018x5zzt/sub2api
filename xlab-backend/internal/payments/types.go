package payments

import (
	"context"
	"encoding/json"
	"time"
)

type ReadSource string

const (
	ReadSourceCore   ReadSource = "core"
	ReadSourceHybrid ReadSource = "hybrid"
	ReadSourceXlab   ReadSource = "xlab"
)

const syncSourcePaymentOrders = "payment_orders"

type ListParams struct {
	Page        int
	PageSize    int
	Status      string
	OrderType   string
	PaymentType string
}

type OrderSnapshot map[string]any

type OrderList struct {
	Items    []OrderSnapshot `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Pages    int             `json:"pages"`
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
	ListOrdersByUser(ctx context.Context, userID int64, params ListParams) (*OrderList, error)
	GetOrderByUser(ctx context.Context, orderID int64, userID int64) (OrderSnapshot, error)
	SyncState(ctx context.Context, sourceName string) (*SyncState, error)
}
