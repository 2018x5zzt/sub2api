package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviterewardrecord "github.com/Wei-Shaw/sub2api/ent/inviterewardrecord"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type InviteUserRepository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByInviteCode(ctx context.Context, code string) (*User, error)
	ExistsByInviteCode(ctx context.Context, code string) (bool, error)
	CountInviteesByInviter(ctx context.Context, inviterID int64) (int64, error)
	UpdateBalance(ctx context.Context, id int64, amount float64) error
}

type InviteRewardRecordRepository interface {
	CreateBatch(ctx context.Context, records []InviteRewardRecord) error
	ListByRewardTarget(ctx context.Context, userID int64, params pagination.PaginationParams) ([]InviteRewardRecord, *pagination.PaginationResult, error)
	ListByAdminActionID(ctx context.Context, adminActionID int64) ([]InviteRewardRecord, error)
	SumBaseRewardsByTargetAndRole(ctx context.Context, userID int64, rewardRole string) (float64, error)
}

type InviteService struct {
	userRepo           InviteUserRepository
	rewardRepo         InviteRewardRecordRepository
	settingService     *SettingService
	entClient          *dbent.Client
	registerURLBase    string
	codeGenerator      func() (string, error)
	baseRewardExistsFn func(ctx context.Context, redeemCodeID int64) (bool, error)
}

func NewInviteService(userRepo InviteUserRepository, rewardRepo InviteRewardRecordRepository, entClient ...*dbent.Client) *InviteService {
	var txClient *dbent.Client
	if len(entClient) > 0 {
		txClient = entClient[0]
	}
	return &InviteService{
		userRepo:        userRepo,
		rewardRepo:      rewardRepo,
		entClient:       txClient,
		registerURLBase: "/register",
		codeGenerator:   defaultInviteCodeGenerator,
	}
}

