ALTER TABLE invoices
  DROP CONSTRAINT IF EXISTS invoices_payment_status_check;

UPDATE invoices
SET payment_status = CASE UPPER(payment_status)
  WHEN 'PAID' THEN 'paid'
  WHEN 'CANCELLED' THEN 'refunded'
  ELSE 'unpaid'
END;

ALTER TABLE invoices
  ADD COLUMN IF NOT EXISTS settlement_status text NOT NULL DEFAULT 'draft',
  ADD COLUMN IF NOT EXISTS settlement_approved_at timestamptz,
  ADD COLUMN IF NOT EXISTS invoice_sent_at timestamptz,
  ADD COLUMN IF NOT EXISTS payment_received_at timestamptz,
  ADD COLUMN IF NOT EXISTS stripe_checkout_session_id text,
  ADD COLUMN IF NOT EXISTS stripe_checkout_url text,
  ADD COLUMN IF NOT EXISTS stripe_payment_intent_id text;

UPDATE invoices
SET settlement_status = CASE UPPER(status)
  WHEN 'PAID' THEN 'paid'
  WHEN 'CANCELLED' THEN 'cancelled'
  ELSE 'draft'
END
WHERE settlement_status NOT IN ('draft', 'approved', 'invoice_sent', 'paid', 'cancelled');

ALTER TABLE invoices
  ADD CONSTRAINT invoices_payment_status_check
  CHECK (payment_status IN ('unpaid', 'payment_pending', 'paid', 'refunded', 'partially_refunded'));

ALTER TABLE invoices
  ADD CONSTRAINT invoices_settlement_status_check
  CHECK (settlement_status IN ('draft', 'approved', 'invoice_sent', 'paid', 'cancelled'));

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_settlement_status
  ON invoices (tenant_id, settlement_status, issued_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_stripe_checkout_session_id
  ON invoices (stripe_checkout_session_id)
  WHERE stripe_checkout_session_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_stripe_payment_intent_id
  ON invoices (stripe_payment_intent_id)
  WHERE stripe_payment_intent_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS payment_webhook_events (
  id uuid PRIMARY KEY,
  provider text NOT NULL,
  event_id text NOT NULL,
  event_type text NOT NULL,
  tenant_id uuid NULL REFERENCES tenants(id) ON DELETE SET NULL,
  invoice_id uuid NULL REFERENCES invoices(id) ON DELETE SET NULL,
  payload jsonb NOT NULL,
  processing_status text NOT NULL DEFAULT 'received',
  error_message text NULL,
  received_at timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz NULL,
  UNIQUE (provider, event_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_webhook_events_received_at
  ON payment_webhook_events (received_at DESC);
