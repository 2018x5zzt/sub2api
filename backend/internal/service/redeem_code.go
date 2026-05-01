package service

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type RedeemCode struct {
	ID         int64
	Code       string
	Type       string
	Value      float64
	Status     string
	SourceType string
	UsedBy     *int64
	UsedAt     *time.Time
	Notes      string
	CreatedAt  time.Time

	GroupID      *int64
	ValidityDays int

	User  *User
	Group *Group
}

func (r *RedeemCode) IsUsed() bool {
	return r.Status == StatusUsed
}

func (r *RedeemCode) CanUse() bool {
	return r.Status == StatusUnused
}

func GenerateRedeemCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func IsValidRedeemSourceType(sourceType string) bool {
	switch strings.TrimSpace(sourceType) {
	case RedeemSourceCommercial, RedeemSourceBenefit, RedeemSourceCompensation, RedeemSourceSystemGrant:
		return true
	default:
		return false
	}
}

func NormalizeRedeemSourceType(sourceType, fallback string) string {
	normalized := strings.TrimSpace(sourceType)
	if normalized == "" {
		normalized = strings.TrimSpace(fallback)
	}
	if normalized == "" {
		return ""
	}
	if !IsValidRedeemSourceType(normalized) {
		return strings.TrimSpace(fallback)
	}
	return normalized
}

func IsCommercialRechargeRedeem(code *RedeemCode) bool {
	if code == nil || NormalizeRedeemSourceType(code.SourceType, "") != RedeemSourceCommercial {
		return false
	}
	return code.Type == RedeemTypeBalance || code.Type == RedeemTypeSubscription
}
