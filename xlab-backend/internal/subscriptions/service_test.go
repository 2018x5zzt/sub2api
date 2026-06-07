package subscriptions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type fakeMirrorRepo struct {
	products []ActiveProduct
	state    *SyncState
	err      error
}

func (f *fakeMirrorRepo) ListActiveProductsByUser(ctx context.Context, userID int64) ([]ActiveProduct, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f *fakeMirrorRepo) SyncState(ctx context.Context, sourceName string) (*SyncState, error) {
	if f.state == nil {
		return nil, sql.ErrNoRows
	}
	return f.state, nil
}

type fakeCoreProxy struct {
	path string
}

func (f *fakeCoreProxy) ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error) {
	f.path = path
	return json.RawMessage(`[{"subscription_id":44,"name":"Core"}]`), nil
}

func TestServiceCoreModeUsesCoreProxy(t *testing.T) {
	coreProxy := &fakeCoreProxy{}
	svc := NewService(ReadSourceCore, &fakeMirrorRepo{}, coreProxy, 10*time.Minute, time.Now)
	out, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
	if _, ok := out.(json.RawMessage); !ok {
		t.Fatalf("output type = %T", out)
	}
}

func TestServiceHybridFreshMirrorUsesRepository(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakeMirrorRepo{
		state:    &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 1},
		products: []ActiveProduct{{ProductID: 10, SubscriptionID: 99, Name: "Mirror", Groups: []Group{}}},
	}
	svc := NewService(ReadSourceHybrid, repo, &fakeCoreProxy{}, 10*time.Minute, func() time.Time { return now })
	out, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	items, ok := out.([]ActiveProduct)
	if !ok || len(items) != 1 || items[0].Name != "Mirror" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestServiceHybridStaleMirrorFallsBackToCore(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	old := now.Add(-time.Hour)
	coreProxy := &fakeCoreProxy{}
	repo := &fakeMirrorRepo{state: &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &old, RowCount: 1}}
	svc := NewService(ReadSourceHybrid, repo, coreProxy, 10*time.Minute, func() time.Time { return now })
	_, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active error: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceXlabFallsBackWhenRepositoryFails(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	coreProxy := &fakeCoreProxy{}
	repo := &fakeMirrorRepo{state: &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 1}, err: errors.New("db down")}
	svc := NewService(ReadSourceXlab, repo, coreProxy, 10*time.Minute, func() time.Time { return now })
	_, err := svc.Active(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Active should fallback instead of failing: %v", err)
	}
	if coreProxy.path != "/subscription-products/active" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceSummaryFallbackUsesCoreSummaryPath(t *testing.T) {
	coreProxy := &fakeCoreProxy{}
	svc := NewService(ReadSourceCore, &fakeMirrorRepo{}, coreProxy, 10*time.Minute, time.Now)
	if _, err := svc.Summary(context.Background(), &core.User{ID: 7}, "tok"); err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	if coreProxy.path != "/subscription-products/summary" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceProgressFallbackUsesCoreProgressPath(t *testing.T) {
	coreProxy := &fakeCoreProxy{}
	svc := NewService(ReadSourceCore, &fakeMirrorRepo{}, coreProxy, 10*time.Minute, time.Now)
	if _, err := svc.Progress(context.Background(), &core.User{ID: 7}, "tok"); err != nil {
		t.Fatalf("Progress error: %v", err)
	}
	if coreProxy.path != "/subscription-products/progress" {
		t.Fatalf("core path = %s", coreProxy.path)
	}
}

func TestServiceSummaryAggregatesFreshMirrorProducts(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakeMirrorRepo{
		state: &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 2},
		products: []ActiveProduct{
			{ProductID: 10, SubscriptionID: 99, Name: "Mirror A", MonthlyUsageUSD: 1.5, MonthlyLimitUSD: 10, Groups: []Group{}},
			{ProductID: 11, SubscriptionID: 100, Name: "Mirror B", MonthlyUsageUSD: 2.5, MonthlyLimitUSD: 20, Groups: []Group{}},
		},
	}
	svc := NewService(ReadSourceXlab, repo, &fakeCoreProxy{}, 10*time.Minute, func() time.Time { return now })
	out, err := svc.Summary(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	summary, ok := out.(Summary)
	if !ok {
		t.Fatalf("output type = %T", out)
	}
	if summary.ActiveCount != 2 || summary.TotalMonthlyUsageUSD != 4 || summary.TotalMonthlyLimitUSD != 30 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestServiceProgressUsesSummaryBehavior(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 0, 0, 0, time.UTC)
	repo := &fakeMirrorRepo{
		state:    &SyncState{SourceName: "product_subscriptions", LastSuccessAt: &now, RowCount: 1},
		products: []ActiveProduct{{ProductID: 10, SubscriptionID: 99, Name: "Mirror", MonthlyUsageUSD: 1.5, MonthlyLimitUSD: 10, Groups: []Group{}}},
	}
	svc := NewService(ReadSourceXlab, repo, &fakeCoreProxy{}, 10*time.Minute, func() time.Time { return now })
	out, err := svc.Progress(context.Background(), &core.User{ID: 7}, "tok")
	if err != nil {
		t.Fatalf("Progress error: %v", err)
	}
	summary, ok := out.(Summary)
	if !ok || summary.ActiveCount != 1 || summary.Products[0].Name != "Mirror" {
		t.Fatalf("unexpected progress output: %#v", out)
	}
}
