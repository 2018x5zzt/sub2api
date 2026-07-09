WITH params AS (
  SELECT TIMESTAMPTZ '2026-07-06 20:38:01.05328+08' AS cutoff
),
target_user AS (
  SELECT id
  FROM users
  WHERE email = 'peng1@laogou.com'
)
SELECT
  (ul.created_at AT TIME ZONE 'Asia/Shanghai')::date AS bill_date,
  COUNT(*) AS request_count,
  COUNT(DISTINCT ul.api_key_id) AS active_key_count,
  ROUND(SUM(ul.actual_cost)::numeric, 6) AS actual_cost_usd,
  ROUND((SUM(ul.actual_cost) * 0.25)::numeric, 6) AS settlement_amount_0_25x
FROM usage_logs ul
JOIN target_user tu
  ON tu.id = ul.user_id
JOIN api_keys ak
  ON ak.id = ul.api_key_id
 AND ak.user_id = tu.id
CROSS JOIN params p
WHERE ul.created_at < p.cutoff
GROUP BY 1
ORDER BY 1;
