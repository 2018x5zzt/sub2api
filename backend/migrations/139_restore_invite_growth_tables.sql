ALTER TABLE users
  ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32),
  ADD COLUMN IF NOT EXISTS invited_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS invite_bound_at TIMESTAMPTZ;

UPDATE users
SET invite_code = CONCAT('U', UPPER(TO_HEX(id)))
WHERE invite_code IS NULL;

CREATE TABLE IF NOT EXISTS invite_code_aliases (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  alias_code VARCHAR(32) NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'migration_139_restore',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invite_code_aliases_alias_code
  ON invite_code_aliases(alias_code);

CREATE INDEX IF NOT EXISTS idx_invite_code_aliases_user_id
  ON invite_code_aliases(user_id);

INSERT INTO invite_code_aliases (user_id, alias_code, source)
SELECT id, invite_code, 'migration_139_restore'
FROM users
WHERE invite_code IS NOT NULL
  AND invite_code !~ '^[A-Za-z]{8}$'
ON CONFLICT (alias_code) DO NOTHING;

DO $$
DECLARE
  alphabet CONSTANT TEXT := 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  target RECORD;
  candidate TEXT;
BEGIN
  FOR target IN
    SELECT id
    FROM users
    WHERE invite_code IS NULL
       OR invite_code !~ '^[A-Za-z]{8}$'
    ORDER BY id
  LOOP
    LOOP
      SELECT string_agg(SUBSTRING(alphabet FROM 1 + FLOOR(random() * LENGTH(alphabet))::INT FOR 1), '')
      INTO candidate
      FROM generate_series(1, 8);

      EXIT WHEN NOT EXISTS (
        SELECT 1
        FROM users
        WHERE invite_code = candidate
      ) AND NOT EXISTS (
        SELECT 1
        FROM invite_code_aliases
        WHERE alias_code = candidate
      );
    END LOOP;

    UPDATE users
    SET invite_code = candidate
    WHERE id = target.id;
  END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_invite_code_not_null
  ON users(invite_code)
  WHERE invite_code IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_invited_by_user_id
  ON users(invited_by_user_id);

ALTER TABLE redeem_codes
  ADD COLUMN IF NOT EXISTS source_type VARCHAR(32);

UPDATE redeem_codes
SET source_type = 'system_grant'
WHERE source_type IS NULL OR BTRIM(source_type) = '';

ALTER TABLE redeem_codes
  ALTER COLUMN source_type SET DEFAULT 'system_grant';

ALTER TABLE redeem_codes
  ALTER COLUMN source_type SET NOT NULL;

CREATE TABLE IF NOT EXISTS invite_admin_actions (
  id BIGSERIAL PRIMARY KEY,
  action_type VARCHAR(32) NOT NULL,
  operator_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  reason TEXT NOT NULL,
  request_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  result_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invite_relationship_events (
  id BIGSERIAL PRIMARY KEY,
  invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  previous_inviter_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  new_inviter_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  event_type VARCHAR(32) NOT NULL,
  effective_at TIMESTAMPTZ NOT NULL,
  operator_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  reason TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_relationship_events_invitee_effective
  ON invite_relationship_events(invitee_user_id, effective_at DESC, id DESC);

INSERT INTO invite_relationship_events (
  invitee_user_id,
  previous_inviter_user_id,
  new_inviter_user_id,
  event_type,
  effective_at,
  operator_user_id,
  reason
)
SELECT
  users.id,
  NULL,
  users.invited_by_user_id,
  'register_bind',
  COALESCE(users.invite_bound_at, users.created_at),
  NULL,
  NULL
FROM users
WHERE users.invited_by_user_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invite_relationship_events
    WHERE invite_relationship_events.invitee_user_id = users.id
      AND invite_relationship_events.event_type = 'register_bind'
  );

CREATE TABLE IF NOT EXISTS invite_reward_records (
  id BIGSERIAL PRIMARY KEY,
  inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  trigger_redeem_code_id BIGINT NULL REFERENCES redeem_codes(id) ON DELETE RESTRICT,
  trigger_redeem_code_value DECIMAL(20,8) NOT NULL DEFAULT 0,
  reward_target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  reward_role VARCHAR(32) NOT NULL,
  reward_type VARCHAR(64) NOT NULL,
  reward_rate DECIMAL(10,8),
  reward_amount DECIMAL(20,8) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'applied',
  notes TEXT,
  admin_action_id BIGINT NULL REFERENCES invite_admin_actions(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE invite_reward_records
  ALTER COLUMN trigger_redeem_code_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS admin_action_id BIGINT NULL REFERENCES invite_admin_actions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invite_reward_records_target_created
  ON invite_reward_records(reward_target_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_invite_reward_records_admin_action
  ON invite_reward_records(admin_action_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invite_reward_records_trigger_role_type
  ON invite_reward_records(trigger_redeem_code_id, reward_role, reward_type);
