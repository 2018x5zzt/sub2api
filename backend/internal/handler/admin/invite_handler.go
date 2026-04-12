package admin

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	adminService service.AdminService
}

type rebindRequest struct {
	InviteeUserID    int64  `json:"invitee_user_id" binding:"required,gt=0"`
	NewInviterUserID int64  `json:"new_inviter_user_id" binding:"required,gt=0"`
	Reason           string `json:"reason" binding:"required"`
}

type manualInviteGrantLineRequest struct {
	InviterUserID      int64   `json:"inviter_user_id" binding:"required,gt=0"`
	InviteeUserID      int64   `json:"invitee_user_id" binding:"required,gt=0"`
	RewardTargetUserID int64   `json:"reward_target_user_id" binding:"required,gt=0"`
	RewardRole         string  `json:"reward_role" binding:"required,oneof=inviter invitee"`
	RewardAmount       float64 `json:"reward_amount" binding:"required,gt=0"`
	Notes              string  `json:"notes"`
}

type manualInviteGrantRequest struct {
	TargetUserID int64                          `json:"target_user_id" binding:"required,gt=0"`
	Reason       string                         `json:"reason" binding:"required"`
	Lines        []manualInviteGrantLineRequest `json:"lines" binding:"required,min=1,dive"`
}

type recomputeRequest struct {
	Reason        string     `json:"reason" binding:"required"`
	InviteeUserID *int64     `json:"invitee_user_id"`
	InviterUserID *int64     `json:"inviter_user_id"`
	StartAt       *time.Time `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
}

type recomputeExecuteRequest struct {
	Reason        string     `json:"reason" binding:"required"`
	InviteeUserID *int64     `json:"invitee_user_id"`
	InviterUserID *int64     `json:"inviter_user_id"`
	StartAt       *time.Time `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
	ScopeHash     string     `json:"scope_hash" binding:"required"`
}

func NewInviteHandler(adminService service.AdminService) *InviteHandler {
	return &InviteHandler{adminService: adminService}
}

func (h *InviteHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetInviteStats(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminInviteStatsFromService(stats))
}

func (h *InviteHandler) ListRelationships(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	filters, err := dto.AdminInviteRelationshipFiltersFromRequest(c)
	if err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	rows, total, err := h.adminService.ListInviteRelationships(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.AdminInviteRelationshipsFromService(rows), total, page, pageSize)
}

func (h *InviteHandler) ListRewards(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	filters, err := dto.AdminInviteRewardFiltersFromRequest(c)
	if err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	rows, total, err := h.adminService.ListInviteRewards(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.AdminInviteRewardsFromService(rows), total, page, pageSize)
}

func (h *InviteHandler) ListActions(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	filters, err := dto.InviteAdminActionFiltersFromRequest(c)
	if err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	rows, total, err := h.adminService.ListInviteActions(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.AdminInviteActionsFromService(rows), total, page, pageSize)
}

func (h *InviteHandler) Rebind(c *gin.Context) {
	var req rebindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	executeAdminIdempotentJSON(c, "admin.invites.rebind", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		err := h.adminService.RebindInviter(ctx, service.RebindInviterInput{
			OperatorUserID:   subject.UserID,
			InviteeUserID:    req.InviteeUserID,
			NewInviterUserID: req.NewInviterUserID,
			Reason:           strings.TrimSpace(req.Reason),
		})
		if err != nil {
			return nil, err
		}
		return gin.H{"message": "invite relationship updated; historical rewards unchanged"}, nil
	})
}

func (h *InviteHandler) CreateManualGrant(c *gin.Context) {
	var req manualInviteGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	executeAdminIdempotentJSON(c, "admin.invites.manual_grants.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		lines := make([]service.ManualInviteGrantLine, 0, len(req.Lines))
		for i := range req.Lines {
			line := req.Lines[i]
			lines = append(lines, service.ManualInviteGrantLine{
				InviterUserID:      line.InviterUserID,
				InviteeUserID:      line.InviteeUserID,
				RewardTargetUserID: line.RewardTargetUserID,
				RewardRole:         strings.TrimSpace(line.RewardRole),
				RewardAmount:       line.RewardAmount,
				Notes:              strings.TrimSpace(line.Notes),
			})
		}
		err := h.adminService.CreateManualInviteGrant(ctx, service.ManualInviteGrantInput{
			OperatorUserID: subject.UserID,
			TargetUserID:   req.TargetUserID,
			Reason:         strings.TrimSpace(req.Reason),
			Lines:          lines,
		})
		if err != nil {
			return nil, err
		}
		return gin.H{"message": "manual invite rewards granted"}, nil
	})
}

func (h *InviteHandler) PreviewRecompute(c *gin.Context) {
	var req recomputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	preview, err := h.adminService.PreviewInviteRecompute(c.Request.Context(), service.InviteRecomputeInput{
		OperatorUserID: subject.UserID,
		Reason:         strings.TrimSpace(req.Reason),
		Scope: service.InviteRecomputeScope{
			InviteeUserID: req.InviteeUserID,
			InviterUserID: req.InviterUserID,
			StartAt:       req.StartAt,
			EndAt:         req.EndAt,
		},
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.AdminInviteRecomputePreviewFromService(preview))
}

func (h *InviteHandler) ExecuteRecompute(c *gin.Context) {
	var req recomputeExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	executeAdminIdempotentJSON(c, "admin.invites.recompute.execute", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		err := h.adminService.ExecuteInviteRecompute(ctx, service.InviteRecomputeExecuteInput{
			OperatorUserID: subject.UserID,
			Reason:         strings.TrimSpace(req.Reason),
			Scope: service.InviteRecomputeScope{
				InviteeUserID: req.InviteeUserID,
				InviterUserID: req.InviterUserID,
				StartAt:       req.StartAt,
				EndAt:         req.EndAt,
			},
			ScopeHash: strings.TrimSpace(req.ScopeHash),
		})
		if err != nil {
			return nil, err
		}
		return gin.H{"message": "invite recompute applied"}, nil
	})
}
