-- xlabapi affiliate missing-inviter compensation (2026-05-25)
-- Goal:
--   1) Backfill user_affiliates.inviter_id for users created during the affected frontend-v2 window
--      when a reliable legacy inviter mapping exists (users.invited_by_user_id).
--   2) Compensate missing tiered affiliate rebates for historical balance orders that were completed
--      before the inviter binding was fixed, while keeping per-invitee cap/freeze behavior.
--
-- IMPORTANT:
--   - This is an ops runbook SQL, not a schema migration.
--   - Replace `affected_from` / `affected_to` before execution.
--   - Run PREVIEW first, validate samples and totals, then run EXECUTE in a low-traffic window.
--
-- ============================================================
-- 0) Parameters (edit before running)
-- ============================================================
-- Suggested window:
--   affected_from = frontend-v2(8081)上线且开始影响注册的时间
--   affected_to   = 修复上线时间
--
-- This script uses these constants in each section:
--   affected_from = TIMESTAMPTZ '2026-05-18 00:00:00+00'
--   affected_to   = TIMESTAMPTZ '2026-05-25 23:59:59+00'

-- ============================================================
-- 1) PREVIEW: users missing affiliate inviter in affected window
-- ============================================================
WITH params AS (
    SELECT
        TIMESTAMPTZ '2026-05-18 00:00:00+00' AS affected_from,
        TIMESTAMPTZ '2026-05-25 23:59:59+00' AS affected_to
),
missing_inviter AS (
    SELECT
        u.id AS invitee_user_id,
        u.email AS invitee_email,
        u.created_at AS invitee_created_at,
        u.invited_by_user_id AS legacy_inviter_user_id,
        ua.inviter_id AS affiliate_inviter_user_id
    FROM users u
    JOIN user_affiliates ua ON ua.user_id = u.id
    CROSS JOIN params p
    WHERE u.created_at >= p.affected_from
      AND u.created_at <= p.affected_to
      AND ua.inviter_id IS NULL
)
SELECT
    COUNT(*) AS missing_affiliate_inviter_count,
    COUNT(*) FILTER (WHERE legacy_inviter_user_id IS NOT NULL) AS can_infer_from_legacy_count,
    COUNT(*) FILTER (WHERE legacy_inviter_user_id IS NULL) AS cannot_infer_count
FROM missing_inviter;

-- Preview unresolved users (need manual mapping if must compensate them):
WITH params AS (
    SELECT
        TIMESTAMPTZ '2026-05-18 00:00:00+00' AS affected_from,
        TIMESTAMPTZ '2026-05-25 23:59:59+00' AS affected_to
)
SELECT
    u.id AS invitee_user_id,
    u.email AS invitee_email,
    u.created_at AS invitee_created_at
FROM users u
JOIN user_affiliates ua ON ua.user_id = u.id
CROSS JOIN params p
WHERE u.created_at >= p.affected_from
  AND u.created_at <= p.affected_to
  AND ua.inviter_id IS NULL
  AND u.invited_by_user_id IS NULL
ORDER BY u.created_at DESC
LIMIT 200;

