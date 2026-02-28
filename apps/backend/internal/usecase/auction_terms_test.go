package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeAuctionTermsRepo struct {
	active       *domain.AuctionTerms
	acceptances  map[string]domain.BidderTermsAcceptance
	createdTerms []*domain.AuctionTerms
}

func (f *fakeAuctionTermsRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID, locale domain.Locale) ([]domain.AuctionTerms, error) {
	return nil, nil
}

func (f *fakeAuctionTermsRepo) GetActiveByAuction(ctx context.Context, tenantID, auctionID uuid.UUID, locale domain.Locale) (*domain.AuctionTerms, error) {
	return f.active, nil
}

func (f *fakeAuctionTermsRepo) GetAcceptance(ctx context.Context, tenantID, auctionID, userID, auctionTermsID uuid.UUID) (*domain.BidderTermsAcceptance, error) {
	if f.acceptances == nil {
		return nil, nil
	}
	item, ok := f.acceptances[auctionTermsID.String()+":"+userID.String()]
	if !ok {
		return nil, nil
	}
	return &item, nil
}

func (f *fakeAuctionTermsRepo) CreateVersion(ctx context.Context, item *domain.AuctionTerms) error {
	item.VersionNumber = len(f.createdTerms) + 1
	item.CreatedAt = time.Now().UTC()
	f.createdTerms = append(f.createdTerms, item)
	return nil
}

func (f *fakeAuctionTermsRepo) PublishVersion(ctx context.Context, tenantID, auctionID, auctionTermsID uuid.UUID, locale domain.Locale) error {
	if f.active != nil && f.active.ID == auctionTermsID {
		f.active.IsActive = true
		return nil
	}
	for _, item := range f.createdTerms {
		item.IsActive = item.ID == auctionTermsID
		if item.IsActive {
			f.active = item
			return nil
		}
	}
	return errors.New("auction terms not found")
}

func (f *fakeAuctionTermsRepo) CreateAcceptance(ctx context.Context, item *domain.BidderTermsAcceptance) error {
	if f.acceptances == nil {
		f.acceptances = map[string]domain.BidderTermsAcceptance{}
	}
	f.acceptances[item.AuctionTermsID.String()+":"+item.UserID.String()] = *item
	return nil
}

func TestPlaceBid_RejectsWithoutTermsAcceptance(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	activeTermsID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{}
	termsRepo := &fakeAuctionTermsRepo{
		active: &domain.AuctionTerms{
			ID:            activeTermsID,
			TenantID:      tenantID,
			AuctionID:     auctionID,
			VersionNumber: 1,
			ContentHash:   "hash-a",
			IsActive:      true,
		},
	}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo, Terms: termsRepo}

	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrTermsRequired) {
		t.Fatalf("expected ErrTermsRequired, got %v", err)
	}
}

func TestPlaceBid_RejectsWhenAcceptedTermsHashIsStale(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	activeTermsID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{}
	termsRepo := &fakeAuctionTermsRepo{
		active: &domain.AuctionTerms{
			ID:            activeTermsID,
			TenantID:      tenantID,
			AuctionID:     auctionID,
			VersionNumber: 2,
			ContentHash:   "hash-b",
			IsActive:      true,
		},
		acceptances: map[string]domain.BidderTermsAcceptance{
			activeTermsID.String() + ":" + userID.String(): {
				ID:                 uuid.New(),
				TenantID:           tenantID,
				AuctionID:          auctionID,
				UserID:             userID,
				AuctionTermsID:     activeTermsID,
				TermsVersionNumber: 1,
				ContentHash:        "hash-a",
				AcceptedAt:         time.Now().UTC(),
				CreatedAt:          time.Now().UTC(),
			},
		},
	}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo, Terms: termsRepo}

	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrTermsRequired) {
		t.Fatalf("expected ErrTermsRequired, got %v", err)
	}
}

func TestAcceptActiveTerms_CapturesImmutableAcceptance(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	activeTermsID := uuid.New()
	termsRepo := &fakeAuctionTermsRepo{
		active: &domain.AuctionTerms{
			ID:            activeTermsID,
			TenantID:      tenantID,
			AuctionID:     auctionID,
			VersionNumber: 3,
			ContentHash:   "hash-c",
			IsActive:      true,
		},
	}
	uc := &AuctionTermsUseCase{Terms: termsRepo}
	now := time.Now().UTC()

	item, active, err := uc.AcceptActive(bidderCtx(tenantID, userID), tenantID, auctionID, userID, "en", now, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("expected acceptance to succeed, got %v", err)
	}
	if active == nil || active.ID != activeTermsID {
		t.Fatalf("expected active terms %s, got %#v", activeTermsID, active)
	}
	if item.AuctionTermsID != activeTermsID || item.ContentHash != "hash-c" {
		t.Fatalf("unexpected acceptance snapshot: %#v", item)
	}
	if item.AcceptedIP != "127.0.0.1" || item.AcceptedUserAgent != "test-agent" {
		t.Fatalf("expected request metadata to be stored, got %#v", item)
	}
}

func TestAcceptActiveTerms_RejectsNonBidderRoles(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	termsRepo := &fakeAuctionTermsRepo{
		active: &domain.AuctionTerms{
			ID:            uuid.New(),
			TenantID:      tenantID,
			AuctionID:     auctionID,
			VersionNumber: 1,
			ContentHash:   "hash-d",
			IsActive:      true,
		},
	}
	uc := &AuctionTermsUseCase{Terms: termsRepo}

	_, _, err := uc.AcceptActive(authCtx(tenantID, domain.RoleViewer), tenantID, auctionID, userID, "en", time.Now().UTC(), "127.0.0.1", "test-agent")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
