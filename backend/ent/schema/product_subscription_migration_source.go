package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductSubscriptionMigrationSource preserves legacy subscription lineage for product backfills.
type ProductSubscriptionMigrationSource struct {
	ent.Schema
}

func (ProductSubscriptionMigrationSource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "product_subscription_migration_sources"},
	}
}

func (ProductSubscriptionMigrationSource) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("product_subscription_id"),
		field.Int64("legacy_user_subscription_id"),
		field.String("migration_batch").
			MaxLen(128),
		field.Int64("legacy_group_id"),
		field.String("legacy_status").
			MaxLen(20),
		field.Time("legacy_starts_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("legacy_expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Float("legacy_daily_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("legacy_weekly_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("legacy_monthly_usage_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("legacy_daily_carryover_in_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Float("legacy_daily_carryover_remaining_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).
			Default(0),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ProductSubscriptionMigrationSource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product_subscription", UserProductSubscription.Type).
			Ref("migration_sources").
			Field("product_subscription_id").
			Unique().
			Required(),
	}
}

func (ProductSubscriptionMigrationSource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_subscription_id"),
		index.Fields("legacy_user_subscription_id"),
		index.Fields("migration_batch"),
		index.Fields("legacy_group_id"),
	}
}