-- ============================================================
-- 2) PREVIEW: compensation totals for inferable users
-- ============================================================
WITH params AS (
    SELECT
        TIMESTAMPTZ '2026-05-18 00:00:00+00' AS affected_from,
        TIMESTAMPTZ '2026-05-25 23:59:59+00' AS affected_to,
        NOW() AS fixed_at
),
bindings AS (
    SELECT
        ua_invitee.user_id AS invitee_user_id,
        u.invited_by_user_id AS inviter_id,
        p.fixed_at
    FROM params p
    JOIN users u
      ON u.created_at >= p.affected_from
     AND u.created_at <= p.affected_to
     AND u.invited_by_user_id IS NOT NULL
     AND u.invited_by_user_id <> u.id
    JOIN user_affiliates ua_invitee
      ON ua_invitee.user_id = u.id
     AND ua_invitee.inviter_id IS NULL
    JOIN user_affiliates ua_inviter
      ON ua_inviter.user_id = u.invited_by_user_id
),
setting_numbers AS (
    SELECT
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_rate'
                 AND value ~ '^-?[0-9]+(\\.[0-9]+)?$'
                THEN value::numeric
            END
        ), 0) AS global_rate_percent,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_freeze_hours'
                 AND value ~ '^-?[0-9]+$'
                THEN value::int
            END
        ), 0) AS freeze_hours,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_per_invitee_cap'
                 AND value ~ '^-?[0-9]+(\\.[0-9]+)?$'
                THEN value::numeric
            END
        ), 0) AS per_invitee_cap
    FROM settings
    WHERE key IN (
        'affiliate_rebate_rate',
        'affiliate_rebate_freeze_hours',
        'affiliate_rebate_per_invitee_cap'
    )
),
tier_source AS (
    SELECT
        CASE
            WHEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]') ~ '^\\s*\\['
            THEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]')::jsonb
            ELSE '[]'::jsonb
        END AS tiers
    FROM settings
),
tier_rows AS (
    SELECT
        GREATEST(
            0,
            COALESCE(
                CASE
                    WHEN elem->>'min_effective_invitees' ~ '^-?[0-9]+$'
                    THEN (elem->>'min_effective_invitees')::int
                END,
                0
            )
        ) AS min_effective_invitees,
        LEAST(
            100,
            GREATEST(
                0,
                COALESCE(
                    CASE
                        WHEN elem->>'rebate_rate' ~ '^-?[0-9]+(\\.[0-9]+)?$'
                        THEN (elem->>'rebate_rate')::numeric
                    END,
                    0
                )
            )
        ) AS rate_percent
    FROM tier_source ts,
         LATERAL jsonb_array_elements(ts.tiers) AS elem
    WHERE jsonb_typeof(elem) = 'object'
),
effective_invitees AS (
    SELECT
        ua.inviter_id AS inviter_id,
        COUNT(DISTINCT ua.user_id) AS effective_invitee_count
    FROM user_affiliates ua
    WHERE ua.inviter_id IN (SELECT DISTINCT inviter_id FROM bindings)
      AND (
            EXISTS (
                SELECT 1
                FROM payment_orders po
                WHERE po.user_id = ua.user_id
                  AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
                  AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
                  AND (po.amount > 0 OR po.pay_amount > 0)
            )
            OR EXISTS (
                SELECT 1
                FROM redeem_codes rc
                WHERE rc.used_by = ua.user_id
                  AND rc.status = 'used'
                  AND rc.used_at IS NOT NULL
                  AND rc.source_type IN ('commercial', 'system_grant')
                  AND rc.type IN ('balance', 'subscription')
                  AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
            )
      )
    GROUP BY ua.inviter_id
),
inviter_rates AS (
    SELECT
        i.inviter_id,
        LEAST(
            100,
            GREATEST(
                0,
                COALESCE(
                    ua.aff_rebate_rate_percent,
                    (
                        SELECT tr.rate_percent
                        FROM tier_rows tr
                        WHERE tr.min_effective_invitees <= COALESCE(ei.effective_invitee_count, 0)
                        ORDER BY tr.min_effective_invitees DESC
                        LIMIT 1
                    ),
                    sn.global_rate_percent
                )
            )
        ) AS rebate_rate_percent,
        GREATEST(0, sn.freeze_hours) AS freeze_hours,
        GREATEST(0, sn.per_invitee_cap) AS per_invitee_cap
    FROM (SELECT DISTINCT inviter_id FROM bindings) AS i
    JOIN user_affiliates ua ON ua.user_id = i.inviter_id
    LEFT JOIN effective_invitees ei ON ei.inviter_id = i.inviter_id
    CROSS JOIN setting_numbers sn
),
orders_base AS (
    SELECT
        po.id AS order_id,
        po.user_id AS invitee_user_id,
        b.inviter_id,
        COALESCE(po.completed_at, po.paid_at, po.created_at) AS order_at,
        po.amount::numeric AS order_amount,
        r.rebate_rate_percent,
        r.freeze_hours,
        r.per_invitee_cap
    FROM bindings b
    JOIN inviter_rates r ON r.inviter_id = b.inviter_id
    JOIN payment_orders po ON po.user_id = b.invitee_user_id
    WHERE po.order_type = 'balance'
      AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
      AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
      AND po.amount > 0
      AND COALESCE(po.completed_at, po.paid_at, po.created_at) < b.fixed_at
      AND NOT EXISTS (
            SELECT 1
            FROM user_affiliate_ledger ual
            WHERE ual.action = 'accrue'
              AND ual.user_id = b.inviter_id
              AND ual.source_user_id = b.invitee_user_id
              AND ual.source_order_id = po.id
      )
),
orders_with_calc AS (
    SELECT
        ob.*,
        ROUND((ob.order_amount * ob.rebate_rate_percent / 100.0)::numeric, 8) AS calc_rebate,
        SUM(ROUND((ob.order_amount * ob.rebate_rate_percent / 100.0)::numeric, 8))
            OVER (
                PARTITION BY ob.inviter_id, ob.invitee_user_id
                ORDER BY ob.order_at, ob.order_id
            ) AS running_calc_rebate
    FROM orders_base ob
),
existing_rebate_by_pair AS (
    SELECT
        ual.user_id AS inviter_id,
        ual.source_user_id AS invitee_user_id,
        COALESCE(SUM(ual.amount), 0)::numeric AS existing_rebate
    FROM user_affiliate_ledger ual
    WHERE ual.action = 'accrue'
    GROUP BY ual.user_id, ual.source_user_id
),
orders_compensation AS (
    SELECT
        owc.order_id,
        owc.inviter_id,
        owc.invitee_user_id,
        owc.order_at,
        owc.order_amount,
        owc.rebate_rate_percent,
        owc.freeze_hours,
        CASE
            WHEN owc.calc_rebate <= 0 THEN 0::numeric
            WHEN owc.per_invitee_cap <= 0 THEN owc.calc_rebate
            ELSE GREATEST(
                LEAST(
                    owc.per_invitee_cap,
                    COALESCE(erp.existing_rebate, 0) + owc.running_calc_rebate
                )
                -
                LEAST(
                    owc.per_invitee_cap,
                    COALESCE(erp.existing_rebate, 0) + owc.running_calc_rebate - owc.calc_rebate
                ),
                0::numeric
            )
        END AS rebate_amount
    FROM orders_with_calc owc
    LEFT JOIN existing_rebate_by_pair erp
      ON erp.inviter_id = owc.inviter_id
     AND erp.invitee_user_id = owc.invitee_user_id
)
SELECT
    COUNT(*) FILTER (WHERE rebate_amount > 0) AS orders_to_compensate,
    COUNT(DISTINCT invitee_user_id) FILTER (WHERE rebate_amount > 0) AS invitees_to_compensate,
    COUNT(DISTINCT inviter_id) FILTER (WHERE rebate_amount > 0) AS inviters_to_credit,
    COALESCE(SUM(rebate_amount) FILTER (WHERE rebate_amount > 0), 0)::numeric(20,8) AS total_rebate_amount,
    COALESCE(SUM(rebate_amount) FILTER (WHERE rebate_amount > 0 AND freeze_hours > 0), 0)::numeric(20,8) AS total_frozen_amount,
    COALESCE(SUM(rebate_amount) FILTER (WHERE rebate_amount > 0 AND freeze_hours <= 0), 0)::numeric(20,8) AS total_available_amount
