//go:build integration

package repository

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
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

func inviteCodeForEmail(email string) string {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(email))
	return fmt.Sprintf("IV%08X", hasher.Sum32())
}
