WITH params AS (
  SELECT TIMESTAMPTZ '2026-07-06 20:38:01.05328+08' AS cutoff
),
target_user AS (
  SELECT id
  FROM users
  WHERE email = 'peng1@laogou.com'
)
SELECT
  ak.id AS api_key_id,
  ak.name AS api_key_name,
  LEFT(ak.key, 12) AS api_key_prefix,
  COALESCE(g.name, '') AS group_name,
  COUNT(*) AS request_count,
  COUNT(DISTINCT (ul.created_at AT TIME ZONE 'Asia/Shanghai')::date) AS active_day_count,
  MIN((ul.created_at AT TIME ZONE 'Asia/Shanghai')::date) AS first_bill_date,
  MAX((ul.created_at AT TIME ZONE 'Asia/Shanghai')::date) AS last_bill_date,
  ROUND(SUM(ul.actual_cost)::numeric, 6) AS actual_cost_usd,
  ROUND((SUM(ul.actual_cost) * 0.25)::numeric, 6) AS settlement_amount_0_25x
FROM usage_logs ul
JOIN target_user tu
  ON tu.id = ul.user_id
JOIN api_keys ak
  ON ak.id = ul.api_key_id
 AND ak.user_id = tu.id
LEFT JOIN groups g
  ON g.id = ak.group_id
CROSS JOIN params p
WHERE ul.created_at < p.cutoff
GROUP BY 1, 2, 3, 4
ORDER BY actual_cost_usd DESC, api_key_id ASC;
