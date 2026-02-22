ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS default_currency text;

UPDATE tenants
SET default_currency = COALESCE(NULLIF(default_currency, ''), 'USD');

ALTER TABLE tenants
  ALTER COLUMN default_currency SET NOT NULL,
  ALTER COLUMN default_currency SET DEFAULT 'USD';

ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS currency text;

UPDATE auctions a
SET currency = COALESCE(NULLIF(a.currency, ''), t.default_currency, 'USD')
FROM tenants t
WHERE t.id = a.tenant_id;

ALTER TABLE auctions
  ALTER COLUMN currency SET NOT NULL,
  ALTER COLUMN currency SET DEFAULT 'USD';

ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS tenants_default_currency_format;
ALTER TABLE tenants
  ADD CONSTRAINT tenants_default_currency_format CHECK (default_currency ~ '^[A-Z]{3}$');

ALTER TABLE auctions
  DROP CONSTRAINT IF EXISTS auctions_currency_format;
ALTER TABLE auctions
  ADD CONSTRAINT auctions_currency_format CHECK (currency ~ '^[A-Z]{3}$');
