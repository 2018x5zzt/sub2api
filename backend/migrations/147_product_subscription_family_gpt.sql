-- Normalize product subscription families to the single supported GPT family.
-- User API key creation no longer accepts family selection, so existing
-- product subscriptions and key bindings must agree on the backend default.

ALTER TABLE subscription_products
    ALTER COLUMN product_family SET DEFAULT 'gpt';

UPDATE subscription_products
SET product_family = 'gpt',
    updated_at = NOW()
WHERE product_family IS DISTINCT FROM 'gpt';

UPDATE api_keys
SET subscription_product_family = 'gpt',
    updated_at = NOW()
WHERE subscription_product_family IS NOT NULL
  AND subscription_product_family IS DISTINCT FROM 'gpt';
