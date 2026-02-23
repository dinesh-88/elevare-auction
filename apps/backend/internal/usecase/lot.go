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

type LotUseCase struct {
	Lots        port.LotRepository
	Auctions    port.AuctionRepository
	Bidders     port.BidderRepository
	Broadcaster port.Broadcaster
	Notifier    *NotificationService
	Audit       port.AuditRepository
}

type CreateLotInput struct {
	TenantID      uuid.UUID
	AuctionID     uuid.UUID
	LotNumber     string
	OrderIndex    int
	Title         string
	Description   string
	Metadata      json.RawMessage
	StartingPrice int64
	ReservePrice  *int64
	MinIncrement  int64
}

func (uc *LotUseCase) Create(ctx context.Context, in CreateLotInput) (*domain.Lot, error) {
	if err := authz.Authorize(ctx, domain.PermissionLotCreate); err != nil {
		return nil, ErrForbidden
	}
	if in.AuctionID == uuid.Nil || in.Title == "" || in.LotNumber == "" || in.MinIncrement <= 0 || in.StartingPrice < 0 {
		return nil, ErrValidation
	}
	lot := &domain.Lot{
		ID:            uuid.New(),
		AuctionID:     in.AuctionID,
		LotNumber:     in.LotNumber,
		OrderIndex:    in.OrderIndex,
		Title:         in.Title,
		Description:   in.Description,
		Metadata:      in.Metadata,
		StartingPrice: in.StartingPrice,
		CurrentPrice:  in.StartingPrice,
		ReservePrice:  in.ReservePrice,
		MinIncrement:  in.MinIncrement,
		Status:        domain.LotPending,
	}
	if len(lot.Metadata) == 0 {
		lot.Metadata = []byte(`{}`)
	}
	if err := uc.Lots.Create(ctx, in.TenantID, lot); err != nil {
		return nil, err
	}
	return lot, nil
}

func (uc *LotUseCase) Update(ctx context.Context, tenantID, lotID uuid.UUID, in CreateLotInput) (*domain.Lot, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionLotUpdate); err != nil {
		return nil, ErrForbidden
	}
	if lotID == uuid.Nil || in.Title == "" || in.LotNumber == "" || in.MinIncrement <= 0 || in.StartingPrice < 0 {
		return nil, ErrValidation
	}
	lot, err := uc.Lots.GetByID(ctx, tenantID, lotID)
	if err != nil {
		return nil, err
	}
	lot.LotNumber = in.LotNumber
	lot.OrderIndex = in.OrderIndex
	lot.Title = in.Title
	lot.Description = in.Description
	if len(in.Metadata) > 0 {
		lot.Metadata = in.Metadata
	}
	lot.StartingPrice = in.StartingPrice
	lot.ReservePrice = in.ReservePrice
	lot.MinIncrement = in.MinIncrement
	if err := uc.Lots.Update(ctx, tenantID, lot); err != nil {
		return nil, err
	}
	return lot, nil
}

func (uc *LotUseCase) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionLotUpdate); err != nil {
		return ErrForbidden
	}
	return uc.Lots.Delete(ctx, tenantID, lotID)
}

func (uc *LotUseCase) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionLotReorder); err != nil {
		return ErrForbidden
	}
	if len(lotIDs) == 0 {
		return ErrValidation
	}
	return uc.Lots.Reorder(ctx, tenantID, auctionID, lotIDs)
}

func (uc *LotUseCase) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionLotStart); err != nil {
		return ErrForbidden
	}
	if err := uc.Lots.Start(ctx, tenantID, lotID, now); err != nil {
		return err
	}
	if uc.Broadcaster != nil {
		uc.Broadcaster.BroadcastLotStatus(ctx, lotID, domain.LotLive)
	}
	return nil
}

func (uc *LotUseCase) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionLotEnd); err != nil {
		return ErrForbidden
	}
	settlement, err := uc.Lots.End(ctx, tenantID, lotID, now)
	if err != nil {
		return err
	}
	if uc.Notifier != nil && settlement != nil && settlement.WinnerUserID != nil && settlement.HammerPrice != nil {
		link := "/lots/" + settlement.LotID.String()
		err := uc.Notifier.Emit(ctx, domain.NotificationEvent{
			TenantID:         tenantID,
			EventType:        domain.NotificationEventLotWon,
			RecipientUserIDs: []uuid.UUID{*settlement.WinnerUserID},
			Payload: map[string]any{
				"auction_title": settlement.AuctionTitle,
				"lot_title":     settlement.LotTitle,
				"lot_id":        settlement.LotID.String(),
				"hammer":        *settlement.HammerPrice,
				"link":          link,
				"dedupe_ref":    "lot_won:" + settlement.LotID.String(),
			},
		})
		if err != nil {
			return err
		}
	}
	if uc.Broadcaster != nil {
		uc.Broadcaster.BroadcastLotStatus(ctx, lotID, domain.LotEnded)
	}
	return nil
}

