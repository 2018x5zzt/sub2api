-- Product subscription hardening:
-- - explicit product family for same-family automatic product fallback
-- - user-level subscription-to-balance fallback settings
-- - subscription group to balance group fallback mapping
-- - standard-billing Team/Plus mixed pool at 1x multiplier

ALTER TABLE subscription_products
    ADD COLUMN IF NOT EXISTS product_family VARCHAR(64) NOT NULL DEFAULT 'default';

CREATE INDEX IF NOT EXISTS idx_subscription_products_family_sort
    ON subscription_products (product_family, sort_order, id)
    WHERE deleted_at IS NULL;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_limit_usd DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS subscription_balance_fallback_used_usd DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS balance_fallback_group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_groups_balance_fallback_group_id
    ON groups(balance_fallback_group_id)
    WHERE deleted_at IS NULL AND balance_fallback_group_id IS NOT NULL;

DO $$
DECLARE
    balance_group_id BIGINT;
BEGIN
    SELECT id
    INTO balance_group_id
    FROM groups
    WHERE deleted_at IS NULL
      AND subscription_type = 'standard'
      AND lower(name) IN (
          lower('Team/Plus 混合余额号池'),
          lower('Team/Plus Mixed Balance Pool'),
          lower('team/plus mixed balance pool')
      )
    ORDER BY id ASC
    LIMIT 1;

    IF balance_group_id IS NULL THEN
        INSERT INTO groups (
            name,
            description,
            platform,
            rate_multiplier,
            is_exclusive,
            status,
            subscription_type,
            default_validity_days,
            sort_order,
            created_at,
            updated_at
        )
        VALUES (
            'Team/Plus 混合余额号池',
            'Balance-mode Team/Plus mixed pool used when product subscription quota is exhausted.',
            'openai',
            1,
            FALSE,
            'active',
            'standard',
            30,
            0,
            NOW(),
            NOW()
        )
        RETURNING id INTO balance_group_id;
    ELSE
        UPDATE groups
        SET rate_multiplier = 1,
            subscription_type = 'standard',
            status = 'active',
            updated_at = NOW()
        WHERE id = balance_group_id;
    END IF;

    UPDATE groups
    SET balance_fallback_group_id = balance_group_id,
        updated_at = NOW()
    WHERE deleted_at IS NULL
      AND subscription_type = 'subscription'
      AND balance_fallback_group_id IS NULL
      AND lower(name) LIKE '%team%'
      AND lower(name) LIKE '%plus%';
END $$;

