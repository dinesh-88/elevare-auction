DROP TABLE IF EXISTS lot_bids;
DROP TABLE IF EXISTS lot_images;
DROP TABLE IF EXISTS lots;

ALTER TABLE auctions
  DROP COLUMN IF EXISTS lot_duration_seconds,
  DROP COLUMN IF EXISTS auto_progress_enabled,
  DROP COLUMN IF EXISTS auction_mode;
