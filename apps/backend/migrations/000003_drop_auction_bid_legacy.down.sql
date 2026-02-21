ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS starting_price bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS current_price bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS reserve_price bigint NULL,
  ADD COLUMN IF NOT EXISTS min_increment bigint NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS bids (
  id uuid PRIMARY KEY,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount bigint NOT NULL CHECK (amount > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bids_auction_created_at ON bids(auction_id, created_at DESC);
