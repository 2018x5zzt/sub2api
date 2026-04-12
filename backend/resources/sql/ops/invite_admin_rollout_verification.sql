-- Pre-deploy and post-deploy invite admin rollout verification checklist.
-- This file is read-only and must not be treated as a migration.

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
