package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type groupHealthSnapshotRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewGroupHealthSnapshotRepository(client *dbent.Client, sqlDB *sql.DB) service.GroupHealthSnapshotRepository {
	return newGroupHealthSnapshotRepositoryWithSQL(client, sqlDB)
}

func newGroupHealthSnapshotRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *groupHealthSnapshotRepository {
	return &groupHealthSnapshotRepository{
		client: client,
		sql:    sqlq,
	}
}

func (r *groupHealthSnapshotRepository) UpsertBatch(ctx context.Context, snapshots []service.GroupHealthSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString(`
		INSERT INTO group_health_snapshots (
			group_id,
			bucket_time,
			health_percent,
			health_state
		)
		VALUES
	`)

	args := make([]any, 0, len(snapshots)*4)
	for i, snapshot := range snapshots {
		if i > 0 {
			query.WriteString(",")
		}

		base := i*4 + 1
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d)", base, base+1, base+2, base+3))
		args = append(args,
			snapshot.GroupID,
			snapshot.BucketTime.UTC(),
			snapshot.HealthPercent,
			snapshot.HealthState,
		)
	}

	query.WriteString(`
		ON CONFLICT (group_id, bucket_time) DO UPDATE SET
			health_percent = EXCLUDED.health_percent,
			health_state = EXCLUDED.health_state
	`)

	_, err := r.sql.ExecContext(ctx, query.String(), args...)
	return err
}

func (r *groupHealthSnapshotRepository) ListRecentByGroupIDs(ctx context.Context, groupIDs []int64, since time.Time) (map[int64][]service.GroupHealthSnapshot, error) {
	result := make(map[int64][]service.GroupHealthSnapshot)
	if len(groupIDs) == 0 {
		return result, nil
	}

	rows, err := r.sql.QueryContext(ctx, `
		SELECT group_id, bucket_time, health_percent, health_state
		FROM group_health_snapshots
		WHERE group_id = ANY($1)
			AND bucket_time >= $2
		ORDER BY group_id ASC, bucket_time ASC
	`, pq.Array(groupIDs), since.UTC())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var snapshot service.GroupHealthSnapshot
		if err := rows.Scan(
			&snapshot.GroupID,
			&snapshot.BucketTime,
			&snapshot.HealthPercent,
			&snapshot.HealthState,
		); err != nil {
			return nil, err
		}
		result[snapshot.GroupID] = append(result[snapshot.GroupID], snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *groupHealthSnapshotRepository) DeleteBefore(ctx context.Context, cutoff time.Time) (int, error) {
	res, err := r.sql.ExecContext(ctx, `
		DELETE FROM group_health_snapshots
		WHERE bucket_time < $1
	`, cutoff.UTC())
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}
