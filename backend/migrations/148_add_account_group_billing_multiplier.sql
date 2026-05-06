-- Store the billing multiplier of an account within a specific group.
-- Dynamic pricing uses this per-binding multiplier when calculating account-side cost.

ALTER TABLE account_groups
    ADD COLUMN IF NOT EXISTS billing_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1.0;

COMMENT ON COLUMN account_groups.billing_multiplier IS '账号在该分组下的扣费倍率，动态定价分组用于账号侧成本计算。';
