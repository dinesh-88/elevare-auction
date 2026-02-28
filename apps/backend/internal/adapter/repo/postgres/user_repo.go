package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	DB *pgxpool.Pool
}

func (r *UserRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	if tenantID == uuid.Nil || user == nil || user.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	const q = `
INSERT INTO users (id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())`
	status := user.Status
	if status == "" {
		status = domain.UserStatusActive
	}
	_, err := r.DB.Exec(ctx, q,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, status, user.EmailVerified, user.ApprovalStatus, user.KYCStatus,
	)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	const q = `
SELECT id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at
FROM users WHERE tenant_id = $1 AND email = $2`
	var u domain.User
	err := r.DB.QueryRow(ctx, q, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.EmailVerified, &u.ApprovalStatus, &u.KYCStatus, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	const q = `
SELECT id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at
FROM users WHERE tenant_id = $1 AND id = $2`
	var u domain.User
	err := r.DB.QueryRow(ctx, q, tenantID, userID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.EmailVerified, &u.ApprovalStatus, &u.KYCStatus, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	const q = `
SELECT id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at
FROM users WHERE tenant_id = $1 AND role = 'BIDDER' ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.EmailVerified, &u.ApprovalStatus, &u.KYCStatus, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	const q = `UPDATE users SET approval_status = $3 WHERE tenant_id = $1 AND id = $2 AND role = 'BIDDER'`
	ct, err := r.DB.Exec(ctx, q, tenantID, userID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	const q = `UPDATE users SET email_verified = $3 WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, userID, verified)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepo) ListTenantLoginCandidatesByEmail(ctx context.Context, email string) ([]port.TenantLoginCandidate, error) {
	const q = `
SELECT t.id, t.name, t.subdomain, t.custom_domain
FROM users u
JOIN tenants t ON t.id = u.tenant_id
WHERE lower(u.email) = lower($1)
  AND t.status = 'ACTIVE'
  AND u.status = 'ACTIVE'
GROUP BY t.id, t.name, t.subdomain, t.custom_domain
ORDER BY MAX(t.created_at) DESC`
	rows, err := r.DB.Query(ctx, q, strings.TrimSpace(email))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]port.TenantLoginCandidate, 0)
	for rows.Next() {
		var item port.TenantLoginCandidate
		if err := rows.Scan(&item.TenantID, &item.TenantName, &item.Subdomain, &item.CustomDomain); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
