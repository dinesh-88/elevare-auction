package usecase

import (
	"context"
	"encoding/json"
	"fmt"
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
	if uc.Bidders != nil {
		if _, err := uc.Bidders.UpdateTenantBidder(ctx, tenantID, userID, port.UpdateTenantBidderInput{
			BidderStatus: domain.BidderOnboardingApproved,
		}); err != nil {
			return err
		}
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
	if err := uc.Users.UpdateApproval(ctx, tenantID, userID, domain.ApprovalRejected); err != nil {
		return err
	}
	if uc.Bidders != nil {
		if _, err := uc.Bidders.UpdateTenantBidder(ctx, tenantID, userID, port.UpdateTenantBidderInput{
			BidderStatus: domain.BidderOnboardingRejected,
		}); err != nil {
			return err
		}
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"approval_status": domain.ApprovalRejected})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "BIDDER_REJECTED",
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

func (uc *BidderUseCase) RequestToBid(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if uc.Bidders == nil || uc.Auctions == nil {
		return nil, ErrValidation
	}
	if err := uc.EnsureBidderReadyForBidding(ctx, tenantID, userID); err != nil {
		return nil, err
	}
	auction, err := uc.Auctions.GetByID(ctx, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if existing, getErr := uc.Bidders.GetAuctionApproval(ctx, tenantID, auctionID, userID); getErr == nil && existing != nil {
		return existing, nil
	}
	var (
		approval *domain.AuctionBidderApproval
		created  bool
	)
	if auction.RegistrationPolicy == domain.AuctionRegistrationOpen || (!auction.RequireBidderApproval && auction.RegistrationPolicy == "") {
		approval, err = uc.Bidders.SetAuctionApproval(ctx, tenantID, auctionID, userID, userID, domain.ApprovalApproved, now)
		if err != nil {
			return nil, err
		}
		created = true
	} else {
		approval, created, err = uc.Bidders.RequestAuctionApproval(ctx, tenantID, auctionID, userID, now)
		if err != nil {
			return nil, err
		}
	}
	if created && uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{"user_id": userID, "status": approval.Status})
		_ = uc.Audit.Append(ctx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: auditCtx.ActorUserID,
			ActorRole:   auditCtx.ActorRole,
			Action:      "AUCTION_REGISTRATION_REQUESTED",
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

func (uc *BidderUseCase) ListMyAuctionParticipations(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.BidderAuctionParticipation, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	return uc.Bidders.ListUserAuctionParticipations(ctx, tenantID, userID)
}

func (uc *BidderUseCase) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionBiddersFilter) ([]domain.AuctionBidderAdminRow, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	return uc.Bidders.ListAuctionBidders(ctx, tenantID, auctionID, filter)
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
			Action:      "AUCTION_REGISTRATION_APPROVED",
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
			Action:      "AUCTION_REGISTRATION_REJECTED",
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

func (uc *BidderUseCase) BulkUpdateAuctionRegistrations(ctx context.Context, tenantID, auctionID, decidedBy uuid.UUID, registrationIDs []uuid.UUID, action string, now time.Time) (int, int, []string, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return 0, 0, nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return 0, 0, nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return 0, 0, nil, ErrValidation
	}
	if len(registrationIDs) == 0 {
		return 0, 0, nil, ErrValidation
	}
	status := domain.ApprovalApproved
	auditAction := "AUCTION_REGISTRATION_APPROVED"
	switch action {
	case "approve":
		status = domain.ApprovalApproved
	case "reject":
		status = domain.ApprovalRejected
		auditAction = "AUCTION_REGISTRATION_REJECTED"
	default:
		return 0, 0, nil, ErrValidation
	}
	unique := make([]uuid.UUID, 0, len(registrationIDs))
	seen := make(map[uuid.UUID]struct{}, len(registrationIDs))
	for _, id := range registrationIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return 0, 0, nil, ErrValidation
	}
	updatedRows, skipped, err := uc.Bidders.BulkSetAuctionApprovalByID(ctx, tenantID, auctionID, unique, decidedBy, status, now)
	if err != nil {
		return 0, 0, nil, err
	}
	if uc.Audit != nil {
		auditCtx := middleware.AuditContextFromContext(ctx)
		for _, row := range updatedRows {
			payload, _ := json.Marshal(map[string]any{
				"registration_id": row.ID,
				"user_id":         row.UserID,
				"status":          row.Status,
			})
			_ = uc.Audit.Append(ctx, port.AuditEntry{
				TenantID:    &tenantID,
				ActorUserID: auditCtx.ActorUserID,
				ActorRole:   auditCtx.ActorRole,
				Action:      auditAction,
				EntityType:  "AUCTION",
				EntityID:    &auctionID,
				RequestID:   auditCtx.RequestID,
				IP:          auditCtx.IP,
				UserAgent:   auditCtx.UserAgent,
				Metadata:    payload,
			})
		}
	}
	errorsOut := make([]string, 0)
	if len(updatedRows)+skipped < len(unique) {
		errorsOut = append(errorsOut, fmt.Sprintf("some registrations could not be processed"))
	}
	return len(updatedRows), skipped, errorsOut, nil
}

func (uc *BidderUseCase) UpdateAuctionBidderApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, in port.UpdateAuctionBidderApprovalInput, now time.Time) (*domain.AuctionBidderApproval, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidderApprove); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	prev, err := uc.Bidders.GetAuctionApproval(ctx, tenantID, auctionID, userID)
	if err != nil {
		prev = nil
	}
	item, err := uc.Bidders.UpdateAuctionApproval(ctx, tenantID, auctionID, userID, in, now)
	if err != nil {
		return nil, err
	}
	uc.appendAuctionBidderControlsAudit(ctx, tenantID, auctionID, userID, prev, item)
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

