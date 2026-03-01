DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'auction_registration_policy') THEN
    CREATE TYPE auction_registration_policy AS ENUM ('open', 'approval_required');
  END IF;
END $$;

ALTER TABLE auctions
  ADD COLUMN IF NOT EXISTS registration_policy auction_registration_policy NOT NULL DEFAULT 'approval_required';

UPDATE auctions
SET registration_policy = CASE
  WHEN require_bidder_approval THEN 'approval_required'::auction_registration_policy
  ELSE 'open'::auction_registration_policy
END
WHERE registration_policy IS NOT NULL;
