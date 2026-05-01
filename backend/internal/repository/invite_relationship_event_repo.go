package repository

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbinviterelationshipevent "github.com/Wei-Shaw/sub2api/ent/inviterelationshipevent"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type inviteRelationshipEventRepository struct {
	client *dbent.Client
}

func NewInviteRelationshipEventRepository(client *dbent.Client) *inviteRelationshipEventRepository {
	return &inviteRelationshipEventRepository{client: client}
}

func (r *inviteRelationshipEventRepository) Create(ctx context.Context, event *service.InviteRelationshipEvent) error {
	if event == nil {
		return nil
	}

	client := clientFromContext(ctx, r.client)
	create := client.InviteRelationshipEvent.Create().
		SetInviteeUserID(event.InviteeUserID).
		SetEventType(event.EventType)
	if event.PreviousInviterUserID != nil {
		create.SetPreviousInviterUserID(*event.PreviousInviterUserID)
	}
	if event.NewInviterUserID != nil {
		create.SetNewInviterUserID(*event.NewInviterUserID)
	}
	if event.EffectiveAt.IsZero() {
		create.SetEffectiveAt(time.Now())
	} else {
		create.SetEffectiveAt(event.EffectiveAt)
	}
	if event.OperatorUserID != nil {
		create.SetOperatorUserID(*event.OperatorUserID)
	}
	create.SetNillableReason(stringPtrOrNil(event.Reason))

	row, err := create.Save(ctx)
	if err != nil {
		return err
	}
	applyInviteRelationshipEventEntityToService(event, row)
	return nil
}

func (r *inviteRelationshipEventRepository) ListByInvitee(ctx context.Context, inviteeUserID int64) ([]service.InviteRelationshipEvent, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.InviteRelationshipEvent.Query().
		Where(dbinviterelationshipevent.InviteeUserIDEQ(inviteeUserID)).
		Order(
			dbinviterelationshipevent.ByEffectiveAt(sql.OrderDesc()),
			dbinviterelationshipevent.ByID(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapInviteRelationshipEventEntities(rows), nil
}

func (r *inviteRelationshipEventRepository) GetEffectiveInviterAt(ctx context.Context, inviteeUserID int64, at time.Time) (*int64, error) {
	client := clientFromContext(ctx, r.client)
	row, err := client.InviteRelationshipEvent.Query().
		Where(
			dbinviterelationshipevent.InviteeUserIDEQ(inviteeUserID),
			dbinviterelationshipevent.EffectiveAtLTE(at),
		).
		Order(
			dbinviterelationshipevent.ByEffectiveAt(sql.OrderDesc()),
			dbinviterelationshipevent.ByID(sql.OrderDesc()),
		).
		First(ctx)
	if err == nil {
		return row.NewInviterUserID, nil
	}
	if !dbent.IsNotFound(err) {
		return nil, err
	}

	userRow, userErr := client.User.Query().
		Where(dbuser.IDEQ(inviteeUserID)).
		Only(ctx)
	if userErr != nil {
		if dbent.IsNotFound(userErr) {
			return nil, service.ErrUserNotFound
		}
		return nil, userErr
	}
	if userRow.InvitedByUserID == nil {
		return nil, nil
	}
	if userRow.InviteBoundAt != nil && userRow.InviteBoundAt.After(at) {
		return nil, nil
	}
	return userRow.InvitedByUserID, nil
}

func applyInviteRelationshipEventEntityToService(dst *service.InviteRelationshipEvent, src *dbent.InviteRelationshipEvent) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.InviteeUserID = src.InviteeUserID
	dst.PreviousInviterUserID = src.PreviousInviterUserID
	dst.NewInviterUserID = src.NewInviterUserID
	dst.EventType = src.EventType
	dst.EffectiveAt = src.EffectiveAt
	dst.OperatorUserID = src.OperatorUserID
	dst.Reason = derefString(src.Reason)
	dst.CreatedAt = src.CreatedAt
}

func mapInviteRelationshipEventEntities(rows []*dbent.InviteRelationshipEvent) []service.InviteRelationshipEvent {
	events := make([]service.InviteRelationshipEvent, 0, len(rows))
	for _, row := range rows {
		var event service.InviteRelationshipEvent
		applyInviteRelationshipEventEntityToService(&event, row)
		events = append(events, event)
	}
	return events
}
