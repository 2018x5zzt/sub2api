-- Legacy group subscriptions do not support daily carryover.
-- Product subscriptions keep their own carryover state in user_product_subscriptions.
UPDATE user_subscriptions
SET daily_carryover_in_usd = 0,
    daily_carryover_remaining_usd = 0,
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND (daily_carryover_in_usd <> 0 OR daily_carryover_remaining_usd <> 0);
