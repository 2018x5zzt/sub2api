-- Add platform discriminator to channel model pricing.
-- Repository code scopes pricing and model visibility by platform.
ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS platform VARCHAR(50) NOT NULL DEFAULT 'anthropic';

COMMENT ON COLUMN channel_model_pricing.platform IS '定价所属平台：anthropic/openai/gemini/antigravity 等';
