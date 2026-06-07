package httpapi

import (
	"context"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/core"
)

func (a *API) getActiveProducts(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Active)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/active")
}

func (a *API) getProductSummary(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Summary)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/summary")
}

func (a *API) getProductProgress(w http.ResponseWriter, r *http.Request) {
	if a.subscriptionReads != nil {
		a.writeReadServiceJSON(w, r, a.subscriptionReads.Progress)
		return
	}
	a.proxyCoreJSON(w, r, "/subscription-products/progress")
}

func (a *API) writeReadServiceJSON(w http.ResponseWriter, r *http.Request, fn func(context.Context, *core.User, string) (any, error)) {
	out, err := fn(r.Context(), userFromContext(r.Context()), tokenFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusBadGateway, "SUBSCRIPTION_READ_UNAVAILABLE", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) proxyCoreJSON(w http.ResponseWriter, r *http.Request, path string) {
	raw, err := a.core.ProxyGET(r.Context(), tokenFromContext(r.Context()), path)
	if err != nil {
		writeError(w, http.StatusBadGateway, "CORE_UNAVAILABLE", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, raw)
}
