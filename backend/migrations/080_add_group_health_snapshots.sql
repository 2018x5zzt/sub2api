CREATE TABLE IF NOT EXISTS group_health_snapshots (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL,
    bucket_time TIMESTAMPTZ NOT NULL,
    health_percent INTEGER NOT NULL CHECK (health_percent >= 0 AND health_percent <= 100),
    health_state VARCHAR(20) NOT NULL CHECK (health_state IN ('healthy', 'degraded', 'down'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_health_snapshots_group_id_bucket_time
    ON group_health_snapshots (group_id, bucket_time);

CREATE INDEX IF NOT EXISTS idx_group_health_snapshots_bucket_time
    ON group_health_snapshots (bucket_time);
