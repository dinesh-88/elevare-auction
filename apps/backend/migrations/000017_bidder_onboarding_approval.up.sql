ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS require_bidder_approval boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS tenant_bidders (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  phone text NULL,
  company_name text NULL,
  address_line1 text NULL,
  address_line2 text NULL,
  city text NULL,
  postal_code text NULL,
  country text NULL,
  status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','SUSPENDED')),
  risk_level text NOT NULL DEFAULT 'NORMAL' CHECK (risk_level IN ('LOW','NORMAL','HIGH')),
  internal_notes text NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_bidders_tenant_user ON tenant_bidders(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_bidders_tenant_status ON tenant_bidders(tenant_id, status);

CREATE TABLE IF NOT EXISTS auction_bidder_approvals (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','REJECTED')),
  decided_by uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  decided_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (auction_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_auction_bidder_approvals_tenant_auction_status
  ON auction_bidder_approvals(tenant_id, auction_id, status);
CREATE INDEX IF NOT EXISTS idx_auction_bidder_approvals_tenant_user
  ON auction_bidder_approvals(tenant_id, user_id);
