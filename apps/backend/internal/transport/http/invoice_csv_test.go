package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type csvTestInvoiceRepo struct {
	tenantExpected uuid.UUID
	rows           []domain.SettlementInvoiceRow
}

func (r *csvTestInvoiceRepo) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	return nil, nil
}
func (r *csvTestInvoiceRepo) GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error) {
	return nil, nil
}
func (r *csvTestInvoiceRepo) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return nil, nil
}
func (r *csvTestInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	return nil
}
func (r *csvTestInvoiceRepo) GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, bool, bool, error) {
	return nil, false, false, nil
}
func (r *csvTestInvoiceRepo) ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (r *csvTestInvoiceRepo) ListTenantInvoices(ctx context.Context, tenantID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	return nil, 0, nil
}
func (r *csvTestInvoiceRepo) ApproveSettlement(ctx context.Context, tenantID, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementApproved}, nil
}
func (r *csvTestInvoiceRepo) RecordInvoiceDispatch(ctx context.Context, tenantID, invoiceID uuid.UUID, session port.CheckoutSession, now time.Time) (*domain.Invoice, error) {
	return &domain.Invoice{ID: invoiceID, TenantID: tenantID, SettlementStatus: domain.InvoiceSettlementInvoiceSent}, nil
}
func (r *csvTestInvoiceRepo) MarkPaidByWebhook(ctx context.Context, event port.PaymentWebhookEvent, now time.Time) (*domain.Invoice, bool, error) {
	return &domain.Invoice{ID: event.InvoiceID, TenantID: event.TenantID, PaymentStatus: domain.InvoicePaymentPaid}, false, nil
}
func (r *csvTestInvoiceRepo) MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error) {
	return nil, nil
}
func (r *csvTestInvoiceRepo) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error) {
	return nil, nil
}
func (r *csvTestInvoiceRepo) StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.StreamAuctionInvoicesFilter, fn func(domain.SettlementInvoiceRow) error) error {
	if tenantID != r.tenantExpected {
		return usecase.ErrForbidden
	}
	for _, row := range r.rows {
		if err := fn(row); err != nil {
			return err
		}
	}
	return nil
}

type csvNoopAudit struct{}

func (n *csvNoopAudit) Append(ctx context.Context, entry port.AuditEntry) error { return nil }
func (n *csvNoopAudit) AppendTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	return nil
}
func (n *csvNoopAudit) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (n *csvNoopAudit) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return nil, nil
}
func (n *csvNoopAudit) ListByEntityPage(
	ctx context.Context,
	tenantID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	query port.AuditListQuery,
) (port.AuditListPage, error) {
	return port.AuditListPage{Items: []domain.AuditLog{}, HasMore: false}, nil
}

func TestExportSettlementCSVHeaderAndLines(t *testing.T) {
	tenantID := uuid.New()
	auctionID := uuid.New()
	repo := &csvTestInvoiceRepo{
		tenantExpected: tenantID,
		rows: []domain.SettlementInvoiceRow{
			{ID: uuid.New(), InvoiceNumber: "INV-1", BidderName: "alpha", BidderEmail: "alpha@example.com", LotSummary: "Lot A", HammerMinor: 1000, FeesMinor: 100, TaxMinor: 10, TotalDueMinor: 1110, Currency: "USD", PaymentStatus: domain.InvoicePaymentUnpaid, IssuedAt: time.Now().UTC()},
			{ID: uuid.New(), InvoiceNumber: "INV-2", BidderName: "beta", BidderEmail: "beta@example.com", LotSummary: "Lot B", HammerMinor: 2000, FeesMinor: 200, TaxMinor: 20, TotalDueMinor: 2220, Currency: "USD", PaymentStatus: domain.InvoicePaymentPaid, IssuedAt: time.Now().UTC()},
		},
	}
	h := &Handler{Invoice: &usecase.InvoiceUseCase{Invoices: repo, Audit: &csvNoopAudit{}}}

	req := httptest.NewRequest(http.MethodGet, "/admin/auctions/"+auctionID.String()+"/export-settlement-csv", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("auctionId", auctionID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = middleware.WithTenant(ctx, &domain.Tenant{ID: tenantID})
	ctx = middleware.WithClaims(ctx, &port.TokenClaims{TenantID: tenantID, UserID: uuid.New(), Role: domain.RoleAccounting})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ExportAuctionSettlementCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rr.Code, rr.Body.String())
	}
	lines := strings.Split(strings.TrimSpace(rr.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 csv lines (header+2), got %d", len(lines))
	}
	if !strings.Contains(lines[0], "invoice_number") {
		t.Fatalf("unexpected header: %s", lines[0])
	}
}

func TestExportSettlementCSVCrossTenantForbidden(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	auctionID := uuid.New()
	repo := &csvTestInvoiceRepo{tenantExpected: tenantB}
	h := &Handler{Invoice: &usecase.InvoiceUseCase{Invoices: repo, Audit: &csvNoopAudit{}}}

	req := httptest.NewRequest(http.MethodGet, "/admin/auctions/"+auctionID.String()+"/export-settlement-csv", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("auctionId", auctionID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = middleware.WithTenant(ctx, &domain.Tenant{ID: tenantB})
	ctx = middleware.WithClaims(ctx, &port.TokenClaims{TenantID: tenantA, UserID: uuid.New(), Role: domain.RoleAccounting})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ExportAuctionSettlementCSV(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-tenant, got %d", rr.Code)
	}
}
