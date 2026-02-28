ALTER TABLE users
  ADD COLUMN IF NOT EXISTS locale varchar(5) NOT NULL DEFAULT 'en';

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_locale_check;

ALTER TABLE users
  ADD CONSTRAINT users_locale_check
  CHECK (locale IN ('en', 'fr', 'de'));

ALTER TABLE auction_terms
  ADD COLUMN IF NOT EXISTS locale varchar(5) NOT NULL DEFAULT 'en';

UPDATE auction_terms
SET locale = 'en'
WHERE locale IS NULL OR locale = '';

DROP INDEX IF EXISTS idx_auction_terms_one_active_per_auction;

DROP INDEX IF EXISTS idx_auction_terms_tenant_auction_created;

ALTER TABLE auction_terms
  DROP CONSTRAINT IF EXISTS auction_terms_auction_id_version_number_key;

ALTER TABLE auction_terms
  ADD CONSTRAINT auction_terms_auction_id_locale_version_number_key
  UNIQUE (auction_id, locale, version_number);

CREATE UNIQUE INDEX idx_auction_terms_one_active_per_auction_locale
  ON auction_terms(auction_id, locale)
  WHERE is_active = true;

CREATE INDEX idx_auction_terms_tenant_auction_locale_created
  ON auction_terms(tenant_id, auction_id, locale, created_at DESC);
