package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SubscriptionProduct holds the schema definition for product-level subscription entitlements.
type SubscriptionProduct struct {
	ent.Schema
}

func (SubscriptionProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_products"},
	}
}

func (SubscriptionProduct) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (SubscriptionProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			MaxLen(64).
			NotEmpty(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("status").
			MaxLen(20).
			Default("draft"),
		field.Int("default_validity_days").
			Default(30),
		field.Float("daily_limit_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("weekly_limit_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("monthly_limit_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Int("sort_order").
			Default(0),
	}
}

func (SubscriptionProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group_bindings", SubscriptionProductGroup.Type),
		edge.To("user_subscriptions", UserProductSubscription.Type),
	}
}

func (SubscriptionProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
		index.Fields("status"),
		index.Fields("sort_order"),
		index.Fields("deleted_at"),
	}
}
