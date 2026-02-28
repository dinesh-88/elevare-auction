CREATE TABLE tax_rules (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name text NOT NULL,
  rate_percent numeric(8,4) NOT NULL CHECK (rate_percent >= 0),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX tax_rules_tenant_name_rate_uidx
  ON tax_rules (tenant_id, lower(name), rate_percent);

CREATE TABLE tax_rule_assignments (
  id uuid PRIMARY KEY,
  scope_type text NOT NULL,
  scope_id uuid NOT NULL,
  tax_rule_id uuid NOT NULL REFERENCES tax_rules(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (scope_type, scope_id, tax_rule_id)
);