FROM orders_compensation;

-- ============================================================
-- 3) EXECUTE: bind inviter + compensate rebate
-- ============================================================
BEGIN;

-- 3.1 Backfill inviter bindings snapshot table
CREATE TABLE IF NOT EXISTS affiliate_bind_backfill_20260525 (
    invitee_user_id BIGINT PRIMARY KEY,
    old_inviter_id BIGINT NULL,
    new_inviter_id BIGINT NOT NULL,
    fixed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

WITH params AS (
    SELECT
        TIMESTAMPTZ '2026-05-18 00:00:00+00' AS affected_from,
        TIMESTAMPTZ '2026-05-25 23:59:59+00' AS affected_to,
        NOW() AS fixed_at
),
binding_candidates AS (
    SELECT
        ua_invitee.user_id AS invitee_user_id,
        ua_invitee.inviter_id AS old_inviter_id,
        u.invited_by_user_id AS new_inviter_id,
        p.fixed_at
    FROM params p
    JOIN users u
      ON u.created_at >= p.affected_from
     AND u.created_at <= p.affected_to
     AND u.invited_by_user_id IS NOT NULL
     AND u.invited_by_user_id <> u.id
    JOIN user_affiliates ua_invitee
      ON ua_invitee.user_id = u.id
     AND ua_invitee.inviter_id IS NULL
    JOIN user_affiliates ua_inviter
      ON ua_inviter.user_id = u.invited_by_user_id
)
INSERT INTO affiliate_bind_backfill_20260525 (invitee_user_id, old_inviter_id, new_inviter_id, fixed_at)
SELECT
    bc.invitee_user_id,
    bc.old_inviter_id,
    bc.new_inviter_id,
    bc.fixed_at
FROM binding_candidates bc
ON CONFLICT (invitee_user_id) DO NOTHING;

-- 3.2 Apply missing inviter binding
UPDATE user_affiliates ua
SET inviter_id = b.new_inviter_id,
    updated_at = b.fixed_at
FROM affiliate_bind_backfill_20260525 b
WHERE ua.user_id = b.invitee_user_id
  AND ua.inviter_id IS NULL;

-- 3.3 Recompute inviter aff_count for affected inviters
WITH affected_inviters AS (
    SELECT DISTINCT new_inviter_id AS inviter_id
    FROM affiliate_bind_backfill_20260525
),
recomputed AS (
    SELECT
        ua.inviter_id AS inviter_id,
        COUNT(*)::int AS invitee_count
    FROM user_affiliates ua
    WHERE ua.inviter_id IN (SELECT inviter_id FROM affected_inviters)
    GROUP BY ua.inviter_id
)
UPDATE user_affiliates ua
SET aff_count = COALESCE(r.invitee_count, 0),
    updated_at = NOW()
FROM recomputed r
WHERE ua.user_id = r.inviter_id;

-- 3.4 Build compensation order snapshot table (idempotent by order_id)
CREATE TABLE IF NOT EXISTS affiliate_rebate_backfill_20260525 (
    order_id BIGINT PRIMARY KEY,
    inviter_user_id BIGINT NOT NULL,
    invitee_user_id BIGINT NOT NULL,
    rebate_amount DECIMAL(20,8) NOT NULL,
    rebate_rate_percent DECIMAL(10,4) NOT NULL,
    freeze_hours INTEGER NOT NULL,
    order_amount DECIMAL(20,8) NOT NULL,
    order_at TIMESTAMPTZ NOT NULL,
    fixed_at TIMESTAMPTZ NOT NULL,
    ledger_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

WITH setting_numbers AS (
    SELECT
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_rate'
                 AND value ~ '^-?[0-9]+(\\.[0-9]+)?$'
                THEN value::numeric
            END
        ), 0) AS global_rate_percent,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_freeze_hours'
                 AND value ~ '^-?[0-9]+$'
                THEN value::int
            END
        ), 0) AS freeze_hours,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_per_invitee_cap'
                 AND value ~ '^-?[0-9]+(\\.[0-9]+)?$'
                THEN value::numeric
            END
        ), 0) AS per_invitee_cap
    FROM settings
    WHERE key IN (
        'affiliate_rebate_rate',
        'affiliate_rebate_freeze_hours',
        'affiliate_rebate_per_invitee_cap'
    )
),
tier_source AS (
    SELECT
        CASE
            WHEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]') ~ '^\\s*\\['
            THEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]')::jsonb
            ELSE '[]'::jsonb
        END AS tiers
    FROM settings
),
tier_rows AS (
    SELECT
        GREATEST(
            0,
            COALESCE(
                CASE
                    WHEN elem->>'min_effective_invitees' ~ '^-?[0-9]+$'
                    THEN (elem->>'min_effective_invitees')::int
                END,
                0
            )
        ) AS min_effective_invitees,
        LEAST(
            100,
            GREATEST(
                0,
                COALESCE(
                    CASE
                        WHEN elem->>'rebate_rate' ~ '^-?[0-9]+(\\.[0-9]+)?$'
                        THEN (elem->>'rebate_rate')::numeric
                    END,
                    0
                )
            )
        ) AS rate_percent
    FROM tier_source ts,
         LATERAL jsonb_array_elements(ts.tiers) AS elem
    WHERE jsonb_typeof(elem) = 'object'
),
bindings AS (
    SELECT
        invitee_user_id,
        new_inviter_id AS inviter_id,
        fixed_at
    FROM affiliate_bind_backfill_20260525
),
effective_invitees AS (
    SELECT
        ua.inviter_id AS inviter_id,
        COUNT(DISTINCT ua.user_id) AS effective_invitee_count
    FROM user_affiliates ua
    WHERE ua.inviter_id IN (SELECT DISTINCT inviter_id FROM bindings)
      AND (
            EXISTS (
                SELECT 1
                FROM payment_orders po
                WHERE po.user_id = ua.user_id
                  AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
                  AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
                  AND (po.amount > 0 OR po.pay_amount > 0)
            )
            OR EXISTS (
                SELECT 1
                FROM redeem_codes rc
                WHERE rc.used_by = ua.user_id
                  AND rc.status = 'used'
                  AND rc.used_at IS NOT NULL
                  AND rc.source_type IN ('commercial', 'system_grant')
                  AND rc.type IN ('balance', 'subscription')
                  AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
            )
      )
    GROUP BY ua.inviter_id
),
inviter_rates AS (
    SELECT
        i.inviter_id,
        LEAST(
            100,
            GREATEST(
                0,
                COALESCE(
                    ua.aff_rebate_rate_percent,
                    (
                        SELECT tr.rate_percent
                        FROM tier_rows tr
                        WHERE tr.min_effective_invitees <= COALESCE(ei.effective_invitee_count, 0)
                        ORDER BY tr.min_effective_invitees DESC
                        LIMIT 1
                    ),
                    sn.global_rate_percent
                )
            )
        ) AS rebate_rate_percent,
        GREATEST(0, sn.freeze_hours) AS freeze_hours,
        GREATEST(0, sn.per_invitee_cap) AS per_invitee_cap
    FROM (SELECT DISTINCT inviter_id FROM bindings) AS i
    JOIN user_affiliates ua ON ua.user_id = i.inviter_id
    LEFT JOIN effective_invitees ei ON ei.inviter_id = i.inviter_id
    CROSS JOIN setting_numbers sn
),
orders_base AS (
    SELECT
        po.id AS order_id,
        po.user_id AS invitee_user_id,
        b.inviter_id,
        COALESCE(po.completed_at, po.paid_at, po.created_at) AS order_at,
        po.amount::numeric AS order_amount,
        r.rebate_rate_percent,
        r.freeze_hours,
        r.per_invitee_cap,
        b.fixed_at
    FROM bindings b
    JOIN inviter_rates r ON r.inviter_id = b.inviter_id
    JOIN payment_orders po ON po.user_id = b.invitee_user_id
    WHERE po.order_type = 'balance'
      AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
      AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
      AND po.amount > 0
      AND COALESCE(po.completed_at, po.paid_at, po.created_at) < b.fixed_at
      AND NOT EXISTS (
            SELECT 1
            FROM user_affiliate_ledger ual
            WHERE ual.action = 'accrue'
              AND ual.user_id = b.inviter_id
              AND ual.source_user_id = b.invitee_user_id
              AND ual.source_order_id = po.id
      )
      AND NOT EXISTS (
            SELECT 1
            FROM affiliate_rebate_backfill_20260525 ab
            WHERE ab.order_id = po.id
      )
),
orders_with_calc AS (
    SELECT
        ob.*,
        ROUND((ob.order_amount * ob.rebate_rate_percent / 100.0)::numeric, 8) AS calc_rebate,
        SUM(ROUND((ob.order_amount * ob.rebate_rate_percent / 100.0)::numeric, 8))
            OVER (
                PARTITION BY ob.inviter_id, ob.invitee_user_id
                ORDER BY ob.order_at, ob.order_id
            ) AS running_calc_rebate
    FROM orders_base ob
),
existing_rebate_by_pair AS (
    SELECT
        ual.user_id AS inviter_id,
        ual.source_user_id AS invitee_user_id,
        COALESCE(SUM(ual.amount), 0)::numeric AS existing_rebate
    FROM user_affiliate_ledger ual
    WHERE ual.action = 'accrue'
    GROUP BY ual.user_id, ual.source_user_id
),
orders_compensation AS (
    SELECT
        owc.order_id,
        owc.inviter_id,
        owc.invitee_user_id,
        owc.order_at,
        owc.order_amount,
        owc.rebate_rate_percent,
        owc.freeze_hours,
        owc.fixed_at,
        CASE
            WHEN owc.calc_rebate <= 0 THEN 0::numeric
            WHEN owc.per_invitee_cap <= 0 THEN owc.calc_rebate
            ELSE GREATEST(
                LEAST(
                    owc.per_invitee_cap,
                    COALESCE(erp.existing_rebate, 0) + owc.running_calc_rebate
                )
                -
                LEAST(
                    owc.per_invitee_cap,
                    COALESCE(erp.existing_rebate, 0) + owc.running_calc_rebate - owc.calc_rebate
                ),
                0::numeric
            )
        END AS rebate_amount
    FROM orders_with_calc owc
    LEFT JOIN existing_rebate_by_pair erp
      ON erp.inviter_id = owc.inviter_id
     AND erp.invitee_user_id = owc.invitee_user_id
)
INSERT INTO affiliate_rebate_backfill_20260525 (
    order_id,
    inviter_user_id,
    invitee_user_id,
    rebate_amount,
    rebate_rate_percent,
    freeze_hours,
    order_amount,
    order_at,
    fixed_at
)
SELECT
    oc.order_id,
    oc.inviter_id,
    oc.invitee_user_id,
    oc.rebate_amount::numeric(20,8),
    oc.rebate_rate_percent::numeric(10,4),
    oc.freeze_hours,
    oc.order_amount::numeric(20,8),
    oc.order_at,
    oc.fixed_at
