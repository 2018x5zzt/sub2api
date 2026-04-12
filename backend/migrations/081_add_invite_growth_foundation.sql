ALTER TABLE users
  ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32),
  ADD COLUMN IF NOT EXISTS invited_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS invite_bound_at TIMESTAMPTZ;

UPDATE users
SET invite_code = CONCAT('U', UPPER(TO_HEX(id)))
WHERE invite_code IS NULL;

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

CREATE TABLE IF NOT EXISTS invite_reward_records (
  id BIGSERIAL PRIMARY KEY,
  inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  trigger_redeem_code_id BIGINT NOT NULL REFERENCES redeem_codes(id) ON DELETE RESTRICT,
  trigger_redeem_code_value DECIMAL(20,8) NOT NULL,
  reward_target_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  reward_role VARCHAR(32) NOT NULL,
  reward_type VARCHAR(64) NOT NULL,
  reward_rate DECIMAL(10,8),
  reward_amount DECIMAL(20,8) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'applied',
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_reward_records_target_created
  ON invite_reward_records(reward_target_user_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invite_reward_records_trigger_role_type
  ON invite_reward_records(trigger_redeem_code_id, reward_role, reward_type);
