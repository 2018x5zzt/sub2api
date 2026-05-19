-- xlabapi affiliate rebate backfill plan (2026-05-19)
-- Goal:
--   Re-label historical payment-fulfillment redeem codes from system_grant -> commercial
--   so they can be counted as "effective invitees" under the tiered affiliate rules.
--
-- Scope policy (conservative):
--   1) only used redeem codes
--   2) only redeem type in ('balance','subscription')
--   3) only source_type currently = 'system_grant'
--   4) only rows that can be linked to a payment order by deterministic recharge_code match
--   5) only orders in paid/completed/recharging lifecycle
--   6) only positive-value/effective-like redeem rows
--
-- Safety:
--   - Run PREVIEW first; verify counts and samples.
--   - Then run BACKUP and UPDATE in one transaction.
--   - Keep backup table for rollback.

-- ============================================================
-- 1) PREVIEW: how many rows will be updated?
-- ============================================================
WITH candidates AS (
    SELECT
        rc.id AS redeem_code_id,
        rc.code,
        rc.used_by,
        rc.type,
        rc.source_type,
        rc.status,
        rc.used_at,
        po.id AS payment_order_id,
        po.out_trade_no,
        po.status AS payment_status
    FROM redeem_codes rc
    JOIN payment_orders po ON po.recharge_code = rc.code
    WHERE rc.source_type = 'system_grant'
      AND rc.status = 'used'
      AND rc.used_at IS NOT NULL
      AND rc.type IN ('balance', 'subscription')
      AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
      AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
      AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
      AND (po.amount > 0 OR po.pay_amount > 0)
)
SELECT COUNT(*) AS candidate_count
FROM candidates;

-- ============================================================
-- 2) PREVIEW SAMPLE: inspect top 50 candidate rows
-- ============================================================
WITH candidates AS (
    SELECT
        rc.id AS redeem_code_id,
        rc.code,
        rc.used_by,
        rc.type,
        rc.source_type,
        rc.status,
        rc.used_at,
        po.id AS payment_order_id,
        po.out_trade_no,
        po.status AS payment_status
    FROM redeem_codes rc
    JOIN payment_orders po ON po.recharge_code = rc.code
    WHERE rc.source_type = 'system_grant'
      AND rc.status = 'used'
      AND rc.used_at IS NOT NULL
      AND rc.type IN ('balance', 'subscription')
      AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
      AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
      AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
      AND (po.amount > 0 OR po.pay_amount > 0)
)
SELECT *
FROM candidates
ORDER BY redeem_code_id DESC
LIMIT 50;

-- ============================================================
-- 3) EXECUTE BACKUP + UPDATE (single transaction)
-- ============================================================
BEGIN;

-- 3.1 Backup rows to be changed
CREATE TABLE IF NOT EXISTS redeem_codes_backfill_20260519 AS
SELECT
    rc.*
FROM redeem_codes rc
JOIN payment_orders po ON po.recharge_code = rc.code
WHERE rc.source_type = 'system_grant'
  AND rc.status = 'used'
  AND rc.used_at IS NOT NULL
  AND rc.type IN ('balance', 'subscription')
  AND (rc.value > 0 OR rc.group_id IS NOT NULL OR rc.validity_days > 0)
  AND po.status IN ('COMPLETED', 'PAID', 'RECHARGING')
  AND (po.paid_at IS NOT NULL OR po.completed_at IS NOT NULL)
  AND (po.amount > 0 OR po.pay_amount > 0);

-- 3.2 Update source_type to commercial for backed-up rows
UPDATE redeem_codes rc
SET source_type = 'commercial'
WHERE rc.id IN (SELECT id FROM redeem_codes_backfill_20260519);

COMMIT;

-- ============================================================
-- 4) POST CHECK: verify update distribution
-- ============================================================
SELECT source_type, COUNT(*) AS cnt
FROM redeem_codes
WHERE id IN (SELECT id FROM redeem_codes_backfill_20260519)
GROUP BY source_type
ORDER BY source_type;

-- ============================================================
-- 5) ROLLBACK PLAN (if needed)
-- ============================================================
-- BEGIN;
-- UPDATE redeem_codes rc
-- SET source_type = b.source_type
-- FROM redeem_codes_backfill_20260519 b
-- WHERE rc.id = b.id;
-- COMMIT;
