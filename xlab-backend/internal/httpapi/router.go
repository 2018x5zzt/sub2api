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

func NewRouter(coreClient CoreClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	api := &API{core: coreClient}
	mux.Handle("/xapi/v1/subscription-products/active", api.auth(http.HandlerFunc(api.getActiveProducts)))
	mux.Handle("/xapi/v1/subscription-products/summary", api.auth(http.HandlerFunc(api.getProductSummary)))
	mux.Handle("/xapi/v1/subscription-products/progress", api.auth(http.HandlerFunc(api.getProductProgress)))
	return mux
}
