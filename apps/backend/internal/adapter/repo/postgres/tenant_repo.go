package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepo struct {
	DB    *pgxpool.Pool
	Audit port.AuditRepository
}

func scanTenant(row pgx.Row) (*domain.Tenant, error) {
	var t domain.Tenant
	err := row.Scan(
		&t.ID, &t.Name, &t.Subdomain, &t.CustomDomain, &t.DefaultCurrency, &t.Status,
		&t.Country, &t.EstimatedAuctionVolume, &t.Branding,
		&t.BrandingPublished, &t.ContentPublished, &t.BrandingDraft, &t.ContentDraft,
		&t.ContentVersion, &t.ContentSchemaVersion,
		&t.RequireAdminApproval, &t.RequireKYC, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) GetByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	normalizedHost := strings.ToLower(strings.TrimSpace(host))
	if cleanHost, _, err := net.SplitHostPort(normalizedHost); err == nil {
		normalizedHost = cleanHost
	}
	normalizedHost = strings.Split(normalizedHost, ":")[0]
	trimmedAPIHost := strings.TrimPrefix(normalizedHost, "api.")
	const q = `
SELECT id, name, subdomain, custom_domain, default_currency, status, country, estimated_auction_volume, branding,
       branding_published, content_published, branding_draft, content_draft, content_version, content_schema_version,
       require_admin_approval, require_kyc, created_at
FROM tenants
WHERE (subdomain = split_part($1, '.', 1) OR custom_domain = $1 OR custom_domain = $2)
  AND status = 'ACTIVE'
LIMIT 1`
	t, err := scanTenant(r.DB.QueryRow(ctx, q, normalizedHost, trimmedAPIHost))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TenantRepo) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	const q = `
SELECT id, name, subdomain, custom_domain, default_currency, status, country, estimated_auction_volume, branding,
       branding_published, content_published, branding_draft, content_draft, content_version, content_schema_version,
       require_admin_approval, require_kyc, created_at
FROM tenants
WHERE id = $1
LIMIT 1`
	t, err := scanTenant(r.DB.QueryRow(ctx, q, tenantID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("tenant not found")
	}
	return t, err
}

func (r *TenantRepo) CreatePendingTenantWithAdmin(ctx context.Context, in port.CreatePendingTenantWithAdminInput) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insTenant = `
INSERT INTO tenants (
  id, name, subdomain, default_currency, status, country, estimated_auction_volume, branding, require_admin_approval, require_kyc, created_at
) VALUES ($1,$2,$3,'USD','PENDING',$4,$5,'{}'::jsonb,$6,$7,now())`
	if _, err := tx.Exec(ctx, insTenant,
		in.TenantID, in.TenantName, in.Subdomain, in.Country, in.EstimatedAuctionVolume, in.RequireAdminApproval, in.RequireKYC,
	); err != nil {
		return err
	}

	const insUser = `
INSERT INTO users (id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at)
VALUES ($1,$2,$3,$4,'ADMIN','ACTIVE',$5,$6,'NONE',now())`
	if _, err := tx.Exec(ctx, insUser,
		in.AdminUserID, in.TenantID, in.AdminEmail, in.AdminPasswordHash, in.AdminEmailVerified, in.AdminApprovalStatus,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TenantRepo) ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	const q = `
SELECT id, name, subdomain, custom_domain, default_currency, status, country, estimated_auction_volume, branding,
       branding_published, content_published, branding_draft, content_draft, content_version, content_schema_version,
       require_admin_approval, require_kyc, created_at
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

func (r *TenantRepo) UpdateDefaultCurrency(ctx context.Context, tenantID uuid.UUID, currency string) error {
	const q = `UPDATE tenants SET default_currency = $2 WHERE id = $1`
	ct, err := r.DB.Exec(ctx, q, tenantID, currency)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tenant not found")
	}
	return nil
}

func (r *TenantRepo) UpdateSiteDraft(ctx context.Context, tenantID uuid.UUID, brandingDraft, contentDraft json.RawMessage) error {
	const q = `
UPDATE tenants
SET branding_draft = $2,
    content_draft = $3,
    content_schema_version = 1
WHERE id = $1`
	ct, err := r.DB.Exec(ctx, q, tenantID, brandingDraft, contentDraft)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("tenant not found")
	}
	return nil
}

func (r *TenantRepo) PublishSiteDraft(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var beforeVersion int
	if err := tx.QueryRow(ctx, `SELECT content_version FROM tenants WHERE id = $1 FOR UPDATE`, tenantID).Scan(&beforeVersion); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, errors.New("tenant not found")
		}
		return 0, 0, err
	}
	afterVersion := beforeVersion + 1
	if _, err := tx.Exec(ctx, `
UPDATE tenants
SET branding_published = branding_draft,
    content_published = content_draft,
    branding = branding_draft,
    content_version = $2,
    content_schema_version = 1
WHERE id = $1`, tenantID, afterVersion); err != nil {
		return 0, 0, err
	}
	if r.Audit != nil {
		base := middleware.AuditContextFromContext(ctx)
		payload := []byte(`{}`)
		payload = mustJSONMap(map[string]any{
			"before_version": beforeVersion,
			"after_version":  afterVersion,
		})
		if err := r.Audit.AppendTx(ctx, tx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: base.ActorUserID,
			ActorRole:   base.ActorRole,
			Action:      "SITE_PUBLISHED",
			EntityType:  "TENANT_SITE",
			EntityID:    &tenantID,
			RequestID:   base.RequestID,
			IP:          base.IP,
			UserAgent:   base.UserAgent,
			Metadata:    payload,
		}); err != nil {
			return 0, 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return beforeVersion, afterVersion, nil
}

func mustJSONMap(v map[string]any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}
