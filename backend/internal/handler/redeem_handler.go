package handler

import (
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RedeemHandler handles redeem code-related requests
type RedeemHandler struct {
	redeemService  *service.RedeemService
	promoService   *service.PromoService
	settingService *service.SettingService
}

// NewRedeemHandler creates a new RedeemHandler
func NewRedeemHandler(redeemService *service.RedeemService, promoService *service.PromoService, settingService *service.SettingService) *RedeemHandler {
	return &RedeemHandler{
		redeemService:  redeemService,
		promoService:   promoService,
		settingService: settingService,
	}
}

// RedeemRequest represents the redeem code request payload
type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

// RedeemResponse represents the redeem response
type RedeemResponse struct {
	Message        string   `json:"message"`
	Type           string   `json:"type"`
	Value          float64  `json:"value"`
	Scene          string   `json:"scene,omitempty"`
	SuccessMessage string   `json:"success_message,omitempty"`
	NewBalance     *float64 `json:"new_balance,omitempty"`
	NewConcurrency *int     `json:"new_concurrency,omitempty"`
	GroupName      string   `json:"group_name,omitempty"`
	ValidityDays   *int     `json:"validity_days,omitempty"`
}

// Redeem handles redeeming a code
// POST /api/v1/redeem
func (h *RedeemHandler) Redeem(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	code := strings.TrimSpace(req.Code)
	if code == "" {
		response.BadRequest(c, "Invalid request: code is required")
		return
	}

	_, err := h.redeemService.GetByCode(ctx, code)
	switch {
	case err == nil:
		result, redeemErr := h.redeemService.Redeem(ctx, subject.UserID, code)
		if redeemErr != nil {
			response.ErrorFrom(c, redeemErr)
			return
		}
		response.Success(c, redeemResponseFromCode(result))
		return
	case !errors.Is(err, service.ErrRedeemCodeNotFound):
		response.ErrorFrom(c, err)
		return
	}

	if h.promoService != nil {
		benefitCode, benefitErr := h.promoService.ValidateBenefitCode(ctx, code)
		switch {
		case benefitErr == nil && benefitCode != nil:
			if applyErr := h.promoService.ApplyBenefitCode(ctx, subject.UserID, code); applyErr != nil {
				response.ErrorFrom(c, applyErr)
				return
			}
			response.Success(c, promoRedeemResponseFromCode(benefitCode))
			return
		case benefitErr != nil && !errors.Is(benefitErr, service.ErrPromoCodeNotFound):
			response.ErrorFrom(c, benefitErr)
			return
		}
	}

	if h.promoService != nil && h.settingService != nil && h.settingService.IsPromoCodeEnabled(ctx) {
		promoCode, promoErr := h.promoService.ValidatePromoCode(ctx, code)
		switch {
		case promoErr == nil && promoCode != nil:
			if applyErr := h.promoService.ApplyPromoCode(ctx, subject.UserID, code); applyErr != nil {
				response.ErrorFrom(c, applyErr)
				return
			}
			response.Success(c, promoRedeemResponseFromCode(promoCode))
			return
		case promoErr != nil && !errors.Is(promoErr, service.ErrPromoCodeNotFound):
			response.ErrorFrom(c, promoErr)
			return
		}
	}

	response.ErrorFrom(c, service.ErrRedeemCodeNotFound)
}

// GetHistory returns the user's redemption history
// GET /api/v1/redeem/history
func (h *RedeemHandler) GetHistory(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Default limit is 25
	limit := 25

	codes, err := h.redeemService.GetUserHistory(c.Request.Context(), subject.UserID, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.RedeemCode, 0, len(codes))
	for i := range codes {
		out = append(out, *dto.RedeemCodeFromService(&codes[i]))
	}
	response.Success(c, out)
}

func redeemResponseFromCode(code *service.RedeemCode) *RedeemResponse {
	if code == nil {
		return nil
	}

	resp := &RedeemResponse{
		Message: "Code redeemed successfully",
		Type:    code.Type,
		Value:   code.Value,
	}

	if code.Group != nil {
		resp.GroupName = code.Group.Name
	}
	if code.Type == service.RedeemTypeSubscription && code.ValidityDays > 0 {
		validityDays := code.ValidityDays
		resp.ValidityDays = &validityDays
	}

	return resp
}

func promoRedeemResponseFromCode(code *service.PromoCode) *RedeemResponse {
	if code == nil {
		return nil
	}

	scene := strings.TrimSpace(code.Scene)
	if scene == "" {
		scene = service.PromoCodeSceneRegister
	}

	return &RedeemResponse{
		Message:        "Promo code redeemed successfully",
		Type:           service.RedeemTypeBalance,
		Value:          code.BonusAmount,
		Scene:          scene,
		SuccessMessage: code.SuccessMessage,
	}
}
