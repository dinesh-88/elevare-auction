ALTER TABLE auctions
  DROP CONSTRAINT IF EXISTS auctions_currency_format;
ALTER TABLE auctions
  DROP COLUMN IF EXISTS currency;

ALTER TABLE tenants
  DROP CONSTRAINT IF EXISTS tenants_default_currency_format;
ALTER TABLE tenants
  DROP COLUMN IF EXISTS default_currency;
