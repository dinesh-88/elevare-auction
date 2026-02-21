ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS invoice_prefix text NOT NULL DEFAULT 'INV',
  ADD COLUMN IF NOT EXISTS invoice_counter bigint NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS invoices (
  id uuid PRIMARY KEY,
  lot_id uuid NOT NULL UNIQUE REFERENCES lots(id) ON DELETE RESTRICT,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  winner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  invoice_number text NOT NULL,
  issued_at timestamptz NOT NULL,
  due_date timestamptz NULL,
  currency text NOT NULL DEFAULT 'USD',
  total_amount bigint NOT NULL CHECK (total_amount >= 0),
  snapshot jsonb NOT NULL,
  status text NOT NULL DEFAULT 'ISSUED' CHECK (status IN ('ISSUED','PAID','CANCELLED')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, invoice_number)
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_winner ON invoices(tenant_id, winner_user_id, created_at DESC);
