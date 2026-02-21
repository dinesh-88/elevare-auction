CREATE TABLE IF NOT EXISTS fee_components (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name text NOT NULL,
  calculation_type text NOT NULL CHECK (calculation_type IN ('PERCENTAGE','FIXED')),
  value numeric NOT NULL CHECK (value >= 0),
  applies_to text NOT NULL CHECK (applies_to IN ('HAMMER','PREMIUM','SUBTOTAL')),
  taxable boolean NOT NULL DEFAULT false,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fee_components_tenant_active ON fee_components(tenant_id, active);

CREATE TABLE IF NOT EXISTS fee_assignments (
  id uuid PRIMARY KEY,
  scope_type text NOT NULL CHECK (scope_type IN ('TENANT','AUCTION','LOT')),
  scope_id uuid NOT NULL,
  fee_component_id uuid NOT NULL REFERENCES fee_components(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (scope_type, scope_id, fee_component_id)
);

CREATE INDEX IF NOT EXISTS idx_fee_assignments_scope ON fee_assignments(scope_type, scope_id);

CREATE OR REPLACE FUNCTION assert_assignment_scope_not_live(p_scope_type text, p_scope_id uuid)
RETURNS void AS $$
DECLARE
  v_status text;
BEGIN
  IF p_scope_type = 'AUCTION' THEN
    SELECT status INTO v_status FROM auctions WHERE id = p_scope_id;
    IF v_status = 'LIVE' THEN
      RAISE EXCEPTION 'cannot modify fee assignments for live auction' USING ERRCODE = '23514';
    END IF;
  ELSIF p_scope_type = 'LOT' THEN
    SELECT a.status INTO v_status
    FROM lots l
    JOIN auctions a ON a.id = l.auction_id
    WHERE l.id = p_scope_id;
    IF v_status = 'LIVE' THEN
      RAISE EXCEPTION 'cannot modify fee assignments for lot in live auction' USING ERRCODE = '23514';
    END IF;
  END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION validate_fee_assignment()
RETURNS trigger AS $$
DECLARE
  v_component_tenant uuid;
  v_scope_tenant uuid;
  v_scope_status text;
BEGIN
  IF TG_OP = 'UPDATE' THEN
    PERFORM assert_assignment_scope_not_live(OLD.scope_type, OLD.scope_id);
  ELSIF TG_OP = 'DELETE' THEN
    PERFORM assert_assignment_scope_not_live(OLD.scope_type, OLD.scope_id);
    RETURN OLD;
  END IF;

  SELECT tenant_id INTO v_component_tenant
  FROM fee_components
  WHERE id = NEW.fee_component_id;
  IF v_component_tenant IS NULL THEN
    RAISE EXCEPTION 'fee_component_id not found' USING ERRCODE = '23503';
  END IF;

  IF NEW.scope_type = 'TENANT' THEN
    IF NEW.scope_id <> v_component_tenant THEN
      RAISE EXCEPTION 'tenant scope_id must match fee component tenant' USING ERRCODE = '23514';
    END IF;
  ELSIF NEW.scope_type = 'AUCTION' THEN
    SELECT tenant_id, status INTO v_scope_tenant, v_scope_status
    FROM auctions
    WHERE id = NEW.scope_id;
    IF v_scope_tenant IS NULL THEN
      RAISE EXCEPTION 'auction scope_id not found' USING ERRCODE = '23503';
    END IF;
    IF v_scope_tenant <> v_component_tenant THEN
      RAISE EXCEPTION 'auction scope tenant mismatch' USING ERRCODE = '23514';
    END IF;
    PERFORM assert_assignment_scope_not_live(NEW.scope_type, NEW.scope_id);
  ELSIF NEW.scope_type = 'LOT' THEN
    SELECT a.tenant_id, a.status INTO v_scope_tenant, v_scope_status
    FROM lots l
    JOIN auctions a ON a.id = l.auction_id
    WHERE l.id = NEW.scope_id;
    IF v_scope_tenant IS NULL THEN
      RAISE EXCEPTION 'lot scope_id not found' USING ERRCODE = '23503';
    END IF;
    IF v_scope_tenant <> v_component_tenant THEN
      RAISE EXCEPTION 'lot scope tenant mismatch' USING ERRCODE = '23514';
    END IF;
    PERFORM assert_assignment_scope_not_live(NEW.scope_type, NEW.scope_id);
  ELSE
    RAISE EXCEPTION 'invalid scope_type' USING ERRCODE = '23514';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_validate_fee_assignment_ins_upd ON fee_assignments;
CREATE TRIGGER trg_validate_fee_assignment_ins_upd
BEFORE INSERT OR UPDATE ON fee_assignments
FOR EACH ROW EXECUTE FUNCTION validate_fee_assignment();

DROP TRIGGER IF EXISTS trg_validate_fee_assignment_del ON fee_assignments;
CREATE TRIGGER trg_validate_fee_assignment_del
BEFORE DELETE ON fee_assignments
FOR EACH ROW EXECUTE FUNCTION validate_fee_assignment();
