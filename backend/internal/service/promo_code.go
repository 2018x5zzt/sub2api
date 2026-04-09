package service

import (
	"time"
)

// PromoCode 注册优惠码
type PromoCode struct {
	ID                    int64
	Code                  string
	Scene                 string
	BonusAmount           float64
	RandomBonusPoolAmount float64
	RandomBonusRemaining  float64
	MaxUses               int
	UsedCount             int
	LeaderboardEnabled    bool
	Status                string
	ExpiresAt             *time.Time
	SuccessMessage        string
	Notes                 string
	CreatedAt             time.Time
	UpdatedAt             time.Time

	// 关联
	UsageRecords []PromoCodeUsage
}

// PromoCodeUsage 优惠码使用记录
type PromoCodeUsage struct {
	ID                int64
	PromoCodeID       int64
	UserID            int64
	BonusAmount       float64
	FixedBonusAmount  float64
	RandomBonusAmount float64
	UsedAt            time.Time

	// 关联
	PromoCode *PromoCode
	User      *User
}

type BenefitCodeRedeemResult struct {
	PromoCode         *PromoCode
	Usage             *PromoCodeUsage
	FixedBonusAmount  float64
	RandomBonusAmount float64
	TotalBonusAmount  float64
}

type BenefitLeaderboardEntry struct {
	Rank              int
	DisplayName       string
	FixedBonusAmount  float64
	RandomBonusAmount float64
	TotalBonusAmount  float64
	UsedAt            time.Time
	IsCurrentUser     bool
}

type BenefitLeaderboardResult struct {
	PromoCode               *PromoCode
	Entries                 []BenefitLeaderboardEntry
	CurrentUserRank         *int
	CurrentUserRandomAmount *float64
	CurrentUserTotalAmount  *float64
}

// CanUse 检查优惠码是否可用
func (p *PromoCode) CanUse() bool {
	if p.Status != PromoCodeStatusActive {
		return false
	}
	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		return false
	}
	if p.MaxUses > 0 && p.UsedCount >= p.MaxUses {
		return false
	}
	return true
}

// IsExpired 检查是否已过期
func (p *PromoCode) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

// CreatePromoCodeInput 创建优惠码输入
type CreatePromoCodeInput struct {
	Code                  string
	Scene                 string
	BonusAmount           float64
	RandomBonusPoolAmount float64
	MaxUses               int
	LeaderboardEnabled    bool
	ExpiresAt             *time.Time
	SuccessMessage        string
	Notes                 string
}

// UpdatePromoCodeInput 更新优惠码输入
type UpdatePromoCodeInput struct {
	Code                  *string
	BonusAmount           *float64
	RandomBonusPoolAmount *float64
	MaxUses               *int
	LeaderboardEnabled    *bool
	Status                *string
	ExpiresAt             *time.Time
	SuccessMessage        *string
	Notes                 *string
}
