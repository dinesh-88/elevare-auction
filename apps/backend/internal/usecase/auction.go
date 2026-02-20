package usecase

import (
	"context"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuctionUseCase struct {
	Auctions    port.AuctionRepository
	Broadcaster port.Broadcaster
}

type CreateAuctionInput struct {
	TenantID                uuid.UUID
	Title                   string
	Description             string
	StartingPrice           int64
	ReservePrice            *int64
	MinIncrement            int64
	StartTime               time.Time
	EndTime                 time.Time
	AllowManualStart        bool
	AllowManualPause        bool
	AllowManualEnd          bool
	AutoExtendEnabled       bool
	SnipingWindowSeconds    int
	SnipingExtensionSeconds int
	BidVisibility           domain.BidVisibility
}

func (uc *AuctionUseCase) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	return uc.Auctions.ListPublic(ctx, tenantID, now)
}

func (uc *AuctionUseCase) Get(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	return uc.Auctions.GetByID(ctx, tenantID, auctionID)
}

func (uc *AuctionUseCase) Create(ctx context.Context, in CreateAuctionInput) (*domain.Auction, error) {
	if in.Title == "" || in.MinIncrement <= 0 || in.StartingPrice < 0 || !in.EndTime.After(in.StartTime) {
		return nil, ErrValidation
	}
	status := domain.AuctionScheduled
	if in.AllowManualStart {
		status = domain.AuctionDraft
	}
	auction := &domain.Auction{
		ID:                      uuid.New(),
		TenantID:                in.TenantID,
		Title:                   in.Title,
		Description:             in.Description,
		StartingPrice:           in.StartingPrice,
		CurrentPrice:            in.StartingPrice,
		ReservePrice:            in.ReservePrice,
		MinIncrement:            in.MinIncrement,
		StartTime:               in.StartTime,
		EndTime:                 in.EndTime,
		Status:                  status,
		AllowManualStart:        in.AllowManualStart,
		AllowManualPause:        in.AllowManualPause,
		AllowManualEnd:          in.AllowManualEnd,
		AutoExtendEnabled:       in.AutoExtendEnabled,
		SnipingWindowSeconds:    in.SnipingWindowSeconds,
		SnipingExtensionSeconds: in.SnipingExtensionSeconds,
		BidVisibility:           in.BidVisibility,
	}
	if err := uc.Auctions.Create(ctx, auction); err != nil {
		return nil, err
	}
	return auction, nil
}

func (uc *AuctionUseCase) Start(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionLive)
}

func (uc *AuctionUseCase) Pause(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionPaused)
}

func (uc *AuctionUseCase) End(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionEnded)
}

func (uc *AuctionUseCase) PlaceBid(ctx context.Context, in port.PlaceBidTxInput) (*port.PlaceBidTxResult, error) {
	res, err := uc.Auctions.PlaceBidTx(ctx, in)
	if err != nil {
		return nil, err
	}
	uc.Broadcaster.BroadcastBid(ctx, res.Auction.ID, res.Bid, res.Auction.CurrentPrice, res.Auction.EndTime)
	return res, nil
}

func (uc *AuctionUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Auctions.RunSchedulerTick(ctx, now)
}
