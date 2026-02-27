DROP INDEX IF EXISTS idx_auction_bidder_approvals_tenant_auction_risk;

ALTER TABLE auction_bidder_approvals
  DROP COLUMN IF EXISTS approved_at,
  DROP COLUMN IF EXISTS approved_by_user_id,
  DROP COLUMN IF EXISTS internal_notes,
  DROP COLUMN IF EXISTS deposit_status,
  DROP COLUMN IF EXISTS deposit_amount_minor,
  DROP COLUMN IF EXISTS deposit_required,
  DROP COLUMN IF EXISTS max_bid_limit_minor,
  DROP COLUMN IF EXISTS risk_level;
