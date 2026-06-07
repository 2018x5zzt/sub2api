package main

import (
	"testing"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/subscriptions"
)

func TestSubscriptionReadSourceMapping(t *testing.T) {
	cases := []struct {
		in   config.SubscriptionReadSource
		want subscriptions.ReadSource
	}{
		{config.SubscriptionReadSourceCore, subscriptions.ReadSourceCore},
		{config.SubscriptionReadSourceHybrid, subscriptions.ReadSourceHybrid},
		{config.SubscriptionReadSourceXlab, subscriptions.ReadSourceXlab},
	}
	for _, tc := range cases {
		if got := subscriptionReadSource(tc.in); got != tc.want {
			t.Fatalf("subscriptionReadSource(%q) = %q", tc.in, got)
		}
	}
}
