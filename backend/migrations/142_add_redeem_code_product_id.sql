-- Add product subscription linkage for redeem/card codes.
--
-- This must stay separate from migration 140 because 140 has already been
-- applied in production and is protected by checksum verification.

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS product_id BIGINT REFERENCES subscription_products(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_redeem_codes_product_id
    ON redeem_codes (product_id);
