DROP INDEX IF EXISTS idx_lot_bids_tenant_lot_created_at;

ALTER TABLE lot_bids
  DROP CONSTRAINT IF EXISTS fk_lot_bids_tenant;

ALTER TABLE lot_bids
  DROP COLUMN IF EXISTS tenant_id;
