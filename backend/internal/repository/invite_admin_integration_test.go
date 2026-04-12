//go:build integration

package repository

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviteadminquery "github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type InviteAdminRepoSuite struct {
	suite.Suite
	ctx                   context.Context
	client                *dbent.Client
	userRepo              *userRepository
	rewardRepo            *inviteRewardRecordRepository
	adminActionRepo       service.InviteAdminActionRepository
	relationshipEventRepo service.InviteRelationshipEventRepository
}

func (s *InviteAdminRepoSuite) SetupTest() {
	s.ctx = context.Background()
	tx := testEntTx(s.T())
	s.client = tx.Client()
	s.userRepo = NewUserRepository(s.client, nil).(*userRepository)
	s.rewardRepo = NewInviteRewardRecordRepository(s.client).(*inviteRewardRecordRepository)
	s.adminActionRepo = NewInviteAdminActionRepository(s.client)
	s.relationshipEventRepo = NewInviteRelationshipEventRepository(s.client)
}

func TestInviteAdminRepoSuite(t *testing.T) {
	suite.Run(t, new(InviteAdminRepoSuite))
}

func (s *InviteAdminRepoSuite) TestBackfilledRegisterBindEventExistsForAlreadyBoundUser() {
	client := testEntClient(s.T())
	userRepo := newUserRepositoryWithSQL(client, integrationDB)
	eventRepo := NewInviteRelationshipEventRepository(client)

	inviter := &service.User{
		Email:        "invite-admin-inviter@example.com",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		InviteCode:   inviteCodeForEmail("invite-admin-inviter@example.com"),
	}
	s.Require().NoError(userRepo.Create(s.ctx, inviter))

	boundAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	invitee, err := client.User.Create().
		SetEmail("invite-admin-invitee@example.com").
		SetPasswordHash("hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		SetInviteCode(inviteCodeForEmail("invite-admin-invitee@example.com")).
		SetInvitedByUserID(inviter.ID).
		SetInviteBoundAt(boundAt).
		Save(s.ctx)
	s.Require().NoError(err)

	events, err := eventRepo.ListByInvitee(s.ctx, invitee.ID)
	s.Require().NoError(err)
	s.Require().Empty(events)

	migrationSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "082_add_invite_admin_ops.sql"))
	s.Require().NoError(err)

	_, err = integrationDB.ExecContext(s.ctx, string(migrationSQL))
	s.Require().NoError(err)

	events, err = eventRepo.ListByInvitee(s.ctx, invitee.ID)
	s.Require().NoError(err)
	s.Require().NotEmpty(events)

	var registerBind *service.InviteRelationshipEvent
	for i := range events {
		if events[i].EventType == service.InviteRelationshipEventTypeRegisterBind {
			registerBind = &events[i]
			break
		}
	}

	s.Require().NotNil(registerBind)
	s.Require().Equal(invitee.ID, registerBind.InviteeUserID)
	s.Require().NotNil(registerBind.NewInviterUserID)
	s.Require().Equal(inviter.ID, *registerBind.NewInviterUserID)
	s.Require().Equal(boundAt, registerBind.EffectiveAt)
}

func (s *InviteAdminRepoSuite) TestManualRewardRecordAllowsNilTriggerAndActionLink() {
	operator := s.createUser("invite-admin-operator@example.com", nil, nil)
	target := s.createUser("invite-admin-target@example.com", nil, nil)
	inviter := s.createUser("invite-admin-manual-inviter@example.com", nil, nil)

	action := &service.InviteAdminAction{
		ActionType:          service.InviteAdminActionTypeManualGrant,
		OperatorUserID:      operator.ID,
		TargetUserID:        target.ID,
		Reason:              "manual correction",
		RequestSnapshotJSON: map[string]any{"source": "integration_test"},
		ResultSnapshotJSON:  map[string]any{"status": "applied"},
	}
	s.Require().NoError(s.adminActionRepo.Create(s.ctx, action))

	s.Require().NoError(s.rewardRepo.CreateBatch(s.ctx, []service.InviteRewardRecord{
		{
			InviterUserID:          inviter.ID,
			InviteeUserID:          target.ID,
			TriggerRedeemCodeID:    nil,
			TriggerRedeemCodeValue: 0,
			RewardTargetUserID:     target.ID,
			RewardRole:             service.InviteRewardRoleInvitee,
			RewardType:             service.InviteRewardTypeManualGrant,
			RewardAmount:           18.5,
			Status:                 "applied",
			Notes:                  "manual grant",
			AdminActionID:          &action.ID,
		},
	}))

	records, err := s.rewardRepo.ListByAdminActionID(s.ctx, action.ID)
	s.Require().NoError(err)
	s.Require().Len(records, 1)

	record := records[0]
	s.Require().Equal(service.InviteRewardTypeManualGrant, record.RewardType)
	s.Require().NotNil(record.AdminActionID)
	s.Require().Equal(action.ID, *record.AdminActionID)
	s.Require().Nil(record.TriggerRedeemCodeID)
}

