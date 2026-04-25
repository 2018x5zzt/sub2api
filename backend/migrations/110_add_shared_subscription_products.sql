-- Add shared subscription products as an additive runtime model.

CREATE TABLE IF NOT EXISTS subscription_products (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    default_validity_days INT NOT NULL DEFAULT 30,
    daily_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    weekly_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    monthly_limit_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS subscription_products_code_unique_active
    ON subscription_products (code)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_products_status
    ON subscription_products (status);
CREATE INDEX IF NOT EXISTS idx_subscription_products_sort_order
    ON subscription_products (sort_order);

CREATE TABLE IF NOT EXISTS subscription_product_groups (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES subscription_products(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    debit_multiplier DECIMAL(10,4) NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT subscription_product_groups_debit_multiplier_positive
        CHECK (debit_multiplier > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS subscription_product_groups_product_group_unique_active
    ON subscription_product_groups (product_id, group_id)
    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS subscription_product_groups_group_unique_active
    ON subscription_product_groups (group_id)
    WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX IF NOT EXISTS idx_subscription_product_groups_product_id
    ON subscription_product_groups (product_id);
CREATE INDEX IF NOT EXISTS idx_subscription_product_groups_group_id
    ON subscription_product_groups (group_id);
CREATE INDEX IF NOT EXISTS idx_subscription_product_groups_status
    ON subscription_product_groups (status);
CREATE INDEX IF NOT EXISTS idx_subscription_product_groups_sort_order
    ON subscription_product_groups (sort_order);

CREATE TABLE IF NOT EXISTS user_product_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES subscription_products(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    daily_window_start TIMESTAMPTZ,
    weekly_window_start TIMESTAMPTZ,
    monthly_window_start TIMESTAMPTZ,
    daily_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    weekly_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    monthly_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    daily_carryover_in_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    daily_carryover_remaining_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    assigned_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS user_product_subscriptions_user_product_unique_active
    ON user_product_subscriptions (user_id, product_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_product_subscriptions_user_id
    ON user_product_subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_product_subscriptions_product_id
    ON user_product_subscriptions (product_id);
CREATE INDEX IF NOT EXISTS idx_user_product_subscriptions_status
    ON user_product_subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_user_product_subscriptions_expires_at
    ON user_product_subscriptions (expires_at);
CREATE INDEX IF NOT EXISTS idx_user_product_subscriptions_assigned_by
    ON user_product_subscriptions (assigned_by);

CREATE TABLE IF NOT EXISTS product_subscription_migration_sources (
    id BIGSERIAL PRIMARY KEY,
    product_subscription_id BIGINT NOT NULL REFERENCES user_product_subscriptions(id) ON DELETE CASCADE,
    legacy_user_subscription_id BIGINT NOT NULL REFERENCES user_subscriptions(id) ON DELETE RESTRICT,
    migration_batch VARCHAR(128) NOT NULL,
    legacy_group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    legacy_status VARCHAR(20) NOT NULL,
    legacy_starts_at TIMESTAMPTZ NOT NULL,
    legacy_expires_at TIMESTAMPTZ NOT NULL,
    legacy_daily_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    legacy_weekly_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    legacy_monthly_usage_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    legacy_daily_carryover_in_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    legacy_daily_carryover_remaining_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS product_subscription_migration_sources_legacy_unique
    ON product_subscription_migration_sources (legacy_user_subscription_id);
CREATE INDEX IF NOT EXISTS idx_product_subscription_migration_sources_product_subscription_id
    ON product_subscription_migration_sources (product_subscription_id);
CREATE INDEX IF NOT EXISTS idx_product_subscription_migration_sources_migration_batch
    ON product_subscription_migration_sources (migration_batch);
CREATE INDEX IF NOT EXISTS idx_product_subscription_migration_sources_legacy_group_id
    ON product_subscription_migration_sources (legacy_group_id);

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_product_id
    ON redeem_codes (product_id);

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS product_subscription_id BIGINT REFERENCES user_product_subscriptions(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS group_debit_multiplier DECIMAL(10,4),
    ADD COLUMN IF NOT EXISTS product_debit_cost DECIMAL(20,10);

CREATE INDEX IF NOT EXISTS idx_usage_logs_product_id
    ON usage_logs (product_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_product_subscription_id
    ON usage_logs (product_subscription_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_product_created
    ON usage_logs (product_id, created_at);
CREATE INDEX IF NOT EXISTS idx_usage_logs_product_subscription_created
    ON usage_logs (product_subscription_id, created_at);
