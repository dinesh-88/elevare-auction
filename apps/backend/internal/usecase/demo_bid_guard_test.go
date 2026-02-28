package usecase

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

type fakeDemoGuardLotRepo struct{}

func (f *fakeDemoGuardLotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	return &port.PlaceLotBidTxResult{
		Lot: domain.Lot{ID: in.LotID, AuctionID: uuid.New()},
		Bid: domain.Bid{ID: uuid.New(), UserID: in.UserID, Amount: *in.Amount, CreatedAt: in.Now},
	}, nil
}
func (f *fakeDemoGuardLotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardLotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	return errors.New("unused")
}

type fakeDemoGuardTenantRepo struct {
	tenant *domain.Tenant
}

func (f *fakeDemoGuardTenantRepo) GetByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	return f.tenant, nil
}
func (f *fakeDemoGuardTenantRepo) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	return f.tenant, nil
}
func (f *fakeDemoGuardTenantRepo) CreatePendingTenantWithAdmin(ctx context.Context, in port.CreatePendingTenantWithAdminInput) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) ListByStatus(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) UpdateStatusAndAdminApproval(ctx context.Context, tenantID uuid.UUID, tenantStatus domain.TenantStatus, adminApproval domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) UpdateProfile(ctx context.Context, tenantID uuid.UUID, name, country string) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) UpdateDefaultCurrency(ctx context.Context, tenantID uuid.UUID, currency string) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) UpdateSiteDraft(ctx context.Context, tenantID uuid.UUID, brandingDraft, contentDraft json.RawMessage) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) PublishSiteDraft(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	return 0, 0, errors.New("unused")
}
func (f *fakeDemoGuardTenantRepo) UpdateOnboardingState(ctx context.Context, tenantID uuid.UUID, status domain.OnboardingStatus, step int, completedAt *time.Time) error {
	return errors.New("unused")
}

type fakeDemoGuardUserRepo struct {
	user *domain.User
}

func (f *fakeDemoGuardUserRepo) Create(ctx context.Context, tenantID uuid.UUID, user *domain.User) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardUserRepo) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	return f.user, nil
}
func (f *fakeDemoGuardUserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	return f.user, nil
}
func (f *fakeDemoGuardUserRepo) GetBidders(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	return nil, errors.New("unused")
}
func (f *fakeDemoGuardUserRepo) UpdateLocale(ctx context.Context, tenantID, userID uuid.UUID, locale domain.Locale) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardUserRepo) UpdateApproval(ctx context.Context, tenantID, userID uuid.UUID, status domain.ApprovalStatus) error {
	return errors.New("unused")
}
func (f *fakeDemoGuardUserRepo) UpdateEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verified bool) error {
	return errors.New("unused")
}

type noopDemoGuardBroadcaster struct{}

func (n *noopDemoGuardBroadcaster) BroadcastLotBid(ctx context.Context, lotID uuid.UUID, bid domain.Bid, currentPrice int64, endedAt *time.Time) {
}
func (n *noopDemoGuardBroadcaster) BroadcastAuctionBid(ctx context.Context, auctionID uuid.UUID, lot domain.Lot, bid domain.Bid, currentPrice int64, endedAt *time.Time) {
}
func (n *noopDemoGuardBroadcaster) BroadcastLotExtended(ctx context.Context, lotID uuid.UUID, newEndTime time.Time) {
}
func (n *noopDemoGuardBroadcaster) BroadcastLotStatus(ctx context.Context, lotID uuid.UUID, status domain.LotStatus) {
}
func (n *noopDemoGuardBroadcaster) BroadcastAuctionStatus(ctx context.Context, auctionID uuid.UUID, status domain.AuctionStatus) {
}

func TestPlaceBid_RejectsSimulatorOutsideDemoTenant(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	lotID := uuid.New()
	amount := int64(100)
	uc := &LotUseCase{
		Lots:           &fakeDemoGuardLotRepo{},
		Tenants:        &fakeDemoGuardTenantRepo{tenant: &domain.Tenant{ID: tenantID, Subdomain: "prod"}},
		Users:          &fakeDemoGuardUserRepo{user: &domain.User{ID: userID, TenantID: tenantID, Email: domain.DemoBidSimulatorEmail}},
		Broadcaster:    &noopDemoGuardBroadcaster{},
		DemoTenantSlug: "demo",
	}

	_, err := uc.PlaceBid(context.Background(), port.PlaceLotBidTxInput{
		TenantID: tenantID,
		LotID:    lotID,
		UserID:   userID,
		Amount:   &amount,
		Now:      time.Now().UTC(),
	})
	if !errors.Is(err, ErrSimulatorBlocked) {
		t.Fatalf("expected ErrSimulatorBlocked, got %v", err)
	}
}
