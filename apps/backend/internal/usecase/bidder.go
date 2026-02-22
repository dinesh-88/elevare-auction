package usecase

import (
	"context"
	"encoding/json"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type BidderUseCase struct {
	Users port.UserRepository
	Audit port.AuditRepository
}

func (uc *BidderUseCase) List(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return uc.Users.GetBidders(ctx, tenantID)
}

func (uc *BidderUseCase) Approve(ctx context.Context, tenantID, userID uuid.UUID) error {
	if err := uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalApproved); err != nil {
		return err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"approval_status": domain.ApprovalApproved})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_APPROVED",
			EntityType:  "BIDDER",
			EntityID:    &userID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return nil
}

func (uc *BidderUseCase) Reject(ctx context.Context, tenantID, userID uuid.UUID) error {
	return uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalRejected)
}
