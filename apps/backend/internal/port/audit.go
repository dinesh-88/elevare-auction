package port

import (
	"context"
	"encoding/json"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuditEntry struct {
	TenantID    *uuid.UUID
	ActorUserID *uuid.UUID
	ActorRole   string
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	RequestID   string
	IP          string
	UserAgent   string
	Metadata    json.RawMessage
}

type AuditRepository interface {
	Append(ctx context.Context, entry AuditEntry) error
	AppendTx(ctx context.Context, tx pgx.Tx, entry AuditEntry) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error)
	ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error)
}
