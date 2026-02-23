package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuditUseCase struct {
	Audit port.AuditRepository
}

type AuditListInput struct {
	Limit           int
	CursorCreatedAt *time.Time
	CursorID        *uuid.UUID
	EventType       string
	ActorUserID     *uuid.UUID
	From            *time.Time
	To              *time.Time
}

type AuditListResult struct {
	Items   []domain.AuditLog
	HasMore bool
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

func (uc *AuditUseCase) ListByEntityPage(
	ctx context.Context,
	tenantID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	input AuditListInput,
) (AuditListResult, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return AuditListResult{}, err
	}
	if err := authz.Authorize(ctx, domain.PermissionAuditView); err != nil {
		return AuditListResult{}, ErrForbidden
	}
	if tenantID == uuid.Nil || entityID == uuid.Nil {
		return AuditListResult{}, ErrValidation
	}
	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	if entityType == "" {
		return AuditListResult{}, ErrValidation
	}
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	input.EventType = strings.ToUpper(strings.TrimSpace(input.EventType))

	page, err := uc.Audit.ListByEntityPage(ctx, tenantID, entityType, entityID, port.AuditListQuery{
		Limit:           input.Limit,
		CursorCreatedAt: input.CursorCreatedAt,
		CursorID:        input.CursorID,
		EventType:       input.EventType,
		ActorUserID:     input.ActorUserID,
		From:            input.From,
		To:              input.To,
	})
	if err != nil {
		return AuditListResult{}, err
	}
	return AuditListResult{Items: page.Items, HasMore: page.HasMore}, nil
}
