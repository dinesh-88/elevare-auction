CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE tenants (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  subdomain text NOT NULL UNIQUE,
  custom_domain text UNIQUE,
  status text NOT NULL CHECK (status IN ('PENDING','ACTIVE','REJECTED')) DEFAULT 'PENDING',
  country text NOT NULL DEFAULT '',
  estimated_auction_volume integer,
  branding jsonb NOT NULL DEFAULT '{}'::jsonb,
  require_admin_approval boolean NOT NULL DEFAULT true,
  require_kyc boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_tenants_status_created_at ON tenants(status, created_at DESC);

CREATE TABLE users (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email text NOT NULL,
  password_hash text NOT NULL,
  role text NOT NULL CHECK (role IN ('ADMIN','BIDDER')),
  email_verified boolean NOT NULL DEFAULT false,
  approval_status text NOT NULL CHECK (approval_status IN ('PENDING','APPROVED','REJECTED')),
  kyc_status text NOT NULL CHECK (kyc_status IN ('NONE','PENDING','VERIFIED','REJECTED')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant_role ON users(tenant_id, role);

CREATE TABLE auctions (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  title text NOT NULL,
  description text NOT NULL DEFAULT '',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  starting_price bigint NOT NULL CHECK (starting_price >= 0),
  current_price bigint NOT NULL CHECK (current_price >= 0),
  reserve_price bigint NULL,
  min_increment bigint NOT NULL CHECK (min_increment > 0),
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  status text NOT NULL CHECK (status IN ('DRAFT','SCHEDULED','LIVE','PAUSED','ENDED','CANCELLED')),
  allow_manual_start boolean NOT NULL DEFAULT false,
  allow_manual_pause boolean NOT NULL DEFAULT false,
  allow_manual_end boolean NOT NULL DEFAULT false,
  auto_extend_enabled boolean NOT NULL DEFAULT false,
  sniping_window_seconds integer NOT NULL DEFAULT 0,
  sniping_extension_seconds integer NOT NULL DEFAULT 0,
  bid_visibility text NOT NULL CHECK (bid_visibility IN ('PUBLIC','ANONYMOUS','SEALED')),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (end_time > start_time)
);

CREATE INDEX idx_auctions_tenant_status_time ON auctions(tenant_id, status, start_time, end_time);

CREATE TABLE bids (
  id uuid PRIMARY KEY,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount bigint NOT NULL CHECK (amount > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_bids_auction_created_at ON bids(auction_id, created_at DESC);
