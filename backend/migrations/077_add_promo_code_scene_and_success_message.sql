ALTER TABLE promo_codes
ADD COLUMN IF NOT EXISTS scene VARCHAR(20) NOT NULL DEFAULT 'register';

ALTER TABLE promo_codes
ADD COLUMN IF NOT EXISTS success_message TEXT;

CREATE INDEX IF NOT EXISTS idx_promo_codes_scene ON promo_codes(scene);

COMMENT ON COLUMN promo_codes.scene IS '优惠码场景: register=注册优惠码, benefit=福利码';
COMMENT ON COLUMN promo_codes.success_message IS '福利码兑换成功后展示给用户的弹窗文案';
