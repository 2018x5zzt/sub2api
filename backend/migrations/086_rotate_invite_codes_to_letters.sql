CREATE TABLE IF NOT EXISTS invite_code_aliases (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  alias_code VARCHAR(32) NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'migration_086_rotation',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invite_code_aliases_alias_code
  ON invite_code_aliases(alias_code);

CREATE INDEX IF NOT EXISTS idx_invite_code_aliases_user_id
  ON invite_code_aliases(user_id);

INSERT INTO invite_code_aliases (user_id, alias_code, source)
SELECT id, invite_code, 'migration_086_rotation'
FROM users
WHERE invite_code IS NOT NULL
  AND invite_code !~ '^[A-Za-z]{8}$'
ON CONFLICT (alias_code) DO NOTHING;

DO $$
DECLARE
  alphabet CONSTANT TEXT := 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  target RECORD;
  candidate TEXT;
BEGIN
  FOR target IN
    SELECT id
    FROM users
    WHERE invite_code IS NULL
       OR invite_code !~ '^[A-Za-z]{8}$'
    ORDER BY id
  LOOP
    LOOP
      SELECT string_agg(SUBSTRING(alphabet FROM 1 + FLOOR(random() * LENGTH(alphabet))::INT FOR 1), '')
      INTO candidate
      FROM generate_series(1, 8);

      EXIT WHEN NOT EXISTS (
        SELECT 1
        FROM users
        WHERE invite_code = candidate
      ) AND NOT EXISTS (
        SELECT 1
        FROM invite_code_aliases
        WHERE alias_code = candidate
      );
    END LOOP;

    UPDATE users
    SET invite_code = candidate
    WHERE id = target.id;
  END LOOP;
END $$;
