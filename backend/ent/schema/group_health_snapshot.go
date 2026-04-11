package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GroupHealthSnapshot struct {
	ent.Schema
}

func (GroupHealthSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_health_snapshots"},
	}
}

func (GroupHealthSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id"),
		field.Time("bucket_time"),
		field.Int("health_percent").
			Validate(func(v int) error {
				if v < 0 || v > 100 {
					return fmt.Errorf("invalid health percent: %d", v)
				}
				return nil
			}),
		field.Enum("health_state").
			Values("healthy", "degraded", "down"),
	}
}

func (GroupHealthSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id", "bucket_time").Unique(),
		index.Fields("bucket_time"),
	}
}
