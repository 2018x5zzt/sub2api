-- Restore user-level subscription-to-balance fallback columns (xlab product subscription).
--
-- The original migrations 144_product_subscription_family_balance_fallback.sql,
-- 145_product_subscription_explicit_fallback_family.sql and
-- 147_product_subscription_family_gpt.sql were dropped during the 0.1.123 kernel
-- re-roll, and those numbers were later reused by unrelated upstream migrations.
-- Existing production databases already have these columns (144/145/147 applied),
-- so the IF NOT EXISTS guards make this migration a safe no-op there. Fresh
-- installs need them recreated for the ent schema and fallback billing flow.
--
-- DDL only. The one-time production data seeding from the originals (mixed balance
-- pool group creation, per-group fallback mapping, family normalization) is not
-- replayed here; fallback groups are configured at runtime via the admin/user UI.

-- Product family on subscription products (terminal state: single 'gpt' family).
ALTER TABLE subscription_products
    ADD COLUMN IF NOT EXISTS product_family VARCHAR(64) NOT NULL DEFAULT 'gpt';

CREATE INDEX IF NOT EXISTS idx_subscription_products_family_sort
    ON subscription_products (product_family, sort_order, id)
    WHERE deleted_at IS NULL;

-- User-level subscription-to-balance fallback settings.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_limit_usd DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_used_usd DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_group_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_users_subscription_balance_fallback_group_id
    ON users(subscription_balance_fallback_group_id)
    WHERE subscription_balance_fallback_group_id IS NOT NULL
      AND deleted_at IS NULL;

-- Per-subscription-group balance fallback mapping.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS balance_fallback_group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_groups_balance_fallback_group_id
    ON groups(balance_fallback_group_id)
    WHERE deleted_at IS NULL AND balance_fallback_group_id IS NOT NULL;

-- API key product-family selection for subscription groups.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS subscription_product_family VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_api_keys_subscription_product_family
    ON api_keys(subscription_product_family)
    WHERE subscription_product_family IS NOT NULL
      AND deleted_at IS NULL;
