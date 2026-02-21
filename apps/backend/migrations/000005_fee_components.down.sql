DROP TRIGGER IF EXISTS trg_validate_fee_assignment_del ON fee_assignments;
DROP TRIGGER IF EXISTS trg_validate_fee_assignment_ins_upd ON fee_assignments;
DROP FUNCTION IF EXISTS validate_fee_assignment();
DROP FUNCTION IF EXISTS assert_assignment_scope_not_live(text, uuid);

DROP TABLE IF EXISTS fee_assignments;
DROP TABLE IF EXISTS fee_components;
