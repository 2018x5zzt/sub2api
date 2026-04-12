package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type InviteSummary struct {
	InviteCode            string  `json:"invite_code"`
	InviteLink            string  `json:"invite_link"`
	InvitedUsersTotal     int64   `json:"invited_users_total"`
	InviteesRechargeTotal float64 `json:"invitees_recharge_total"`
	BaseRewardsTotal      float64 `json:"base_rewards_total"`
}

type InviteRewardRecord struct {
	RewardRole   string    `json:"reward_role"`
	RewardType   string    `json:"reward_type"`
	RewardAmount float64   `json:"reward_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

func InviteSummaryFromService(in *service.InviteSummary) *InviteSummary {
	if in == nil {
		return nil
	}
	return &InviteSummary{
		InviteCode:            in.InviteCode,
		InviteLink:            in.InviteLink,
		InvitedUsersTotal:     in.InvitedUsersTotal,
		InviteesRechargeTotal: in.InviteesRechargeTotal,
		BaseRewardsTotal:      in.BaseRewardsTotal,
	}
}

func InviteRewardFromService(in *service.InviteRewardRecord) *InviteRewardRecord {
	if in == nil {
		return nil
	}
	return &InviteRewardRecord{
		RewardRole:   in.RewardRole,
		RewardType:   in.RewardType,
		RewardAmount: in.RewardAmount,
		CreatedAt:    in.CreatedAt,
	}
}
