DROP INDEX IF EXISTS idx_lot_bids_lot_max_created;

ALTER TABLE lot_bids
  DROP COLUMN IF EXISTS max_bid_minor;
