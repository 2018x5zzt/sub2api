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

// SubscriptionProductGroup binds a subscription product to a real routing group.
type SubscriptionProductGroup struct {
	ent.Schema
}

func (SubscriptionProductGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_product_groups"},
	}
}

func (SubscriptionProductGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (SubscriptionProductGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("product_id"),
		field.Int64("group_id"),
		field.Float("debit_multiplier").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Default(1),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.Int("sort_order").
			Default(0),
	}
}

func (SubscriptionProductGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", SubscriptionProduct.Type).
			Ref("group_bindings").
			Field("product_id").
			Unique().
			Required(),
	}
}

func (SubscriptionProductGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("group_id"),
		index.Fields("status"),
		index.Fields("sort_order"),
		index.Fields("deleted_at"),
		index.Fields("product_id", "group_id"),
	}
}
