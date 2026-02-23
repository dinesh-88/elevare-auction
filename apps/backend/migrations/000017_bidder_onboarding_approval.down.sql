DROP INDEX IF EXISTS idx_auction_bidder_approvals_tenant_user;
DROP INDEX IF EXISTS idx_auction_bidder_approvals_tenant_auction_status;
DROP TABLE IF EXISTS auction_bidder_approvals;

DROP INDEX IF EXISTS idx_tenant_bidders_tenant_status;
DROP INDEX IF EXISTS idx_tenant_bidders_tenant_user;
DROP TABLE IF EXISTS tenant_bidders;

ALTER TABLE auctions
  DROP COLUMN IF EXISTS require_bidder_approval;
