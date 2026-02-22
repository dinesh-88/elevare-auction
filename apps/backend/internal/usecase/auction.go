package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type AuctionUseCase struct {
	Auctions    port.AuctionRepository
	Tenants     port.TenantRepository
	Broadcaster port.Broadcaster
}

type CreateAuctionInput struct {
	TenantID                uuid.UUID
	Currency                string
	Title                   string
	Description             string
	Metadata                json.RawMessage
	StartTime               time.Time
	EndTime                 time.Time
	AllowManualStart        bool
	AllowManualPause        bool
	AllowManualEnd          bool
	AutoExtendEnabled       bool
	AuctionMode             domain.AuctionMode
	AutoProgressEnabled     bool
	LotDurationSeconds      int
	SnipingWindowSeconds    int
	SnipingExtensionSeconds int
	BidVisibility           domain.BidVisibility
}

func (uc *AuctionUseCase) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	return uc.Auctions.ListPublic(ctx, tenantID, now)
}

func (uc *AuctionUseCase) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	return uc.Auctions.ListAdmin(ctx, tenantID)
}

func (uc *AuctionUseCase) Get(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	return uc.Auctions.GetByID(ctx, tenantID, auctionID)
}

func (uc *AuctionUseCase) Create(ctx context.Context, in CreateAuctionInput) (*domain.Auction, error) {
	if in.Title == "" || !in.EndTime.After(in.StartTime) {
		return nil, ErrValidation
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		tenant, err := uc.Tenants.GetByID(ctx, in.TenantID)
		if err != nil {
			return nil, err
		}
		currency = strings.ToUpper(strings.TrimSpace(tenant.DefaultCurrency))
	}
	if len(currency) != 3 {
		return nil, ErrValidation
	}
	status := domain.AuctionScheduled
	if in.AllowManualStart {
		status = domain.AuctionDraft
	}
	auction := &domain.Auction{
		ID:                      uuid.New(),
		TenantID:                in.TenantID,
		Currency:                currency,
		Title:                   in.Title,
		Description:             in.Description,
		Metadata:                in.Metadata,
		StartTime:               in.StartTime,
		EndTime:                 in.EndTime,
		Status:                  status,
		AllowManualStart:        in.AllowManualStart,
		AllowManualPause:        in.AllowManualPause,
		AllowManualEnd:          in.AllowManualEnd,
		AutoExtendEnabled:       in.AutoExtendEnabled,
		AuctionMode:             in.AuctionMode,
		AutoProgressEnabled:     in.AutoProgressEnabled,
		LotDurationSeconds:      in.LotDurationSeconds,
		SnipingWindowSeconds:    in.SnipingWindowSeconds,
		SnipingExtensionSeconds: in.SnipingExtensionSeconds,
		BidVisibility:           in.BidVisibility,
	}
	if len(auction.Metadata) == 0 {
		auction.Metadata = []byte(`{}`)
	}
	if auction.AuctionMode == "" {
		auction.AuctionMode = domain.AuctionModeTimed
	}
	if err := uc.Auctions.Create(ctx, auction); err != nil {
		return nil, err
	}
	return auction, nil
}

func (uc *AuctionUseCase) Update(ctx context.Context, in CreateAuctionInput, auctionID uuid.UUID) (*domain.Auction, error) {
	if in.Title == "" || !in.EndTime.After(in.StartTime) {
		return nil, ErrValidation
	}
	current, err := uc.Auctions.GetByID(ctx, in.TenantID, auctionID)
	if err != nil {
		return nil, err
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = current.Currency
	}
	if len(currency) != 3 {
		return nil, ErrValidation
	}
	if current.Status == domain.AuctionLive && currency != current.Currency {
		return nil, errors.New("currency cannot be changed while auction is live")
	}
	current.Currency = currency
	current.Title = in.Title
	current.Description = in.Description
	current.Metadata = in.Metadata
	current.StartTime = in.StartTime
	current.EndTime = in.EndTime
	current.AllowManualStart = in.AllowManualStart
	current.AllowManualPause = in.AllowManualPause
	current.AllowManualEnd = in.AllowManualEnd
	current.AutoExtendEnabled = in.AutoExtendEnabled
	current.AuctionMode = in.AuctionMode
	current.AutoProgressEnabled = in.AutoProgressEnabled
	current.LotDurationSeconds = in.LotDurationSeconds
	current.SnipingWindowSeconds = in.SnipingWindowSeconds
	current.SnipingExtensionSeconds = in.SnipingExtensionSeconds
	current.BidVisibility = in.BidVisibility
	if len(current.Metadata) == 0 {
		current.Metadata = []byte(`{}`)
	}
	if current.AuctionMode == "" {
		current.AuctionMode = domain.AuctionModeTimed
	}
	if err := uc.Auctions.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (uc *AuctionUseCase) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return uc.Auctions.Delete(ctx, tenantID, auctionID)
}

func (uc *AuctionUseCase) Start(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionLive); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionLive)
	return nil
}

func (uc *AuctionUseCase) Pause(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionPaused); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionPaused)
	return nil
}

func (uc *AuctionUseCase) End(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	if err := uc.Auctions.UpdateStatus(ctx, tenantID, auctionID, domain.AuctionEnded); err != nil {
		return err
	}
	uc.Broadcaster.BroadcastAuctionStatus(ctx, auctionID, domain.AuctionEnded)
	return nil
}

func (uc *AuctionUseCase) SchedulerTick(ctx context.Context, now time.Time) error {
	return uc.Auctions.RunSchedulerTick(ctx, now)
}
