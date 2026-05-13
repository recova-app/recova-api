DROP INDEX IF EXISTS uq_users_username;

ALTER TABLE users
  DROP COLUMN IF EXISTS password_hash,
  DROP COLUMN IF EXISTS username;

UPDATE users
SET google_id = 'manual-' || id::text
WHERE google_id IS NULL;

ALTER TABLE users
  ALTER COLUMN google_id SET NOT NULL;
