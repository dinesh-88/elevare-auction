ALTER TABLE lot_bids
  ADD COLUMN IF NOT EXISTS tenant_id uuid;

UPDATE lot_bids b
SET tenant_id = a.tenant_id
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE b.lot_id = l.id
  AND b.tenant_id IS NULL;

ALTER TABLE lot_bids
  ALTER COLUMN tenant_id SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_lot_bids_tenant'
  ) THEN
    ALTER TABLE lot_bids
      ADD CONSTRAINT fk_lot_bids_tenant
      FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_lot_bids_tenant_lot_created_at
  ON lot_bids(tenant_id, lot_id, created_at DESC);
