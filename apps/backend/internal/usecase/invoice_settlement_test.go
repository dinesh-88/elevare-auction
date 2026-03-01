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
	"github.com/jackc/pgx/v5"
)

type fakeInvoiceSettlementRepo struct {
	calledSummary  bool
	calledMarkPaid bool
	summary        *domain.AuctionSettlementSummary
	mixedCurrency  bool
}

func (f *fakeInvoiceSettlementRepo) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, WinnerUserID: uuid.New()}, nil
}
func (f *fakeInvoiceSettlementRepo) GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error) {
	return &domain.Invoice{ID: uuid.New(), TenantID: tenantID, LotID: lotID, WinnerUserID: uuid.New()}, nil
}
func (f *fakeInvoiceSettlementRepo) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceSettlementRepo) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	return nil
}
func (f *fakeInvoiceSettlementRepo) GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, bool, bool, error) {
	f.calledSummary = true
	if f.summary == nil {
		return &domain.AuctionSettlementSummary{AuctionID: auctionID}, false, false, nil
	}
	return f.summary, true, f.mixedCurrency, nil
}
func (f *fakeInvoiceSettlementRepo) ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceSettlementRepo) ListTenantInvoices(ctx context.Context, tenantID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (f *fakeInvoiceSettlementRepo) ApproveSettlement(ctx context.Context, tenantID, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementApproved}, nil
}
func (f *fakeInvoiceSettlementRepo) RecordInvoiceDispatch(ctx context.Context, tenantID, invoiceID uuid.UUID, session port.CheckoutSession, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementInvoiceSent}, nil
}
func (f *fakeInvoiceSettlementRepo) MarkPaidByWebhook(ctx context.Context, event port.PaymentWebhookEvent, now time.Time) (*domain.Invoice, bool, error) {
	return &domain.Invoice{ID: event.InvoiceID, TenantID: event.TenantID, PaymentStatus: domain.InvoicePaymentPaid}, false, nil
}
func (f *fakeInvoiceSettlementRepo) MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error) {
	f.calledMarkPaid = true
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, PaymentStatus: domain.InvoicePaymentPaid}, nil
}
func (f *fakeInvoiceSettlementRepo) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error) {
	return nil, errors.New("unused")
}
func (f *fakeInvoiceSettlementRepo) StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.StreamAuctionInvoicesFilter, fn func(domain.SettlementInvoiceRow) error) error {
	return nil
}

type noopAuditRepo struct{}

func (n *noopAuditRepo) Append(ctx context.Context, entry port.AuditEntry) error { return nil }
func (n *noopAuditRepo) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	return nil
}
func (n *noopAuditRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (n *noopAuditRepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (n *noopAuditRepo) ListByEntityPage(
	ctx context.Context,
	tenantID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	query port.AuditListQuery,
) (port.AuditListPage, error) {
	return port.AuditListPage{Items: []domain.AuditLog{}, HasMore: false}, nil
}

func TestInvoiceSettlementSummaryCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeInvoiceSettlementRepo{}
	uc := &InvoiceUseCase{Invoices: repo, Audit: &noopAuditRepo{}}
	ctx := middleware.WithClaims(context.Background(), &port.TokenClaims{TenantID: tenantA, UserID: uuid.New(), Role: domain.RoleAccounting})
	_, err := uc.GetAuctionSettlementSummary(ctx, tenantB, uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledSummary {
		t.Fatal("expected summary repo not called")
	}
}

func TestInvoiceMarkPaidCrossTenantForbiddenBeforeRepo(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	repo := &fakeInvoiceSettlementRepo{}
	uc := &InvoiceUseCase{Invoices: repo, Audit: &noopAuditRepo{}}
	ctx := middleware.WithClaims(context.Background(), &port.TokenClaims{TenantID: tenantA, UserID: uuid.New(), Role: domain.RoleAccounting})
	_, err := uc.MarkPaid(ctx, tenantB, uuid.New(), nil, time.Now().UTC())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.calledMarkPaid {
		t.Fatal("expected mark paid repo not called")
	}
}

func TestInvoiceSettlementOutstandingComputed(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	repo := &fakeInvoiceSettlementRepo{summary: &domain.AuctionSettlementSummary{
		AuctionID:      auctionID,
		TotalDueMinor:  2000,
		TotalPaidMinor: 500,
		Currency:       "USD",
	}}
	uc := &InvoiceUseCase{Invoices: repo, Audit: &noopAuditRepo{}}
	ctx := middleware.WithClaims(context.Background(), &port.TokenClaims{TenantID: tenantID, UserID: uuid.New(), Role: domain.RoleAccounting})
	out, err := uc.GetAuctionSettlementSummary(ctx, tenantID, auctionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.OutstandingMinor != 1500 {
		t.Fatalf("expected outstanding 1500, got %d", out.OutstandingMinor)
	}
}
