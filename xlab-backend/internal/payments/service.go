package payments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

const (
	fallbackReasonCoreMode             = "core_mode"
	fallbackReasonStaleOrMissingSync   = "stale_or_missing_sync"
	fallbackReasonRepoError            = "repo_error"
	fallbackReasonHybridEmptyFirstPage = "hybrid_empty_first_page"
	fallbackReasonMissingCoreProxy     = "missing_core_proxy"
)

var errMissingCoreProxy = errors.New("payments service missing core proxy")

type Service struct {
	source     ReadSource
	repo       MirrorRepository
	core       CoreProxy
	staleAfter time.Duration
	now        func() time.Time
}

func NewService(source ReadSource, repo MirrorRepository, coreProxy CoreProxy, staleAfter time.Duration, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{source: source, repo: repo, core: coreProxy, staleAfter: staleAfter, now: now}
}

func (s *Service) ListOrders(ctx context.Context, user *core.User, token string, params ListParams) (any, error) {
	params = normalizeListParams(params)
	list, fallbackReason, err := s.mirrorListOrders(ctx, user, params)
	if err != nil || fallbackReason != "" {
		return s.proxyCore(ctx, token, listOrdersPath(params), fallbackReason)
	}
	return list, nil
}

func (s *Service) GetOrder(ctx context.Context, user *core.User, token string, orderID int64) (any, error) {
	order, fallbackReason, err := s.mirrorOrder(ctx, user, orderID)
	if err != nil || fallbackReason != "" {
		return s.proxyCore(ctx, token, fmt.Sprintf("/payment/orders/%d", orderID), fallbackReason)
	}
	return order, nil
}

func (s *Service) mirrorListOrders(ctx context.Context, user *core.User, params ListParams) (*OrderList, string, error) {
	if s.source == ReadSourceCore {
		return nil, fallbackReasonCoreMode, nil
	}
	if user == nil || s.repo == nil {
		return nil, fallbackReasonStaleOrMissingSync, nil
	}
	if !s.mirrorFresh(ctx) {
		return nil, fallbackReasonStaleOrMissingSync, nil
	}
	list, err := s.repo.ListOrdersByUser(ctx, user.ID, params)
	if err != nil {
		return nil, fallbackReasonRepoError, err
	}
	if s.source == ReadSourceHybrid && isUnfilteredFirstPage(params) && isEmptyOrderList(list) {
		return nil, fallbackReasonHybridEmptyFirstPage, nil
	}
	return list, "", nil
}

func (s *Service) mirrorOrder(ctx context.Context, user *core.User, orderID int64) (OrderSnapshot, string, error) {
	if s.source == ReadSourceCore {
		return nil, fallbackReasonCoreMode, nil
	}
	if user == nil || s.repo == nil {
		return nil, fallbackReasonStaleOrMissingSync, nil
	}
	if !s.mirrorFresh(ctx) {
		return nil, fallbackReasonStaleOrMissingSync, nil
	}
	order, err := s.repo.GetOrderByUser(ctx, orderID, user.ID)
	if err != nil {
		return nil, fallbackReasonRepoError, err
	}
	if order == nil {
		return nil, fallbackReasonRepoError, nil
	}
	return order, "", nil
}

func (s *Service) mirrorFresh(ctx context.Context) bool {
	state, err := s.repo.SyncState(ctx, syncSourcePaymentOrders)
	if err != nil || state == nil || state.LastSuccessAt == nil {
		return false
	}
	return s.now().Sub(*state.LastSuccessAt) <= s.staleAfter
}

func (s *Service) proxyCore(ctx context.Context, token string, path string, reason string) (any, error) {
	if s.core == nil {
		logFallback(ctx, fallbackReasonMissingCoreProxy)
		return nil, errMissingCoreProxy
	}
	logFallback(ctx, reason)
	return s.core.ProxyGET(ctx, token, path)
}

func logFallback(ctx context.Context, reason string) {
	slog.WarnContext(ctx, "payment_orders_core_fallback", "reason", reason)
}

func normalizeListParams(params ListParams) ListParams {
	page, pageSize := normalizePagination(params)
	params.Page = page
	params.PageSize = pageSize
	return params
}

func listOrdersPath(params ListParams) string {
	query := []string{
		"page=" + strconv.Itoa(params.Page),
		"page_size=" + strconv.Itoa(params.PageSize),
	}
	if params.Status != "" {
		query = append(query, "status="+url.QueryEscape(params.Status))
	}
	if params.OrderType != "" {
		query = append(query, "order_type="+url.QueryEscape(params.OrderType))
	}
	if params.PaymentType != "" {
		query = append(query, "payment_type="+url.QueryEscape(params.PaymentType))
	}
	return "/payment/orders/my?" + strings.Join(query, "&")
}

func isUnfilteredFirstPage(params ListParams) bool {
	return params.Page == 1 && params.Status == "" && params.OrderType == "" && params.PaymentType == ""
}

func isEmptyOrderList(list *OrderList) bool {
	return list == nil || len(list.Items) == 0
}
