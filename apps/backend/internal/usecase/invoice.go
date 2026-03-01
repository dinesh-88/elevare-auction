package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type InvoiceUseCase struct {
	Invoices    port.InvoiceRepository
	Audit       port.AuditRepository
	Notifier    *NotificationService
	EmailOutbox *EmailOutboxService
	Payments    port.PaymentGateway
}

type InvoiceListFilter struct {
	Status *domain.InvoicePaymentStatus
	Search string
	Limit  int
	Offset int
}

type InvoiceStreamFilter struct {
	Status *domain.InvoicePaymentStatus
	Search string
}

func (uc *InvoiceUseCase) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceSend); err != nil {
		return nil, ErrForbidden
	}
	inv, err := uc.Invoices.GenerateForLot(ctx, tenantID, lotID, now)
	if err != nil {
		return nil, err
	}
	if uc.Notifier != nil {
		link := "/invoices/" + inv.ID.String()
		if err := uc.Notifier.Emit(ctx, domain.NotificationEvent{
			TenantID:         tenantID,
			EventType:        domain.NotificationEventInvoiceIssued,
			RecipientUserIDs: []uuid.UUID{inv.WinnerUserID},
			Payload: map[string]any{
				"auction_id":     inv.AuctionID.String(),
				"invoice_id":     inv.ID.String(),
				"invoice_number": inv.InvoiceNumber,
				"lot_id":         inv.LotID.String(),
				"total_due":      inv.TotalDueMinor,
				"link":           link,
				"dedupe_ref":     "invoice_issued:" + inv.ID.String(),
			},
		}); err != nil {
			return nil, err
		}
	}
	return inv, nil
}

func (uc *InvoiceUseCase) GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceView); err != nil {
		return nil, ErrForbidden
	}
	summary, hasRows, mixedCurrency, err := uc.Invoices.GetAuctionSettlementSummary(ctx, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	if !hasRows {
		return &domain.AuctionSettlementSummary{AuctionID: auctionID}, nil
	}
	if mixedCurrency {
		if uc.Audit != nil {
			auditCtx := middleware.AuditContextFromContext(ctx)
			payload, _ := json.Marshal(map[string]any{"auction_id": auctionID, "reason": "MIXED_CURRENCY"})
			_ = uc.Audit.Append(ctx, port.AuditEntry{
				TenantID:    &tenantID,
				ActorUserID: auditCtx.ActorUserID,
				ActorRole:   auditCtx.ActorRole,
				Action:      "INVOICE_CURRENCY_MISMATCH",
				EntityType:  "AUCTION",
				EntityID:    &auctionID,
				RequestID:   auditCtx.RequestID,
				IP:          auditCtx.IP,
				UserAgent:   auditCtx.UserAgent,
				Metadata:    payload,
			})
		}
		return nil, ErrValidation
	}
	summary.OutstandingMinor = summary.TotalDueMinor - summary.TotalPaidMinor
	return summary, nil
}

