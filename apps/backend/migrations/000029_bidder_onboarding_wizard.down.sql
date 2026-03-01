DROP INDEX IF EXISTS idx_tenant_bidders_tenant_bidder_status;

ALTER TABLE tenant_bidders
  DROP COLUMN IF EXISTS privacy_accepted_at,
  DROP COLUMN IF EXISTS terms_accepted_at,
  DROP COLUMN IF EXISTS vat_classification,
  DROP COLUMN IF EXISTS bidder_status;

ALTER TABLE tenants
  DROP COLUMN IF EXISTS require_bidder_onboarding_approval;

DROP TYPE IF EXISTS bidder_onboarding_status;
