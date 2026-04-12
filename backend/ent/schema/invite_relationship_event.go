package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type InviteRelationshipEvent struct {
	ent.Schema
}

func (InviteRelationshipEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invite_relationship_events"},
	}
}

func (InviteRelationshipEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("invitee_user_id"),
		field.Int64("previous_inviter_user_id").
			Optional().
			Nillable(),
		field.Int64("new_inviter_user_id").
			Optional().
			Nillable(),
		field.String("event_type").
			MaxLen(32),
		field.Time("effective_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("operator_user_id").
			Optional().
			Nillable(),
		field.String("reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (InviteRelationshipEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("invitee_user_id", "effective_at", "id"),
	}
}
