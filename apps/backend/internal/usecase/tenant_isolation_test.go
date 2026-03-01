package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func authCtx(tenantID uuid.UUID, role domain.Role) context.Context {
	return middleware.WithClaims(context.Background(), &port.TokenClaims{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Role:     role,
	})
}

type fakeAuctionRepo struct {
	calledGet bool
}

func (f *fakeAuctionRepo) Create(ctx context.Context, auction *domain.Auction) error { return nil }
func (f *fakeAuctionRepo) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	return nil, nil
}
func (f *fakeAuctionRepo) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	return nil, nil
}
func (f *fakeAuctionRepo) GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	f.calledGet = true
	return &domain.Auction{ID: auctionID, TenantID: tenantID}, nil
}
func (f *fakeAuctionRepo) Update(ctx context.Context, auction *domain.Auction) error { return nil }
func (f *fakeAuctionRepo) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	return nil
}
func (f *fakeAuctionRepo) UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error {
	return nil
}
func (f *fakeAuctionRepo) UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error {
	return nil
}
func (f *fakeAuctionRepo) RunSchedulerTick(ctx context.Context, now time.Time) error { return nil }

type fakeLotRepo struct {
	calledGet bool
}

func (f *fakeLotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return nil
}
func (f *fakeLotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return nil
}
func (f *fakeLotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error { return nil }
func (f *fakeLotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	return nil
}
func (f *fakeLotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return nil, nil
}
func (f *fakeLotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	f.calledGet = true
	return &domain.Lot{ID: lotID}, nil
}
func (f *fakeLotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return nil
}
func (f *fakeLotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	return nil, nil
}
func (f *fakeLotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return nil
}
func (f *fakeLotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	return nil, nil
}
func (f *fakeLotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	return nil, nil
}
func (f *fakeLotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return nil, nil
}
func (f *fakeLotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error { return nil }

type fakeInvoiceRepo struct {
	calledGet bool
}

func (f *fakeInvoiceRepo) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	f.calledGet = true
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, WinnerUserID: uuid.New()}, nil
}
func (f *fakeInvoiceRepo) GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error) {
	return &domain.Invoice{ID: uuid.New(), LotID: lotID, TenantID: tenantID, WinnerUserID: uuid.New()}, nil
}
func (f *fakeInvoiceRepo) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	return nil
}
func (f *fakeInvoiceRepo) GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, bool, bool, error) {
	return &domain.AuctionSettlementSummary{AuctionID: auctionID}, true, false, nil
}
func (f *fakeInvoiceRepo) ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceRepo) ListTenantInvoices(ctx context.Context, tenantID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceRepo) ApproveSettlement(ctx context.Context, tenantID, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementApproved}, nil
}
func (f *fakeInvoiceRepo) RecordInvoiceDispatch(ctx context.Context, tenantID, invoiceID uuid.UUID, session port.CheckoutSession, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementInvoiceSent}, nil
}
func (f *fakeInvoiceRepo) MarkPaidByWebhook(ctx context.Context, event port.PaymentWebhookEvent, now time.Time) (*domain.Invoice, bool, error) {
	return &domain.Invoice{ID: event.InvoiceID, TenantID: event.TenantID, PaymentStatus: domain.InvoicePaymentPaid}, false, nil
}
func (f *fakeInvoiceRepo) MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID}, nil
}
func (f *fakeInvoiceRepo) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID}, nil
}
func (f *fakeInvoiceRepo) StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.StreamAuctionInvoicesFilter, fn func(domain.SettlementInvoiceRow) error) error {
	return nil
}

type fakeAuditRepo struct {
	calledListByTenant bool
}

func (f *fakeAuditRepo) Append(ctx context.Context, entry port.AuditEntry) error { return nil }
func (f *fakeAuditRepo) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	return nil
}
func (f *fakeAuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	f.calledListByTenant = true
	return []domain.AuditLog{{ID: uuid.New(), TenantID: &tenantID}}, nil
}
func (f *fakeAuditRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (f *fakeAuditRepo) ListByEntityPage(
	ctx context.Context,
	tenantID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	query port.AuditListQuery,
) (port.AuditListPage, error) {
	return port.AuditListPage{Items: []domain.AuditLog{}, HasMore: false}, nil
}

func TestEnsureTenantScopeRejectsCrossTenant(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	ctx := authCtx(tenantA, domain.RoleAdmin)
	if err := ensureTenantScope(ctx, tenantB); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAuctionGetCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeAuctionRepo{}
	uc := &AuctionUseCase{Auctions: repo}

	_, err := uc.Get(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledGet {
		t.Fatal("expected repo GetByID not to be called")
	}
}

func TestLotGetByIDCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeLotRepo{}
	uc := &LotUseCase{Lots: repo}

	_, err := uc.GetByID(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledGet {
		t.Fatal("expected repo GetByID not to be called")
	}
}

func TestInvoiceGetAuthorizedCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeInvoiceRepo{}
	uc := &InvoiceUseCase{Invoices: repo}

	_, err := uc.GetAuthorized(authCtx(tenantA, domain.RoleAdmin), tenantB, uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledGet {
		t.Fatal("expected repo GetByID not to be called")
	}
}

func TestAuditListByTenantCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeAuditRepo{}
	uc := &AuditUseCase{Audit: repo}

	_, err := uc.ListByTenant(authCtx(tenantA, domain.RoleAdmin), tenantB, 10)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledListByTenant {
		t.Fatal("expected repo ListByTenant not to be called")
	}
}

func TestAuditListByTenantSameTenantAllowed(t *testing.T) {
	tenantA := uuid.New()
	repo := &fakeAuditRepo{}
	uc := &AuditUseCase{Audit: repo}

	rows, err := uc.ListByTenant(authCtx(tenantA, domain.RoleAdmin), tenantA, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !repo.calledListByTenant {
		t.Fatal("expected repo ListByTenant to be called")
	}
}
