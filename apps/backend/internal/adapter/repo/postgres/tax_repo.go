package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaxRepo struct {
	DB *pgxpool.Pool
}

func (r *TaxRepo) CreateRule(ctx context.Context, rule *domain.TaxRule) error {
	const q = `
INSERT INTO tax_rules (id, tenant_id, name, rate_percent, active, created_at)
VALUES ($1,$2,$3,$4,$5,now())`
	_, err := r.DB.Exec(ctx, q, rule.ID, rule.TenantID, rule.Name, rule.RatePercent, rule.Active)
	return mapDBError(err, "A tax rule with these details already exists.")
}

func (r *TaxRepo) UpdateRule(ctx context.Context, rule *domain.TaxRule) error {
	const q = `
UPDATE tax_rules
SET name = $3,
    rate_percent = $4
WHERE tenant_id = $1
  AND id = $2`
	ct, err := r.DB.Exec(ctx, q, rule.TenantID, rule.ID, rule.Name, rule.RatePercent)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tax rule not found")
	}
	return nil
}

func (r *TaxRepo) ListRules(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.TaxRule, error) {
	q := `
SELECT id, tenant_id, name, rate_percent::double precision, active, created_at
FROM tax_rules
WHERE tenant_id = $1`
	if activeOnly {
		q += ` AND active = true`
	}
	q += ` ORDER BY created_at DESC`

	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.TaxRule, 0)
	for rows.Next() {
		var rule domain.TaxRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.RatePercent, &rule.Active, &rule.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *TaxRepo) GetRuleByID(ctx context.Context, tenantID, ruleID uuid.UUID) (*domain.TaxRule, error) {
	const q = `
SELECT id, tenant_id, name, rate_percent::double precision, active, created_at
FROM tax_rules
WHERE tenant_id = $1
  AND id = $2`
	var rule domain.TaxRule
	if err := r.DB.QueryRow(ctx, q, tenantID, ruleID).Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.RatePercent, &rule.Active, &rule.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("tax rule not found")
		}
		return nil, err
	}
	return &rule, nil
}

func (r *TaxRepo) DeactivateRule(ctx context.Context, tenantID, ruleID uuid.UUID) error {
	const q = `UPDATE tax_rules SET active = false WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, ruleID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tax rule not found")
	}
	return nil
}

func (r *TaxRepo) CreateAssignment(ctx context.Context, tenantID uuid.UUID, assignment *domain.TaxRuleAssignment) error {
	const q = `
INSERT INTO tax_rule_assignments (id, scope_type, scope_id, tax_rule_id, created_at)
VALUES ($1,$2,$3,$4,now())`
	_, err := r.DB.Exec(ctx, q, assignment.ID, assignment.ScopeType, assignment.ScopeID, assignment.TaxRuleID)
	return mapDBError(err, "This tax rule is already assigned to that scope.")
}

func (r *TaxRepo) ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.TaxRuleScopeType, scopeID uuid.UUID) ([]domain.TaxRuleAssignment, error) {
	const q = `
SELECT ta.id, ta.scope_type, ta.scope_id, ta.tax_rule_id, ta.created_at
FROM tax_rule_assignments ta
JOIN tax_rules tr ON tr.id = ta.tax_rule_id
WHERE ta.scope_type = $1
  AND ta.scope_id = $2
  AND tr.tenant_id = $3
ORDER BY ta.created_at DESC`
	rows, err := r.DB.Query(ctx, q, scopeType, scopeID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.TaxRuleAssignment, 0)
	for rows.Next() {
		var assignment domain.TaxRuleAssignment
		if err := rows.Scan(&assignment.ID, &assignment.ScopeType, &assignment.ScopeID, &assignment.TaxRuleID, &assignment.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, assignment)
	}
	return out, rows.Err()
}

func (r *TaxRepo) DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error {
	const q = `
DELETE FROM tax_rule_assignments ta
USING tax_rules tr
WHERE ta.tax_rule_id = tr.id
  AND tr.tenant_id = $1
  AND ta.id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, assignmentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tax rule assignment not found")
	}
	return nil
}
