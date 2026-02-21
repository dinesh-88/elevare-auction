package usecase

import (
	"context"
	"encoding/json"
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
	Metadata      json.RawMessage
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
	return uc.Lots.Delete(ctx, tenantID, lotID)
}

func (uc *LotUseCase) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	if len(lotIDs) == 0 {
		return ErrValidation
	}
	return uc.Lots.Reorder(ctx, tenantID, auctionID, lotIDs)
}

func (uc *LotUseCase) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	if err := uc.Lots.Start(ctx, tenantID, lotID, now); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastLotStatus(ctx, lotID, domain.LotLive)
	return nil
}

func (uc *LotUseCase) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	if err := uc.Lots.End(ctx, tenantID, lotID, now); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastLotStatus(ctx, lotID, domain.LotEnded)
	return nil
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

func (uc *LotUseCase) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	return uc.Lots.ListBidsByLot(ctx, tenantID, lotID, limit)
}

func (uc *LotUseCase) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return uc.Lots.ListWonByUser(ctx, tenantID, userID, limit)
}

func (uc *LotUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Lots.RunSchedulerTick(ctx, now)
}
