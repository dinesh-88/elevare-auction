ALTER TABLE tenants
  DROP COLUMN IF EXISTS onboarding_completed_at,
  DROP COLUMN IF EXISTS onboarding_step,
  DROP COLUMN IF EXISTS onboarding_status;

DROP TYPE IF EXISTS tenant_onboarding_status;
