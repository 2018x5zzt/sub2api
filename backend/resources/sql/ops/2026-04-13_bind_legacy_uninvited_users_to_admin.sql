-- One-off operator script: bind legacy active users without an inviter to a chosen admin inviter.
-- This is intentionally not a schema migration.
--
-- Usage example:
--   psql -v ON_ERROR_STOP=1 \
--     -v target_invite_code='aiGpAvid' \
--     -f backend/resources/sql/ops/2026-04-13_bind_legacy_uninvited_users_to_admin.sql
--
-- Operator notes:
--   1) This only targets users created before the fixed cutoff below.
--   2) It mirrors the admin rebind write path by inserting invite_admin_actions,
--      invite_relationship_events, and updating users.invited_by_user_id.
--   3) Historical invite rewards are not recomputed.

\if :{?target_invite_code}
\else
\echo 'target_invite_code psql variable is required'
\quit 1
\endif

SELECT id AS target_admin_user_id
FROM users
WHERE invite_code = :'target_invite_code'
  AND role = 'admin'
  AND deleted_at IS NULL;

BEGIN;

WITH operation AS (
  SELECT
    'Bind legacy active users without inviter to admin inviter for future recharge rewards'::text AS reason,
    TIMESTAMPTZ '2026-04-13 10:15:00+02:00' AS created_before
),
admin_user AS (
  SELECT id AS admin_user_id, invite_code
  FROM users
  WHERE invite_code = :'target_invite_code'
    AND role = 'admin'
    AND deleted_at IS NULL
),
candidates AS (
  SELECT
    u.id AS invitee_user_id,
    a.admin_user_id,
    a.invite_code,
    op.reason
  FROM users u
  CROSS JOIN admin_user a
  CROSS JOIN operation op
  WHERE u.invited_by_user_id IS NULL
    AND u.deleted_at IS NULL
    AND u.status = 'active'
    AND u.role = 'user'
    AND u.created_at < op.created_before
    AND u.id <> a.admin_user_id
),
inserted_actions AS (
  INSERT INTO invite_admin_actions (
    action_type,
    operator_user_id,
    target_user_id,
    reason,
    request_snapshot_json,
    result_snapshot_json
  )
  SELECT
    'rebind_inviter',
    c.admin_user_id,
    c.invitee_user_id,
    c.reason,
    jsonb_build_object(
      'invitee_user_id', c.invitee_user_id,
      'new_inviter_user_id', c.admin_user_id,
      'target_invite_code', c.invite_code,
      'bulk_backfill', true
    ),
    jsonb_build_object(
      'historical_rewards_changed', false,
      'bulk_backfill', true
    )
  FROM candidates c
  RETURNING
    id,
    operator_user_id AS admin_user_id,
    target_user_id AS invitee_user_id,
    reason
),
updated_users AS (
  UPDATE users u
  SET
    invited_by_user_id = a.admin_user_id,
    updated_at = NOW()
  FROM inserted_actions a
  WHERE u.id = a.invitee_user_id
  RETURNING
    u.id AS invitee_user_id,
    a.admin_user_id,
    a.reason
),
inserted_events AS (
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
    u.invitee_user_id,
    NULL,
    u.admin_user_id,
    'admin_rebind',
    NOW(),
    u.admin_user_id,
    u.reason
  FROM updated_users u
  RETURNING invitee_user_id
)
SELECT
  (SELECT COUNT(*) FROM inserted_actions) AS admin_actions_inserted,
  (SELECT COUNT(*) FROM updated_users) AS users_updated,
  (SELECT COUNT(*) FROM inserted_events) AS relationship_events_inserted;

COMMIT;

WITH operation AS (
  SELECT
    'Bind legacy active users without inviter to admin inviter for future recharge rewards'::text AS reason,
    TIMESTAMPTZ '2026-04-13 10:15:00+02:00' AS created_before
),
admin_user AS (
  SELECT id AS admin_user_id
  FROM users
  WHERE invite_code = :'target_invite_code'
    AND role = 'admin'
    AND deleted_at IS NULL
)
SELECT
  COUNT(*) FILTER (WHERE u.invited_by_user_id = a.admin_user_id) AS bound_to_target_admin,
  COUNT(*) FILTER (WHERE u.invited_by_user_id IS NULL) AS still_unbound_in_scope
FROM users u
CROSS JOIN admin_user a
CROSS JOIN operation op
WHERE u.deleted_at IS NULL
  AND u.status = 'active'
  AND u.role = 'user'
  AND u.created_at < op.created_before;
