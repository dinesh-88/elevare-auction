ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
  ADD CONSTRAINT users_role_check
  CHECK (role IN ('OWNER','ADMIN','STAFF','ACCOUNTING','BIDDER'));

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS status text;

UPDATE users
SET status = 'ACTIVE'
WHERE status IS NULL OR status = '';

ALTER TABLE users
  ALTER COLUMN status SET NOT NULL,
  ALTER COLUMN status SET DEFAULT 'ACTIVE';

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_status_check;

ALTER TABLE users
  ADD CONSTRAINT users_status_check
  CHECK (status IN ('ACTIVE','SUSPENDED'));
