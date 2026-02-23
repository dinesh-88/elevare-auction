package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type BidderUseCase struct {
	Users    port.UserRepository
	Bidders  port.BidderRepository
	Auctions port.AuctionRepository
	Audit    port.AuditRepository
}

func (uc *BidderUseCase) List(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Users.GetBidders(ctx, tenantID)
}

func (uc *BidderUseCase) Approve(ctx context.Context, tenantID, userID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return ErrForbidden
	}
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
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderSuspend); err != nil {
		return ErrForbidden
	}
	return uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalRejected)
}

func (uc *BidderUseCase) RequestToBid(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if uc.Bidders == nil || uc.Auctions == nil {
		return nil, ErrValidation
	}
	auction, err := uc.Auctions.GetByID(ctx, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if !auction.RequireBidderApproval {
		return &domain.AuctionBidderApproval{
			ID:        uuid.Nil,
			TenantID:  tenantID,
			AuctionID: auctionID,
			UserID:    userID,
			Status:    domain.ApprovalApproved,
		}, nil
	}
	approval, created, err := uc.Bidders.RequestAuctionApproval(ctx, tenantID, auctionID, userID, now)
	if err != nil {
		return nil, err
	}
	if created && uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"user_id": userID, "status": domain.ApprovalPending})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_APPROVAL_REQUESTED",
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return approval, nil
}

func (uc *BidderUseCase) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, status *domain.ApprovalStatus) ([]domain.AuctionBidderAdminRow, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	return uc.Bidders.ListAuctionBidders(ctx, tenantID, auctionID, status)
}

func (uc *BidderUseCase) ApproveAuctionBidder(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	item, err := uc.Bidders.SetAuctionApproval(ctx, tenantID, auctionID, userID, decidedBy, domain.ApprovalApproved, now)
	if err != nil {
		return nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"user_id": userID, "status": domain.ApprovalApproved})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_APPROVED",
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return item, nil
}

func (uc *BidderUseCase) RejectAuctionBidder(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	item, err := uc.Bidders.SetAuctionApproval(ctx, tenantID, auctionID, userID, decidedBy, domain.ApprovalRejected, now)
	if err != nil {
		return nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"user_id": userID, "status": domain.ApprovalRejected})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_REJECTED",
			EntityType:  "AUCTION",
			EntityID:    &auctionID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return item, nil
}

func (uc *BidderUseCase) GetProfile(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	return uc.Bidders.EnsureTenantBidder(ctx, tenantID, userID)
}

func (uc *BidderUseCase) UpdateProfile(ctx context.Context, tenantID, userID uuid.UUID, in port.UpdateTenantBidderInput) (*domain.TenantBidderProfile, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderSuspend); err != nil {
		return nil, ErrForbidden
	}
	prev, err := uc.Bidders.EnsureTenantBidder(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	item, err := uc.Bidders.UpdateTenantBidder(ctx, tenantID, userID, in)
	if err != nil {
		return nil, err
	}
	if uc.Audit != nil && prev.Status != item.Status {
		auditCtx := middleware.AuditContextFromContext(ctx)
		action := "BIDDER_UNSUSPENDED"
		if item.Status == domain.TenantBidderSuspended {
			action = "BIDDER_SUSPENDED"
		}
		payload, _ := json.Marshal(map[string]any{"status": item.Status})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      action,
			EntityType:  "USER",
			EntityID:    &userID,
			RequestID:   auditCtx.RequestID,
			IP:          auditCtx.IP,
			UserAgent:   auditCtx.UserAgent,
			Metadata:    payload,
		})
	}
	return item, nil
}
