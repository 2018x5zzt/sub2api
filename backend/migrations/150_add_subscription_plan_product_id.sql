ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_subscription_plans_product_id
    ON subscription_plans(product_id);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS subscription_product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_payment_orders_subscription_product_id
    ON payment_orders(subscription_product_id);