func (s *InviteService) withInviteWriteTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil || s.entClient == nil {
		return fn(ctx)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (s *InviteService) existsBaseRewardByRedeemCodeID(ctx context.Context, redeemCodeID int64) (bool, error) {
	if redeemCodeID <= 0 {
		return false, nil
	}
	if s.baseRewardExistsFn != nil {
		return s.baseRewardExistsFn(ctx, redeemCodeID)
	}

	var client *dbent.Client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	} else {
		client = s.entClient
	}
	if client == nil {
		return false, nil
	}

	count, err := client.InviteRewardRecord.Query().
		Where(
			dbinviterewardrecord.TriggerRedeemCodeIDEQ(redeemCodeID),
			dbinviterewardrecord.RewardTypeEQ(InviteRewardTypeBase),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func defaultInviteCodeGenerator() (string, error) {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(raw)), nil
}

func (s *InviteService) GenerateUniqueInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		code, err := s.codeGenerator()
		if err != nil {
			return "", err
		}
		exists, err := s.userRepo.ExistsByInviteCode(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", infraerrors.ServiceUnavailable("INVITE_CODE_GENERATION_FAILED", "failed to generate unique invite code")
}

func (s *InviteService) ResolveInviterByCode(ctx context.Context, code string) (*User, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return nil, ErrUserNotFound
	}
	inviter, err := s.userRepo.GetByInviteCode(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if inviter == nil || !inviter.IsActive() {
		return nil, ErrUserNotFound
	}
	return inviter, nil
}

func (s *InviteService) resolveRegisterURLBase(ctx context.Context) string {
	if s.settingService != nil {
		if frontendURL := strings.TrimSpace(s.settingService.GetFrontendURL(ctx)); frontendURL != "" {
			return strings.TrimRight(frontendURL, "/") + "/register"
		}
	}

	base := strings.TrimSpace(s.registerURLBase)
	if base == "" {
		return "/register"
	}
	return base
}

func (s *InviteService) buildSummaryForUser(ctx context.Context, user *User, invitedUsersTotal int64, inviteesRechargeTotal, baseRewardsTotal float64) *InviteSummary {
	return &InviteSummary{
		InviteCode:            user.InviteCode,
		InviteLink:            s.resolveRegisterURLBase(ctx) + "?invite=" + url.QueryEscape(user.InviteCode),
		InvitedUsersTotal:     invitedUsersTotal,
		InviteesRechargeTotal: inviteesRechargeTotal,
		BaseRewardsTotal:      baseRewardsTotal,
	}
}

func (s *InviteService) GetSummary(ctx context.Context, userID int64) (*InviteSummary, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	invitedUsersTotal, err := s.userRepo.CountInviteesByInviter(ctx, userID)
	if err != nil {
		return nil, err
	}

	baseRewardsTotal, err := s.rewardRepo.SumBaseRewardsByTargetAndRole(ctx, userID, InviteRewardRoleInviter)
	if err != nil {
		return nil, err
	}

	inviteesRechargeTotal := 0.0
	if InviteBaseRewardRate > 0 {
		inviteesRechargeTotal = baseRewardsTotal / InviteBaseRewardRate
	}

	return s.buildSummaryForUser(ctx, user, invitedUsersTotal, inviteesRechargeTotal, baseRewardsTotal), nil
}

func (s *InviteService) ListRewards(ctx context.Context, userID int64, page, pageSize int) ([]InviteRewardRecord, int64, error) {
	records, result, err := s.rewardRepo.ListByRewardTarget(ctx, userID, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	return records, result.Total, nil
}

func (s *InviteService) ApplyBaseRechargeRewards(ctx context.Context, inviteeID int64, redeemCode *RedeemCode) error {
	if redeemCode == nil || redeemCode.Type != RedeemTypeBalance {
		return nil
	}

	sourceType := NormalizeRedeemSourceType(redeemCode.SourceType, "")
	if sourceType != RedeemSourceCommercial {
		return nil
	}
	exists, err := s.existsBaseRewardByRedeemCodeID(ctx, redeemCode.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	invitee, err := s.userRepo.GetByID(ctx, inviteeID)
	if err != nil {
		return err
	}
	if invitee.InvitedByUserID == nil || *invitee.InvitedByUserID == inviteeID {
		return nil
	}

	inviterID := *invitee.InvitedByUserID
	inviterRewardRate := InviteBaseRewardRate
	inviteeRewardRate := InviteBaseRewardRate
	inviterRewardAmount := redeemCode.Value * inviterRewardRate
	inviteeRewardAmount := redeemCode.Value * inviteeRewardRate

	records := []InviteRewardRecord{
		{
			InviterUserID:          inviterID,
			InviteeUserID:          inviteeID,
			TriggerRedeemCodeID:    &redeemCode.ID,
			TriggerRedeemCodeValue: redeemCode.Value,
			RewardTargetUserID:     inviterID,
			RewardRole:             InviteRewardRoleInviter,
			RewardType:             InviteRewardTypeBase,
			RewardRate:             &inviterRewardRate,
			RewardAmount:           inviterRewardAmount,
			Status:                 "applied",
		},
		{
			InviterUserID:          inviterID,
			InviteeUserID:          inviteeID,
			TriggerRedeemCodeID:    &redeemCode.ID,
			TriggerRedeemCodeValue: redeemCode.Value,
			RewardTargetUserID:     inviteeID,
			RewardRole:             InviteRewardRoleInvitee,
			RewardType:             InviteRewardTypeBase,
			RewardRate:             &inviteeRewardRate,
			RewardAmount:           inviteeRewardAmount,
			Status:                 "applied",
		},
	}

	return s.withInviteWriteTx(ctx, func(txCtx context.Context) error {
		if err := s.rewardRepo.CreateBatch(txCtx, records); err != nil {
			return err
		}

		if err := s.userRepo.UpdateBalance(txCtx, inviterID, inviterRewardAmount); err != nil {
			return err
		}
		return s.userRepo.UpdateBalance(txCtx, inviteeID, inviteeRewardAmount)
	})
}
