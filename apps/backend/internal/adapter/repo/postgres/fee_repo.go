package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeeRepo struct {
	DB *pgxpool.Pool
}

func (r *FeeRepo) CreateComponent(ctx context.Context, component *domain.FeeComponent) error {
	const q = `
INSERT INTO fee_components (id, tenant_id, name, calculation_type, value, applies_to, taxable, active, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now())`
	_, err := r.DB.Exec(ctx, q,
		component.ID, component.TenantID, component.Name, component.CalculationType,
		component.Value, component.AppliesTo, component.Taxable, component.Active,
	)
	return mapDBError(err, "A fee component with these details already exists.")
}

func (r *FeeRepo) UpdateComponent(ctx context.Context, component *domain.FeeComponent) error {
	const q = `
UPDATE fee_components
SET name = $3,
    calculation_type = $4,
    value = $5,
    applies_to = $6,
    taxable = $7
WHERE tenant_id = $1
  AND id = $2`
	ct, err := r.DB.Exec(ctx, q,
		component.TenantID, component.ID, component.Name, component.CalculationType,
		component.Value, component.AppliesTo, component.Taxable,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("fee component not found")
	}
	return nil
}

func (r *FeeRepo) ListComponents(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]domain.FeeComponent, error) {
	q := `
SELECT id, tenant_id, name, calculation_type, value::double precision, applies_to, taxable, active, created_at
FROM fee_components
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
	out := make([]domain.FeeComponent, 0)
	for rows.Next() {
		var c domain.FeeComponent
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.CalculationType, &c.Value, &c.AppliesTo, &c.Taxable, &c.Active, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *FeeRepo) GetComponentByID(ctx context.Context, tenantID, componentID uuid.UUID) (*domain.FeeComponent, error) {
	const q = `
SELECT id, tenant_id, name, calculation_type, value::double precision, applies_to, taxable, active, created_at
FROM fee_components
WHERE tenant_id = $1
  AND id = $2`
	var c domain.FeeComponent
	if err := r.DB.QueryRow(ctx, q, tenantID, componentID).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.CalculationType, &c.Value, &c.AppliesTo, &c.Taxable, &c.Active, &c.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("fee component not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *FeeRepo) DeactivateComponent(ctx context.Context, tenantID, componentID uuid.UUID) error {
	const q = `UPDATE fee_components SET active = false WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, componentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("fee component not found")
	}
	return nil
}

func (r *FeeRepo) CreateAssignment(ctx context.Context, tenantID uuid.UUID, assignment *domain.FeeAssignment) error {
	const q = `
INSERT INTO fee_assignments (id, scope_type, scope_id, fee_component_id, created_at)
VALUES ($1,$2,$3,$4,now())`
	_, err := r.DB.Exec(ctx, q, assignment.ID, assignment.ScopeType, assignment.ScopeID, assignment.FeeComponentID)
	return mapDBError(err, "This fee is already assigned to that scope.")
}

func (r *FeeRepo) ListAssignments(ctx context.Context, tenantID uuid.UUID, scopeType domain.FeeScopeType, scopeID uuid.UUID) ([]domain.FeeAssignment, error) {
	const q = `
SELECT fa.id, fa.scope_type, fa.scope_id, fa.fee_component_id, fa.created_at
FROM fee_assignments fa
JOIN fee_components fc ON fc.id = fa.fee_component_id
WHERE fa.scope_type = $1
  AND fa.scope_id = $2
  AND fc.tenant_id = $3
ORDER BY fa.created_at DESC`
	rows, err := r.DB.Query(ctx, q, scopeType, scopeID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FeeAssignment, 0)
	for rows.Next() {
		var a domain.FeeAssignment
		if err := rows.Scan(&a.ID, &a.ScopeType, &a.ScopeID, &a.FeeComponentID, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *FeeRepo) DeleteAssignment(ctx context.Context, tenantID, assignmentID uuid.UUID) error {
	const q = `
DELETE FROM fee_assignments fa
USING fee_components fc
WHERE fa.fee_component_id = fc.id
  AND fc.tenant_id = $1
  AND fa.id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, assignmentID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("fee assignment not found")
	}
	return nil
}
