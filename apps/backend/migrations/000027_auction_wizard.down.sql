ALTER TABLE auctions
  DROP COLUMN IF EXISTS published_at,
  DROP COLUMN IF EXISTS auction_status,
  DROP COLUMN IF EXISTS wizard_step;
