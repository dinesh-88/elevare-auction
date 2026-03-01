package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeBidderApprovalLotRepo struct {
	lot domain.Lot
}

func (f *fakeBidderApprovalLotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	l := f.lot
	l.ID = lotID
	return &l, nil
}
func (f *fakeBidderApprovalLotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	return &port.PlaceLotBidTxResult{
		Lot: domain.Lot{ID: in.LotID, CurrentPrice: 200},
		Bid: domain.Bid{ID: uuid.New(), Amount: 200, CreatedAt: in.Now},
	}, nil
}
func (f *fakeBidderApprovalLotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalLotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	return errors.New("unused")
}

type fakeBidderApprovalAuctionRepo struct {
	auction domain.Auction
}

func (f *fakeBidderApprovalAuctionRepo) Create(ctx context.Context, auction *domain.Auction) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	a := f.auction
	a.ID = auctionID
	a.TenantID = tenantID
	return &a, nil
}
func (f *fakeBidderApprovalAuctionRepo) Update(ctx context.Context, auction *domain.Auction) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalAuctionRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	return errors.New("unused")
}

type fakeBidderApprovalRepo struct {
	profiles        map[uuid.UUID]domain.TenantBidderProfile
	approvals       map[string]domain.AuctionBidderApproval
	requestCalls    int
	lastRequestMade bool
}

func approvalKey(auctionID, userID uuid.UUID) string {
	return auctionID.String() + ":" + userID.String()
}

func (f *fakeBidderApprovalRepo) EnsureTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	if f.profiles == nil {
		f.profiles = map[uuid.UUID]domain.TenantBidderProfile{}
	}
	p, ok := f.profiles[userID]
	if !ok {
		p = domain.TenantBidderProfile{
			ID:           uuid.New(),
			TenantID:     tenantID,
			UserID:       userID,
			Status:       domain.TenantBidderActive,
			RiskLevel:    domain.TenantBidderRiskNormal,
			BidderStatus: domain.BidderOnboardingIncomplete,
		}
		f.profiles[userID] = p
	}
	return &p, nil
}
func (f *fakeBidderApprovalRepo) GetTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	p, ok := f.profiles[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return &p, nil
}
func (f *fakeBidderApprovalRepo) UpdateTenantBidder(ctx context.Context, tenantID, userID uuid.UUID, in port.UpdateTenantBidderInput) (*domain.TenantBidderProfile, error) {
	p, err := f.EnsureTenantBidder(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if in.Status != "" {
		p.Status = in.Status
	}
	if in.BidderStatus != "" {
		p.BidderStatus = in.BidderStatus
	}
	if in.VATClassification != nil {
		p.VATClassification = in.VATClassification
	}
	if in.TermsAcceptedAt != nil {
		p.TermsAcceptedAt = in.TermsAcceptedAt
	}
	if in.PrivacyAcceptedAt != nil {
		p.PrivacyAcceptedAt = in.PrivacyAcceptedAt
	}
	if in.Phone != nil {
		p.Phone = in.Phone
	}
	if in.AddressLine1 != nil {
		p.AddressLine1 = in.AddressLine1
	}
	if in.City != nil {
		p.City = in.City
	}
	if in.PostalCode != nil {
		p.PostalCode = in.PostalCode
	}
	if in.Country != nil {
		p.Country = in.Country
	}
	f.profiles[userID] = *p
	return p, nil
}
func (f *fakeBidderApprovalRepo) ListUserAuctionParticipations(ctx context.Context, tenantID, userID uuid.UUID) ([]domain.BidderAuctionParticipation, error) {
	return nil, nil
}
func (f *fakeBidderApprovalRepo) RequestAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, bool, error) {
	f.requestCalls++
	if f.approvals == nil {
		f.approvals = map[string]domain.AuctionBidderApproval{}
	}
	key := approvalKey(auctionID, userID)
	if existing, ok := f.approvals[key]; ok {
		return &existing, false, nil
	}
	item := domain.AuctionBidderApproval{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AuctionID:     auctionID,
		UserID:        userID,
		Status:        domain.ApprovalPending,
		RiskLevel:     domain.AuctionBidderRiskLow,
		DepositStatus: domain.AuctionBidderDepositNotRequired,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	f.approvals[key] = item
	f.lastRequestMade = true
	return &item, true, nil
}
func (f *fakeBidderApprovalRepo) GetAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID) (*domain.AuctionBidderApproval, error) {
	key := approvalKey(auctionID, userID)
	item, ok := f.approvals[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return &item, nil
}
func (f *fakeBidderApprovalRepo) SetAuctionApproval(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, status domain.ApprovalStatus, now time.Time) (*domain.AuctionBidderApproval, error) {
	if f.approvals == nil {
		f.approvals = map[string]domain.AuctionBidderApproval{}
	}
	key := approvalKey(auctionID, userID)
	item := domain.AuctionBidderApproval{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AuctionID:     auctionID,
		UserID:        userID,
		Status:        status,
		RiskLevel:     domain.AuctionBidderRiskLow,
		DepositStatus: domain.AuctionBidderDepositNotRequired,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	f.approvals[key] = item
	return &item, nil
}
func (f *fakeBidderApprovalRepo) BulkSetAuctionApprovalByID(ctx context.Context, tenantID, auctionID uuid.UUID, registrationIDs []uuid.UUID, decidedBy uuid.UUID, status domain.ApprovalStatus, now time.Time) ([]domain.AuctionBidderApproval, int, error) {
	if f.approvals == nil {
		return nil, len(registrationIDs), nil
	}
	updated := make([]domain.AuctionBidderApproval, 0, len(registrationIDs))
	skipped := 0
	for _, registrationID := range registrationIDs {
		matched := false
		for key, item := range f.approvals {
			if item.ID != registrationID || item.TenantID != tenantID || item.AuctionID != auctionID {
				continue
			}
			matched = true
			if item.Status != domain.ApprovalPending {
				skipped++
				break
			}
			item.Status = status
			item.DecidedBy = &decidedBy
			item.DecidedAt = &now
			if status == domain.ApprovalApproved {
				item.ApprovedByUserID = &decidedBy
				item.ApprovedAt = &now
			} else {
				item.ApprovedByUserID = nil
				item.ApprovedAt = nil
			}
			item.UpdatedAt = now
			f.approvals[key] = item
			updated = append(updated, item)
			break
		}
		if !matched {
			skipped++
		}
	}
	return updated, skipped, nil
}
func (f *fakeBidderApprovalRepo) UpdateAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, in port.UpdateAuctionBidderApprovalInput, now time.Time) (*domain.AuctionBidderApproval, error) {
	if f.approvals == nil {
		f.approvals = map[string]domain.AuctionBidderApproval{}
	}
	key := approvalKey(auctionID, userID)
	item := f.approvals[key]
	if item.ID == uuid.Nil {
		item = domain.AuctionBidderApproval{
			ID:            uuid.New(),
			TenantID:      tenantID,
			AuctionID:     auctionID,
			UserID:        userID,
			Status:        domain.ApprovalPending,
			DepositStatus: domain.AuctionBidderDepositNotRequired,
			CreatedAt:     now,
		}
	}
	item.RiskLevel = in.RiskLevel
	item.MaxBidLimitMinor = in.MaxBidLimitMinor
	item.DepositRequired = in.DepositRequired
	item.DepositAmountMinor = in.DepositAmountMinor
	item.DepositStatus = in.DepositStatus
	item.InternalNotes = in.InternalNotes
	item.UpdatedAt = now
	f.approvals[key] = item
	return &item, nil
}
func (f *fakeBidderApprovalRepo) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionBiddersFilter) ([]domain.AuctionBidderAdminRow, error) {
	return nil, nil
}

