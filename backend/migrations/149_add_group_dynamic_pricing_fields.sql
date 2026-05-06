-- Add group-level dynamic pricing metadata.
-- pricing_mode controls whether a group uses fixed or dynamic pricing.
-- default_budget_multiplier is the default user/API-key budget multiplier for dynamic groups.

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS pricing_mode VARCHAR(20) NOT NULL DEFAULT 'fixed',
    ADD COLUMN IF NOT EXISTS default_budget_multiplier DECIMAL(10,4);

CREATE INDEX IF NOT EXISTS idx_groups_pricing_mode ON groups(pricing_mode);

COMMENT ON COLUMN groups.pricing_mode IS '分组定价模式：fixed 或 dynamic。';
COMMENT ON COLUMN groups.default_budget_multiplier IS '动态定价分组的默认用户预算倍率。';
