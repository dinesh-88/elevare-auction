ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS branding_published jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS content_published jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS branding_draft jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS content_draft jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS content_version integer NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS content_schema_version integer NOT NULL DEFAULT 1;

UPDATE tenants
SET branding_published = COALESCE(NULLIF(branding, 'null'::jsonb), '{}'::jsonb)
WHERE branding_published = '{}'::jsonb;

UPDATE tenants
SET branding_draft = branding_published
WHERE branding_draft = '{}'::jsonb;

CREATE TABLE IF NOT EXISTS tenant_assets (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  kind text NOT NULL,
  url text NOT NULL,
  width integer,
  height integer,
  size_bytes bigint,
  mime text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (kind IN ('logo','banner','image')),
  CHECK (size_bytes IS NULL OR size_bytes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_tenant_assets_tenant_kind_created_at ON tenant_assets(tenant_id, kind, created_at DESC);
