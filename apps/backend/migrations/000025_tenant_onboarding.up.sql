DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_onboarding_status') THEN
    CREATE TYPE tenant_onboarding_status AS ENUM ('not_started', 'in_progress', 'completed');
  END IF;
END $$;

ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS onboarding_status tenant_onboarding_status NOT NULL DEFAULT 'not_started',
  ADD COLUMN IF NOT EXISTS onboarding_step INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS onboarding_completed_at TIMESTAMPTZ NULL;
