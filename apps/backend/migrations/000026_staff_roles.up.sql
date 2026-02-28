ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
  ADD CONSTRAINT users_role_check
  CHECK (role IN ('OWNER','ADMIN','STAFF','ACCOUNTING','VIEWER','BIDDER'));

CREATE TABLE IF NOT EXISTS tenant_users (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('OWNER','ADMIN','STAFF','ACCOUNTING','VIEWER')),
  is_primary_owner boolean NOT NULL DEFAULT false,
  invited_by uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_users_primary_owner
  ON tenant_users(tenant_id)
  WHERE role = 'OWNER' AND is_primary_owner = true;

CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_role ON tenant_users(tenant_id, role);

CREATE TABLE IF NOT EXISTS tenant_invitations (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email text NOT NULL,
  role text NOT NULL CHECK (role IN ('ADMIN','STAFF','ACCOUNTING','VIEWER')),
  token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz NULL,
  invited_by uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_invitations_token_hash ON tenant_invitations(token_hash);
CREATE INDEX IF NOT EXISTS idx_tenant_invitations_tenant_email ON tenant_invitations(tenant_id, lower(email));

INSERT INTO tenant_users (id, tenant_id, user_id, role, is_primary_owner, invited_by, created_at)
SELECT gen_random_uuid(), u.tenant_id, u.id, u.role,
       CASE WHEN u.role = 'OWNER' THEN true ELSE false END,
       NULL,
       u.created_at
FROM users u
WHERE u.role IN ('OWNER','ADMIN','STAFF','ACCOUNTING','VIEWER')
  AND NOT EXISTS (
    SELECT 1 FROM tenant_users tu WHERE tu.tenant_id = u.tenant_id AND tu.user_id = u.id
  );
