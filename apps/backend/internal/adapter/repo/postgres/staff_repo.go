package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StaffRepo struct {
	DB *pgxpool.Pool
}

func (r *StaffRepo) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission domain.Permission) (bool, error) {
	role, err := r.GetEffectiveRole(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	return domain.HasPermission(role, permission), nil
}

func (r *StaffRepo) GetEffectiveRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.Role, error) {
	const q = `
SELECT COALESCE(
  (SELECT tu.role FROM tenant_users tu WHERE tu.tenant_id = $1 AND tu.user_id = $2 LIMIT 1),
  (SELECT u.role FROM users u WHERE u.tenant_id = $1 AND u.id = $2 LIMIT 1)
)`
	var role domain.Role
	if err := r.DB.QueryRow(ctx, q, tenantID, userID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("user not found")
		}
		return "", err
	}
	return role, nil
}

func (r *StaffRepo) CreateTenantUser(ctx context.Context, item *domain.TenantUser) error {
	const q = `
INSERT INTO tenant_users (id, tenant_id, user_id, role, is_primary_owner, invited_by, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.DB.Exec(ctx, q, item.ID, item.TenantID, item.UserID, item.Role, item.IsPrimaryOwner, item.InvitedBy, item.CreatedAt)
	return mapDBError(err, "")
}

func (r *StaffRepo) ListTenantUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantUser, error) {
	const q = `
SELECT id, tenant_id, user_id, role, is_primary_owner, invited_by, created_at
FROM tenant_users
WHERE tenant_id = $1
ORDER BY is_primary_owner DESC, created_at ASC`
	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TenantUser, 0)
	for rows.Next() {
		var item domain.TenantUser
		if err := rows.Scan(&item.ID, &item.TenantID, &item.UserID, &item.Role, &item.IsPrimaryOwner, &item.InvitedBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *StaffRepo) GetTenantUserByID(ctx context.Context, tenantID, tenantUserID uuid.UUID) (*domain.TenantUser, error) {
	const q = `
SELECT id, tenant_id, user_id, role, is_primary_owner, invited_by, created_at
FROM tenant_users
WHERE tenant_id = $1 AND id = $2`
	var item domain.TenantUser
	if err := r.DB.QueryRow(ctx, q, tenantID, tenantUserID).Scan(&item.ID, &item.TenantID, &item.UserID, &item.Role, &item.IsPrimaryOwner, &item.InvitedBy, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("staff member not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *StaffRepo) GetTenantUserByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantUser, error) {
	const q = `
SELECT id, tenant_id, user_id, role, is_primary_owner, invited_by, created_at
FROM tenant_users
WHERE tenant_id = $1 AND user_id = $2`
	var item domain.TenantUser
	if err := r.DB.QueryRow(ctx, q, tenantID, userID).Scan(&item.ID, &item.TenantID, &item.UserID, &item.Role, &item.IsPrimaryOwner, &item.InvitedBy, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("staff member not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *StaffRepo) UpdateTenantUserRole(ctx context.Context, tenantID, tenantUserID uuid.UUID, role domain.Role) error {
	const q = `UPDATE tenant_users SET role = $3, is_primary_owner = false WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, tenantUserID, role)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("staff member not found")
	}
	return nil
}

func (r *StaffRepo) DeleteTenantUser(ctx context.Context, tenantID, tenantUserID uuid.UUID) error {
	const q = `DELETE FROM tenant_users WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, tenantUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("staff member not found")
	}
	return nil
}

func (r *StaffRepo) CreateInvitation(ctx context.Context, item *domain.TenantInvitation) error {
	const q = `
INSERT INTO tenant_invitations (id, tenant_id, email, role, token_hash, expires_at, accepted_at, invited_by, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.DB.Exec(ctx, q, item.ID, item.TenantID, strings.ToLower(strings.TrimSpace(item.Email)), item.Role, item.TokenHash, item.ExpiresAt, item.AcceptedAt, item.InvitedBy, item.CreatedAt)
	return mapDBError(err, "")
}

func (r *StaffRepo) ListInvitations(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantInvitation, error) {
	const q = `
SELECT id, tenant_id, email, role, token_hash, expires_at, accepted_at, invited_by, created_at
FROM tenant_invitations
WHERE tenant_id = $1 AND accepted_at IS NULL
ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TenantInvitation, 0)
	for rows.Next() {
		var item domain.TenantInvitation
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Email, &item.Role, &item.TokenHash, &item.ExpiresAt, &item.AcceptedAt, &item.InvitedBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *StaffRepo) GetInvitationByID(ctx context.Context, tenantID, invitationID uuid.UUID) (*domain.TenantInvitation, error) {
	const q = `
SELECT id, tenant_id, email, role, token_hash, expires_at, accepted_at, invited_by, created_at
FROM tenant_invitations
WHERE tenant_id = $1 AND id = $2`
	var item domain.TenantInvitation
	if err := r.DB.QueryRow(ctx, q, tenantID, invitationID).Scan(&item.ID, &item.TenantID, &item.Email, &item.Role, &item.TokenHash, &item.ExpiresAt, &item.AcceptedAt, &item.InvitedBy, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *StaffRepo) GetInvitationByTokenHash(ctx context.Context, tenantID uuid.UUID, tokenHash string) (*domain.TenantInvitation, error) {
	const q = `
SELECT id, tenant_id, email, role, token_hash, expires_at, accepted_at, invited_by, created_at
FROM tenant_invitations
WHERE tenant_id = $1 AND token_hash = $2`
	var item domain.TenantInvitation
	if err := r.DB.QueryRow(ctx, q, tenantID, tokenHash).Scan(&item.ID, &item.TenantID, &item.Email, &item.Role, &item.TokenHash, &item.ExpiresAt, &item.AcceptedAt, &item.InvitedBy, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *StaffRepo) MarkInvitationAccepted(ctx context.Context, tenantID, invitationID uuid.UUID, acceptedAt time.Time) error {
	const q = `UPDATE tenant_invitations SET accepted_at = $3 WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, invitationID, acceptedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("invitation not found")
	}
	return nil
}

func (r *StaffRepo) DeleteInvitation(ctx context.Context, tenantID, invitationID uuid.UUID) error {
	const q = `DELETE FROM tenant_invitations WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, invitationID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("invitation not found")
	}
	return nil
}