func (uc *LotUseCase) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	return uc.Lots.OnAuctionStart(ctx, tenantID, auctionID, now)
}

func (uc *LotUseCase) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Lots.ListByAuction(ctx, tenantID, auctionID)
}

func (uc *LotUseCase) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Lots.GetByID(ctx, tenantID, lotID)
}

func (uc *LotUseCase) PlaceBid(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	if err := ensureTenantScope(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if uc.Bidders != nil {
		profile, err := uc.Bidders.EnsureTenantBidder(ctx, in.TenantID, in.UserID)
		if err != nil {
			return nil, err
		}
		if profile.Status == domain.TenantBidderSuspended {
			return nil, ErrBidderSuspended
		}
	}
	if uc.Auctions != nil && uc.Bidders != nil {
		lot, err := uc.Lots.GetByID(ctx, in.TenantID, in.LotID)
		if err != nil {
			return nil, err
		}
		auction, err := uc.Auctions.GetByID(ctx, in.TenantID, lot.AuctionID)
		if err != nil {
			return nil, err
		}
		if auction.RequireBidderApproval {
			approval, err := uc.Bidders.GetAuctionApproval(ctx, in.TenantID, lot.AuctionID, in.UserID)
			if err != nil {
				createdApproval, created, reqErr := uc.Bidders.RequestAuctionApproval(ctx, in.TenantID, lot.AuctionID, in.UserID, in.Now)
				if reqErr != nil {
					return nil, reqErr
				}
				if created && uc.Audit != nil {
					auditCtx := middleware.AuditContextFromContext(ctx)
					payload, _ := json.Marshal(map[string]any{"user_id": in.UserID, "status": createdApproval.Status})
					_ = uc.Audit.Append(ctx, port.AuditEntry{
						TenantID:    &in.TenantID,
						ActorUserID: auditCtx.ActorUserID,
						ActorRole:   auditCtx.ActorRole,
						Action:      "BIDDER_APPROVAL_REQUESTED",
						EntityType:  "AUCTION",
						EntityID:    &lot.AuctionID,
						RequestID:   auditCtx.RequestID,
						IP:          auditCtx.IP,
						UserAgent:   auditCtx.UserAgent,
						Metadata:    payload,
					})
				}
				return nil, ErrApprovalRequired
			}
			switch approval.Status {
			case domain.ApprovalApproved:
			case domain.ApprovalPending:
				return nil, ErrPendingApproval
			default:
				return nil, ErrNotApproved
			}
		}
	}
	res, err := uc.Lots.PlaceBidTx(ctx, in)
	if err != nil {
		return nil, err
	}
	if uc.Notifier != nil && res.PreviousHighestBidder != nil && *res.PreviousHighestBidder != in.UserID {
		link := "/lots/" + res.Lot.ID.String()
		if err := uc.Notifier.Emit(ctx, domain.NotificationEvent{
			TenantID:         in.TenantID,
			EventType:        domain.NotificationEventOutbidded,
			RecipientUserIDs: []uuid.UUID{*res.PreviousHighestBidder},
			Payload: map[string]any{
				"auction_title": res.AuctionTitle,
				"lot_title":     res.Lot.Title,
				"lot_id":        res.Lot.ID.String(),
				"amount":        res.Bid.Amount,
				"bid_id":        res.Bid.ID.String(),
				"link":          link,
				"dedupe_ref":    "outbid:" + res.Bid.ID.String() + ":" + res.PreviousHighestBidder.String(),
			},
		}); err != nil {
			return nil, err
		}
	}
	if uc.Broadcaster != nil {
		uc.Broadcaster.BroadcastLotBid(ctx, res.Lot.ID, res.Bid, res.Lot.CurrentPrice, res.Lot.EndedAt)
		if res.WasExtended && res.ExtensionEndTime != nil {
			uc.Broadcaster.BroadcastLotExtended(ctx, res.Lot.ID, *res.ExtensionEndTime)
		}
	}
	return res, nil
}

func (uc *LotUseCase) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Lots.ListBidsByLot(ctx, tenantID, lotID, limit)
}

func (uc *LotUseCase) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	return uc.Lots.ListWonByUser(ctx, tenantID, userID, limit)
}

func (uc *LotUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Lots.RunSchedulerTick(ctx, now)
}
