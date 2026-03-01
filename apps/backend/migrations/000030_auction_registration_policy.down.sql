ALTER TABLE auctions
  DROP COLUMN IF EXISTS registration_policy;

DROP TYPE IF EXISTS auction_registration_policy;
