package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type subscriptionProductAdminService interface {
	ListProducts(ctx context.Context) ([]service.SubscriptionProduct, error)
	CreateProduct(ctx context.Context, input *service.CreateSubscriptionProductInput) (*service.SubscriptionProduct, error)
	UpdateProduct(ctx context.Context, productID int64, input *service.UpdateSubscriptionProductInput) (*service.SubscriptionProduct, error)
	ListProductBindings(ctx context.Context, productID int64) ([]service.SubscriptionProductBindingDetail, error)
	SyncProductBindings(ctx context.Context, productID int64, inputs []service.SubscriptionProductBindingInput) ([]service.SubscriptionProductBindingDetail, error)
	ListProductSubscriptions(ctx context.Context, productID int64) ([]service.UserProductSubscription, error)
	ListUserProductSubscriptionsForAdmin(ctx context.Context, params service.AdminProductSubscriptionListParams) ([]service.AdminProductSubscriptionListItem, *pagination.PaginationResult, error)
	AssignOrExtendProductSubscription(ctx context.Context, input *service.AssignProductSubscriptionInput) (*service.UserProductSubscription, bool, error)
	AdjustProductSubscription(ctx context.Context, subscriptionID int64, input *service.AdjustProductSubscriptionInput) (*service.UserProductSubscription, error)
	ResetProductSubscriptionQuota(ctx context.Context, subscriptionID int64, input *service.ResetProductSubscriptionQuotaInput) (*service.UserProductSubscription, error)
	RevokeProductSubscription(ctx context.Context, subscriptionID int64) error
}

type SubscriptionProductHandler struct {
	subscriptionProductService subscriptionProductAdminService
}

func NewSubscriptionProductHandler(subscriptionProductService *service.SubscriptionProductService) *SubscriptionProductHandler {
	return &SubscriptionProductHandler{subscriptionProductService: subscriptionProductService}
}

type createSubscriptionProductRequest struct {
	Code                string  `json:"code" binding:"required,max=64"`
	Name                string  `json:"name" binding:"required,max=255"`
	Description         string  `json:"description"`
	Status              string  `json:"status" binding:"omitempty,oneof=draft active disabled"`
	ProductFamily       string  `json:"product_family" binding:"omitempty,max=64"`
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
	ProductFamily       *string  `json:"product_family" binding:"omitempty,max=64"`
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

type assignProductSubscriptionRequest struct {
	UserID       int64  `json:"user_id" binding:"required,gt=0"`
	ValidityDays int    `json:"validity_days" binding:"omitempty,min=1,max=36500"`
	Notes        string `json:"notes"`
}

type adjustProductSubscriptionRequest struct {
	ExpiresAt *string `json:"expires_at"`
	Status    *string `json:"status" binding:"omitempty,oneof=active expired revoked"`
	Notes     *string `json:"notes"`
}

type resetProductSubscriptionQuotaRequest struct {
	Daily   bool `json:"daily"`
	Weekly  bool `json:"weekly"`
	Monthly bool `json:"monthly"`
}

func (h *SubscriptionProductHandler) List(c *gin.Context) {
	products, err := h.subscriptionProductService.ListProducts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductsFromService(products))
}

func (h *SubscriptionProductHandler) ListAllSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	productID, _ := strconv.ParseInt(c.DefaultQuery("product_id", "0"), 10, 64)
	userID, _ := strconv.ParseInt(c.DefaultQuery("user_id", "0"), 10, 64)

	items, pageResult, err := h.subscriptionProductService.ListUserProductSubscriptionsForAdmin(c.Request.Context(), service.AdminProductSubscriptionListParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    c.Query("search"),
		ProductID: productID,
		UserID:    userID,
		Status:    c.Query("status"),
		SortBy:    c.DefaultQuery("sort_by", "expires_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.AdminProductSubscriptionListItemsFromService(items), pageResult.Total, pageResult.Page, pageResult.PageSize)
}

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
		ProductFamily:       req.ProductFamily,
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
		ProductFamily:       req.ProductFamily,
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

func (h *SubscriptionProductHandler) ListBindings(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid subscription product ID")
		return
	}

	bindings, err := h.subscriptionProductService.ListProductBindings(c.Request.Context(), productID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminSubscriptionProductBindingsFromService(bindings))
}

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

func (h *SubscriptionProductHandler) Assign(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid subscription product ID")
		return
	}

	var req assignProductSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	subscription, reused, err := h.subscriptionProductService.AssignOrExtendProductSubscription(c.Request.Context(), &service.AssignProductSubscriptionInput{
		UserID:       req.UserID,
		ProductID:    productID,
		ValidityDays: req.ValidityDays,
		Notes:        req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"subscription": dto.AdminUserProductSubscriptionsFromService([]service.UserProductSubscription{*subscription})[0],
		"reused":       reused,
	})
}

func (h *SubscriptionProductHandler) AdjustSubscription(c *gin.Context) {
	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid product subscription ID")
		return
	}
	var req adjustProductSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			parsedDate, dateErr := time.Parse("2006-01-02", *req.ExpiresAt)
			if dateErr != nil {
				response.BadRequest(c, "Invalid expires_at format")
				return
			}
			parsed = parsedDate
		}
		expiresAt = &parsed
	}
	subscription, err := h.subscriptionProductService.AdjustProductSubscription(c.Request.Context(), subscriptionID, &service.AdjustProductSubscriptionInput{
		ExpiresAt: expiresAt,
		Status:    req.Status,
		Notes:     req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminUserProductSubscriptionsFromService([]service.UserProductSubscription{*subscription})[0])
}

func (h *SubscriptionProductHandler) ResetSubscriptionQuota(c *gin.Context) {
	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid product subscription ID")
		return
	}
	var req resetProductSubscriptionQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subscription, err := h.subscriptionProductService.ResetProductSubscriptionQuota(c.Request.Context(), subscriptionID, &service.ResetProductSubscriptionQuotaInput{
		Daily:   req.Daily,
		Weekly:  req.Weekly,
		Monthly: req.Monthly,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminUserProductSubscriptionsFromService([]service.UserProductSubscription{*subscription})[0])
}

func (h *SubscriptionProductHandler) RevokeSubscription(c *gin.Context) {
	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid product subscription ID")
		return
	}
	if err := h.subscriptionProductService.RevokeProductSubscription(c.Request.Context(), subscriptionID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Product subscription revoked"})
}
