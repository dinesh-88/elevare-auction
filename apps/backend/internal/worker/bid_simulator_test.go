package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type fakeBidSimTenantRepo struct {
	tenant    *domain.Tenant
	slugCalls int
}

func (f *fakeBidSimTenantRepo) GetByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimTenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	f.slugCalls++
	return f.tenant, nil
}
func (f *fakeBidSimTenantRepo) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	return f.tenant, nil
}
func (f *fakeBidSimTenantRepo) CreatePendingTenantWithAdmin(ctx context.Context, in port.CreatePendingTenantWithAdminInput) error {
	return errors.New("unused")
}
func (f *fakeBidSimTenantRepo) ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimTenantRepo) UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeBidSimTenantRepo) UpdateDefaultCurrency(ctx context.Context, tenantID uuid.UUID, currency string) error {
	return errors.New("unused")
}
func (f *fakeBidSimTenantRepo) UpdateSiteDraft(ctx context.Context, tenantID uuid.UUID, brandingDraft, contentDraft json.RawMessage) error {
	return errors.New("unused")
}
func (f *fakeBidSimTenantRepo) PublishSiteDraft(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	return 0, 0, errors.New("unused")
}

type fakeBidSimAuctionRepo struct {
	auctions []domain.Auction
}

func (f *fakeBidSimAuctionRepo) Create(ctx context.Context, auction *domain.Auction) error {
	return errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	return f.auctions, nil
}
func (f *fakeBidSimAuctionRepo) GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) Update(ctx context.Context, auction *domain.Auction) error {
	return errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error {
	return errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error {
	return errors.New("unused")
}
func (f *fakeBidSimAuctionRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	return errors.New("unused")
}

type fakeBidSimLotRepo struct {
	lots       []domain.Lot
	bidsByLot  map[uuid.UUID][]domain.Bid
	bidsCalled int
}

func (f *fakeBidSimLotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return f.lots, nil
}
func (f *fakeBidSimLotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimLotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimLotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeBidSimLotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimLotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	f.bidsCalled++
	return f.bidsByLot[lotID], nil
}
func (f *fakeBidSimLotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimLotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	return errors.New("unused")
}

type fakeBidSimUserRepo struct {
	user        *domain.User
	createCalls int
}

func (f *fakeBidSimUserRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	f.createCalls++
	f.user = user
	return nil
}
func (f *fakeBidSimUserRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	return f.user, nil
}
func (f *fakeBidSimUserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	if f.user == nil {
		return nil, errors.New("user not found")
	}
	return f.user, nil
}
func (f *fakeBidSimUserRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimUserRepo) UpdateLocale(ctx context.Context, tenantID, userID uuid.UUID, locale domain.Locale) error {
	return errors.New("unused")
}
func (f *fakeBidSimUserRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeBidSimUserRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	return errors.New("unused")
}

type fakeBidSimBidderRepo struct {
	ensureCalls   int
	approvalCalls int
}

func (f *fakeBidSimBidderRepo) EnsureTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	f.ensureCalls++
	return &domain.TenantBidderProfile{TenantID: tenantID, UserID: userID, Status: domain.TenantBidderActive}, nil
}
func (f *fakeBidSimBidderRepo) GetTenantBidder(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantBidderProfile, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimBidderRepo) UpdateTenantBidder(ctx context.Context, tenantID, userID uuid.UUID, in port.UpdateTenantBidderInput) (*domain.TenantBidderProfile, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimBidderRepo) RequestAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, now time.Time) (*domain.AuctionBidderApproval, bool, error) {
	return nil, false, errors.New("unused")
}
func (f *fakeBidSimBidderRepo) GetAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID) (*domain.AuctionBidderApproval, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimBidderRepo) SetAuctionApproval(ctx context.Context, tenantID, auctionID, userID, decidedBy uuid.UUID, status domain.ApprovalStatus, now time.Time) (*domain.AuctionBidderApproval, error) {
	f.approvalCalls++
	return &domain.AuctionBidderApproval{
		TenantID:      tenantID,
		AuctionID:     auctionID,
		UserID:        userID,
		Status:        status,
		RiskLevel:     domain.AuctionBidderRiskLow,
		DepositStatus: domain.AuctionBidderDepositNotRequired,
	}, nil
}
func (f *fakeBidSimBidderRepo) UpdateAuctionApproval(ctx context.Context, tenantID, auctionID, userID uuid.UUID, in port.UpdateAuctionBidderApprovalInput, now time.Time) (*domain.AuctionBidderApproval, error) {
	return nil, errors.New("unused")
}
func (f *fakeBidSimBidderRepo) ListAuctionBidders(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionBiddersFilter) ([]domain.AuctionBidderAdminRow, error) {
	return nil, errors.New("unused")
}

type fakeBidPlacer struct {
	calls int
	last  port.PlaceLotBidTxInput
}

func (f *fakeBidPlacer) PlaceBid(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	f.calls++
	f.last = in
	return &port.PlaceLotBidTxResult{}, nil
}

