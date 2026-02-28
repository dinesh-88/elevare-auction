DROP TABLE IF EXISTS tenant_invitations;
DROP INDEX IF EXISTS idx_tenant_users_primary_owner;
DROP TABLE IF EXISTS tenant_users;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
  ADD CONSTRAINT users_role_check
  CHECK (role IN ('OWNER','ADMIN','STAFF','ACCOUNTING','BIDDER'));
