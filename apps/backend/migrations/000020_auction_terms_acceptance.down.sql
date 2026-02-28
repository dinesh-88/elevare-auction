DROP INDEX IF EXISTS idx_bidder_terms_acceptance_tenant_auction_user;
DROP TABLE IF EXISTS bidder_terms_acceptance;

DROP INDEX IF EXISTS idx_auction_terms_tenant_auction_created;
DROP INDEX IF EXISTS idx_auction_terms_one_active_per_auction;
DROP TABLE IF EXISTS auction_terms;
