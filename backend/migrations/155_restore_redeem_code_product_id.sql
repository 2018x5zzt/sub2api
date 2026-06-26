-- Restore product subscription linkage on redeem codes (xlab product redeem codes).
--
-- The original migration 142_add_redeem_code_product_id.sql was dropped from the
-- tree during the 0.1.123 kernel re-roll, so fresh installs no longer create the
-- redeem_codes.product_id column that the ent schema and product-redeem flow rely on.
-- Existing production databases already have this column (142 applied), so the
-- IF NOT EXISTS guards make this migration a safe no-op there.

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_product_id
    ON redeem_codes (product_id);
