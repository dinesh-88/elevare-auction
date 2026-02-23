DROP INDEX IF EXISTS idx_invoices_tenant_status_issued;
DROP INDEX IF EXISTS idx_invoices_tenant_auction_status;
DROP INDEX IF EXISTS idx_invoices_tenant_auction;

ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_auction_id_fkey;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_cancel_reason_len_check;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_payment_reference_len_check;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_payment_status_check;

ALTER TABLE invoices
  DROP COLUMN IF EXISTS cancel_reason,
  DROP COLUMN IF EXISTS payment_reference,
  DROP COLUMN IF EXISTS cancelled_at,
  DROP COLUMN IF EXISTS paid_at,
  DROP COLUMN IF EXISTS payment_status,
  DROP COLUMN IF EXISTS total_due_minor,
  DROP COLUMN IF EXISTS tax_total_minor,
  DROP COLUMN IF EXISTS fees_total_minor,
  DROP COLUMN IF EXISTS hammer_price_minor,
  DROP COLUMN IF EXISTS auction_id;
