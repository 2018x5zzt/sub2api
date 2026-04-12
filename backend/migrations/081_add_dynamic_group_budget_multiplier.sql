ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS pricing_mode VARCHAR(20) NOT NULL DEFAULT 'fixed',
    ADD COLUMN IF NOT EXISTS default_budget_multiplier DECIMAL(10,4);

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS budget_multiplier DECIMAL(10,4);

UPDATE groups
SET pricing_mode = 'dynamic',
    default_budget_multiplier = 8.0
WHERE name = 'claude-dynamic';

UPDATE api_keys
SET budget_multiplier = 8.0
WHERE budget_multiplier IS NULL
  AND group_id IN (
      SELECT id
      FROM groups
      WHERE name = 'claude-dynamic'
        AND pricing_mode = 'dynamic'
        AND default_budget_multiplier = 8.0
  );