func (uc *InvoiceUseCase) ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter InvoiceListFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, 0, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceView); err != nil {
		return nil, 0, ErrForbidden
	}
	rows, total, err := uc.Invoices.ListAuctionInvoices(ctx, tenantID, auctionID, port.ListAuctionInvoicesFilter{
		Status: filter.Status,
		Search: filter.Search,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
	return rows, total, err
}

func (uc *InvoiceUseCase) ListReconciliation(ctx context.Context, tenantID uuid.UUID, filter InvoiceListFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, 0, err
	}
	if err := authz.Authorize(ctx, domain.PermissionFinancialView); err != nil {
		return nil, 0, ErrForbidden
	}
	return uc.Invoices.ListTenantInvoices(ctx, tenantID, port.ListAuctionInvoicesFilter{
		Status: filter.Status,
		Search: filter.Search,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (uc *InvoiceUseCase) ListWinnerSettlements(ctx context.Context, tenantID, userID uuid.UUID, filter InvoiceListFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, 0, err
	}
	return uc.Invoices.ListWinnerInvoices(ctx, tenantID, userID, port.ListAuctionInvoicesFilter{
		Status: filter.Status,
		Search: filter.Search,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (uc *InvoiceUseCase) StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter InvoiceStreamFilter, fn func(domain.SettlementInvoiceRow) error) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceMarkPaid); err != nil {
		return ErrForbidden
	}
	return uc.Invoices.StreamAuctionInvoices(ctx, tenantID, auctionID, port.StreamAuctionInvoicesFilter{
		Status: filter.Status,
		Search: filter.Search,
	}, fn)
}

func (uc *InvoiceUseCase) AuditCSVExport(ctx context.Context, tenantID, auctionID uuid.UUID) {
	if uc.Audit == nil {
		return
	}
	auditCtx := middleware.AuditContextFromContext(ctx)
	payload, _ := json.Marshal(map[string]any{"auction_id": auctionID})
	_ = uc.Audit.Append(ctx, port.AuditEntry{
		TenantID:    &tenantID,
		ActorUserID: auditCtx.ActorUserID,
		ActorRole:   auditCtx.ActorRole,
		Action:      "SETTLEMENT_CSV_EXPORTED",
		EntityType:  "AUCTION",
		EntityID:    &auctionID,
		RequestID:   auditCtx.RequestID,
		IP:          auditCtx.IP,
		UserAgent:   auditCtx.UserAgent,
		Metadata:    payload,
	})
}

func (uc *InvoiceUseCase) MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceMarkPaid); err != nil {
		return nil, ErrForbidden
	}
	if paymentReference != nil && len(strings.TrimSpace(*paymentReference)) > 128 {
		return nil, ErrValidation
	}
	return uc.Invoices.MarkPaid(ctx, tenantID, invoiceID, paymentReference, now)
}

func (uc *InvoiceUseCase) ApproveSettlement(ctx context.Context, tenantID, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionSettlementApprove); err != nil {
		return nil, ErrForbidden
	}
	return uc.Invoices.ApproveSettlement(ctx, tenantID, invoiceID, now)
}

func (uc *InvoiceUseCase) SendInvoice(ctx context.Context, tenant domain.Tenant, invoiceID uuid.UUID, successURL, cancelURL string, now time.Time) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenant.ID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceSend); err != nil {
		return nil, ErrForbidden
	}
	if uc.Payments == nil {
		return nil, ErrValidation
	}
	inv, err := uc.Invoices.GetByID(ctx, tenant.ID, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.SettlementStatus != domain.InvoiceSettlementApproved && inv.SettlementStatus != domain.InvoiceSettlementInvoiceSent {
		return nil, ErrValidation
	}
	session, err := uc.Payments.CreateCheckoutSession(ctx, port.CreateCheckoutSessionInput{
		InvoiceID:     inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		TenantID:      tenant.ID,
		AmountMinor:   inv.TotalDueMinor,
		Currency:      inv.Currency,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		CustomerEmail: inv.WinnerEmail,
		Description:   strings.TrimSpace(tenant.Name) + " invoice " + inv.InvoiceNumber,
	})
	if err != nil {
		return nil, err
	}
	inv, err = uc.Invoices.RecordInvoiceDispatch(ctx, tenant.ID, invoiceID, session, now)
	if err != nil {
		return nil, err
	}
	if uc.EmailOutbox != nil && strings.TrimSpace(inv.WinnerEmail) != "" {
		_, err = uc.EmailOutbox.Queue(ctx, QueueEmailInput{
			TenantID:   tenant.ID,
			UserID:     &inv.WinnerUserID,
			ToEmail:    inv.WinnerEmail,
			Subject:    fmt.Sprintf("%s invoice %s", strings.TrimSpace(tenant.Name), inv.InvoiceNumber),
			HTMLBody:   buildInvoiceEmailHTML(strings.TrimSpace(tenant.Name), inv, session.URL),
			TextBody:   buildInvoiceEmailText(strings.TrimSpace(tenant.Name), inv, session.URL),
			Type:       "invoice_payment_request",
			EntityType: "INVOICE",
			EntityID:   &inv.ID,
			Critical:   true,
		})
		if err != nil {
			return nil, err
		}
	}
	return inv, nil
}

