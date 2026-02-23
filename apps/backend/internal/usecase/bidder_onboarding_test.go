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
		p = domain.TenantBidderProfile{ID: uuid.New(), TenantID: tenantID, UserID: userID, Status: domain.TenantBidderActive, RiskLevel: domain.TenantBidderRiskNormal}
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
	f.profiles[userID] = *p
	return p, nil
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
		ID:        uuid.New(),
		TenantID:  tenantID,
		AuctionID: auctionID,
		UserID:    userID,
		Status:    domain.ApprovalPending,
		CreatedAt: now,
		UpdatedAt: now,
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
	item := domain.AuctionBidderApproval{ID: uuid.New(), TenantID: tenantID, AuctionID: auctionID, UserID: userID, Status: status, CreatedAt: now, UpdatedAt: now}
	f.approvals[key] = item
	return &item, nil
}
func (f *fakeBidderApprovalRepo) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, status *domain.ApprovalStatus) ([]domain.AuctionBidderAdminRow, error) {
	return nil, nil
}

func bidderCtx(tenantID, userID uuid.UUID) context.Context {
	return middleware.WithClaims(context.Background(), &port.TokenClaims{
		TenantID: tenantID,
		UserID:   userID,
		Role:     domain.RoleBidder,
	})
}

func TestPlaceBid_AllowsWhenApprovalNotRequired(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	lotRepo := &fakeBidderApprovalLotRepo{lot: domain.Lot{AuctionID: auctionID}}
	auctionRepo := &fakeBidderApprovalAuctionRepo{auction: domain.Auction{RequireBidderApproval: false}}
	bidderRepo := &fakeBidderApprovalRepo{}
	uc := &LotUseCase{Lots: lotRepo, Auctions: auctionRepo, Bidders: bidderRepo}
	_, err := uc.PlaceBid(bidderCtx(tenantID, userID), port.PlaceLotBidTxInput{
		TenantID: tenantID, LotID: uuid.New(), UserID: userID, Amount: int64Ptr(100), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("expected bid to succeed, got %v", err)
	}
}

func TestPlaceBid_CreatesPendingWhenApprovalRequired(t *testing.T) {
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
	if !errors.Is(err, ErrApprovalRequired) {
		t.Fatalf("expected ErrApprovalRequired, got %v", err)
	}
	if !bidderRepo.lastRequestMade {
		t.Fatal("expected approval request to be created")
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

func TestRequestToBid_Idempotent(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	userID := uuid.New()
	repo := &fakeBidderApprovalRepo{}
	uc := &BidderUseCase{
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
	if repo.requestCalls != 2 {
		t.Fatalf("expected 2 request calls, got %d", repo.requestCalls)
	}
	if len(repo.approvals) != 1 {
		t.Fatalf("expected single approval row, got %d", len(repo.approvals))
	}
}

func TestAuctionBidderListCrossTenantForbidden(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	uc := &BidderUseCase{Bidders: &fakeBidderApprovalRepo{}}
	_, err := uc.ListAuctionBidders(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New(), nil)
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
