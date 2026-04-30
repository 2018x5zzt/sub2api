-- xlabapi channel pricing and affiliate compatibility on top of upstream/main.
-- Upstream owns the base affiliate tables; this migration keeps the local
-- compatibility columns, seed settings, and legacy invite bridge idempotent.

ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS platform VARCHAR(50) NOT NULL DEFAULT 'anthropic';

CREATE INDEX IF NOT EXISTS idx_channel_model_pricing_platform
    ON channel_model_pricing (platform);

COMMENT ON COLUMN channel_model_pricing.platform IS '定价所属平台：anthropic/openai/gemini/antigravity 等';

UPDATE channels
SET model_mapping = jsonb_build_object('anthropic', model_mapping)
WHERE model_mapping IS NOT NULL
  AND model_mapping::text NOT IN ('{}', 'null', '')
  AND NOT EXISTS (
      SELECT 1 FROM jsonb_each(model_mapping) AS kv
      WHERE jsonb_typeof(kv.value) = 'object'
      LIMIT 1
  );

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

ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS aff_frozen_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS aff_rebate_rate_percent DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS aff_code_custom BOOLEAN NOT NULL DEFAULT false;

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

ALTER TABLE user_affiliate_ledger
    ADD COLUMN IF NOT EXISTS frozen_until TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS legacy_invite_reward_record_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_user_id ON user_affiliate_ledger(user_id);
CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_action ON user_affiliate_ledger(action);
CREATE INDEX IF NOT EXISTS idx_ual_frozen_thaw
    ON user_affiliate_ledger (user_id, frozen_until)
    WHERE frozen_until IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_ual_legacy_invite_reward_record_id
    ON user_affiliate_ledger (legacy_invite_reward_record_id)
    WHERE legacy_invite_reward_record_id IS NOT NULL;

COMMENT ON COLUMN user_affiliates.aff_frozen_quota IS 'Rebate quota currently frozen (pending thaw after freeze period)';
COMMENT ON COLUMN user_affiliates.aff_rebate_rate_percent IS '专属返利比例（百分比 0-100，NULL 表示沿用全局）';
COMMENT ON COLUMN user_affiliates.aff_code_custom IS '邀请码是否由管理员改写过（用于专属用户筛选）';
COMMENT ON COLUMN user_affiliate_ledger.frozen_until IS 'Rebate frozen until this time; NULL means already thawed or never frozen';
COMMENT ON COLUMN user_affiliate_ledger.legacy_invite_reward_record_id IS 'xlabapi legacy invite_reward_records.id carried during affiliate compatibility migration';

DO $$
BEGIN
    IF to_regclass('public.invite_reward_records') IS NOT NULL THEN
        EXECUTE $sql$
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
              )
        $sql$;

        EXECUTE $sql$
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
            WHERE ua.user_id = legacy_totals.user_id
        $sql$;
    END IF;
END $$;

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
