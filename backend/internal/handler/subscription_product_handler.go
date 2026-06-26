package handler

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type subscriptionProductUserService interface {
	ListActiveUserProducts(ctx context.Context, userID int64) ([]service.ActiveSubscriptionProduct, error)
	GetUserProductSummary(ctx context.Context, userID int64) (*service.SubscriptionProductSummary, error)
	GetUserProductProgress(ctx context.Context, userID int64) (*service.SubscriptionProductSummary, error)
}

type SubscriptionProductHandler struct {
	subscriptionProductService subscriptionProductUserService
}

func NewSubscriptionProductHandler(subscriptionProductService *service.SubscriptionProductService) *SubscriptionProductHandler {
	return &SubscriptionProductHandler{subscriptionProductService: subscriptionProductService}
}

func (h *SubscriptionProductHandler) GetActive(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	products, err := h.subscriptionProductService.ListActiveUserProducts(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.ActiveSubscriptionProductsFromService(products))
}

func (h *SubscriptionProductHandler) GetSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	summary, err := h.subscriptionProductService.GetUserProductSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionProductSummaryFromService(summary))
}

func (h *SubscriptionProductHandler) GetProgress(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	progress, err := h.subscriptionProductService.GetUserProductProgress(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionProductSummaryFromService(progress))
}