func TestBidSimulator_ProcessOnceNoopWithoutDemoTenant(t *testing.T) {
	sim := &BidSimulator{
		Tenants:   &fakeBidSimTenantRepo{},
		Auctions:  &fakeBidSimAuctionRepo{},
		Lots:      &fakeBidSimLotRepo{},
		Users:     &fakeBidSimUserRepo{},
		Bidders:   &fakeBidSimBidderRepo{},
		BidPlacer: &fakeBidPlacer{},
		Now:       func() time.Time { return time.Unix(1_000, 0).UTC() },
	}

	if err := sim.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if sim.Tenants.(*fakeBidSimTenantRepo).slugCalls != 0 {
		t.Fatal("expected tenant lookup to be skipped")
	}
}

func TestBidSimulator_ProcessOnceNoopWhenNoLiveAuction(t *testing.T) {
	tenantID := uuid.New()
	tenants := &fakeBidSimTenantRepo{tenant: &domain.Tenant{ID: tenantID, Subdomain: "demo"}}
	placer := &fakeBidPlacer{}
	sim := &BidSimulator{
		Tenants:        tenants,
		Auctions:       &fakeBidSimAuctionRepo{auctions: []domain.Auction{{ID: uuid.New(), TenantID: tenantID, Status: domain.AuctionScheduled}}},
		Lots:           &fakeBidSimLotRepo{},
		Users:          &fakeBidSimUserRepo{},
		Bidders:        &fakeBidSimBidderRepo{},
		BidPlacer:      placer,
		DemoTenantSlug: "demo",
		Now:            func() time.Time { return time.Unix(1_000, 0).UTC() },
	}

	if err := sim.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if placer.calls != 0 {
		t.Fatal("expected no bid placement")
	}
}

func TestBidSimulator_RespectsIdleThreshold(t *testing.T) {
	now := time.Unix(2_000, 0).UTC()
	tenantID := uuid.New()
	auctionID := uuid.New()
	lotID := uuid.New()
	placer := &fakeBidPlacer{}
	sim := &BidSimulator{
		Tenants: &fakeBidSimTenantRepo{tenant: &domain.Tenant{ID: tenantID, Subdomain: "demo"}},
		Auctions: &fakeBidSimAuctionRepo{auctions: []domain.Auction{{
			ID: auctionID, TenantID: tenantID, Status: domain.AuctionLive, EndTime: now.Add(time.Hour),
		}}},
		Lots: &fakeBidSimLotRepo{
			lots: []domain.Lot{{ID: lotID, AuctionID: auctionID, LotNumber: "12", Status: domain.LotLive, CurrentPrice: 100, MinIncrement: 10, EndedAt: ptrTime(now.Add(time.Minute))}},
			bidsByLot: map[uuid.UUID][]domain.Bid{
				lotID: []domain.Bid{{ID: uuid.New(), UserID: uuid.New(), Amount: 110, CreatedAt: now.Add(-10 * time.Second)}},
			},
		},
		Users:          &fakeBidSimUserRepo{},
		Bidders:        &fakeBidSimBidderRepo{},
		BidPlacer:      placer,
		DemoTenantSlug: "demo",
		IdleThreshold:  20 * time.Second,
		Now:            func() time.Time { return now },
	}

	if err := sim.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if placer.calls != 0 {
		t.Fatal("expected recent bid to suppress simulation")
	}
}

func TestBidSimulator_UsesPlaceBidApplicationFlow(t *testing.T) {
	now := time.Unix(3_000, 0).UTC()
	tenantID := uuid.New()
	auctionID := uuid.New()
	lotID := uuid.New()
	placer := &fakeBidPlacer{}
	sim := &BidSimulator{
		Tenants: &fakeBidSimTenantRepo{tenant: &domain.Tenant{ID: tenantID, Subdomain: "demo"}},
		Auctions: &fakeBidSimAuctionRepo{auctions: []domain.Auction{{
			ID: auctionID, TenantID: tenantID, Status: domain.AuctionLive, EndTime: now.Add(time.Hour),
		}}},
		Lots: &fakeBidSimLotRepo{
			lots:      []domain.Lot{{ID: lotID, AuctionID: auctionID, LotNumber: "7", Status: domain.LotLive, CurrentPrice: 500, MinIncrement: 50, ReservePrice: ptrInt64(650), EndedAt: ptrTime(now.Add(time.Minute))}},
			bidsByLot: map[uuid.UUID][]domain.Bid{},
		},
		Users:          &fakeBidSimUserRepo{},
		Bidders:        &fakeBidSimBidderRepo{},
		BidPlacer:      placer,
		DemoTenantSlug: "demo",
		Now:            func() time.Time { return now },
	}

	if err := sim.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce failed: %v", err)
	}
	if placer.calls != 1 {
		t.Fatalf("expected one bid placement, got %d", placer.calls)
	}
	if placer.last.TenantID != tenantID || placer.last.LotID != lotID {
		t.Fatal("expected simulator to bid on the live demo lot")
	}
	if placer.last.Amount == nil || *placer.last.Amount != 650 {
		t.Fatalf("expected reserve-respecting bid amount, got %+v", placer.last.Amount)
	}
}

func ptrTime(v time.Time) *time.Time {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
