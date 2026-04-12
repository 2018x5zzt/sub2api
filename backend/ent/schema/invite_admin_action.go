package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type InviteAdminAction struct {
	ent.Schema
}

func (InviteAdminAction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invite_admin_actions"},
	}
}

func (InviteAdminAction) Fields() []ent.Field {
	return []ent.Field{
		field.String("action_type").
			MaxLen(32),
		field.Int64("operator_user_id"),
		field.Int64("target_user_id"),
		field.String("reason").
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.JSON("request_snapshot_json", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("result_snapshot_json", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}
