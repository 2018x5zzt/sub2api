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

type InviteRewardRecord struct {
	ent.Schema
}

func (InviteRewardRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invite_reward_records"},
	}
}

func (InviteRewardRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("inviter_user_id"),
		field.Int64("invitee_user_id"),
		field.Int64("trigger_redeem_code_id").
			Optional().
			Nillable(),
		field.Float("trigger_redeem_code_value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Int64("reward_target_user_id"),
		field.String("reward_role").
			MaxLen(32),
		field.String("reward_type").
			MaxLen(64),
		field.Float("reward_rate").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,8)"}),
		field.Float("reward_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("status").
			MaxLen(32).
			Default("applied"),
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int64("admin_action_id").
			Optional().
			Nillable(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (InviteRewardRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reward_target_user_id", "created_at"),
		index.Fields("admin_action_id"),
		index.Fields("trigger_redeem_code_id", "reward_role", "reward_type").Unique(),
	}
}
