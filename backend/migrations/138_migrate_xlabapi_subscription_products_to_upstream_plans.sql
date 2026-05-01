-- Migrate xlabapi product subscriptions into upstream subscription tables.
--
-- xlabapi introduced subscription_products, subscription_product_groups, and
-- user_product_subscriptions. Upstream v0.1.121 uses subscription_plans for
-- sale configuration and user_subscriptions for active entitlements. Keep the
-- legacy tables for audit, but copy their effective data into upstream tables.

DO $$
DECLARE
    has_product_tables BOOLEAN;
    has_redeem_product_id BOOLEAN;
BEGIN
    SELECT
        to_regclass('public.subscription_products') IS NOT NULL
        AND to_regclass('public.subscription_product_groups') IS NOT NULL
        AND to_regclass('public.user_product_subscriptions') IS NOT NULL
    INTO has_product_tables;

    IF NOT has_product_tables THEN
        RETURN;
    END IF;

    INSERT INTO subscription_plans (
        group_id,
        name,
        description,
        price,
        original_price,
        validity_days,
        validity_unit,
        features,
        product_name,
        for_sale,
        sort_order,
        created_at,
        updated_at
    )
    SELECT
        spg.group_id,
        LEFT(sp.name, 100),
        COALESCE(sp.description, ''),
        0,
        NULL,
        GREATEST(sp.default_validity_days, 1),
        'day',
        CONCAT(
            'Migrated from xlabapi subscription product ',
            sp.code,
            '; product daily limit USD=',
            COALESCE(sp.daily_limit_usd, 0)::TEXT,
            '; debit multiplier=',
            COALESCE(spg.debit_multiplier, 1)::TEXT
        ),
        LEFT(sp.name, 100),
        sp.status = 'active',
        COALESCE(sp.sort_order, 0) * 100 + COALESCE(spg.sort_order, 0),
        COALESCE(sp.created_at, NOW()),
        NOW()
    FROM subscription_products sp
    JOIN subscription_product_groups spg
        ON spg.product_id = sp.id
    JOIN groups g
        ON g.id = spg.group_id
    WHERE sp.deleted_at IS NULL
      AND spg.deleted_at IS NULL
      AND COALESCE(g.deleted_at, NULL) IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM subscription_plans existing
          WHERE existing.group_id = spg.group_id
            AND existing.product_name = LEFT(sp.name, 100)
            AND existing.features LIKE CONCAT('%xlabapi subscription product ', sp.code, '%')
      );

    INSERT INTO user_subscriptions (
        user_id,
        group_id,
        starts_at,
        expires_at,
        status,
        daily_window_start,
        weekly_window_start,
        monthly_window_start,
        daily_usage_usd,
        weekly_usage_usd,
        monthly_usage_usd,
        assigned_by,
        assigned_at,
        notes,
        created_at,
        updated_at
    )
    SELECT
        ups.user_id,
        spg.group_id,
        ups.starts_at,
        ups.expires_at,
        ups.status,
        ups.daily_window_start,
        ups.weekly_window_start,
        ups.monthly_window_start,
        ups.daily_usage_usd,
        ups.weekly_usage_usd,
        ups.monthly_usage_usd,
        ups.assigned_by,
        ups.assigned_at,
        CONCAT(
            COALESCE(NULLIF(ups.notes, ''), ''),
            CASE WHEN COALESCE(NULLIF(ups.notes, ''), '') = '' THEN '' ELSE E'\n' END,
            'Migrated from xlabapi product subscription #',
            ups.id::TEXT,
            ' product=',
            sp.code
        ),
        COALESCE(ups.created_at, NOW()),
        NOW()
    FROM user_product_subscriptions ups
    JOIN subscription_products sp
        ON sp.id = ups.product_id
    JOIN subscription_product_groups spg
        ON spg.product_id = sp.id
    JOIN users u
        ON u.id = ups.user_id
    JOIN groups g
        ON g.id = spg.group_id
    WHERE ups.deleted_at IS NULL
      AND sp.deleted_at IS NULL
      AND spg.deleted_at IS NULL
      AND COALESCE(g.deleted_at, NULL) IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM user_subscriptions existing
          WHERE existing.user_id = ups.user_id
            AND existing.group_id = spg.group_id
            AND existing.deleted_at IS NULL
      );

    WITH product_entitlements AS (
        SELECT
            ups.user_id,
            spg.group_id,
            MAX(ups.expires_at) AS max_expires_at,
            MAX(ups.daily_usage_usd) AS daily_usage_usd,
            MAX(ups.weekly_usage_usd) AS weekly_usage_usd,
            MAX(ups.monthly_usage_usd) AS monthly_usage_usd
        FROM user_product_subscriptions ups
        JOIN subscription_products sp
            ON sp.id = ups.product_id
        JOIN subscription_product_groups spg
            ON spg.product_id = sp.id
        WHERE ups.deleted_at IS NULL
          AND sp.deleted_at IS NULL
          AND spg.deleted_at IS NULL
        GROUP BY ups.user_id, spg.group_id
    )
    UPDATE user_subscriptions us
    SET
        expires_at = GREATEST(us.expires_at, product_entitlements.max_expires_at),
        daily_usage_usd = GREATEST(us.daily_usage_usd, product_entitlements.daily_usage_usd),
        weekly_usage_usd = GREATEST(us.weekly_usage_usd, product_entitlements.weekly_usage_usd),
        monthly_usage_usd = GREATEST(us.monthly_usage_usd, product_entitlements.monthly_usage_usd),
        updated_at = NOW()
    FROM product_entitlements
    WHERE us.user_id = product_entitlements.user_id
      AND us.group_id = product_entitlements.group_id
      AND us.deleted_at IS NULL;

    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'redeem_codes'
          AND column_name = 'product_id'
    )
    INTO has_redeem_product_id;

    IF has_redeem_product_id THEN
        UPDATE redeem_codes rc
        SET
            group_id = picked.group_id,
            validity_days = picked.default_validity_days
        FROM (
            SELECT DISTINCT ON (sp.id)
                sp.id AS product_id,
                sp.default_validity_days,
                spg.group_id
            FROM subscription_products sp
            JOIN subscription_product_groups spg
                ON spg.product_id = sp.id
            WHERE sp.deleted_at IS NULL
              AND spg.deleted_at IS NULL
            ORDER BY sp.id, spg.sort_order, spg.id
        ) picked
        WHERE rc.product_id = picked.product_id
          AND rc.group_id IS NULL;
    END IF;
END $$;