FROM orders_compensation oc
WHERE oc.rebate_amount > 0
ON CONFLICT (order_id) DO NOTHING;

-- 3.5 Apply inviter quota deltas for rows not yet materialized to ledger
WITH pending AS (
    SELECT *
    FROM affiliate_rebate_backfill_20260525
    WHERE ledger_id IS NULL
),
agg AS (
    SELECT
        inviter_user_id,
        COALESCE(SUM(CASE WHEN freeze_hours > 0 THEN rebate_amount ELSE 0 END), 0)::numeric AS frozen_delta,
        COALESCE(SUM(CASE WHEN freeze_hours <= 0 THEN rebate_amount ELSE 0 END), 0)::numeric AS available_delta,
        COALESCE(SUM(rebate_amount), 0)::numeric AS history_delta
    FROM pending
    GROUP BY inviter_user_id
)
UPDATE user_affiliates ua
SET aff_frozen_quota = ua.aff_frozen_quota + agg.frozen_delta,
    aff_quota = ua.aff_quota + agg.available_delta,
    aff_history_quota = ua.aff_history_quota + agg.history_delta,
    updated_at = NOW()
FROM agg
WHERE ua.user_id = agg.inviter_user_id;

-- 3.6 Insert affiliate ledger rows and map ledger_id back to snapshot rows
WITH pending AS (
    SELECT *
    FROM affiliate_rebate_backfill_20260525
    WHERE ledger_id IS NULL
),
inserted AS (
    INSERT INTO user_affiliate_ledger (
        user_id,
        action,
        amount,
        source_user_id,
        source_order_id,
        frozen_until,
        created_at,
        updated_at
    )
    SELECT
        p.inviter_user_id,
        'accrue',
        p.rebate_amount,
        p.invitee_user_id,
        p.order_id,
        CASE
            WHEN p.freeze_hours > 0 THEN NOW() + make_interval(hours => p.freeze_hours)
            ELSE NULL
        END,
        NOW(),
        NOW()
    FROM pending p
    RETURNING id, source_order_id
)
UPDATE affiliate_rebate_backfill_20260525 b
SET ledger_id = i.id
FROM inserted i
WHERE b.order_id = i.source_order_id
  AND b.ledger_id IS NULL;

