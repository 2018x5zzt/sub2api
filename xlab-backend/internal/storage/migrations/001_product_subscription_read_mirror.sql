CREATE TABLE IF NOT EXISTS xlab_subscription_products (
    core_product_id BIGINT PRIMARY KEY,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    product_family TEXT NOT NULL DEFAULT 'gpt',
    daily_limit_usd NUMERIC(18,6),
    weekly_limit_usd NUMERIC(18,6),
    monthly_limit_usd NUMERIC(18,6),
    daily_carryover_enabled BOOLEAN NOT NULL DEFAULT false,
    daily_carryover_limit_usd NUMERIC(18,6),
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_products_status
    ON xlab_subscription_products(status);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_products_synced_at
    ON xlab_subscription_products(synced_at);

CREATE TABLE IF NOT EXISTS xlab_subscription_product_groups (
    core_binding_id BIGINT PRIMARY KEY,
    core_product_id BIGINT NOT NULL REFERENCES xlab_subscription_products(core_product_id) ON DELETE CASCADE,
    core_group_id BIGINT NOT NULL,
    group_name TEXT NOT NULL,
    group_platform TEXT,
    balance_fallback_group_id BIGINT,
    balance_fallback_group_name TEXT,
    debit_multiplier NUMERIC(18,6) NOT NULL DEFAULT 1,
    status TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_product_groups_product
    ON xlab_subscription_product_groups(core_product_id, sort_order, core_group_id);

CREATE INDEX IF NOT EXISTS idx_xlab_subscription_product_groups_group
    ON xlab_subscription_product_groups(core_group_id);

CREATE TABLE IF NOT EXISTS xlab_user_product_subscriptions (
    core_subscription_id BIGINT PRIMARY KEY,
    core_user_id BIGINT NOT NULL,
    core_product_id BIGINT NOT NULL REFERENCES xlab_subscription_products(core_product_id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    daily_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    weekly_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    monthly_usage_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    daily_limit_usd NUMERIC(18,6),
    weekly_limit_usd NUMERIC(18,6),
    monthly_limit_usd NUMERIC(18,6),
    daily_carryover_in_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    daily_carryover_remaining_usd NUMERIC(18,6) NOT NULL DEFAULT 0,
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_user_active
    ON xlab_user_product_subscriptions(core_user_id, status, expires_at);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_product
    ON xlab_user_product_subscriptions(core_product_id);

CREATE INDEX IF NOT EXISTS idx_xlab_user_product_subscriptions_synced_at
    ON xlab_user_product_subscriptions(synced_at);

CREATE TABLE IF NOT EXISTS xlab_sync_state (
    source_name TEXT PRIMARY KEY,
    last_success_at TIMESTAMPTZ,
    last_watermark TEXT,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,
    row_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
