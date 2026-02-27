package worker

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type BidPlacer interface {
	PlaceBid(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error)
}

type BidSimulator struct {
	Tenants            port.TenantRepository
	Auctions           port.AuctionRepository
	Lots               port.LotRepository
	Users              port.UserRepository
	Bidders            port.BidderRepository
	BidPlacer          BidPlacer
	DemoTenantSlug     string
	Interval           time.Duration
	IdleThreshold      time.Duration
	MaxEventsPerMinute int
	Now                func() time.Time
	Logger             *log.Logger

	eventTimes []time.Time
}

func (w *BidSimulator) Enabled() bool {
	return strings.TrimSpace(w.DemoTenantSlug) != ""
}

func (w *BidSimulator) Run(ctx context.Context) error {
	if !w.Enabled() {
		return nil
	}
	interval := w.Interval
	if interval <= 0 {
		interval = 12 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := w.ProcessOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			w.printf("bid simulator tick failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *BidSimulator) ProcessOnce(ctx context.Context) error {
	if !w.Enabled() {
		return nil
	}
	now := w.now()
	if !w.allowAnotherEvent(now) {
		return nil
	}

	tenant, err := w.Tenants.GetBySubdomain(ctx, w.DemoTenantSlug)
	if err != nil || tenant == nil {
		return err
	}

	auction, ok, err := w.findLiveAuction(ctx, tenant.ID, now)
	if err != nil || !ok {
		return err
	}

	simulatorUser, err := w.ensureSimulatorUser(ctx, tenant.ID, auction.ID, auction.RequireBidderApproval, now)
	if err != nil {
		return err
	}

	lots, err := w.Lots.ListByAuction(ctx, tenant.ID, auction.ID)
	if err != nil {
		return err
	}
	if len(lots) == 0 {
		return nil
	}

	lastBidAt, err := w.lastAuctionBidAt(ctx, tenant.ID, lots)
	if err != nil {
		return err
	}
	idleThreshold := w.IdleThreshold
	if idleThreshold <= 0 {
		idleThreshold = 20 * time.Second
	}
	if !lastBidAt.IsZero() && now.Sub(lastBidAt) < idleThreshold {
		return nil
	}

	openLots := make([]domain.Lot, 0, len(lots))
	for _, lot := range lots {
		if lot.Status != domain.LotLive {
			continue
		}
		if lot.EndedAt != nil && !now.Before(*lot.EndedAt) {
			continue
		}
		openLots = append(openLots, lot)
	}
	if len(openLots) == 0 {
		return nil
	}

	lot := openLots[rand.IntN(len(openLots))]
	amount := lot.CurrentPrice + lot.MinIncrement
	if lot.ReservePrice != nil && *lot.ReservePrice > amount {
		amount = *lot.ReservePrice
	}
	if _, err := w.BidPlacer.PlaceBid(ctx, port.PlaceLotBidTxInput{
		TenantID: tenant.ID,
		LotID:    lot.ID,
		UserID:   simulatorUser.ID,
		Amount:   &amount,
		Now:      now,
	}); err != nil {
		w.printf("bid simulator skipped failed bid for lot %s: %v", lot.ID, err)
		return nil
	}

	w.eventTimes = append(w.eventTimes, now)
	return nil
}

func (w *BidSimulator) findLiveAuction(ctx context.Context, tenantID uuid.UUID, now time.Time) (*domain.Auction, bool, error) {
	auctions, err := w.Auctions.ListAdmin(ctx, tenantID)
	if err != nil {
		return nil, false, err
	}
	for i := range auctions {
		if auctions[i].Status != domain.AuctionLive {
			continue
		}
		if !auctions[i].EndTime.IsZero() && !now.Before(auctions[i].EndTime) {
			continue
		}
		return &auctions[i], true, nil
	}
	return nil, false, nil
}

func (w *BidSimulator) ensureSimulatorUser(ctx context.Context, tenantID, auctionID uuid.UUID, requireAuctionApproval bool, now time.Time) (*domain.User, error) {
	user, err := w.Users.GetByEmail(ctx, tenantID, domain.DemoBidSimulatorEmail)
	if err != nil || user == nil {
		user = &domain.User{
			ID:             uuid.New(),
			TenantID:       tenantID,
			Email:          domain.DemoBidSimulatorEmail,
			PasswordHash:   "disabled",
			Role:           domain.RoleBidder,
			Status:         domain.UserStatusActive,
			EmailVerified:  true,
			ApprovalStatus: domain.ApprovalApproved,
			KYCStatus:      domain.KYCNone,
		}
		if createErr := w.Users.Create(ctx, tenantID, user); createErr != nil {
			return nil, createErr
		}
	}
	if _, err := w.Bidders.EnsureTenantBidder(ctx, tenantID, user.ID); err != nil {
		return nil, err
	}
	if requireAuctionApproval {
		if _, err := w.Bidders.SetAuctionApproval(ctx, tenantID, auctionID, user.ID, user.ID, domain.ApprovalApproved, now); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (w *BidSimulator) lastAuctionBidAt(ctx context.Context, tenantID uuid.UUID, lots []domain.Lot) (time.Time, error) {
	var latest time.Time
	for _, lot := range lots {
		bids, err := w.Lots.ListBidsByLot(ctx, tenantID, lot.ID, 1)
		if err != nil {
			return time.Time{}, err
		}
		if len(bids) == 0 {
			continue
		}
		if bids[0].CreatedAt.After(latest) {
			latest = bids[0].CreatedAt
		}
	}
	return latest, nil
}

func (w *BidSimulator) allowAnotherEvent(now time.Time) bool {
	maxEvents := w.MaxEventsPerMinute
	if maxEvents <= 0 {
		maxEvents = 6
	}
	cutoff := now.Add(-time.Minute)
	kept := w.eventTimes[:0]
	for _, ts := range w.eventTimes {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	w.eventTimes = kept
	return len(w.eventTimes) < maxEvents
}

func (w *BidSimulator) now() time.Time {
	if w.Now != nil {
		return w.Now().UTC()
	}
	return time.Now().UTC()
}

func (w *BidSimulator) printf(format string, args ...any) {
	if w.Logger != nil {
		w.Logger.Printf(format, args...)
		return
	}
	log.Printf(format, args...)
}