-- 3.7 Add payment audit logs for compensated orders
INSERT INTO payment_audit_logs (order_id, action, detail, operator, created_at)
SELECT
    b.order_id::text,
    'AFFILIATE_REBATE_APPLIED',
    jsonb_build_object(
        'source', 'manual_affiliate_backfill_20260525',
        'inviteeUserID', b.invitee_user_id,
        'inviterUserID', b.inviter_user_id,
        'rebateAmount', b.rebate_amount,
        'rebateRatePercent', b.rebate_rate_percent
    )::text,
    'system',
    NOW()
FROM affiliate_rebate_backfill_20260525 b
WHERE b.ledger_id IS NOT NULL
ON CONFLICT (order_id, action) DO NOTHING;

COMMIT;

-- ============================================================
-- 4) POST-CHECK
-- ============================================================
SELECT
    COUNT(*) AS bound_invitees
FROM affiliate_bind_backfill_20260525;

SELECT
    COUNT(*) AS compensated_orders,
    COUNT(*) FILTER (WHERE freeze_hours > 0) AS compensated_orders_frozen,
    COALESCE(SUM(rebate_amount), 0)::numeric(20,8) AS compensated_total_rebate
FROM affiliate_rebate_backfill_20260525
WHERE ledger_id IS NOT NULL;

