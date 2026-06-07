package httpapi

import "net/http"

func (a *API) getActiveProducts(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/active")
}

func (a *API) getProductSummary(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/summary")
}

func (a *API) getProductProgress(w http.ResponseWriter, r *http.Request) {
	a.proxyCoreJSON(w, r, "/subscription-products/progress")
}

func (a *API) proxyCoreJSON(w http.ResponseWriter, r *http.Request, path string) {
	raw, err := a.core.ProxyGET(r.Context(), tokenFromContext(r.Context()), path)
	if err != nil {
		writeError(w, http.StatusBadGateway, "CORE_UNAVAILABLE", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, raw)
}
