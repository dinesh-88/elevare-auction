ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS auction_mode text NOT NULL DEFAULT 'TIMED' CHECK (auction_mode IN ('TIMED','SEQUENTIAL')),
  ADD COLUMN IF NOT EXISTS auto_progress_enabled boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS lot_duration_seconds integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS lots (
  id uuid PRIMARY KEY,
  auction_id uuid NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
  lot_number text NOT NULL,
  order_index integer NOT NULL,
  title text NOT NULL,
  description text NOT NULL DEFAULT '',
  starting_price bigint NOT NULL CHECK (starting_price >= 0),
  current_price bigint NOT NULL CHECK (current_price >= 0),
  reserve_price bigint NULL,
  min_increment bigint NOT NULL CHECK (min_increment > 0),
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL CHECK (status IN ('PENDING','LIVE','PAUSED','ENDED','CANCELLED')),
  started_at timestamptz NULL,
  ended_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lots_auction_order ON lots(auction_id, order_index);
CREATE INDEX IF NOT EXISTS idx_lots_status ON lots(status);

CREATE TABLE IF NOT EXISTS lot_images (
  id uuid PRIMARY KEY,
  lot_id uuid NOT NULL REFERENCES lots(id) ON DELETE CASCADE,
  url text NOT NULL,
  position integer NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_lot_images_lot_position ON lot_images(lot_id, position);

CREATE TABLE IF NOT EXISTS lot_bids (
  id uuid PRIMARY KEY,
  lot_id uuid NOT NULL REFERENCES lots(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id),
  amount bigint NOT NULL CHECK (amount > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lot_bids_lot_created_at ON lot_bids(lot_id, created_at DESC);
