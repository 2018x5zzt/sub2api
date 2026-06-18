CREATE TABLE IF NOT EXISTS xlab_payment_orders (
    core_order_id BIGINT PRIMARY KEY,
    core_user_id BIGINT NOT NULL,
    out_trade_no TEXT,
    status TEXT NOT NULL,
    order_type TEXT,
    payment_type TEXT,
    amount NUMERIC(18,6),
    pay_amount NUMERIC(18,6),
    currency TEXT,
    provider_instance_id TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    source_created_at TIMESTAMPTZ,
    source_updated_at TIMESTAMPTZ,
    response_snapshot JSONB NOT NULL,
    synced_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_user_created
    ON xlab_payment_orders(core_user_id, created_at DESC, core_order_id DESC);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_user_status
    ON xlab_payment_orders(core_user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_user_order_type
    ON xlab_payment_orders(core_user_id, order_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_user_payment_type
    ON xlab_payment_orders(core_user_id, payment_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_out_trade_no
    ON xlab_payment_orders(out_trade_no);

CREATE INDEX IF NOT EXISTS idx_xlab_payment_orders_synced_at
    ON xlab_payment_orders(synced_at);
