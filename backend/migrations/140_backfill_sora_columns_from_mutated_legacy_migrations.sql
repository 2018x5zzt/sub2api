-- Migration: 140_backfill_sora_columns_from_mutated_legacy_migrations
-- Backfill columns that were added to historical migrations after some
-- deployments had already recorded those migration filenames.

ALTER TABLE groups
	ADD COLUMN IF NOT EXISTS sora_image_price_360 decimal(20,8),
	ADD COLUMN IF NOT EXISTS sora_image_price_540 decimal(20,8),
	ADD COLUMN IF NOT EXISTS sora_video_price_per_request decimal(20,8),
	ADD COLUMN IF NOT EXISTS sora_video_price_per_request_hd decimal(20,8),
	ADD COLUMN IF NOT EXISTS sora_storage_quota_bytes BIGINT NOT NULL DEFAULT 0;

ALTER TABLE users
	ADD COLUMN IF NOT EXISTS sora_storage_quota_bytes BIGINT NOT NULL DEFAULT 0,
	ADD COLUMN IF NOT EXISTS sora_storage_used_bytes BIGINT NOT NULL DEFAULT 0;

ALTER TABLE usage_logs
	ADD COLUMN IF NOT EXISTS media_type VARCHAR(16);
