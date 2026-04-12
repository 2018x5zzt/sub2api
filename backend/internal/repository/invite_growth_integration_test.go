//go:build integration

package repository

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type InviteGrowthRepoSuite struct {
	suite.Suite
	ctx        context.Context
	client     *dbent.Client
	userRepo   *userRepository
	redeemRepo *redeemCodeRepository
}

func (s *InviteGrowthRepoSuite) SetupTest() {
	s.ctx = context.Background()
	tx := testEntTx(s.T())
	s.client = tx.Client()
	s.userRepo = NewUserRepository(s.client, nil).(*userRepository)
	s.redeemRepo = NewRedeemCodeRepository(s.client).(*redeemCodeRepository)
}

func TestInviteGrowthRepoSuite(t *testing.T) {
	suite.Run(t, new(InviteGrowthRepoSuite))
}

func (s *InviteGrowthRepoSuite) TestUserInviteFieldsRoundTrip() {
	inviter := &service.User{
		Email:        "inviter@example.com",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		InviteCode:   "INVITER42",
	}
	s.Require().NoError(s.userRepo.Create(s.ctx, inviter))

	inviterID := inviter.ID
	user := &service.User{
		Email:           "invite-roundtrip@example.com",
		PasswordHash:    "hash",
		Role:            service.RoleUser,
		Status:          service.StatusActive,
		InviteCode:      "INVITE42",
		InvitedByUserID: &inviterID,
	}

	s.Require().NoError(s.userRepo.Create(s.ctx, user))

	got, err := s.userRepo.GetByID(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().Equal("INVITE42", got.InviteCode)
	s.Require().NotNil(got.InvitedByUserID)
	s.Require().Equal(inviter.ID, *got.InvitedByUserID)
	s.Require().NotNil(got.InviteBoundAt)
}

func (s *InviteGrowthRepoSuite) TestRedeemCodeSourceTypeRoundTrip() {
	code := &service.RedeemCode{
		Code:       "COMMERCIAL-001",
		Type:       service.RedeemTypeBalance,
		Value:      100,
		Status:     service.StatusUnused,
		SourceType: service.RedeemSourceCommercial,
	}

	s.Require().NoError(s.redeemRepo.Create(s.ctx, code))

	got, err := s.redeemRepo.GetByCode(s.ctx, "COMMERCIAL-001")
	s.Require().NoError(err)
	s.Require().Equal(service.RedeemSourceCommercial, got.SourceType)
}
