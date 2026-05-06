package service

import "time"

type AccountGroup struct {
	AccountID         int64
	GroupID           int64
	Priority          int
	BillingMultiplier float64
	CreatedAt         time.Time

	Account *Account
	Group   *Group
}

type AccountGroupBindingInput struct {
	GroupID           int64    `json:"group_id"`
	BillingMultiplier *float64 `json:"billing_multiplier,omitempty"`
}

func (in AccountGroupBindingInput) EffectiveBillingMultiplier() float64 {
	if in.BillingMultiplier == nil || *in.BillingMultiplier <= 0 {
		return 1.0
	}
	return *in.BillingMultiplier
}

func (ag *AccountGroup) EffectiveBillingMultiplier() float64 {
	if ag == nil || ag.BillingMultiplier <= 0 {
		return 1.0
	}
	return ag.BillingMultiplier
}
