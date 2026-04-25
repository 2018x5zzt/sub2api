package service

import (
	"reflect"
	"testing"
)

func TestUserSubscriptionRepositoryIncludesAdvanceDailyWindow(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf((*UserSubscriptionRepository)(nil)).Elem()
	if _, ok := typ.MethodByName("AdvanceDailyWindow"); !ok {
		t.Fatalf("expected UserSubscriptionRepository to include AdvanceDailyWindow")
	}
}
