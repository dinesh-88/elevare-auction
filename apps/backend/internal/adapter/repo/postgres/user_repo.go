package postgres

import (
	"context"
	"errors"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	DB *pgxpool.Pool
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const q = `
INSERT INTO users (id, tenant_id, email, password_hash, role, email_verified, approval_status, kyc_status, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,now())`
	_, err := r.DB.Exec(ctx, q,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, user.EmailVerified, user.ApprovalStatus, user.KYCStatus,
	)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	const q = `
SELECT id, tenant_id, email, password_hash, role, email_verified, approval_status, kyc_status, created_at
FROM users WHERE tenant_id = $1 AND email = $2`
	var u domain.User
	err := r.DB.QueryRow(ctx, q, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.EmailVerified, &u.ApprovalStatus, &u.KYCStatus, &u.CreatedAt,
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
SELECT id, tenant_id, email, password_hash, role, email_verified, approval_status, kyc_status, created_at
FROM users WHERE tenant_id = $1 AND role = 'BIDDER' ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.EmailVerified, &u.ApprovalStatus, &u.KYCStatus, &u.CreatedAt); err != nil {
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
