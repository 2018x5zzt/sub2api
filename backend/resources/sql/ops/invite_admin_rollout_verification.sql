-- Pre-deploy and post-deploy invite admin rollout verification checklist.
-- This file is read-only and must not be treated as a migration.
-- Statements must remain simple: semicolons act only as terminators for the naive splitter shared by the integration test harness.
-- Statement-order contract (fixed for current rollout stage):
--   1) overview metrics
--   2) binding alignment metrics
--   3) missing register_bind samples
--   4) duplicate register_bind samples
--   5) inviter/effective_at mismatch samples
-- Later tasks must append new statements after statement 5, never insert before statements 1-5.

WITH metrics AS (
  SELECT 1 AS ord, 'bound_users_total' AS metric_name, COUNT(*)::text AS metric_value
  FROM users
  WHERE invited_by_user_id IS NOT NULL

  UNION ALL

  SELECT 2, 'register_bind_event_rows_total', COUNT(*)::text
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'

  UNION ALL

  SELECT 3, 'register_bind_distinct_invitees_total', COUNT(DISTINCT invitee_user_id)::text
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'

  UNION ALL

  SELECT 4, 'base_invite_reward_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'

  UNION ALL

  SELECT 5, 'base_invite_reward_amount_total', COALESCE(SUM(reward_amount), 0)::text
  FROM invite_reward_records
  WHERE reward_type = 'base_invite_reward'

  UNION ALL

  SELECT 6, 'admin_rebind_event_rows_total', COUNT(*)::text
  FROM invite_relationship_events
  WHERE event_type = 'admin_rebind'

  UNION ALL

  SELECT 7, 'manual_invite_grant_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'manual_invite_grant'

  UNION ALL

  SELECT 8, 'recompute_delta_rows_total', COUNT(*)::text
  FROM invite_reward_records
  WHERE reward_type = 'recompute_delta'
)
SELECT metric_name, metric_value
FROM metrics
ORDER BY ord;

WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
),
register_bind_events AS (
  SELECT
    invitee_user_id,
    new_inviter_user_id AS event_inviter_user_id,
    effective_at AS event_effective_at
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'
),
register_bind_counts AS (
  SELECT invitee_user_id, COUNT(*) AS register_bind_count
  FROM register_bind_events
  GROUP BY invitee_user_id
),
single_register_bind_events AS (
  SELECT e.invitee_user_id, e.event_inviter_user_id, e.event_effective_at
  FROM register_bind_events e
  INNER JOIN register_bind_counts c ON c.invitee_user_id = e.invitee_user_id
  WHERE c.register_bind_count = 1
),
metrics AS (
  SELECT 1 AS ord, 'bound_users_missing_register_bind_total' AS metric_name, COUNT(*)::text AS metric_value
  FROM bound_users b
  LEFT JOIN register_bind_events e ON e.invitee_user_id = b.invitee_user_id
  WHERE e.invitee_user_id IS NULL

  UNION ALL

  SELECT 2, 'register_bind_without_bound_user_total', COUNT(DISTINCT e.invitee_user_id)::text
  FROM register_bind_events e
  LEFT JOIN bound_users b ON b.invitee_user_id = e.invitee_user_id
  WHERE b.invitee_user_id IS NULL

  UNION ALL

  SELECT 3, 'register_bind_duplicate_invitee_total', COUNT(*)::text
  FROM register_bind_counts
  WHERE register_bind_count > 1

  UNION ALL

  -- Duplicate-policy: invitees with duplicate register_bind events are counted by
  -- register_bind_duplicate_invitee_total and excluded from mismatch metrics because
  -- there is no single authoritative register_bind event to compare.
  SELECT 4, 'register_bind_inviter_mismatch_total', COUNT(*)::text
  FROM bound_users b
  INNER JOIN single_register_bind_events e ON e.invitee_user_id = b.invitee_user_id
  WHERE b.current_inviter_user_id IS DISTINCT FROM e.event_inviter_user_id

  UNION ALL

  SELECT 5, 'register_bind_effective_at_mismatch_total', COUNT(*)::text
  FROM bound_users b
  INNER JOIN single_register_bind_events e ON e.invitee_user_id = b.invitee_user_id
  WHERE b.expected_effective_at IS DISTINCT FROM e.event_effective_at
)
SELECT metric_name, metric_value
FROM metrics
ORDER BY ord;

WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
)
SELECT
  b.invitee_user_id,
  b.current_inviter_user_id,
  b.expected_effective_at
FROM bound_users b
LEFT JOIN invite_relationship_events e
  ON e.invitee_user_id = b.invitee_user_id
 AND e.event_type = 'register_bind'
WHERE e.invitee_user_id IS NULL
ORDER BY b.invitee_user_id;

SELECT
  invitee_user_id,
  COUNT(*) AS register_bind_count,
  MIN(effective_at) AS first_effective_at,
  MAX(effective_at) AS last_effective_at
FROM invite_relationship_events
WHERE event_type = 'register_bind'
GROUP BY invitee_user_id
HAVING COUNT(*) > 1
ORDER BY invitee_user_id;

WITH bound_users AS (
  SELECT
    id AS invitee_user_id,
    invited_by_user_id AS current_inviter_user_id,
    COALESCE(invite_bound_at, created_at) AS expected_effective_at
  FROM users
  WHERE invited_by_user_id IS NOT NULL
),
register_bind_events AS (
  SELECT
    invitee_user_id,
    new_inviter_user_id AS event_inviter_user_id,
    effective_at AS event_effective_at
  FROM invite_relationship_events
  WHERE event_type = 'register_bind'
),
register_bind_counts AS (
  SELECT invitee_user_id, COUNT(*) AS register_bind_count
  FROM register_bind_events
  GROUP BY invitee_user_id
),
single_register_bind_events AS (
  SELECT e.invitee_user_id, e.event_inviter_user_id, e.event_effective_at
  FROM register_bind_events e
  INNER JOIN register_bind_counts c ON c.invitee_user_id = e.invitee_user_id
  WHERE c.register_bind_count = 1
)
-- Duplicate-policy: mismatch samples intentionally include only invitees with exactly
-- one register_bind event. Duplicate invitees are surfaced by statement 4 instead.
SELECT
  b.invitee_user_id,
  b.current_inviter_user_id,
  e.event_inviter_user_id,
  b.expected_effective_at,
  e.event_effective_at
FROM bound_users b
INNER JOIN single_register_bind_events e ON e.invitee_user_id = b.invitee_user_id
WHERE b.current_inviter_user_id IS DISTINCT FROM e.event_inviter_user_id
   OR b.expected_effective_at IS DISTINCT FROM e.event_effective_at
ORDER BY b.invitee_user_id;
