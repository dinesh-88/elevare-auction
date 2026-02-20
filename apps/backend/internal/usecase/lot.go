package usecase

import (
	"context"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type LotUseCase struct {
	Lots        port.LotRepository
	Broadcaster port.Broadcaster
}

type CreateLotInput struct {
	TenantID      uuid.UUID
	AuctionID     uuid.UUID
	LotNumber     string
	OrderIndex    int
	Title         string
	Description   string
	StartingPrice int64
	ReservePrice  *int64
	MinIncrement  int64
}

func (uc *LotUseCase) Create(ctx context.Context, in CreateLotInput) (*domain.Lot, error) {
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
		StartingPrice: in.StartingPrice,
		CurrentPrice:  in.StartingPrice,
		ReservePrice:  in.ReservePrice,
		MinIncrement:  in.MinIncrement,
		Status:        domain.LotPending,
	}
	if err := uc.Lots.Create(ctx, in.TenantID, lot); err != nil {
		return nil, err
	}
	return lot, nil
}

func (uc *LotUseCase) Update(ctx context.Context, tenantID, lotID uuid.UUID, in CreateLotInput) (*domain.Lot, error) {
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
	lot.StartingPrice = in.StartingPrice
	lot.ReservePrice = in.ReservePrice
	lot.MinIncrement = in.MinIncrement
	if err := uc.Lots.Update(ctx, tenantID, lot); err != nil {
		return nil, err
	}
	return lot, nil
}

func (uc *LotUseCase) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	return uc.Lots.Delete(ctx, tenantID, lotID)
}

func (uc *LotUseCase) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	if len(lotIDs) == 0 {
		return ErrValidation
	}
	return uc.Lots.Reorder(ctx, tenantID, auctionID, lotIDs)
}

func (uc *LotUseCase) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return uc.Lots.Start(ctx, tenantID, lotID, now)
}

func (uc *LotUseCase) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return uc.Lots.End(ctx, tenantID, lotID, now)
}

func (uc *LotUseCase) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return uc.Lots.OnAuctionStart(ctx, tenantID, auctionID, now)
}

func (uc *LotUseCase) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return uc.Lots.ListByAuction(ctx, tenantID, auctionID)
}

func (uc *LotUseCase) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	return uc.Lots.GetByID(ctx, tenantID, lotID)
}

func (uc *LotUseCase) PlaceBid(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	res, err := uc.Lots.PlaceBidTx(ctx, in)
	if err != nil {
		return nil, err
	}
	uc.Broadcaster.BroadcastLotBid(ctx, res.Lot.ID, res.Bid, res.Lot.CurrentPrice, res.Lot.EndedAt)
	return res, nil
}

func (uc *LotUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Lots.RunSchedulerTick(ctx, now)
}
