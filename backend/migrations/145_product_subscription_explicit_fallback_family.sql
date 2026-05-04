-- Product subscription closure:
-- - user-selected balance fallback group
-- - API key product-family selection for subscription groups

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_group_id BIGINT;

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS subscription_product_family VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_users_subscription_balance_fallback_group_id
    ON users(subscription_balance_fallback_group_id)
    WHERE subscription_balance_fallback_group_id IS NOT NULL
      AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_api_keys_subscription_product_family
    ON api_keys(subscription_product_family)
    WHERE subscription_product_family IS NOT NULL
      AND deleted_at IS NULL;
