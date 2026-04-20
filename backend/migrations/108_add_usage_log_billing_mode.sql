-- Add billing_mode to usage_logs so pricing-path snapshots can persist token/per_request/image mode.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS billing_mode VARCHAR(20);
