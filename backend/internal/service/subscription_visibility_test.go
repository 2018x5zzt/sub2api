package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type migratedVisibilityRepoStub struct {
	userSubRepoNoop

	all      []UserSubscription
	active   []UserSubscription
	migrated map[int64]struct{}
}

func (r *migratedVisibilityRepoStub) ListByUserID(context.Context, int64) ([]UserSubscription, error) {
	return append([]UserSubscription(nil), r.all...), nil
}

func (r *migratedVisibilityRepoStub) ListActiveByUserID(context.Context, int64) ([]UserSubscription, error) {
	return append([]UserSubscription(nil), r.active...), nil
}

func (r *migratedVisibilityRepoStub) ListMigratedLegacySubscriptionIDs(context.Context, int64) (map[int64]struct{}, error) {
	return r.migrated, nil
}

func TestListVisibleUserSubscriptions_HidesMigratedLegacySubscriptions(t *testing.T) {
	repo := &migratedVisibilityRepoStub{
		all: []UserSubscription{
			{ID: 10, UserID: 42, GroupID: 100, Status: SubscriptionStatusActive},
			{ID: 11, UserID: 42, GroupID: 101, Status: SubscriptionStatusActive},
		},
		migrated: map[int64]struct{}{10: {}},
	}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)

	visible, err := svc.ListVisibleUserSubscriptions(context.Background(), 42)

	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, int64(11), visible[0].ID)
}

func TestListVisibleActiveUserSubscriptions_HidesMigratedLegacySubscriptions(t *testing.T) {
	repo := &migratedVisibilityRepoStub{
		active: []UserSubscription{
			{ID: 20, UserID: 42, GroupID: 200, Status: SubscriptionStatusActive},
			{ID: 21, UserID: 42, GroupID: 201, Status: SubscriptionStatusActive},
		},
		migrated: map[int64]struct{}{20: {}},
	}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)

	visible, err := svc.ListVisibleActiveUserSubscriptions(context.Background(), 42)

	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, int64(21), visible[0].ID)
}
