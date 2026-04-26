-- xlabapi compatibility bridge:
-- - materialize upstream affiliate tables on the xlabapi migration line
-- - preserve legacy invite bindings and reward history
-- - avoid double-crediting old invite rewards that were already paid into balance

CREATE TABLE IF NOT EXISTS user_affiliates (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    aff_code VARCHAR(32) NOT NULL UNIQUE,
    inviter_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    aff_count INTEGER NOT NULL DEFAULT 0,
    aff_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    aff_frozen_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    aff_history_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    aff_rebate_rate_percent DECIMAL(5,2),
    aff_code_custom BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_affiliates_inviter_id ON user_affiliates(inviter_id);
CREATE INDEX IF NOT EXISTS idx_user_affiliates_aff_quota ON user_affiliates(aff_quota);
CREATE INDEX IF NOT EXISTS idx_user_affiliates_admin_settings
    ON user_affiliates (updated_at)
    WHERE aff_code_custom = true OR aff_rebate_rate_percent IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_affiliate_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(32) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    source_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    frozen_until TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_user_id ON user_affiliate_ledger(user_id);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_action ON user_affiliate_ledger(action);
CREATE INDEX IF NOT EXISTS idx_ual_frozen_thaw
    ON user_affiliate_ledger (user_id, frozen_until)
    WHERE frozen_until IS NOT NULL;

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS legacy_invite_reward_record_id BIGINT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ual_legacy_invite_reward_record_id
    ON user_affiliate_ledger (legacy_invite_reward_record_id)
    WHERE legacy_invite_reward_record_id IS NOT NULL;

COMMENT ON TABLE user_affiliates IS '用户邀请返利信息';
COMMENT ON COLUMN user_affiliates.aff_code IS '用户邀请代码';
COMMENT ON COLUMN user_affiliates.inviter_id IS '邀请人用户ID';
COMMENT ON COLUMN user_affiliates.aff_count IS '累计邀请人数';
COMMENT ON COLUMN user_affiliates.aff_quota IS '当前可提取返利金额';
COMMENT ON COLUMN user_affiliates.aff_frozen_quota IS 'Rebate quota currently frozen (pending thaw after freeze period)';
COMMENT ON COLUMN user_affiliates.aff_history_quota IS '累计返利历史金额';
COMMENT ON COLUMN user_affiliates.aff_rebate_rate_percent IS '专属返利比例（百分比 0-100，NULL 表示沿用全局）';
COMMENT ON COLUMN user_affiliates.aff_code_custom IS '邀请码是否由管理员改写过（用于专属用户筛选）';
COMMENT ON TABLE user_affiliate_ledger IS '邀请返利资金流水（累计/转入）';
COMMENT ON COLUMN user_affiliate_ledger.action IS 'accrue|transfer';
COMMENT ON COLUMN user_affiliate_ledger.frozen_until IS 'Rebate frozen until this time; NULL means already thawed or never frozen';
COMMENT ON COLUMN user_affiliate_ledger.legacy_invite_reward_record_id IS 'xlabapi legacy invite_reward_records.id carried during affiliate compatibility migration';

DROP TABLE IF EXISTS _xlabapi_affiliate_code_map;

CREATE TEMP TABLE _xlabapi_affiliate_code_map (
    user_id BIGINT PRIMARY KEY,
    aff_code VARCHAR(32) NOT NULL UNIQUE
) ON COMMIT DROP;

DO $$
DECLARE
    target RECORD;
    candidate TEXT;
    attempt INTEGER;
BEGIN
    FOR target IN
        WITH legacy_codes AS (
            SELECT
                u.id AS user_id,
                UPPER(BTRIM(u.invite_code)) AS legacy_code,
                CASE
                    WHEN UPPER(BTRIM(u.invite_code)) ~ '^[A-Z0-9_-]{4,32}$' THEN
                        ROW_NUMBER() OVER (
                            PARTITION BY UPPER(BTRIM(u.invite_code))
                            ORDER BY u.id
                        )
                    ELSE NULL
                END AS legacy_rank
            FROM users u
            WHERE NOT EXISTS (
                SELECT 1
                FROM user_affiliates ua
                WHERE ua.user_id = u.id
            )
        )
        SELECT
            user_id,
            legacy_code,
            legacy_rank = 1 AS use_legacy_code
        FROM legacy_codes
        ORDER BY user_id
    LOOP
        candidate := NULL;

        IF target.use_legacy_code THEN
            candidate := target.legacy_code;
        END IF;

        IF candidate IS NULL
            OR EXISTS (SELECT 1 FROM _xlabapi_affiliate_code_map WHERE aff_code = candidate)
            OR EXISTS (SELECT 1 FROM user_affiliates WHERE aff_code = candidate)
        THEN
            attempt := 0;
            LOOP
                candidate := 'X' || UPPER(SUBSTRING(MD5(target.user_id::TEXT || ':' || attempt::TEXT) FROM 1 FOR 11));
                EXIT WHEN NOT EXISTS (SELECT 1 FROM _xlabapi_affiliate_code_map WHERE aff_code = candidate)
                    AND NOT EXISTS (SELECT 1 FROM user_affiliates WHERE aff_code = candidate);
                attempt := attempt + 1;
            END LOOP;
        END IF;

        INSERT INTO _xlabapi_affiliate_code_map (user_id, aff_code)
        VALUES (target.user_id, candidate);
    END LOOP;
END $$;

INSERT INTO user_affiliates (
    user_id,
    aff_code,
    inviter_id,
    aff_count,
    aff_quota,
    aff_frozen_quota,
    aff_history_quota,
    aff_code_custom,
    created_at,
    updated_at
)
SELECT
    u.id,
    m.aff_code,
    CASE
        WHEN u.invited_by_user_id IS NOT NULL
            AND u.invited_by_user_id <> u.id
            AND EXISTS (SELECT 1 FROM users inviter WHERE inviter.id = u.invited_by_user_id)
        THEN u.invited_by_user_id
        ELSE NULL
    END,
    0,
    0,
    0,
    0,
    false,
    COALESCE(u.created_at, NOW()),
    NOW()
FROM users u
JOIN _xlabapi_affiliate_code_map m ON m.user_id = u.id
ON CONFLICT (user_id) DO NOTHING;

UPDATE user_affiliates ua
SET inviter_id = u.invited_by_user_id,
    updated_at = NOW()
FROM users u
WHERE ua.user_id = u.id
  AND ua.inviter_id IS NULL
  AND u.invited_by_user_id IS NOT NULL
  AND u.invited_by_user_id <> u.id
  AND EXISTS (SELECT 1 FROM users inviter WHERE inviter.id = u.invited_by_user_id);

WITH invite_counts AS (
    SELECT invited_by_user_id AS inviter_id, COUNT(*)::INTEGER AS invitees
    FROM users
    WHERE invited_by_user_id IS NOT NULL
      AND invited_by_user_id <> id
    GROUP BY invited_by_user_id
)
UPDATE user_affiliates ua
SET aff_count = GREATEST(ua.aff_count, invite_counts.invitees),
    updated_at = NOW()
FROM invite_counts
WHERE ua.user_id = invite_counts.inviter_id;

INSERT INTO user_affiliate_ledger (
    user_id,
    action,
    amount,
    source_user_id,
    legacy_invite_reward_record_id,
    created_at,
    updated_at
)
SELECT
    irr.reward_target_user_id,
    'accrue',
    irr.reward_amount,
    irr.invitee_user_id,
    irr.id,
    COALESCE(irr.created_at, NOW()),
    NOW()
FROM invite_reward_records irr
WHERE irr.reward_role = 'inviter'
  AND irr.status = 'applied'
  AND irr.reward_amount > 0
  AND EXISTS (SELECT 1 FROM user_affiliates ua WHERE ua.user_id = irr.reward_target_user_id)
  AND EXISTS (SELECT 1 FROM users u WHERE u.id = irr.invitee_user_id)
  AND NOT EXISTS (
      SELECT 1
      FROM user_affiliate_ledger ual
      WHERE ual.legacy_invite_reward_record_id = irr.id
  );

WITH legacy_totals AS (
    SELECT user_id, COALESCE(SUM(amount), 0) AS total
    FROM user_affiliate_ledger
    WHERE legacy_invite_reward_record_id IS NOT NULL
      AND action = 'accrue'
    GROUP BY user_id
)
UPDATE user_affiliates ua
SET aff_history_quota = GREATEST(ua.aff_history_quota, legacy_totals.total),
    updated_at = NOW()
FROM legacy_totals
WHERE ua.user_id = legacy_totals.user_id;

INSERT INTO settings (key, value, updated_at)
VALUES
    ('affiliate_enabled', 'true', NOW()),
    ('affiliate_rebate_rate', '3', NOW()),
    ('affiliate_rebate_freeze_hours', '0', NOW()),
    ('affiliate_rebate_duration_days', '0', NOW()),
    ('affiliate_rebate_per_invitee_cap', '0', NOW()),
    ('available_channels_enabled', 'true', NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