func (uc *InvoiceUseCase) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionInvoiceMarkPaid); err != nil {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(reason) == "" {
		return nil, ErrValidation
	}
	return uc.Invoices.Cancel(ctx, tenantID, invoiceID, reason, now)
}

func (uc *InvoiceUseCase) HandlePaymentWebhook(ctx context.Context, payload []byte, signatureHeader string, now time.Time) (*domain.Invoice, bool, error) {
	if uc.Payments == nil {
		return nil, false, ErrValidation
	}
	event, err := uc.Payments.ParseWebhook(payload, signatureHeader)
	if err != nil {
		return nil, false, err
	}
	if event == nil {
		return nil, false, nil
	}
	return uc.Invoices.MarkPaidByWebhook(ctx, *event, now)
}

func (uc *InvoiceUseCase) GetAuthorized(ctx context.Context, tenantID, invoiceID, userID uuid.UUID) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	inv, err := uc.Invoices.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	if authz.HasPermission(ctx, userID, tenantID, domain.PermissionInvoiceView) || inv.WinnerUserID == userID {
		return inv, nil
	}
	return nil, ErrForbidden
}

func (uc *InvoiceUseCase) GetByLotAuthorized(ctx context.Context, tenantID, lotID, userID uuid.UUID) (*domain.Invoice, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	inv, err := uc.Invoices.GetByLotID(ctx, tenantID, lotID)
	if err != nil {
		return nil, err
	}
	if authz.HasPermission(ctx, userID, tenantID, domain.PermissionInvoiceView) || inv.WinnerUserID == userID {
		return inv, nil
	}
	return nil, ErrForbidden
}

func (uc *InvoiceUseCase) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	switch status {
	case domain.InvoicePaid:
		_, err := uc.MarkPaid(ctx, tenantID, invoiceID, nil, time.Now().UTC())
		return err
	case domain.InvoiceCancelled:
		_, err := uc.Cancel(ctx, tenantID, invoiceID, "Cancelled from legacy endpoint", time.Now().UTC())
		return err
	case domain.InvoiceIssued:
		return ErrValidation
	default:
		return ErrValidation
	}
}

func ParseInvoiceSnapshot(raw json.RawMessage) (domain.LotFinancialSnapshot, error) {
	var snap domain.LotFinancialSnapshot
	if len(raw) == 0 {
		return snap, nil
	}
	err := json.Unmarshal(raw, &snap)
	return snap, err
}

func buildInvoiceEmailHTML(tenantName string, inv *domain.Invoice, checkoutURL string) string {
	return fmt.Sprintf(
		"<p>%s has issued invoice <strong>%s</strong> for %s %s.</p><p><a href=\"%s\">Complete payment</a></p>",
		htmlEscape(tenantName),
		htmlEscape(inv.InvoiceNumber),
		htmlEscape(strings.ToUpper(inv.Currency)),
		formatMinorMoney(inv.TotalDueMinor),
		htmlEscape(checkoutURL),
	)
}

func buildInvoiceEmailText(tenantName string, inv *domain.Invoice, checkoutURL string) string {
	return fmt.Sprintf("%s issued invoice %s for %s %s.\nPay here: %s", tenantName, inv.InvoiceNumber, strings.ToUpper(inv.Currency), formatMinorMoney(inv.TotalDueMinor), checkoutURL)
}

func formatMinorMoney(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s%d.%02d", sign, amount/100, amount%100)
}

func htmlEscape(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}
