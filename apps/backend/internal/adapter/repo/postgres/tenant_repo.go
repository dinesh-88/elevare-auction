package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepo struct {
	DB *pgxpool.Pool
}

func scanTenant(row pgx.Row) (*domain.Tenant, error) {
	var t domain.Tenant
	err := row.Scan(
		&t.ID, &t.Name, &t.Subdomain, &t.CustomDomain, &t.Status,
		&t.Country, &t.EstimatedAuctionVolume, &t.Branding,
		&t.RequireAdminApproval, &t.RequireKYC, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) GetByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	const q = `
SELECT id, name, subdomain, custom_domain, status, country, estimated_auction_volume, branding, require_admin_approval, require_kyc, created_at
FROM tenants
WHERE (subdomain = split_part($1, '.', 1) OR custom_domain = $1)
  AND status = 'ACTIVE'
LIMIT 1`
	t, err := scanTenant(r.DB.QueryRow(ctx, q, host))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TenantRepo) CreatePendingTenantWithAdmin(ctx context.Context, in port.CreatePendingTenantWithAdminInput) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insTenant = `
INSERT INTO tenants (
  id, name, subdomain, status, country, estimated_auction_volume, branding, require_admin_approval, require_kyc, created_at
) VALUES ($1,$2,$3,'PENDING',$4,$5,'{}'::jsonb,$6,$7,now())`
	if _, err := tx.Exec(ctx, insTenant,
		in.TenantID, in.TenantName, in.Subdomain, in.Country, in.EstimatedAuctionVolume, in.RequireAdminApproval, in.RequireKYC,
	); err != nil {
		return err
	}

	const insUser = `
INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, approval_status, kyc_status, created_at)
VALUES ($1,$2,$3,$4,'ADMIN',$5,$6,'NONE',now())`
	if _, err := tx.Exec(ctx, insUser,
		in.AdminUserID, in.TenantID, in.AdminEmail, in.AdminPasswordHash, in.AdminEmailVerified, in.AdminApprovalStatus,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TenantRepo) ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	const q = `
SELECT id, name, subdomain, custom_domain, status, country, estimated_auction_volume, branding, require_admin_approval, require_kyc, created_at
FROM tenants
WHERE status = $1
ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, q, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Tenant, 0)
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *TenantRepo) UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const updTenant = `UPDATE tenants SET status = $2 WHERE id = $1`
	ct, err := tx.Exec(ctx, updTenant, tenantID, tenantStatus)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tenant not found")
	}

	const updAdmin = `UPDATE users SET approval_status = $2 WHERE tenant_id = $1 AND role = 'ADMIN'`
	if _, err := tx.Exec(ctx, updAdmin, tenantID, adminApproval); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
