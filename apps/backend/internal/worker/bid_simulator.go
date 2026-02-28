package worker

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type BidPlacer interface {
	PlaceBid(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error)
}

type BidSimulator struct {
	Tenants            port.TenantRepository
	Auctions           port.AuctionRepository
	Terms              port.AuctionTermsRepository
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

	tenant, err := w.Tenants.GetBySubdomain(ctx, w.DemoTenantSlug)
	if err != nil || tenant == nil {
		return err
	}
	return w.processTenant(ctx, tenant, now, false)
}

func (w *BidSimulator) TriggerForTenant(ctx context.Context, tenant *domain.Tenant) error {
	if !w.Enabled() {
		return errors.New("bid simulator is disabled")
	}
	if tenant == nil {
		return errors.New("tenant not found")
	}
	if !strings.EqualFold(strings.TrimSpace(tenant.Subdomain), strings.TrimSpace(w.DemoTenantSlug)) {
		return errors.New("simulator is restricted to the configured demo tenant")
	}
	return w.processTenant(ctx, tenant, w.now(), true)
}

func (w *BidSimulator) processTenant(ctx context.Context, tenant *domain.Tenant, now time.Time, force bool) error {
	if tenant == nil {
		return nil
	}
	if !force && !w.allowAnotherEvent(now) {
		return nil
	}

	auction, ok, err := w.findLiveAuction(ctx, tenant.ID, now)
	if err != nil {
		return err
	}
	if !ok {
		auction, ok, err = w.ensureAuctionAvailable(ctx, tenant.ID, now)
		if err != nil || !ok {
			return err
		}
	}

	simulatorUser, err := w.ensureSimulatorUser(ctx, tenant.ID, auction.ID, auction.RequireBidderApproval, now)
	if err != nil {
		return err
	}
	bidCtx, err := w.prepareBidContext(ctx, tenant.ID, auction.ID, simulatorUser, now)
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
	if !force && !lastBidAt.IsZero() && now.Sub(lastBidAt) < idleThreshold {
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
		if err := w.ensureLotsAvailable(ctx, tenant.ID, auction.ID, now); err != nil {
			return err
		}
		lots, err = w.Lots.ListByAuction(ctx, tenant.ID, auction.ID)
		if err != nil {
			return err
		}
		openLots = w.filterOpenLots(lots, now)
		if len(openLots) == 0 {
			return nil
		}
	}

	lot := openLots[rand.IntN(len(openLots))]
	amount := lot.CurrentPrice + lot.MinIncrement
	if lot.ReservePrice != nil && *lot.ReservePrice > amount {
		amount = *lot.ReservePrice
	}
	if _, err := w.BidPlacer.PlaceBid(bidCtx, port.PlaceLotBidTxInput{
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

func (w *BidSimulator) ensureAuctionAvailable(ctx context.Context, tenantID uuid.UUID, now time.Time) (*domain.Auction, bool, error) {
	auctions, err := w.Auctions.ListAdmin(ctx, tenantID)
	if err != nil {
		return nil, false, err
	}
	var candidate *domain.Auction
	for i := range auctions {
		if auctions[i].Status == domain.AuctionCancelled {
			continue
		}
		candidate = &auctions[i]
		break
	}
	if candidate == nil {
		return nil, false, nil
	}
	if !candidate.EndTime.After(now) || candidate.StartTime.After(now) {
		candidate.StartTime = minTime(candidate.StartTime, now.Add(-time.Minute))
		if !candidate.EndTime.After(now) {
			candidate.EndTime = now.Add(30 * time.Minute)
		}
		if err := w.Auctions.Update(ctx, candidate); err != nil {
			return nil, false, err
		}
	}
	if candidate.Status != domain.AuctionLive {
		if err := w.Auctions.UpdateStatus(ctx, tenantID, candidate.ID, domain.AuctionLive); err != nil {
			return nil, false, err
		}
		candidate.Status = domain.AuctionLive
	}
	if err := w.Lots.OnAuctionStart(ctx, tenantID, candidate.ID, now); err != nil {
		return nil, false, err
	}
	return candidate, true, nil
}

func (w *BidSimulator) ensureLotsAvailable(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return w.Lots.OnAuctionStart(ctx, tenantID, auctionID, now)
}

func (w *BidSimulator) filterOpenLots(lots []domain.Lot, now time.Time) []domain.Lot {
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
	return openLots
}

func minTime(a, b time.Time) time.Time {
	if a.IsZero() || a.After(b) {
		return b
	}
	return a
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

func (w *BidSimulator) prepareBidContext(ctx context.Context, tenantID, auctionID uuid.UUID, user *domain.User, now time.Time) (context.Context, error) {
	locale, err := w.ensureSimulatorTermsAccepted(ctx, tenantID, auctionID, user, now)
	if err != nil {
		return nil, err
	}
	claims := &port.TokenClaims{
		TenantID: tenantID,
		UserID:   user.ID,
		Role:     domain.RoleBidder,
		Locale:   locale,
	}
	return middleware.WithClaims(ctx, claims), nil
}

func (w *BidSimulator) ensureSimulatorTermsAccepted(ctx context.Context, tenantID, auctionID uuid.UUID, user *domain.User, now time.Time) (domain.Locale, error) {
	locale := domain.NormalizeLocale(string(user.Locale))
	if locale == "" {
		locale = domain.LocaleEnglish
	}
	if w.Terms == nil {
		return locale, nil
	}

	active, activeLocale, err := w.findActiveTerms(ctx, tenantID, auctionID, locale)
	if err != nil {
		return "", err
	}
	if active == nil {
		return activeLocale, nil
	}

	acceptance, err := w.Terms.GetAcceptance(ctx, tenantID, auctionID, user.ID, active.ID)
	if err != nil {
		return "", err
	}
	if acceptance != nil && acceptance.ContentHash == active.ContentHash {
		return activeLocale, nil
	}

	item := &domain.BidderTermsAcceptance{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		AuctionID:          auctionID,
		UserID:             user.ID,
		AuctionTermsID:     active.ID,
		TermsVersionNumber: active.VersionNumber,
		ContentHash:        active.ContentHash,
		AcceptedAt:         now,
		AcceptedIP:         "simulator",
		AcceptedUserAgent:  "bid-simulator",
		CreatedAt:          now,
	}
	if err := w.Terms.CreateAcceptance(ctx, item); err != nil {
		return "", err
	}
	return activeLocale, nil
}

func (w *BidSimulator) findActiveTerms(ctx context.Context, tenantID, auctionID uuid.UUID, preferred domain.Locale) (*domain.AuctionTerms, domain.Locale, error) {
	locales := []domain.Locale{preferred}
	for _, candidate := range []domain.Locale{domain.LocaleEnglish, domain.LocaleFrench, domain.LocaleGerman} {
		if candidate != preferred {
			locales = append(locales, candidate)
		}
	}
	for _, locale := range locales {
		active, err := w.Terms.GetActiveByAuction(ctx, tenantID, auctionID, locale)
		if err != nil {
			return nil, "", err
		}
		if active != nil {
			return active, locale, nil
		}
	}
	return nil, preferred, nil
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
