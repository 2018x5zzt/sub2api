package handler

import (
	"errors"
	"strings"
	"time"

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
	Message            string   `json:"message"`
	Type               string   `json:"type"`
	Value              float64  `json:"value"`
	FixedValue         float64  `json:"fixed_value,omitempty"`
	RandomValue        float64  `json:"random_value,omitempty"`
	TotalValue         float64  `json:"total_value,omitempty"`
	Scene              string   `json:"scene,omitempty"`
	SuccessMessage     string   `json:"success_message,omitempty"`
	LeaderboardEnabled bool     `json:"leaderboard_enabled,omitempty"`
	NewBalance         *float64 `json:"new_balance,omitempty"`
	NewConcurrency     *int     `json:"new_concurrency,omitempty"`
	GroupName          string   `json:"group_name,omitempty"`
	ValidityDays       *int     `json:"validity_days,omitempty"`
}

type BenefitLeaderboardResponse struct {
	Code                   string                            `json:"code"`
	FixedValue             float64                           `json:"fixed_value"`
	RandomPoolValue        float64                           `json:"random_pool_value"`
	RandomRemainingValue   float64                           `json:"random_remaining_value"`
	MaxUses                int                               `json:"max_uses"`
	UsedCount              int                               `json:"used_count"`
	Entries                []BenefitLeaderboardEntryResponse `json:"entries"`
	CurrentUserRank        *int                              `json:"current_user_rank,omitempty"`
	CurrentUserRandomValue *float64                          `json:"current_user_random_value,omitempty"`
	CurrentUserTotalValue  *float64                          `json:"current_user_total_value,omitempty"`
}

type BenefitLeaderboardEntryResponse struct {
	Rank          int     `json:"rank"`
	DisplayName   string  `json:"display_name"`
	FixedValue    float64 `json:"fixed_value"`
	RandomValue   float64 `json:"random_value"`
	TotalValue    float64 `json:"total_value"`
	UsedAt        string  `json:"used_at"`
	IsCurrentUser bool    `json:"is_current_user"`
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
		benefitResult, benefitErr := h.promoService.RedeemBenefitCode(ctx, subject.UserID, code)
		switch {
		case benefitErr == nil && benefitResult != nil:
			response.Success(c, promoRedeemResponseFromBenefitResult(benefitResult))
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

// GetBenefitLeaderboard returns the lucky leaderboard for a redeemed benefit red packet code.
// POST /api/v1/redeem/benefit-leaderboard
func (h *RedeemHandler) GetBenefitLeaderboard(c *gin.Context) {
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

	code := strings.TrimSpace(req.Code)
	if code == "" {
		response.BadRequest(c, "Invalid request: code is required")
		return
	}

	result, err := h.promoService.GetBenefitLeaderboard(c.Request.Context(), subject.UserID, code, 0)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, benefitLeaderboardResponseFromResult(result))
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
		FixedValue:     code.BonusAmount,
		TotalValue:     code.BonusAmount,
		Scene:          scene,
		SuccessMessage: code.SuccessMessage,
	}
}

func promoRedeemResponseFromBenefitResult(result *service.BenefitCodeRedeemResult) *RedeemResponse {
	if result == nil || result.PromoCode == nil {
		return nil
	}

	scene := strings.TrimSpace(result.PromoCode.Scene)
	if scene == "" {
		scene = service.PromoCodeSceneBenefit
	}

	return &RedeemResponse{
		Message:            "Promo code redeemed successfully",
		Type:               service.RedeemTypeBalance,
		Value:              result.TotalBonusAmount,
		FixedValue:         result.FixedBonusAmount,
		RandomValue:        result.RandomBonusAmount,
		TotalValue:         result.TotalBonusAmount,
		Scene:              scene,
		SuccessMessage:     result.PromoCode.SuccessMessage,
		LeaderboardEnabled: result.PromoCode.LeaderboardEnabled && result.PromoCode.RandomBonusPoolAmount > 0,
	}
}

func benefitLeaderboardResponseFromResult(result *service.BenefitLeaderboardResult) *BenefitLeaderboardResponse {
	if result == nil || result.PromoCode == nil {
		return nil
	}

	entries := make([]BenefitLeaderboardEntryResponse, 0, len(result.Entries))
	for i := range result.Entries {
		entry := result.Entries[i]
		entries = append(entries, BenefitLeaderboardEntryResponse{
			Rank:          entry.Rank,
			DisplayName:   entry.DisplayName,
			FixedValue:    entry.FixedBonusAmount,
			RandomValue:   entry.RandomBonusAmount,
			TotalValue:    entry.TotalBonusAmount,
			UsedAt:        entry.UsedAt.UTC().Format(time.RFC3339),
			IsCurrentUser: entry.IsCurrentUser,
		})
	}

	return &BenefitLeaderboardResponse{
		Code:                   result.PromoCode.Code,
		FixedValue:             result.PromoCode.BonusAmount,
		RandomPoolValue:        result.PromoCode.RandomBonusPoolAmount,
		RandomRemainingValue:   result.PromoCode.RandomBonusRemaining,
		MaxUses:                result.PromoCode.MaxUses,
		UsedCount:              result.PromoCode.UsedCount,
		Entries:                entries,
		CurrentUserRank:        result.CurrentUserRank,
		CurrentUserRandomValue: result.CurrentUserRandomAmount,
		CurrentUserTotalValue:  result.CurrentUserTotalAmount,
	}
}
