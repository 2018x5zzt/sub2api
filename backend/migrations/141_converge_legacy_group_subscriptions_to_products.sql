-- Converge legacy per-group subscription entitlements back into xlabapi shared
-- product subscriptions.
--
-- Migration 138 expanded product subscriptions into user_subscriptions for
-- upstream compatibility. xlabapi needs the inverse at runtime: one shared
-- user_product_subscriptions quota pool consumed through several bound groups.
-- This migration is idempotent and soft-deletes only rows it can map to a
-- product subscription.

DO $$
DECLARE
    has_product_tables BOOLEAN;
    has_source_table BOOLEAN;
BEGIN
    SELECT
        to_regclass('public.subscription_products') IS NOT NULL
        AND to_regclass('public.subscription_product_groups') IS NOT NULL
        AND to_regclass('public.user_product_subscriptions') IS NOT NULL
    INTO has_product_tables;

    IF NOT has_product_tables THEN
        RETURN;
    END IF;

    SELECT to_regclass('public.product_subscription_migration_sources') IS NOT NULL
    INTO has_source_table;

    IF NOT has_source_table THEN
        CREATE TABLE product_subscription_migration_sources (
            id BIGSERIAL PRIMARY KEY,
            product_subscription_id BIGINT NOT NULL REFERENCES user_product_subscriptions(id) ON DELETE CASCADE,
            legacy_user_subscription_id BIGINT NOT NULL REFERENCES user_subscriptions(id) ON DELETE CASCADE,
            migration_batch VARCHAR(80) NOT NULL,
            legacy_group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
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
    END IF;

    CREATE UNIQUE INDEX IF NOT EXISTS product_subscription_migration_sources_legacy_unique
        ON product_subscription_migration_sources (legacy_user_subscription_id);
    CREATE INDEX IF NOT EXISTS idx_product_subscription_migration_sources_product
        ON product_subscription_migration_sources (product_subscription_id);

    ALTER TABLE user_subscriptions
        ADD COLUMN IF NOT EXISTS daily_carryover_in_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
        ADD COLUMN IF NOT EXISTS daily_carryover_remaining_usd DECIMAL(20,10) NOT NULL DEFAULT 0;

    ALTER TABLE user_product_subscriptions
        ADD COLUMN IF NOT EXISTS daily_carryover_in_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
        ADD COLUMN IF NOT EXISTS daily_carryover_remaining_usd DECIMAL(20,10) NOT NULL DEFAULT 0;

    -- Legacy sale groups still have active API keys. Bind them to their
    -- equivalent product so those keys continue to authenticate after their
    -- user_subscriptions rows are soft-deleted.
    INSERT INTO subscription_product_groups (
        product_id,
        group_id,
        debit_multiplier,
        status,
        sort_order,
        created_at,
        updated_at
    )
    SELECT
        sp.id,
        g.id,
        COALESCE(NULLIF(g.rate_multiplier, 0), 1),
        'active',
        100 + g.id,
        NOW(),
        NOW()
    FROM groups g
    JOIN subscription_products sp
        ON sp.deleted_at IS NULL
        AND sp.status = 'active'
        AND (
            (g.id = 32 AND sp.code = 'gpt_daily_150')
            OR (g.id = 22 AND sp.code = 'gpt_daily_225')
            OR (g.id = 23 AND sp.code = 'gpt_daily_450')
        )
    WHERE g.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM subscription_product_groups existing
          WHERE existing.product_id = sp.id
            AND existing.group_id = g.id
            AND existing.deleted_at IS NULL
      );

    WITH legacy_candidates AS (
        SELECT
            us.*,
            COALESCE(src_existing.product_subscription_id, ups_existing.id, user_product_pick.id) AS existing_product_subscription_id,
            target_product.id AS target_product_id,
            target_product.code AS target_product_code
        FROM user_subscriptions us
        JOIN groups g
            ON g.id = us.group_id
            AND g.deleted_at IS NULL
        LEFT JOIN product_subscription_migration_sources src_existing
            ON src_existing.legacy_user_subscription_id = us.id
        LEFT JOIN user_product_subscriptions ups_existing
            ON ups_existing.id = src_existing.product_subscription_id
            AND ups_existing.deleted_at IS NULL
        LEFT JOIN LATERAL (
            SELECT ups.id, ups.product_id
            FROM user_product_subscriptions ups
            JOIN subscription_products sp
                ON sp.id = ups.product_id
                AND sp.deleted_at IS NULL
                AND sp.status = 'active'
            WHERE ups.user_id = us.user_id
              AND ups.deleted_at IS NULL
              AND ups.status = 'active'
              AND ups.expires_at > NOW()
            ORDER BY
                ABS(EXTRACT(EPOCH FROM (ups.expires_at - us.expires_at))) ASC,
                ups.expires_at DESC,
                ups.id DESC
            LIMIT 1
        ) user_product_pick ON TRUE
        LEFT JOIN LATERAL (
            SELECT sp.id, sp.code
            FROM subscription_products sp
            WHERE sp.deleted_at IS NULL
              AND sp.status = 'active'
              AND (
                  (us.group_id IN (21, 33, 36) AND sp.id = COALESCE(ups_existing.product_id, user_product_pick.product_id))
                  OR (us.group_id = 21 AND COALESCE(ups_existing.product_id, user_product_pick.product_id) IS NULL AND sp.code = 'gpt_daily_45')
                  OR (us.group_id = 32 AND sp.code = 'gpt_daily_150')
                  OR (us.group_id = 22 AND sp.code = 'gpt_daily_225')
                  OR (us.group_id = 23 AND sp.code = 'gpt_daily_450')
              )
            ORDER BY sp.id
            LIMIT 1
        ) target_product ON TRUE
        WHERE us.deleted_at IS NULL
          AND us.status = 'active'
          AND us.expires_at > NOW()
          AND us.group_id IN (21, 22, 23, 32, 33, 36)
    ),
    inserted AS (
        INSERT INTO user_product_subscriptions (
            user_id,
            product_id,
            starts_at,
            expires_at,
            status,
            daily_window_start,
            weekly_window_start,
            monthly_window_start,
            daily_usage_usd,
            weekly_usage_usd,
            monthly_usage_usd,
            daily_carryover_in_usd,
            daily_carryover_remaining_usd,
            assigned_by,
            assigned_at,
            notes,
            created_at,
            updated_at
        )
        SELECT DISTINCT ON (lc.user_id, lc.target_product_id)
            lc.user_id,
            lc.target_product_id,
            lc.starts_at,
            lc.expires_at,
            lc.status,
            lc.daily_window_start,
            lc.weekly_window_start,
            lc.monthly_window_start,
            lc.daily_usage_usd,
            lc.weekly_usage_usd,
            lc.monthly_usage_usd,
            lc.daily_carryover_in_usd,
            lc.daily_carryover_remaining_usd,
            lc.assigned_by,
            COALESCE(lc.assigned_at, NOW()),
            CONCAT(
                COALESCE(NULLIF(lc.notes, ''), ''),
                CASE WHEN COALESCE(NULLIF(lc.notes, ''), '') = '' THEN '' ELSE E'\n' END,
                'Converged from legacy user_subscription #',
                lc.id::TEXT,
                ' group=',
                lc.group_id::TEXT,
                ' product=',
                lc.target_product_code
            ),
            COALESCE(lc.created_at, NOW()),
            NOW()
        FROM legacy_candidates lc
        WHERE lc.target_product_id IS NOT NULL
          AND lc.existing_product_subscription_id IS NULL
          AND NOT EXISTS (
              SELECT 1
              FROM user_product_subscriptions existing
              WHERE existing.user_id = lc.user_id
                AND existing.product_id = lc.target_product_id
                AND existing.deleted_at IS NULL
          )
        ORDER BY lc.user_id, lc.target_product_id, lc.expires_at DESC, lc.id DESC
        RETURNING id, user_id, product_id
    ),
    resolved AS (
        SELECT
            lc.*,
            COALESCE(
                lc.existing_product_subscription_id,
                inserted.id,
                ups_by_user_product.id
            ) AS product_subscription_id
        FROM legacy_candidates lc
        LEFT JOIN inserted
            ON inserted.user_id = lc.user_id
            AND inserted.product_id = lc.target_product_id
        LEFT JOIN user_product_subscriptions ups_by_user_product
            ON ups_by_user_product.user_id = lc.user_id
            AND ups_by_user_product.product_id = lc.target_product_id
            AND ups_by_user_product.deleted_at IS NULL
        WHERE lc.target_product_id IS NOT NULL
    ),
    merged_product_usage AS (
        SELECT
            product_subscription_id,
            MIN(starts_at) AS starts_at,
            MAX(expires_at) AS expires_at,
            MAX(daily_window_start) AS daily_window_start,
            MAX(weekly_window_start) AS weekly_window_start,
            MAX(monthly_window_start) AS monthly_window_start,
            MAX(daily_usage_usd) AS daily_usage_usd,
            MAX(weekly_usage_usd) AS weekly_usage_usd,
            MAX(monthly_usage_usd) AS monthly_usage_usd,
            MAX(daily_carryover_in_usd) AS daily_carryover_in_usd,
            MAX(daily_carryover_remaining_usd) AS daily_carryover_remaining_usd
        FROM resolved
        WHERE product_subscription_id IS NOT NULL
        GROUP BY product_subscription_id
    ),
    updated_products AS (
        UPDATE user_product_subscriptions ups
        SET
            starts_at = LEAST(ups.starts_at, merged_product_usage.starts_at),
            expires_at = GREATEST(ups.expires_at, merged_product_usage.expires_at),
            daily_window_start = COALESCE(ups.daily_window_start, merged_product_usage.daily_window_start),
            weekly_window_start = COALESCE(ups.weekly_window_start, merged_product_usage.weekly_window_start),
            monthly_window_start = COALESCE(ups.monthly_window_start, merged_product_usage.monthly_window_start),
            daily_usage_usd = GREATEST(ups.daily_usage_usd, merged_product_usage.daily_usage_usd),
            weekly_usage_usd = GREATEST(ups.weekly_usage_usd, merged_product_usage.weekly_usage_usd),
            monthly_usage_usd = GREATEST(ups.monthly_usage_usd, merged_product_usage.monthly_usage_usd),
            daily_carryover_in_usd = GREATEST(ups.daily_carryover_in_usd, merged_product_usage.daily_carryover_in_usd),
            daily_carryover_remaining_usd = GREATEST(ups.daily_carryover_remaining_usd, merged_product_usage.daily_carryover_remaining_usd),
            status = CASE
                WHEN ups.status <> 'active' AND merged_product_usage.expires_at > NOW() THEN 'active'
                ELSE ups.status
            END,
            updated_at = NOW()
        FROM merged_product_usage
        WHERE ups.id = merged_product_usage.product_subscription_id
          AND ups.deleted_at IS NULL
        RETURNING ups.id
    ),
    upserted_sources AS (
        INSERT INTO product_subscription_migration_sources (
            product_subscription_id,
            legacy_user_subscription_id,
            migration_batch,
            legacy_group_id,
            legacy_status,
            legacy_starts_at,
            legacy_expires_at,
            legacy_daily_usage_usd,
            legacy_weekly_usage_usd,
            legacy_monthly_usage_usd,
            legacy_daily_carryover_in_usd,
            legacy_daily_carryover_remaining_usd,
            created_at
        )
        SELECT
            r.product_subscription_id,
            r.id,
            'converge-legacy-group-subscriptions-20260502',
            r.group_id,
            r.status,
            r.starts_at,
            r.expires_at,
            r.daily_usage_usd,
            r.weekly_usage_usd,
            r.monthly_usage_usd,
            r.daily_carryover_in_usd,
            r.daily_carryover_remaining_usd,
            NOW()
        FROM resolved r
        WHERE r.product_subscription_id IS NOT NULL
        ON CONFLICT (legacy_user_subscription_id) DO NOTHING
        RETURNING legacy_user_subscription_id
    )
    UPDATE user_subscriptions us
    SET
        deleted_at = NOW(),
        status = 'expired',
        notes = CONCAT(
            COALESCE(NULLIF(us.notes, ''), ''),
            CASE WHEN COALESCE(NULLIF(us.notes, ''), '') = '' THEN '' ELSE E'\n' END,
            'Soft-deleted after convergence to product subscription on 2026-05-02'
        ),
        updated_at = NOW()
    FROM resolved r
    WHERE us.id = r.id
      AND r.product_subscription_id IS NOT NULL
      AND EXISTS (SELECT 1 FROM updated_products)
      AND us.deleted_at IS NULL;
END $$;
