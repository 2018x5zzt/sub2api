package service

import (
	"reflect"
	"testing"
)

func TestUserSubscriptionIncludesDailyCarryoverFields(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(UserSubscription{})

	if _, ok := typ.FieldByName("DailyCarryoverInUSD"); !ok {
		t.Fatalf("expected UserSubscription to include DailyCarryoverInUSD")
	}

	if _, ok := typ.FieldByName("DailyCarryoverRemainingUSD"); !ok {
		t.Fatalf("expected UserSubscription to include DailyCarryoverRemainingUSD")
	}
}
