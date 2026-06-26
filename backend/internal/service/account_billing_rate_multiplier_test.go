package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_BillingRateMultiplier_DefaultsToOneWhenNil(t *testing.T) {
	var a Account
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"name":"acc","status":"active"}`), &a))
	require.Nil(t, a.RateMultiplier)
	require.Equal(t, 1.0, a.BillingRateMultiplier())
}

func TestAccount_BillingRateMultiplier_AllowsZero(t *testing.T) {
	v := 0.0
	a := Account{RateMultiplier: &v}
	require.Equal(t, 0.0, a.BillingRateMultiplier())
}

func TestAccount_BillingRateMultiplier_NegativeFallsBackToOne(t *testing.T) {
	v := -1.0
	a := Account{RateMultiplier: &v}
	require.Equal(t, 1.0, a.BillingRateMultiplier())
}

func TestAccount_GroupBillingMultiplier_DefaultsToOne(t *testing.T) {
	groupID := int64(10)
	var a Account
	require.Equal(t, 1.0, a.GroupBillingMultiplier(&groupID))
}

func TestAccount_GroupBillingMultiplier_UsesMatchingBinding(t *testing.T) {
	groupID := int64(10)
	a := Account{
		AccountGroups: []AccountGroup{
			{GroupID: 9, BillingMultiplier: 1.1},
			{GroupID: 10, BillingMultiplier: 1.35},
		},
	}
	require.Equal(t, 1.35, a.GroupBillingMultiplier(&groupID))
}

func TestAccount_GroupBillingMultiplier_InvalidFallsBackToOne(t *testing.T) {
	groupID := int64(10)
	a := Account{
		AccountGroups: []AccountGroup{
			{GroupID: 10, BillingMultiplier: 0},
		},
	}
	require.Equal(t, 1.0, a.GroupBillingMultiplier(&groupID))
}
