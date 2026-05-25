-- xlabapi affiliate rebate compensation (redeem-delta model, 2026-05-25)
-- Goal:
--   Compensate inviters that should have received level-1 affiliate rebate
--   but currently have lower accrued totals than expected from qualifying redeemed codes.
--
-- Why this model:
--   - Current production has redeem_codes + affiliate ledger data, but payment_orders is empty.
--   - Historical affiliate accrual rows mostly have source_order_id = NULL.
--   - So we compute expected rebate from redeemed codes and compensate pair-level delta.
--
-- Qualifying redeemed codes:
--   type in ('balance','subscription')
--   source_type in ('commercial','system_grant')
--   status='used' and used_by/used_at present
--   effective value rule: value > 0 OR group_id IS NOT NULL OR validity_days > 0
--
-- Rebate rate model:
--   1) inviter aff_rebate_rate_percent (custom) if set
--   2) else tiered rate (affiliate_rebate_tiers) by effective invitee count
--   3) else global affiliate_rebate_rate
--   (fallback defaults aligned with backend service:
--      global rate=20, tiers=AffiliateRebateTiersDefault)
--
-- Compensation policy:
--   - Compute pair-level expected total under current rebate settings snapshot.
--   - Compare with current accrued_total in user_affiliate_ledger(action='accrue').
--   - Insert delta only when expected_total > accrued_total.
--   - Respect freeze policy:
--       freeze_hours > 0  => credit aff_frozen_quota (+ history), ledger.frozen_until populated.
--       freeze_hours <= 0 => credit aff_quota (+ history), ledger.frozen_until NULL.
--   - Insert one synthetic accrue ledger row per (inviter, invitee) pair.
--
-- ============================================================
-- 1) PREVIEW
-- ============================================================
WITH setting_numbers AS (
    SELECT
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_rate'
                 AND value ~ '^-?[0-9]+([.][0-9]+)?$'
                THEN LEAST(100, GREATEST(0, value::numeric))
            END
        ), 20.0) AS global_rate_percent,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_per_invitee_cap'
                 AND value ~ '^-?[0-9]+([.][0-9]+)?$'
                THEN GREATEST(0, value::numeric)
            END
        ), 0) AS per_invitee_cap,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_freeze_hours'
                 AND value ~ '^-?[0-9]+$'
                THEN GREATEST(0, value::int)
            END
        ), 0) AS freeze_hours,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_enabled'
                THEN CASE WHEN LOWER(TRIM(value)) = 'true' THEN 1 ELSE 0 END
            END
        ), 0) AS affiliate_enabled
    FROM settings
    WHERE key IN (
        'affiliate_rebate_rate',
        'affiliate_rebate_per_invitee_cap',
        'affiliate_rebate_freeze_hours',
        'affiliate_enabled'
    )
),
tier_source AS (
    SELECT
        CASE
            WHEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]') ~ '^[[:space:]]*[[]'
            THEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]')::jsonb
            ELSE '[{"min_effective_invitees":1,"rebate_rate":5},{"min_effective_invitees":3,"rebate_rate":8},{"min_effective_invitees":10,"rebate_rate":12},{"min_effective_invitees":30,"rebate_rate":15},{"min_effective_invitees":50,"rebate_rate":20}]'::jsonb
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
                        WHEN elem->>'rebate_rate' ~ '^-?[0-9]+([.][0-9]+)?$'
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
    WHERE ua.inviter_id IS NOT NULL
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
        GREATEST(0, sn.per_invitee_cap) AS per_invitee_cap,
        GREATEST(0, sn.freeze_hours) AS freeze_hours
    FROM (SELECT DISTINCT inviter_id FROM user_affiliates WHERE inviter_id IS NOT NULL) AS i
    JOIN user_affiliates ua ON ua.user_id = i.inviter_id
    LEFT JOIN effective_invitees ei ON ei.inviter_id = i.inviter_id
    CROSS JOIN setting_numbers sn
),
redeem_rebates AS (
    SELECT
        ua.inviter_id AS inviter_id,
        rc.used_by AS invitee_user_id,
        MAX(rc.used_at) AS last_used_at,
        ir.rebate_rate_percent,
        ir.per_invitee_cap,
        ir.freeze_hours,
        COALESCE(SUM(ROUND((rc.value::numeric * ir.rebate_rate_percent / 100.0), 8)), 0)::numeric AS calc_rebate_total
    FROM redeem_codes rc
    JOIN user_affiliates ua ON ua.user_id = rc.used_by
    JOIN inviter_rates ir ON ir.inviter_id = ua.inviter_id
    WHERE ua.inviter_id IS NOT NULL
      AND rc.status = 'used'
      AND rc.used_by IS NOT NULL
      AND rc.used_at IS NOT NULL
      AND rc.source_type IN ('commercial', 'system_grant')
      AND rc.type IN ('balance', 'subscription')
      AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
    GROUP BY ua.inviter_id, rc.used_by, ir.rebate_rate_percent, ir.per_invitee_cap, ir.freeze_hours
),
expected_totals AS (
    SELECT
        rr.inviter_id,
        rr.invitee_user_id,
        rr.last_used_at,
        rr.rebate_rate_percent,
        rr.per_invitee_cap,
        rr.freeze_hours,
        CASE
            WHEN rr.per_invitee_cap > 0 THEN LEAST(rr.per_invitee_cap, rr.calc_rebate_total)
            ELSE rr.calc_rebate_total
        END AS expected_total
    FROM redeem_rebates rr
),
accrued_totals AS (
    SELECT
        ual.user_id AS inviter_id,
        ual.source_user_id AS invitee_user_id,
        COALESCE(SUM(ual.amount), 0)::numeric AS accrued_total
    FROM user_affiliate_ledger ual
    WHERE ual.action = 'accrue'
    GROUP BY ual.user_id, ual.source_user_id
),
compared AS (
    SELECT
        e.inviter_id,
        e.invitee_user_id,
        e.last_used_at,
        e.rebate_rate_percent,
        e.per_invitee_cap,
        e.freeze_hours,
        ROUND(e.expected_total, 8) AS expected_total,
        ROUND(COALESCE(a.accrued_total, 0), 8) AS accrued_total
    FROM expected_totals e
    LEFT JOIN accrued_totals a
      ON a.inviter_id = e.inviter_id
     AND a.invitee_user_id = e.invitee_user_id
)
SELECT
    (SELECT affiliate_enabled FROM setting_numbers) AS affiliate_enabled,
    COUNT(*) FILTER (WHERE expected_total > accrued_total) AS underpaid_pairs,
    COUNT(DISTINCT inviter_id) FILTER (WHERE expected_total > accrued_total) AS underpaid_inviters,
    COUNT(DISTINCT invitee_user_id) FILTER (WHERE expected_total > accrued_total) AS underpaid_invitees,
    COALESCE(SUM(expected_total - accrued_total) FILTER (WHERE expected_total > accrued_total), 0)::numeric(20,8) AS total_compensation_amount,
    COALESCE(SUM(expected_total - accrued_total) FILTER (WHERE expected_total > accrued_total AND freeze_hours > 0), 0)::numeric(20,8) AS frozen_compensation_amount,
    COALESCE(SUM(expected_total - accrued_total) FILTER (WHERE expected_total > accrued_total AND freeze_hours <= 0), 0)::numeric(20,8) AS available_compensation_amount
