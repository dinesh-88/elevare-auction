DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bidder_onboarding_status') THEN
    CREATE TYPE bidder_onboarding_status AS ENUM ('incomplete', 'pending_review', 'approved', 'rejected');
  END IF;
END $$;

ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS require_bidder_onboarding_approval boolean NOT NULL DEFAULT false;

ALTER TABLE tenant_bidders
  ADD COLUMN IF NOT EXISTS bidder_status bidder_onboarding_status NOT NULL DEFAULT 'incomplete',
  ADD COLUMN IF NOT EXISTS vat_classification text NULL,
  ADD COLUMN IF NOT EXISTS terms_accepted_at timestamptz NULL,
  ADD COLUMN IF NOT EXISTS privacy_accepted_at timestamptz NULL;

CREATE INDEX IF NOT EXISTS idx_tenant_bidders_tenant_bidder_status
  ON tenant_bidders(tenant_id, bidder_status);
