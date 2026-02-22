package usecase

import (
	"context"

	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/google/uuid"
)

func ensureTenantScope(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return ErrValidation
	}
	uc, ok := middleware.UserContextFromContext(ctx)
	if !ok {
		return nil
	}
	if uc.TenantID != tenantID {
		return ErrForbidden
	}
	return nil
}
