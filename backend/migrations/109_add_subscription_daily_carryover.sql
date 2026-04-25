ALTER TABLE user_subscriptions
  ADD COLUMN IF NOT EXISTS daily_carryover_in_usd DECIMAL(20,10) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS daily_carryover_remaining_usd DECIMAL(20,10) NOT NULL DEFAULT 0;
