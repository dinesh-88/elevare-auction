DROP INDEX IF EXISTS idx_tenant_assets_tenant_kind_created_at;
DROP TABLE IF EXISTS tenant_assets;

ALTER TABLE tenants
  DROP COLUMN IF EXISTS content_schema_version,
  DROP COLUMN IF EXISTS content_version,
  DROP COLUMN IF EXISTS content_draft,
  DROP COLUMN IF EXISTS branding_draft,
  DROP COLUMN IF EXISTS content_published,
  DROP COLUMN IF EXISTS branding_published;
