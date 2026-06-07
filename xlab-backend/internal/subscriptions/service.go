package subscriptions

import (
	"context"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

const syncSourceProductSubscriptions = "product_subscriptions"

type Service struct {
	source     ReadSource
	repo       MirrorRepository
	core       CoreProxy
	staleAfter time.Duration
	now        func() time.Time
}

func NewService(source ReadSource, repo MirrorRepository, core CoreProxy, staleAfter time.Duration, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{source: source, repo: repo, core: core, staleAfter: staleAfter, now: now}
}

func (s *Service) Active(ctx context.Context, user *core.User, token string) (any, error) {
	items, ok, err := s.mirrorActiveProducts(ctx, user)
	if err != nil || !ok {
		return s.core.ProxyGET(ctx, token, "/subscription-products/active")
	}
	return items, nil
}

func (s *Service) Summary(ctx context.Context, user *core.User, token string) (any, error) {
	items, ok, err := s.mirrorActiveProducts(ctx, user)
	if err != nil || !ok {
		return s.core.ProxyGET(ctx, token, "/subscription-products/summary")
	}
	return SummaryFromActiveProducts(items), nil
}

func (s *Service) Progress(ctx context.Context, user *core.User, token string) (any, error) {
	items, ok, err := s.mirrorActiveProducts(ctx, user)
	if err != nil || !ok {
		return s.core.ProxyGET(ctx, token, "/subscription-products/progress")
	}
	return SummaryFromActiveProducts(items), nil
}

func (s *Service) mirrorActiveProducts(ctx context.Context, user *core.User) ([]ActiveProduct, bool, error) {
	if s.source == ReadSourceCore || user == nil || s.repo == nil {
		return nil, false, nil
	}
	if !s.mirrorFresh(ctx) {
		return nil, false, nil
	}
	items, err := s.repo.ListActiveProductsByUser(ctx, user.ID)
	if err != nil {
		return nil, false, err
	}
	if s.source == ReadSourceHybrid && len(items) == 0 {
		return nil, false, nil
	}
	return items, true, nil
}

func (s *Service) mirrorFresh(ctx context.Context) bool {
	state, err := s.repo.SyncState(ctx, syncSourceProductSubscriptions)
	if err != nil || state == nil || state.LastSuccessAt == nil {
		return false
	}
	return s.now().Sub(*state.LastSuccessAt) <= s.staleAfter
}