type fakeBidderApprovalUserRepo struct {
	users map[uuid.UUID]domain.User
}

func (f *fakeBidderApprovalUserRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	if f.users == nil {
		return nil, errors.New("not found")
	}
	user, ok := f.users[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return &user, nil
}
func (f *fakeBidderApprovalUserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) UpdateLocale(ctx context.Context, tenantID, userID uuid.UUID, locale domain.Locale) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	return errors.New("unused")
}
func (f *fakeBidderApprovalUserRepo) UpdateRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.Role) error {
	return errors.New("unused")
}

func bidderCtx(tenantID, userID uuid.UUID) context.Context {
	return middleware.WithClaims(context.Background(), &port.TokenClaims{
		TenantID: tenantID,
		UserID:   userID,
		Role:     domain.RoleBidder,
	})
}

func TestPlaceBid_AllowsWhenRegistrationApproved(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{
		approvals: map[string]domain.AuctionBidderApproval{
			approvalKey(auctionID, userID): {
				ID:        uuid.New(),
				TenantID:  tenantID,
				AuctionID: auctionID,
				UserID:    userID,
				Status:    domain.ApprovalApproved,
			},
		},
	}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("expected bid to succeed, got %v", err)
	}
}

func TestPlaceBid_RequiresRegistrationBeforeBidding(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: true}}
	bidderRepo := &fakeBidderApprovalRepo{}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrRegistrationRequired) {
		t.Fatalf("expected ErrRegistrationRequired, got %v", err)
	}
	if bidderRepo.lastRequestMade {
		t.Fatal("did not expect registration to be created during PlaceBid")
	}
}

func TestPlaceBid_RejectsSuspendedBidder(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{
		profiles: map[uuid.UUID]domain.TenantBidderProfile{
			userID: {ID: uuid.New(), TenantID: tenantID, UserID: userID, Status: domain.TenantBidderSuspended, RiskLevel: domain.TenantBidderRiskNormal},
		},
	}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrBidderSuspended) {
		t.Fatalf("expected ErrBidderSuspended, got %v", err)
	}
}

