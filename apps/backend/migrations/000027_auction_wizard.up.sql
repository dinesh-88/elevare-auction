ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS wizard_step integer NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS auction_status text NOT NULL DEFAULT 'draft' CHECK (auction_status IN ('draft','published','closed','cancelled')),
  ADD COLUMN IF NOT EXISTS published_at timestamptz NULL;
