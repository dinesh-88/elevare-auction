DROP TABLE IF EXISTS invoices;

ALTER TABLE tenants
  DROP COLUMN IF EXISTS invoice_counter,
  DROP COLUMN IF EXISTS invoice_prefix;
