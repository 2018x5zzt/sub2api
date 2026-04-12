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
  ON invite_relationship_events (invitee_user_id, effective_at DESC, id DESC);

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

ALTER TABLE invite_reward_records
  ALTER COLUMN trigger_redeem_code_id DROP NOT NULL;

ALTER TABLE invite_reward_records
  ADD COLUMN IF NOT EXISTS admin_action_id BIGINT NULL REFERENCES invite_admin_actions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_invite_reward_records_admin_action
  ON invite_reward_records(admin_action_id);
