package admin

import (
	"context"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type subscriptionProductAdminService interface {
	ListProducts(ctx context.Context) ([]service.SubscriptionProduct, error)
	CreateProduct(ctx context.Context, input *service.CreateSubscriptionProductInput) (*service.SubscriptionProduct, error)
	UpdateProduct(ctx context.Context, productID int64, input *service.UpdateSubscriptionProductInput) (*service.SubscriptionProduct, error)
	SyncProductBindings(ctx context.Context, productID int64, inputs []service.SubscriptionProductBindingInput) ([]service.SubscriptionProductBindingDetail, error)
	ListProductSubscriptions(ctx context.Context, productID int64) ([]service.UserProductSubscription, error)
}

type SubscriptionProductHandler struct {
	subscriptionProductService subscriptionProductAdminService
}

func NewSubscriptionProductHandler(subscriptionProductService subscriptionProductAdminService) *SubscriptionProductHandler {
	return &SubscriptionProductHandler{subscriptionProductService: subscriptionProductService}
}

func ProvideSubscriptionProductHandler(subscriptionProductService *service.SubscriptionProductAdminService) *SubscriptionProductHandler {
	return NewSubscriptionProductHandler(subscriptionProductService)
}

type createSubscriptionProductRequest struct {
	Code                string  `json:"code" binding:"required,max=64"`
	Name                string  `json:"name" binding:"required,max=255"`
	Description         string  `json:"description"`
	Status              string  `json:"status" binding:"omitempty,oneof=draft active disabled"`
	DefaultValidityDays int     `json:"default_validity_days" binding:"omitempty,min=1,max=36500"`
	DailyLimitUSD       float64 `json:"daily_limit_usd" binding:"omitempty,min=0"`
	WeeklyLimitUSD      float64 `json:"weekly_limit_usd" binding:"omitempty,min=0"`
	MonthlyLimitUSD     float64 `json:"monthly_limit_usd" binding:"omitempty,min=0"`
	SortOrder           int     `json:"sort_order"`
}

type updateSubscriptionProductRequest struct {
	Code                *string  `json:"code" binding:"omitempty,max=64"`
	Name                *string  `json:"name" binding:"omitempty,max=255"`
	Description         *string  `json:"description"`
	Status              *string  `json:"status" binding:"omitempty,oneof=draft active disabled"`
	DefaultValidityDays *int     `json:"default_validity_days" binding:"omitempty,min=1,max=36500"`
	DailyLimitUSD       *float64 `json:"daily_limit_usd" binding:"omitempty,min=0"`
	WeeklyLimitUSD      *float64 `json:"weekly_limit_usd" binding:"omitempty,min=0"`
	MonthlyLimitUSD     *float64 `json:"monthly_limit_usd" binding:"omitempty,min=0"`
	SortOrder           *int     `json:"sort_order"`
}

type syncSubscriptionProductBindingsRequest struct {
	Bindings []subscriptionProductBindingRequest `json:"bindings"`
}

type subscriptionProductBindingRequest struct {
	GroupID         int64   `json:"group_id" binding:"required"`
	DebitMultiplier float64 `json:"debit_multiplier" binding:"omitempty,min=0"`
	Status          string  `json:"status" binding:"omitempty,oneof=active inactive"`
	SortOrder       int     `json:"sort_order"`
}

// List handles listing subscription products.
// GET /api/v1/admin/subscription-products
func (h *SubscriptionProductHandler) List(c *gin.Context) {
	products, err := h.subscriptionProductService.ListProducts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductsFromService(products))
}

// Create handles creating a subscription product.
// POST /api/v1/admin/subscription-products
func (h *SubscriptionProductHandler) Create(c *gin.Context) {
	var req createSubscriptionProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	product, err := h.subscriptionProductService.CreateProduct(c.Request.Context(), &service.CreateSubscriptionProductInput{
		Code:                req.Code,
		Name:                req.Name,
		Description:         req.Description,
		Status:              req.Status,
		DefaultValidityDays: req.DefaultValidityDays,
		DailyLimitUSD:       req.DailyLimitUSD,
		WeeklyLimitUSD:      req.WeeklyLimitUSD,
		MonthlyLimitUSD:     req.MonthlyLimitUSD,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductFromService(product))
}

// Update handles updating a subscription product.
// PUT /api/v1/admin/subscription-products/:id
func (h *SubscriptionProductHandler) Update(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid subscription product ID")
		return
	}

	var req updateSubscriptionProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	product, err := h.subscriptionProductService.UpdateProduct(c.Request.Context(), productID, &service.UpdateSubscriptionProductInput{
		Code:                req.Code,
		Name:                req.Name,
		Description:         req.Description,
		Status:              req.Status,
		DefaultValidityDays: req.DefaultValidityDays,
		DailyLimitUSD:       req.DailyLimitUSD,
		WeeklyLimitUSD:      req.WeeklyLimitUSD,
		MonthlyLimitUSD:     req.MonthlyLimitUSD,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductFromService(product))
}

// SyncBindings replaces the desired group bindings for a subscription product.
// PUT /api/v1/admin/subscription-products/:id/bindings
func (h *SubscriptionProductHandler) SyncBindings(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid subscription product ID")
		return
	}

	var req syncSubscriptionProductBindingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	inputs := make([]service.SubscriptionProductBindingInput, 0, len(req.Bindings))
	for _, binding := range req.Bindings {
		inputs = append(inputs, service.SubscriptionProductBindingInput{
			GroupID:         binding.GroupID,
			DebitMultiplier: binding.DebitMultiplier,
			Status:          binding.Status,
			SortOrder:       binding.SortOrder,
		})
	}

	bindings, err := h.subscriptionProductService.SyncProductBindings(c.Request.Context(), productID, inputs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductBindingsFromService(bindings))
}

// ListSubscriptions handles listing user product subscriptions by product.
// GET /api/v1/admin/subscription-products/:id/subscriptions
func (h *SubscriptionProductHandler) ListSubscriptions(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid subscription product ID")
		return
	}

	subscriptions, err := h.subscriptionProductService.ListProductSubscriptions(c.Request.Context(), productID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminUserProductSubscriptionsFromService(subscriptions))
}
