package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

type CoreClient interface {
	CurrentUser(ctx context.Context, token string) (*core.User, error)
	ProxyGET(ctx context.Context, token string, path string) (json.RawMessage, error)
}

type SubscriptionReadService interface {
	Active(ctx context.Context, user *core.User, token string) (any, error)
	Summary(ctx context.Context, user *core.User, token string) (any, error)
	Progress(ctx context.Context, user *core.User, token string) (any, error)
}

func NewRouter(coreClient CoreClient, readService SubscriptionReadService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	api := &API{core: coreClient, subscriptionReads: readService}
	mux.Handle("/xapi/v1/subscription-products/active", api.auth(http.HandlerFunc(api.getActiveProducts)))
	mux.Handle("/xapi/v1/subscription-products/summary", api.auth(http.HandlerFunc(api.getProductSummary)))
	mux.Handle("/xapi/v1/subscription-products/progress", api.auth(http.HandlerFunc(api.getProductProgress)))
	return mux
}
