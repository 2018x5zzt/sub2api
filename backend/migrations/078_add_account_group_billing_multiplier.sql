-- Add a binding-level billing multiplier on account_groups.
-- This multiplier is multiplied into the existing user/group billing path.
ALTER TABLE account_groups
    ADD COLUMN IF NOT EXISTS billing_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1;

COMMENT ON COLUMN account_groups.billing_multiplier IS '账号在该分组下的扣费乘数，默认 1.0，乘入用户实际扣费链路';
