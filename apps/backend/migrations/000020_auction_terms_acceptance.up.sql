CREATE TABLE auction_terms (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  version_number integer NOT NULL CHECK (version_number > 0),
  content_markdown text NOT NULL,
  content_hash varchar(128) NOT NULL,
  is_active boolean NOT NULL DEFAULT false,
  created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (auction_id, version_number)
);

CREATE UNIQUE INDEX idx_auction_terms_one_active_per_auction
  ON auction_terms(auction_id)
  WHERE is_active = true;

CREATE INDEX idx_auction_terms_tenant_auction_created
  ON auction_terms(tenant_id, auction_id, created_at DESC);

CREATE TABLE bidder_terms_acceptance (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  auction_terms_id uuid NOT NULL REFERENCES auction_terms(id) ON DELETE RESTRICT,
  terms_version_number integer NOT NULL CHECK (terms_version_number > 0),
  content_hash varchar(128) NOT NULL,
  accepted_at timestamptz NOT NULL,
  accepted_ip varchar(128),
  accepted_user_agent text,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (auction_id, user_id, auction_terms_id)
);

CREATE INDEX idx_bidder_terms_acceptance_tenant_auction_user
  ON bidder_terms_acceptance(tenant_id, auction_id, user_id, created_at DESC);