func (s *InviteAdminRepoSuite) TestGetStatsUsesDatabaseAggregates() {
	baselineRepo := NewInviteAdminQueryRepository(s.client).(*inviteAdminQueryRepository)
	baselineStats, err := baselineRepo.GetStats(s.ctx)
	s.Require().NoError(err)

	inviter := s.createUser("invite-stats-inviter@example.com", nil, nil)
	inviteeA := s.createUser("invite-stats-a@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-2*time.Hour)))
	inviteeB := s.createUser("invite-stats-b@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-time.Hour)))

	s.Require().NoError(s.rewardRepo.CreateBatch(s.ctx, []service.InviteRewardRecord{
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeA.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       3,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeB.ID,
			RewardTargetUserID: inviteeB.ID,
			RewardRole:         service.InviteRewardRoleInvitee,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       4,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeB.ID,
			RewardTargetUserID: inviteeB.ID,
			RewardRole:         service.InviteRewardRoleInvitee,
			RewardType:         service.InviteRewardTypeManualGrant,
			RewardAmount:       2,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeA.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeRecomputeDelta,
			RewardAmount:       -1,
			Status:             "applied",
		},
	}))

	repo, logs := s.newInviteAdminQueryRepoWithDebugLogs()
	stats, err := repo.GetStats(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(baselineStats.TotalInvitedUsers+2, stats.TotalInvitedUsers)
	s.Require().Equal(baselineStats.QualifiedRewardUsersTotal+2, stats.QualifiedRewardUsersTotal)
	s.Require().Equal(baselineStats.BaseRewardsTotal+7.0, stats.BaseRewardsTotal)
	s.Require().Equal(baselineStats.ManualGrantsTotal+2.0, stats.ManualGrantsTotal)
	s.Require().Equal(baselineStats.RecomputeAdjustmentsTotal-1.0, stats.RecomputeAdjustmentsTotal)

	sqlLogs := logs()
	s.Contains(sqlLogs, "COUNT(DISTINCT")
	s.Contains(sqlLogs, "SUM(")
}

func (s *InviteAdminRepoSuite) TestListRelationshipsUsesDatabasePaginationForSearch() {
	inviter := s.createUser("alpha-inviter@example.com", nil, nil)
	otherInviter := s.createUser("beta-inviter@example.com", nil, nil)
	oldInvitee := s.createUser("alpha-old@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-2*time.Hour)))
	newInvitee := s.createUser("alpha-new@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-time.Hour)))
	_ = s.createUser("beta-user@example.com", &otherInviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-30*time.Minute)))

	s.Require().NoError(s.relationshipEventRepo.Create(s.ctx, &service.InviteRelationshipEvent{
		InviteeUserID:    newInvitee.ID,
		NewInviterUserID: &inviter.ID,
		EventType:        service.InviteRelationshipEventTypeRegisterBind,
		EffectiveAt:      time.Now().UTC().Add(-45 * time.Minute),
	}))
	s.Require().NoError(s.relationshipEventRepo.Create(s.ctx, &service.InviteRelationshipEvent{
		InviteeUserID:    oldInvitee.ID,
		NewInviterUserID: &inviter.ID,
		EventType:        service.InviteRelationshipEventTypeRegisterBind,
		EffectiveAt:      time.Now().UTC().Add(-100 * time.Minute),
	}))

	repo, logs := s.newInviteAdminQueryRepoWithDebugLogs()
	rows, page, err := repo.ListRelationships(s.ctx, dbinviteadminquery.PaginationParams{Page: 2, PageSize: 1}, service.AdminInviteRelationshipFilters{
		Search: "alpha-inviter",
	})
	s.Require().NoError(err)
	s.Require().Len(rows, 1)
	s.Require().Equal(int64(2), page.Total)
	s.Require().Equal(oldInvitee.ID, rows[0].InviteeUserID)
	s.Require().Equal("alpha-inviter@example.com", rows[0].CurrentInviterEmail)

	sqlLogs := logs()
	s.Contains(sqlLogs, "COUNT(")
	s.Contains(sqlLogs, "LIMIT 1")
	s.Contains(sqlLogs, "OFFSET 1")
}

func (s *InviteAdminRepoSuite) TestListRewardsUsesDatabasePaginationForSearch() {
	inviter := s.createUser("reward-inviter@example.com", nil, nil)
	otherInviter := s.createUser("other-inviter@example.com", nil, nil)
	inviteeA := s.createUser("reward-a@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-2*time.Hour)))
	inviteeB := s.createUser("reward-b@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-time.Hour)))
	otherInvitee := s.createUser("reward-c@example.com", &otherInviter.ID, inviteBoundAtPtr(time.Now().UTC().Add(-30*time.Minute)))

	s.Require().NoError(s.rewardRepo.CreateBatch(s.ctx, []service.InviteRewardRecord{
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeA.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       2,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      inviteeB.ID,
			RewardTargetUserID: inviteeB.ID,
			RewardRole:         service.InviteRewardRoleInvitee,
			RewardType:         service.InviteRewardTypeManualGrant,
			RewardAmount:       3,
			Status:             "applied",
		},
		{
			InviterUserID:      otherInviter.ID,
			InviteeUserID:      otherInvitee.ID,
			RewardTargetUserID: otherInviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       4,
			Status:             "applied",
		},
	}))

	repo, logs := s.newInviteAdminQueryRepoWithDebugLogs()
	rows, page, err := repo.ListRewards(s.ctx, dbinviteadminquery.PaginationParams{Page: 2, PageSize: 1}, service.AdminInviteRewardFilters{
		Search: "reward-inviter",
	})
	s.Require().NoError(err)
	s.Require().Len(rows, 1)
	s.Require().Equal(int64(2), page.Total)
	s.Require().Equal("reward-inviter@example.com", rows[0].InviterEmail)

	sqlLogs := logs()
	s.Contains(sqlLogs, "COUNT(")
	s.Contains(sqlLogs, "LIMIT 1")
	s.Contains(sqlLogs, "OFFSET 1")
}

func (s *InviteAdminRepoSuite) TestSumBaseRewardsByTargetAndRoleUsesDatabaseAggregation() {
	inviter := s.createUser("sum-inviter@example.com", nil, nil)
	invitee := s.createUser("sum-invitee@example.com", &inviter.ID, inviteBoundAtPtr(time.Now().UTC()))

	s.Require().NoError(s.rewardRepo.CreateBatch(s.ctx, []service.InviteRewardRecord{
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       1.5,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeBase,
			RewardAmount:       2.25,
			Status:             "applied",
		},
		{
			InviterUserID:      inviter.ID,
			InviteeUserID:      invitee.ID,
			RewardTargetUserID: inviter.ID,
			RewardRole:         service.InviteRewardRoleInviter,
			RewardType:         service.InviteRewardTypeManualGrant,
			RewardAmount:       99,
			Status:             "applied",
		},
	}))

	repo, logs := s.newInviteRewardRepoWithDebugLogs()
	total, err := repo.SumBaseRewardsByTargetAndRole(s.ctx, inviter.ID, service.InviteRewardRoleInviter)
	s.Require().NoError(err)
	s.Require().Equal(3.75, total)

	sqlLogs := logs()
	s.Contains(sqlLogs, "SUM(")
}

func (s *InviteAdminRepoSuite) TestSumBaseRewardsByTargetAndRoleReturnsZeroWhenNoRows() {
	user := s.createUser("sum-empty@example.com", nil, nil)

	total, err := s.rewardRepo.SumBaseRewardsByTargetAndRole(s.ctx, user.ID, service.InviteRewardRoleInviter)
	s.Require().NoError(err)
	s.Require().Equal(0.0, total)
}

func (s *InviteAdminRepoSuite) TestListActionsUsesDatabasePagination() {
	operator := s.createUser("action-operator@example.com", nil, nil)
	target := s.createUser("action-target@example.com", nil, nil)

	older := &service.InviteAdminAction{
		ActionType:          service.InviteAdminActionTypeManualGrant,
		OperatorUserID:      operator.ID,
		TargetUserID:        target.ID,
		Reason:              "older action",
		RequestSnapshotJSON: map[string]any{"idx": 1},
		ResultSnapshotJSON:  map[string]any{"status": "ok"},
	}
	s.Require().NoError(s.adminActionRepo.Create(s.ctx, older))

	newer := &service.InviteAdminAction{
		ActionType:          service.InviteAdminActionTypeManualGrant,
		OperatorUserID:      operator.ID,
		TargetUserID:        target.ID,
		Reason:              "newer action",
		RequestSnapshotJSON: map[string]any{"idx": 2},
		ResultSnapshotJSON:  map[string]any{"status": "ok"},
	}
	s.Require().NoError(s.adminActionRepo.Create(s.ctx, newer))

	other := &service.InviteAdminAction{
		ActionType:          service.InviteAdminActionTypeRebind,
		OperatorUserID:      operator.ID,
		TargetUserID:        target.ID,
		Reason:              "other action",
		RequestSnapshotJSON: map[string]any{"idx": 3},
		ResultSnapshotJSON:  map[string]any{"status": "ok"},
	}
	s.Require().NoError(s.adminActionRepo.Create(s.ctx, other))

	repo, logs := s.newInviteAdminQueryRepoWithDebugLogs()
	rows, page, err := repo.ListActions(s.ctx, dbinviteadminquery.PaginationParams{Page: 2, PageSize: 1}, service.InviteAdminActionFilters{
		ActionType: service.InviteAdminActionTypeManualGrant,
	})
	s.Require().NoError(err)
	s.Require().Len(rows, 1)
	s.Require().Equal(int64(2), page.Total)
	s.Require().Equal(older.ID, rows[0].ID)
	s.Require().Equal("older action", rows[0].Reason)

	sqlLogs := logs()
	s.Contains(sqlLogs, "COUNT(")
	s.Contains(sqlLogs, "LIMIT 1")
	s.Contains(sqlLogs, "OFFSET 1")
}

func (s *InviteAdminRepoSuite) createUser(email string, invitedByUserID *int64, inviteBoundAt *time.Time) *service.User {
	user := &service.User{
		Email:           email,
		PasswordHash:    "hash",
		Role:            service.RoleUser,
		Status:          service.StatusActive,
		InviteCode:      inviteCodeForEmail(email),
		InvitedByUserID: invitedByUserID,
		InviteBoundAt:   inviteBoundAt,
	}
	s.Require().NoError(s.userRepo.Create(s.ctx, user))
	return user
}

func (s *InviteAdminRepoSuite) newInviteAdminQueryRepoWithDebugLogs() (*inviteAdminQueryRepository, func() string) {
	debugClient, collect := newDebugEntClientWithLogs(s.client)
	return NewInviteAdminQueryRepository(debugClient).(*inviteAdminQueryRepository), collect
}

func (s *InviteAdminRepoSuite) newInviteRewardRepoWithDebugLogs() (*inviteRewardRecordRepository, func() string) {
	debugClient, collect := newDebugEntClientWithLogs(s.client)
	return NewInviteRewardRecordRepository(debugClient).(*inviteRewardRecordRepository), collect
}

func newDebugEntClientWithLogs(client *dbent.Client) (*dbent.Client, func() string) {
	var (
		mu   sync.Mutex
		logs []string
	)

	debugClient := dbent.NewClient(
		dbent.Driver(client.Driver()),
		dbent.Debug(),
		dbent.Log(func(args ...any) {
			mu.Lock()
			defer mu.Unlock()
			logs = append(logs, fmt.Sprint(args...))
		}),
	)

	return debugClient, func() string {
		mu.Lock()
		defer mu.Unlock()
		return strings.Join(logs, "\n")
	}
}

func inviteBoundAtPtr(value time.Time) *time.Time {
	return &value
}

func inviteCodeForEmail(email string) string {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(email))
	return fmt.Sprintf("IV%08X", hasher.Sum32())
}
