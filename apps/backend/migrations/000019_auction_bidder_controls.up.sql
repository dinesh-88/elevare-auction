ALTER TABLE auction_bidder_approvals
  ADD COLUMN IF NOT EXISTS risk_level text NOT NULL DEFAULT 'low' CHECK (risk_level IN ('low','medium','high')),
  ADD COLUMN IF NOT EXISTS max_bid_limit_minor bigint NULL,
  ADD COLUMN IF NOT EXISTS deposit_required boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS deposit_amount_minor bigint NULL,
  ADD COLUMN IF NOT EXISTS deposit_status text NOT NULL DEFAULT 'not_required' CHECK (deposit_status IN ('not_required','required','pending','received','waived')),
  ADD COLUMN IF NOT EXISTS internal_notes text NULL,
  ADD COLUMN IF NOT EXISTS approved_by_user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS approved_at timestamptz NULL;

CREATE INDEX IF NOT EXISTS idx_auction_bidder_approvals_tenant_auction_risk
  ON auction_bidder_approvals(tenant_id, auction_id, risk_level);