func (uc *BidderUseCase) GetMyOnboarding(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidPlace); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	return uc.Bidders.EnsureTenantBidder(ctx, tenantID, userID)
}

func (uc *BidderUseCase) SaveMyOnboarding(ctx context.Context, tenantID, userID uuid.UUID, in port.UpdateTenantBidderInput) (*domain.TenantBidderProfile, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidPlace); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil {
		return nil, ErrValidation
	}
	return uc.Bidders.UpdateTenantBidder(ctx, tenantID, userID, in)
}

func (uc *BidderUseCase) SubmitMyOnboarding(ctx context.Context, tenantID, userID uuid.UUID, requireApproval bool, now time.Time) (*domain.TenantBidderProfile, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionBidPlace); err != nil {
		return nil, ErrForbidden
	}
	if uc.Bidders == nil || uc.Users == nil {
		return nil, ErrValidation
	}
	user, err := uc.Users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	profile, err := uc.Bidders.EnsureTenantBidder(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if !user.EmailVerified || !bidderProfileComplete(profile) || profile.VATClassification == nil || *profile.VATClassification == "" ||
		profile.TermsAcceptedAt == nil || profile.PrivacyAcceptedAt == nil {
		return nil, ErrValidation
	}
	nextStatus := domain.BidderOnboardingApproved
	if requireApproval {
		nextStatus = domain.BidderOnboardingPendingReview
	}
	return uc.Bidders.UpdateTenantBidder(ctx, tenantID, userID, port.UpdateTenantBidderInput{
		BidderStatus: nextStatus,
	})
}

func (uc *BidderUseCase) EnsureBidderReadyForBidding(ctx context.Context, tenantID, userID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if uc.Bidders == nil || uc.Users == nil {
		return ErrValidation
	}
	user, err := uc.Users.GetByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	profile, err := uc.Bidders.EnsureTenantBidder(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if !user.EmailVerified || !bidderProfileComplete(profile) || profile.BidderStatus != domain.BidderOnboardingApproved {
		return ErrBidderOnboardingRequired
	}
	return nil
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

func (uc *BidderUseCase) appendAuctionBidderControlsAudit(
	ctx context.Context,
	tenantID, auctionID, userID uuid.UUID,
	before, after *domain.AuctionBidderApproval,
) {
	if uc.Audit == nil || after == nil {
		return
	}
	changes := map[string]map[string]any{}
	addChange := func(name string, oldValue, newValue any) {
		if oldValue == newValue {
			return
		}
		changes[name] = map[string]any{"before": oldValue, "after": newValue}
	}

	var beforeRisk domain.AuctionBidderRiskLevel
	var beforeLimit *int64
	var beforeDepositRequired bool
	var beforeDepositAmount *int64
	var beforeDepositStatus domain.AuctionBidderDepositStatus
	var beforeNotes *string
	if before != nil {
		beforeRisk = before.RiskLevel
		beforeLimit = before.MaxBidLimitMinor
		beforeDepositRequired = before.DepositRequired
		beforeDepositAmount = before.DepositAmountMinor
		beforeDepositStatus = before.DepositStatus
		beforeNotes = before.InternalNotes
	}
	addChange("risk_level", string(beforeRisk), string(after.RiskLevel))
	addChange("max_bid_limit_minor", nullableInt64Value(beforeLimit), nullableInt64Value(after.MaxBidLimitMinor))
	addChange("deposit_required", beforeDepositRequired, after.DepositRequired)
	addChange("deposit_amount_minor", nullableInt64Value(beforeDepositAmount), nullableInt64Value(after.DepositAmountMinor))
	addChange("deposit_status", string(beforeDepositStatus), string(after.DepositStatus))
	addChange("internal_notes", nullableStringValue(beforeNotes), nullableStringValue(after.InternalNotes))
	if len(changes) == 0 {
		return
	}

	auditCtx := middleware.AuditContextFromContext(ctx)
	payload, _ := json.Marshal(map[string]any{
		"user_id":  userID,
		"changes":  changes,
		"approval": after.Status,
	})
	_ = uc.Audit.Append(ctx, port.AuditEntry{
		TenantID:    &tenantID,
		ActorUserID: auditCtx.ActorUserID,
		ActorRole:   auditCtx.ActorRole,
		Action:      "AUCTION_BIDDER_CONTROLS_UPDATED",
		EntityType:  "AUCTION",
		EntityID:    &auctionID,
		RequestID:   auditCtx.RequestID,
		IP:          auditCtx.IP,
		UserAgent:   auditCtx.UserAgent,
		Metadata:    payload,
	})
}

func nullableInt64Value(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullableStringValue(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func bidderProfileComplete(profile *domain.TenantBidderProfile) bool {
	if profile == nil {
		return false
	}
	required := []*string{
		profile.Phone,
		profile.AddressLine1,
		profile.City,
		profile.PostalCode,
		profile.Country,
	}
	for _, value := range required {
		if value == nil || *value == "" {
			return false
		}
	}
	return true
}
