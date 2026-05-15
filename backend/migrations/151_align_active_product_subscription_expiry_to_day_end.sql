-- Align active product subscriptions to natural-day expiry boundaries.
--
-- Existing active subscriptions that may still carry rolling timestamps are
-- normalized so expires_at always lands at HH:MM:SS = 23:59:59 while
-- preserving the original calendar date in the configured DB timezone.
UPDATE user_product_subscriptions ups
SET expires_at = date_trunc('day', ups.expires_at) + INTERVAL '23 hours 59 minutes 59 seconds',
    updated_at = NOW()
FROM users u
WHERE ups.user_id = u.id
  AND ups.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND ups.status = 'active'
  AND u.status = 'active'
  AND ups.expires_at <> date_trunc('day', ups.expires_at) + INTERVAL '23 hours 59 minutes 59 seconds';