func TestPlaceBid_RejectsNonBidderRoles(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}

	_, err := uc.PlaceBid(authCtx(tenantID, domain.RoleViewer), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: uuid.New(), Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestPlaceBid_RejectsAmountAboveAuthorisedLimit(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	limit := int64(150)
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: true}}
	bidderRepo := &fakeBidderApprovalRepo{
		approvals: map[string]domain.AuctionBidderApproval{
			approvalKey(auctionID, userID): {
				ID:               uuid.New(),
				TenantID:         tenantID,
				AuctionID:        auctionID,
				UserID:           userID,
				Status:           domain.ApprovalApproved,
				RiskLevel:        domain.AuctionBidderRiskMedium,
				MaxBidLimitMinor: &limit,
				DepositStatus:    domain.AuctionBidderDepositNotRequired,
			},
		},
	}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(200), Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrBidLimitExceeded) {
		t.Fatalf("expected ErrBidLimitExceeded, got %v", err)
	}
}

func TestPlaceBid_RejectsProxyBidUntilDepositReceived(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: true}}
	bidderRepo := &fakeBidderApprovalRepo{
		approvals: map[string]domain.AuctionBidderApproval{
			approvalKey(auctionID, userID): {
				ID:              uuid.New(),
				TenantID:        tenantID,
				AuctionID:       auctionID,
				UserID:          userID,
				Status:          domain.ApprovalApproved,
				RiskLevel:       domain.AuctionBidderRiskHigh,
				DepositRequired: true,
				DepositStatus:   domain.AuctionBidderDepositPending,
			},
		},
	}
	maxBid := int64(500)
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, MaxBid: &maxBid, Now: time.Now().UTC(),
	})
	if !errors.Is(err, ErrDepositNeeded) {
		t.Fatalf("expected ErrDepositNeeded, got %v", err)
	}
}

func TestRequestToBid_Idempotent(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	repo := &fakeBidderApprovalRepo{}
	phone := "123"
	address := "Street"
	city := "Berlin"
	postalCode := "10115"
	country := "DE"
	repo.profiles = map[uuid.UUID]domain.TenantBidderProfile{
		userID: {
			ID:           uuid.New(),
			TenantID:     tenantID,
			UserID:       userID,
			Status:       domain.TenantBidderActive,
			RiskLevel:    domain.TenantBidderRiskNormal,
			BidderStatus: domain.BidderOnboardingApproved,
			Phone:        &phone,
			AddressLine1: &address,
			City:         &city,
			PostalCode:   &postalCode,
			Country:      &country,
		},
	}
	uc := &BidderUseCase{
		Users: &fakeBidderApprovalUserRepo{
			users: map[uuid.UUID]domain.User{
				userID: {ID: userID, TenantID: tenantID, Role: domain.RoleBidder, EmailVerified: true},
			},
		},
		Bidders:  repo,
		Auctions: &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: true}},
	}
	ctx := bidderCtx(tenantID, userID)
	if _, err := uc.RequestToBid(ctx, tenantID, auctionID, userID, time.Now().UTC()); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if _, err := uc.RequestToBid(ctx, tenantID, auctionID, userID, time.Now().UTC()); err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if repo.requestCalls != 1 {
		t.Fatalf("expected 1 request call, got %d", repo.requestCalls)
	}
	if len(repo.approvals) != 1 {
		t.Fatalf("expected single approval row, got %d", len(repo.approvals))
	}
}

func TestBulkUpdateAuctionRegistrations_OnlyPendingUpdated(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	actorID := uuid.New()
	pendingID := uuid.New()
	approvedID := uuid.New()
	repo := &fakeBidderApprovalRepo{
		approvals: map[string]domain.AuctionBidderApproval{
			approvalKey(auctionID, uuid.New()): {
				ID:        pendingID,
				TenantID:  tenantID,
				AuctionID: auctionID,
				UserID:    uuid.New(),
				Status:    domain.ApprovalPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			approvalKey(auctionID, uuid.New()): {
				ID:        approvedID,
				TenantID:  tenantID,
				AuctionID: auctionID,
				UserID:    uuid.New(),
				Status:    domain.ApprovalApproved,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}
	uc := &BidderUseCase{Bidders: repo}
	updated, skipped, errs, err := uc.BulkUpdateAuctionRegistrations(
		authCtx(tenantID, domain.RoleAdmin),
		tenantID,
		auctionID,
		actorID,
		[]uuid.UUID{pendingID, approvedID, uuid.New()},
		"approve",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected 1 updated, got %d", updated)
	}
	if skipped != 2 {
		t.Fatalf("expected 2 skipped, got %d", skipped)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestAuctionBidderListCrossTenantForbidden(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	uc := &BidderUseCase{Bidders: &fakeBidderApprovalRepo{}}
	_, err := uc.ListAuctionBidders(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New(), port.ListAuctionBiddersFilter{})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestApproveAuctionBidderCrossTenantForbidden(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	uc := &BidderUseCase{Bidders: &fakeBidderApprovalRepo{}}
	_, err := uc.ApproveAuctionBidder(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New(), uuid.New(), uuid.New(), time.Now().UTC())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func int64Ptr(v int64) *int64 { return &v }
