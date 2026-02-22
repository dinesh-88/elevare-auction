package postgres

import (
	"context"
	"encoding/json"
	"net"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct {
	DB *pgxpool.Pool
}

func (r *AuditRepo) Append(ctx context.Context, entry port.AuditEntry) error {
	_, err := r.DB.Exec(ctx, insertAuditSQL(),
		uuid.New(),
		entry.TenantID,
		entry.ActorUserID,
		emptyToNil(entry.ActorRole),
		emptyToUnknown(entry.Action),
		emptyToUnknown(entry.EntityType),
		entry.EntityID,
		emptyToNil(entry.RequestID),
		parseIPOrNil(entry.IP),
		emptyToNil(entry.UserAgent),
		normalizeMetadata(entry.Metadata),
	)
	return err
}

func (r *AuditRepo) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	_, err := tx.Exec(ctx, insertAuditSQL(),
		uuid.New(),
		entry.TenantID,
		entry.ActorUserID,
		emptyToNil(entry.ActorRole),
		emptyToUnknown(entry.Action),
		emptyToUnknown(entry.EntityType),
		entry.EntityID,
		emptyToNil(entry.RequestID),
		parseIPOrNil(entry.IP),
		emptyToNil(entry.UserAgent),
		normalizeMetadata(entry.Metadata),
	)
	return err
}

func (r *AuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := r.DB.Query(ctx, `
SELECT id, tenant_id, actor_user_id, actor_role,
       action, entity_type, entity_id,
       request_id, host(ip), user_agent, metadata, created_at
FROM audit_logs
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.AuditLog, 0, limit)
	for rows.Next() {
		item, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *AuditRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := r.DB.Query(ctx, `
SELECT id, tenant_id, actor_user_id, actor_role,
       action, entity_type, entity_id,
       request_id, host(ip), user_agent, metadata, created_at
FROM audit_logs
WHERE tenant_id = $1
  AND entity_type = $2
  AND entity_id = $3
ORDER BY created_at DESC
LIMIT $4`, tenantID, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.AuditLog, 0, limit)
	for rows.Next() {
		item, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func insertAuditSQL() string {
	return `
INSERT INTO audit_logs (
  id, tenant_id, actor_user_id, actor_role,
  action, entity_type, entity_id,
  request_id, ip, user_agent, metadata
) VALUES (
  $1,$2,$3,$4,
  $5,$6,$7,
  $8,$9,$10,$11
)`
}

func scanAuditLog(row pgx.Row) (*domain.AuditLog, error) {
	var item domain.AuditLog
	var tenantID pgtype.UUID
	var actorUserID pgtype.UUID
	var entityID pgtype.UUID
	var actorRole pgtype.Text
	var requestID pgtype.Text
	var ip pgtype.Text
	var userAgent pgtype.Text
	if err := row.Scan(
		&item.ID,
		&tenantID,
		&actorUserID,
		&actorRole,
		&item.Action,
		&item.EntityType,
		&entityID,
		&requestID,
		&ip,
		&userAgent,
		&item.Metadata,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	if tenantID.Valid {
		id, _ := uuid.FromBytes(tenantID.Bytes[:])
		item.TenantID = &id
	}
	if actorUserID.Valid {
		id, _ := uuid.FromBytes(actorUserID.Bytes[:])
		item.ActorUserID = &id
	}
	if entityID.Valid {
		id, _ := uuid.FromBytes(entityID.Bytes[:])
		item.EntityID = &id
	}
	if actorRole.Valid {
		item.ActorRole = actorRole.String
	}
	if requestID.Valid {
		item.RequestID = requestID.String
	}
	if ip.Valid {
		item.IP = ip.String
	}
	if userAgent.Valid {
		item.UserAgent = userAgent.String
	}
	if len(item.Metadata) == 0 {
		item.Metadata = []byte(`{}`)
	}
	return &item, nil
}

func parseIPOrNil(raw string) any {
	if raw == "" {
		return nil
	}
	parsed := net.ParseIP(raw)
	if parsed == nil {
		return nil
	}
	return parsed
}

func emptyToNil(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func emptyToUnknown(v string) string {
	if v == "" {
		return "UNKNOWN"
	}
	return v
}

func normalizeMetadata(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return raw
}
