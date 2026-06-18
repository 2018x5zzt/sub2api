package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/2018x5zzt/xlab-backend/internal/payments"
)

const paymentOrderRoutePrefix = "/xapi/v1/payment/orders/"

func (a *API) listPaymentOrders(w http.ResponseWriter, r *http.Request) {
	params := paymentListParams(r)
	if a.paymentReads != nil {
		out, err := a.paymentReads.ListOrders(r.Context(), userFromContext(r.Context()), tokenFromContext(r.Context()), params)
		if err != nil {
			writeError(w, http.StatusBadGateway, "PAYMENT_READ_UNAVAILABLE", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	a.proxyCoreJSON(w, r, paymentOrdersListPath(params))
}

func (a *API) getPaymentOrder(w http.ResponseWriter, r *http.Request) {
	orderID, ok := paymentOrderID(r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ORDER_ID", "Payment order id must be a positive integer")
		return
	}
	if a.paymentReads != nil {
		out, err := a.paymentReads.GetOrder(r.Context(), userFromContext(r.Context()), tokenFromContext(r.Context()), orderID)
		if err != nil {
			writeError(w, http.StatusBadGateway, "PAYMENT_READ_UNAVAILABLE", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	a.proxyCoreJSON(w, r, fmt.Sprintf("/payment/orders/%d", orderID))
}

func paymentListParams(r *http.Request) payments.ListParams {
	query := r.URL.Query()
	return payments.ListParams{
		Page:        positiveIntParam(query.Get("page"), 1, 0),
		PageSize:    positiveIntParam(query.Get("page_size"), 20, 100),
		Status:      query.Get("status"),
		OrderType:   query.Get("order_type"),
		PaymentType: query.Get("payment_type"),
	}
}

func positiveIntParam(raw string, defaultValue int, maxValue int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		value = defaultValue
	}
	if maxValue > 0 && value > maxValue {
		value = maxValue
	}
	return value
}

func paymentOrdersListPath(params payments.ListParams) string {
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

func paymentOrderID(path string) (int64, bool) {
	rawID := strings.TrimPrefix(path, paymentOrderRoutePrefix)
	if rawID == "" || strings.Contains(rawID, "/") {
		return 0, false
	}
	orderID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || orderID <= 0 {
		return 0, false
	}
	return orderID, true
}