FROM compared;

-- Optional sample of largest underpaid pairs:
-- WITH ...same CTE as above...
-- SELECT
--   c.inviter_id, inv.email AS inviter_email,
--   c.invitee_user_id, ie.email AS invitee_email,
--   c.rebate_rate_percent, c.expected_total, c.accrued_total,
--   (c.expected_total - c.accrued_total)::numeric(20,8) AS compensation_amount,
--   c.last_used_at
-- FROM compared c
-- JOIN users inv ON inv.id = c.inviter_id
-- JOIN users ie ON ie.id = c.invitee_user_id
-- WHERE c.expected_total > c.accrued_total
-- ORDER BY compensation_amount DESC
-- LIMIT 100;

-- ============================================================
-- 2) EXECUTE (idempotent)
-- ============================================================
BEGIN;

CREATE TABLE IF NOT EXISTS affiliate_rebate_compensation_pairs_20260525 (
    inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rebate_rate_percent DECIMAL(10,4) NOT NULL,
    per_invitee_cap DECIMAL(20,8) NOT NULL,
    freeze_hours INTEGER NOT NULL DEFAULT 0,
    expected_total DECIMAL(20,8) NOT NULL,
    accrued_before DECIMAL(20,8) NOT NULL,
    compensation_amount DECIMAL(20,8) NOT NULL,
    last_used_at TIMESTAMPTZ NULL,
    ledger_id BIGINT NULL REFERENCES user_affiliate_ledger(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ NULL,
    PRIMARY KEY (inviter_user_id, invitee_user_id)
);

ALTER TABLE affiliate_rebate_compensation_pairs_20260525
    ADD COLUMN IF NOT EXISTS freeze_hours INTEGER NOT NULL DEFAULT 0;

WITH setting_numbers AS (
    SELECT
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_rate'
                 AND value ~ '^-?[0-9]+([.][0-9]+)?$'
                THEN LEAST(100, GREATEST(0, value::numeric))
            END
        ), 20.0) AS global_rate_percent,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_per_invitee_cap'
                 AND value ~ '^-?[0-9]+([.][0-9]+)?$'
                THEN GREATEST(0, value::numeric)
            END
        ), 0) AS per_invitee_cap,
        COALESCE(MAX(
            CASE
                WHEN key = 'affiliate_rebate_freeze_hours'
                 AND value ~ '^-?[0-9]+$'
                THEN GREATEST(0, value::int)
            END
        ), 0) AS freeze_hours
    FROM settings
    WHERE key IN (
        'affiliate_rebate_rate',
        'affiliate_rebate_per_invitee_cap',
        'affiliate_rebate_freeze_hours'
    )
),
tier_source AS (
    SELECT
        CASE
            WHEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]') ~ '^[[:space:]]*[[]'
            THEN COALESCE(MAX(CASE WHEN key = 'affiliate_rebate_tiers' THEN value END), '[]')::jsonb
            ELSE '[{"min_effective_invitees":1,"rebate_rate":5},{"min_effective_invitees":3,"rebate_rate":8},{"min_effective_invitees":10,"rebate_rate":12},{"min_effective_invitees":30,"rebate_rate":15},{"min_effective_invitees":50,"rebate_rate":20}]'::jsonb
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
                        WHEN elem->>'rebate_rate' ~ '^-?[0-9]+([.][0-9]+)?$'
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
    WHERE ua.inviter_id IS NOT NULL
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
        GREATEST(0, sn.per_invitee_cap) AS per_invitee_cap,
        GREATEST(0, sn.freeze_hours) AS freeze_hours
    FROM (SELECT DISTINCT inviter_id FROM user_affiliates WHERE inviter_id IS NOT NULL) AS i
    JOIN user_affiliates ua ON ua.user_id = i.inviter_id
    LEFT JOIN effective_invitees ei ON ei.inviter_id = i.inviter_id
    CROSS JOIN setting_numbers sn
),
redeem_rebates AS (
    SELECT
        ua.inviter_id AS inviter_id,
        rc.used_by AS invitee_user_id,
        MAX(rc.used_at) AS last_used_at,
        ir.rebate_rate_percent,
        ir.per_invitee_cap,
        ir.freeze_hours,
        COALESCE(SUM(ROUND((rc.value::numeric * ir.rebate_rate_percent / 100.0), 8)), 0)::numeric AS calc_rebate_total
    FROM redeem_codes rc
    JOIN user_affiliates ua ON ua.user_id = rc.used_by
    JOIN inviter_rates ir ON ir.inviter_id = ua.inviter_id
    WHERE ua.inviter_id IS NOT NULL
      AND rc.status = 'used'
      AND rc.used_by IS NOT NULL
      AND rc.used_at IS NOT NULL
      AND rc.source_type IN ('commercial', 'system_grant')
      AND rc.type IN ('balance', 'subscription')
      AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
    GROUP BY ua.inviter_id, rc.used_by, ir.rebate_rate_percent, ir.per_invitee_cap, ir.freeze_hours
),
expected_totals AS (
    SELECT
        rr.inviter_id,
        rr.invitee_user_id,
        rr.last_used_at,
        rr.rebate_rate_percent,
        rr.per_invitee_cap,
        rr.freeze_hours,
        CASE
            WHEN rr.per_invitee_cap > 0 THEN LEAST(rr.per_invitee_cap, rr.calc_rebate_total)
            ELSE rr.calc_rebate_total
        END AS expected_total
    FROM redeem_rebates rr
),
accrued_totals AS (
    SELECT
        ual.user_id AS inviter_id,
        ual.source_user_id AS invitee_user_id,
        COALESCE(SUM(ual.amount), 0)::numeric AS accrued_total
    FROM user_affiliate_ledger ual
    WHERE ual.action = 'accrue'
    GROUP BY ual.user_id, ual.source_user_id
),
underpaid AS (
    SELECT
        e.inviter_id,
        e.invitee_user_id,
        e.last_used_at,
        e.rebate_rate_percent,
        e.per_invitee_cap,
        e.freeze_hours,
        ROUND(e.expected_total, 8) AS expected_total,
        ROUND(COALESCE(a.accrued_total, 0), 8) AS accrued_total,
        ROUND(e.expected_total - COALESCE(a.accrued_total, 0), 8) AS compensation_amount
    FROM expected_totals e
    LEFT JOIN accrued_totals a
      ON a.inviter_id = e.inviter_id
     AND a.invitee_user_id = e.invitee_user_id
    WHERE e.expected_total > COALESCE(a.accrued_total, 0)
)
INSERT INTO affiliate_rebate_compensation_pairs_20260525 (
    inviter_user_id,
    invitee_user_id,
    rebate_rate_percent,
    per_invitee_cap,
    freeze_hours,
    expected_total,
    accrued_before,
    compensation_amount,
    last_used_at
)
SELECT
    u.inviter_id,
    u.invitee_user_id,
    u.rebate_rate_percent::numeric(10,4),
    u.per_invitee_cap::numeric(20,8),
    u.freeze_hours,
    u.expected_total::numeric(20,8),
    u.accrued_total::numeric(20,8),
    u.compensation_amount::numeric(20,8),
    u.last_used_at
