ALTER TABLE promo_codes
ADD COLUMN IF NOT EXISTS random_bonus_pool_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE promo_codes
ADD COLUMN IF NOT EXISTS random_bonus_remaining DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE promo_codes
ADD COLUMN IF NOT EXISTS leaderboard_enabled BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE promo_code_usages
ADD COLUMN IF NOT EXISTS fixed_bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

ALTER TABLE promo_code_usages
ADD COLUMN IF NOT EXISTS random_bonus_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

COMMENT ON COLUMN promo_codes.random_bonus_pool_amount IS '随机红包总池金额，仅 benefit 场景使用';
COMMENT ON COLUMN promo_codes.random_bonus_remaining IS '随机红包剩余金额，仅 benefit 场景使用';
COMMENT ON COLUMN promo_codes.leaderboard_enabled IS '是否启用手气排行榜，仅 benefit 场景使用';
COMMENT ON COLUMN promo_code_usages.fixed_bonus_amount IS '固定赠送金额';
COMMENT ON COLUMN promo_code_usages.random_bonus_amount IS '随机红包金额';
