package usecase

import (
	"context"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type BidderUseCase struct {
	Users port.UserRepository
}

func (uc *BidderUseCase) List(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return uc.Users.GetBidders(ctx, tenantID)
}

func (uc *BidderUseCase) Approve(ctx context.Context, tenantID, userID uuid.UUID) error {
	return uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalApproved)
}

func (uc *BidderUseCase) Reject(ctx context.Context, tenantID, userID uuid.UUID) error {
	return uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalRejected)
}