FROM underpaid u
ON CONFLICT (inviter_user_id, invitee_user_id) DO UPDATE
SET rebate_rate_percent = EXCLUDED.rebate_rate_percent,
    per_invitee_cap = EXCLUDED.per_invitee_cap,
    freeze_hours = EXCLUDED.freeze_hours,
    expected_total = EXCLUDED.expected_total,
    accrued_before = EXCLUDED.accrued_before,
    compensation_amount = EXCLUDED.compensation_amount,
    last_used_at = EXCLUDED.last_used_at
WHERE affiliate_rebate_compensation_pairs_20260525.ledger_id IS NULL;

WITH pending AS (
    SELECT
        inviter_user_id,
        freeze_hours,
        compensation_amount
    FROM affiliate_rebate_compensation_pairs_20260525
    WHERE ledger_id IS NULL
      AND compensation_amount > 0
),
agg AS (
    SELECT
        inviter_user_id,
        COALESCE(SUM(CASE WHEN freeze_hours > 0 THEN compensation_amount ELSE 0 END), 0)::numeric AS frozen_delta,
        COALESCE(SUM(CASE WHEN freeze_hours <= 0 THEN compensation_amount ELSE 0 END), 0)::numeric AS available_delta,
        COALESCE(SUM(compensation_amount), 0)::numeric AS history_delta
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

WITH pending AS (
    SELECT
        inviter_user_id,
        invitee_user_id,
        freeze_hours,
        compensation_amount
    FROM affiliate_rebate_compensation_pairs_20260525
    WHERE ledger_id IS NULL
      AND compensation_amount > 0
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
        p.compensation_amount,
        p.invitee_user_id,
        NULL,
        CASE
            WHEN p.freeze_hours > 0 THEN NOW() + make_interval(hours => p.freeze_hours)
            ELSE NULL
        END,
        NOW(),
        NOW()
    FROM pending p
    RETURNING id, user_id, source_user_id
)
UPDATE affiliate_rebate_compensation_pairs_20260525 c
SET ledger_id = i.id,
    applied_at = NOW()
FROM inserted i
WHERE c.ledger_id IS NULL
  AND c.inviter_user_id = i.user_id
  AND c.invitee_user_id = i.source_user_id;

COMMIT;

-- ============================================================
-- 3) POST CHECK
-- ============================================================
SELECT
    COUNT(*) AS total_pairs_in_snapshot,
    COUNT(*) FILTER (WHERE ledger_id IS NOT NULL) AS applied_pairs,
    COUNT(*) FILTER (WHERE ledger_id IS NULL) AS unapplied_pairs,
    COUNT(*) FILTER (WHERE freeze_hours > 0) AS frozen_pairs_in_snapshot,
    COUNT(*) FILTER (WHERE freeze_hours <= 0) AS available_pairs_in_snapshot,
    COALESCE(SUM(compensation_amount), 0)::numeric(20,8) AS total_snapshot_compensation,
    COALESCE(SUM(compensation_amount) FILTER (WHERE ledger_id IS NOT NULL), 0)::numeric(20,8) AS applied_compensation,
    COALESCE(SUM(compensation_amount) FILTER (WHERE ledger_id IS NOT NULL AND freeze_hours > 0), 0)::numeric(20,8) AS applied_frozen_compensation,
    COALESCE(SUM(compensation_amount) FILTER (WHERE ledger_id IS NOT NULL AND freeze_hours <= 0), 0)::numeric(20,8) AS applied_available_compensation
FROM affiliate_rebate_compensation_pairs_20260525;

-- ============================================================
-- 4) ROLLBACK PLAN (if needed)
-- ============================================================
-- BEGIN;
--
-- WITH applied AS (
--     SELECT *
--     FROM affiliate_rebate_compensation_pairs_20260525
--     WHERE ledger_id IS NOT NULL
-- ),
-- agg AS (
--     SELECT
--         inviter_user_id,
--         COALESCE(SUM(CASE WHEN freeze_hours > 0 THEN compensation_amount ELSE 0 END), 0)::numeric AS frozen_delta,
--         COALESCE(SUM(CASE WHEN freeze_hours <= 0 THEN compensation_amount ELSE 0 END), 0)::numeric AS available_delta,
--         COALESCE(SUM(compensation_amount), 0)::numeric AS history_delta
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
-- DELETE FROM user_affiliate_ledger
-- WHERE id IN (
--     SELECT ledger_id
--     FROM affiliate_rebate_compensation_pairs_20260525
--     WHERE ledger_id IS NOT NULL
-- );
--
-- UPDATE affiliate_rebate_compensation_pairs_20260525
-- SET ledger_id = NULL,
--     applied_at = NULL;
--
-- COMMIT;
