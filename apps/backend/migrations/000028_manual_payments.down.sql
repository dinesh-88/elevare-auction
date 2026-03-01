DROP TABLE IF EXISTS payment_webhook_events;

DROP INDEX IF EXISTS idx_invoices_stripe_payment_intent_id;
DROP INDEX IF EXISTS idx_invoices_stripe_checkout_session_id;
DROP INDEX IF EXISTS idx_invoices_tenant_settlement_status;

ALTER TABLE invoices
  DROP CONSTRAINT IF EXISTS invoices_settlement_status_check,
  DROP CONSTRAINT IF EXISTS invoices_payment_status_check;

UPDATE invoices
SET payment_status = CASE payment_status
  WHEN 'paid' THEN 'PAID'
  WHEN 'refunded' THEN 'CANCELLED'
  WHEN 'partially_refunded' THEN 'PAID'
  WHEN 'payment_pending' THEN 'UNPAID'
  ELSE 'UNPAID'
END;

ALTER TABLE invoices
  ADD CONSTRAINT invoices_payment_status_check
  CHECK (payment_status IN ('UNPAID', 'PAID', 'CANCELLED'));

ALTER TABLE invoices
  DROP COLUMN IF EXISTS stripe_payment_intent_id,
  DROP COLUMN IF EXISTS stripe_checkout_url,
  DROP COLUMN IF EXISTS stripe_checkout_session_id,
  DROP COLUMN IF EXISTS payment_received_at,
  DROP COLUMN IF EXISTS invoice_sent_at,
  DROP COLUMN IF EXISTS settlement_approved_at,
  DROP COLUMN IF EXISTS settlement_status;