-- ============================================================
-- 5) ROLLBACK PLAN (if needed)
-- ============================================================
-- BEGIN;
--
-- -- 5.1 Revert audit log rows generated by this backfill marker
-- DELETE FROM payment_audit_logs pal
-- WHERE pal.action = 'AFFILIATE_REBATE_APPLIED'
--   AND pal.order_id IN (
--       SELECT order_id::text
--       FROM affiliate_rebate_backfill_20260525
--       WHERE ledger_id IS NOT NULL
--   )
--   AND pal.detail LIKE '%manual_affiliate_backfill_20260525%';
--
-- -- 5.2 Revert inviter quota deltas (based on snapshot rows that were materialized)
-- WITH applied AS (
--     SELECT *
--     FROM affiliate_rebate_backfill_20260525
--     WHERE ledger_id IS NOT NULL
-- ),
-- agg AS (
--     SELECT
--         inviter_user_id,
--         COALESCE(SUM(CASE WHEN freeze_hours > 0 THEN rebate_amount ELSE 0 END), 0)::numeric AS frozen_delta,
--         COALESCE(SUM(CASE WHEN freeze_hours <= 0 THEN rebate_amount ELSE 0 END), 0)::numeric AS available_delta,
--         COALESCE(SUM(rebate_amount), 0)::numeric AS history_delta
--     FROM applied
--     GROUP BY inviter_user_id
-- )
-- UPDATE user_affiliates ua
-- SET aff_frozen_quota = GREATEST(0, ua.aff_frozen_quota - agg.frozen_delta),
--     aff_quota = GREATEST(0, ua.aff_quota - agg.available_delta),
--     aff_history_quota = GREATEST(0, ua.aff_history_quota - agg.history_delta),
--     updated_at = NOW()
-- FROM agg
-- WHERE ua.user_id = agg.inviter_user_id;
--
-- -- 5.3 Delete inserted affiliate ledger rows
-- DELETE FROM user_affiliate_ledger ual
-- WHERE ual.id IN (
--     SELECT ledger_id
--     FROM affiliate_rebate_backfill_20260525
--     WHERE ledger_id IS NOT NULL
-- );
--
-- -- 5.4 Revert inviter binding
-- UPDATE user_affiliates ua
-- SET inviter_id = b.old_inviter_id,
--     updated_at = NOW()
-- FROM affiliate_bind_backfill_20260525 b
-- WHERE ua.user_id = b.invitee_user_id;
--
-- -- 5.5 Recompute aff_count for affected inviters (new and old)
-- WITH affected AS (
--     SELECT DISTINCT new_inviter_id AS inviter_id FROM affiliate_bind_backfill_20260525
--     UNION
--     SELECT DISTINCT old_inviter_id AS inviter_id
--     FROM affiliate_bind_backfill_20260525
--     WHERE old_inviter_id IS NOT NULL
-- ),
-- recalc AS (
--     SELECT ua.inviter_id, COUNT(*)::int AS invitee_count
--     FROM user_affiliates ua
--     WHERE ua.inviter_id IN (SELECT inviter_id FROM affected)
--     GROUP BY ua.inviter_id
-- )
-- UPDATE user_affiliates ua
-- SET aff_count = COALESCE(r.invitee_count, 0),
--     updated_at = NOW()
-- FROM recalc r
-- WHERE ua.user_id = r.inviter_id;
--
-- COMMIT;

