ALTER TABLE invoices
  ADD COLUMN IF NOT EXISTS auction_id uuid,
  ADD COLUMN IF NOT EXISTS hammer_price_minor bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fees_total_minor bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS tax_total_minor bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS total_due_minor bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS payment_status text NOT NULL DEFAULT 'UNPAID',
  ADD COLUMN IF NOT EXISTS paid_at timestamptz,
  ADD COLUMN IF NOT EXISTS cancelled_at timestamptz,
  ADD COLUMN IF NOT EXISTS payment_reference text,
  ADD COLUMN IF NOT EXISTS cancel_reason text;

UPDATE invoices i
SET auction_id = l.auction_id
FROM lots l
WHERE i.lot_id = l.id
  AND i.auction_id IS NULL;

UPDATE invoices i
SET hammer_price_minor = COALESCE((i.snapshot ->> 'hammer_price')::bigint, l.hammer_price, 0),
    fees_total_minor = COALESCE((
      SELECT SUM((fee ->> 'amount')::bigint)
      FROM jsonb_array_elements(COALESCE(i.snapshot -> 'fees', '[]'::jsonb)) fee
    ), 0),
    tax_total_minor = 0,
    total_due_minor = COALESCE((i.snapshot ->> 'total_due')::bigint, i.total_amount, 0)
FROM lots l
WHERE i.lot_id = l.id;

UPDATE invoices
SET payment_status = CASE status
  WHEN 'PAID' THEN 'PAID'
  WHEN 'CANCELLED' THEN 'CANCELLED'
  ELSE 'UNPAID'
END
WHERE payment_status NOT IN ('UNPAID','PAID','CANCELLED')
   OR payment_status IS NULL;

ALTER TABLE invoices
  ADD CONSTRAINT invoices_payment_status_check
  CHECK (payment_status IN ('UNPAID','PAID','CANCELLED'));

ALTER TABLE invoices
  ADD CONSTRAINT invoices_payment_reference_len_check
  CHECK (payment_reference IS NULL OR char_length(payment_reference) <= 128);

ALTER TABLE invoices
  ADD CONSTRAINT invoices_cancel_reason_len_check
  CHECK (cancel_reason IS NULL OR char_length(cancel_reason) <= 1024);

ALTER TABLE invoices
  ALTER COLUMN auction_id SET NOT NULL;

ALTER TABLE invoices
  ADD CONSTRAINT invoices_auction_id_fkey
  FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_auction
  ON invoices (tenant_id, auction_id);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_auction_status
  ON invoices (tenant_id, auction_id, payment_status);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status_issued
  ON invoices (tenant_id, payment_status, issued_at DESC);
