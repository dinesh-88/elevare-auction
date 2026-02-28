DROP INDEX IF EXISTS idx_auction_terms_tenant_auction_locale_created;

DROP INDEX IF EXISTS idx_auction_terms_one_active_per_auction_locale;

ALTER TABLE auction_terms
  DROP CONSTRAINT IF EXISTS auction_terms_auction_id_locale_version_number_key;

ALTER TABLE auction_terms
  ADD CONSTRAINT auction_terms_auction_id_version_number_key
  UNIQUE (auction_id, version_number);

CREATE UNIQUE INDEX idx_auction_terms_one_active_per_auction
  ON auction_terms(auction_id)
  WHERE is_active = true;

CREATE INDEX idx_auction_terms_tenant_auction_created
  ON auction_terms(tenant_id, auction_id, created_at DESC);

ALTER TABLE auction_terms
  DROP COLUMN IF EXISTS locale;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_locale_check;

ALTER TABLE users
  DROP COLUMN IF EXISTS locale;
