package usecase

import (
	"context"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuditUseCase struct {
	Audit port.AuditRepository
}

func (uc *AuditUseCase) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuditView); err != nil {
		return nil, ErrForbidden
	}
	if tenantID == uuid.Nil {
		return nil, ErrValidation
	}
	return uc.Audit.ListByTenant(ctx, tenantID, limit)
}

func (uc *AuditUseCase) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuditView); err != nil {
		return nil, ErrForbidden
	}
	if tenantID == uuid.Nil || entityID == uuid.Nil {
		return nil, ErrValidation
	}
	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	if entityType == "" {
		return nil, ErrValidation
	}
	return uc.Audit.ListByEntity(ctx, tenantID, entityType, entityID, limit)
}
