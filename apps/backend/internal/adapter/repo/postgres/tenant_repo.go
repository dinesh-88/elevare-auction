package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepo struct {
	DB *pgxpool.Pool
}

func (r *TenantRepo) GetByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	const q = `
SELECT id, name, subdomain, custom_domain, branding, require_admin_approval, require_kyc, created_at
FROM tenants
WHERE subdomain = split_part($1, '.', 1) OR custom_domain = $1
LIMIT 1`
	var t domain.Tenant
	err := r.DB.QueryRow(ctx, q, host).Scan(
		&t.ID, &t.Name, &t.Subdomain, &t.CustomDomain, &t.Branding,
		&t.RequireAdminApproval, &t.RequireKYC, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
