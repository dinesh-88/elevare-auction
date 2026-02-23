ALTER TABLE lot_bids
  ADD COLUMN IF NOT EXISTS max_bid_minor bigint NULL CHECK (max_bid_minor > 0);

CREATE INDEX IF NOT EXISTS idx_lot_bids_lot_max_created
  ON lot_bids(lot_id, max_bid_minor DESC NULLS LAST, created_at ASC);
