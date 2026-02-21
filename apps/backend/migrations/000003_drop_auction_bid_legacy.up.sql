DROP TABLE IF EXISTS bids;

ALTER TABLE auctions
  DROP COLUMN IF EXISTS starting_price,
  DROP COLUMN IF EXISTS current_price,
  DROP COLUMN IF EXISTS reserve_price,
  DROP COLUMN IF EXISTS min_increment;
