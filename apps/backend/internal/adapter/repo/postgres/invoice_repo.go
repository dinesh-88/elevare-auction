package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepo struct {
	DB    *pgxpool.Pool
	Audit port.AuditRepository
}

func scanInvoice(row pgx.Row) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := row.Scan(
		&inv.ID, &inv.LotID, &inv.AuctionID, &inv.TenantID, &inv.WinnerUserID,
		&inv.InvoiceNumber, &inv.IssuedAt, &inv.DueDate, &inv.Currency,
		&inv.TotalAmount, &inv.HammerMinor, &inv.FeesMinor, &inv.TaxMinor,
		&inv.TotalDueMinor, &inv.SettlementStatus, &inv.PaymentStatus,
		&inv.SettlementApprovedAt, &inv.InvoiceSentAt, &inv.PaidAt, &inv.PaymentReceivedAt, &inv.CancelledAt,
		&inv.PaymentRef, &inv.CancelReason, &inv.StripeCheckoutSessionID, &inv.StripeCheckoutURL, &inv.StripePaymentIntentID,
		&inv.Snapshot, &inv.Status, &inv.CreatedAt,
		&inv.LotNumber, &inv.LotTitle, &inv.AuctionTitle, &inv.WinnerEmail,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func invoiceSelectByID() string {
	return `
SELECT i.id, i.lot_id, i.auction_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.hammer_price_minor, i.fees_total_minor, i.tax_total_minor,
       i.total_due_minor, i.settlement_status, i.payment_status,
       i.settlement_approved_at, i.invoice_sent_at, i.paid_at, i.payment_received_at, i.cancelled_at,
       i.payment_reference, i.cancel_reason, i.stripe_checkout_session_id, i.stripe_checkout_url, i.stripe_payment_intent_id,
       i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title, u.email
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = i.auction_id
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
WHERE i.tenant_id = $1
  AND i.id = $2`
}

func (r *InvoiceRepo) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	inv, err := scanInvoice(r.DB.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invoice not found")
	}
	return inv, err
}

func (r *InvoiceRepo) GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error) {
	const q = `
SELECT i.id, i.lot_id, i.auction_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.hammer_price_minor, i.fees_total_minor, i.tax_total_minor,
       i.total_due_minor, i.settlement_status, i.payment_status,
       i.settlement_approved_at, i.invoice_sent_at, i.paid_at, i.payment_received_at, i.cancelled_at,
       i.payment_reference, i.cancel_reason, i.stripe_checkout_session_id, i.stripe_checkout_url, i.stripe_payment_intent_id,
       i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title, u.email
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = i.auction_id
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
WHERE i.tenant_id = $1
  AND i.lot_id = $2`
	inv, err := scanInvoice(r.DB.QueryRow(ctx, q, tenantID, lotID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invoice not found")
	}
	return inv, err
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	paymentStatus := domain.InvoicePaymentUnpaid
	settlementStatus := domain.InvoiceSettlementDraft
	switch status {
	case domain.InvoicePaid:
		paymentStatus = domain.InvoicePaymentPaid
		settlementStatus = domain.InvoiceSettlementPaid
	case domain.InvoiceCancelled:
		paymentStatus = domain.InvoicePaymentRefunded
		settlementStatus = domain.InvoiceSettlementCancelled
	case domain.InvoiceIssued:
		paymentStatus = domain.InvoicePaymentUnpaid
		settlementStatus = domain.InvoiceSettlementDraft
	default:
		return errors.New("invalid invoice status")
	}
	const q = `
UPDATE invoices
SET status = $3,
    settlement_status = $4,
    payment_status = $5,
    paid_at = CASE WHEN $5 = 'paid' THEN NOW() ELSE NULL END,
    payment_received_at = CASE WHEN $5 = 'paid' THEN NOW() ELSE NULL END,
    cancelled_at = CASE WHEN $4 = 'cancelled' THEN NOW() ELSE NULL END
WHERE tenant_id = $1
  AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, invoiceID, status, settlementStatus, paymentStatus)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("invoice not found")
	}
	return nil
}

func (r *InvoiceRepo) GenerateForLot(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var existingID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM invoices WHERE lot_id = $1 AND tenant_id = $2`, lotID, tenantID).Scan(&existingID); err == nil {
		inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, existingID))
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return inv, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var status domain.LotStatus
	var winnerUserID *uuid.UUID
	var totalAmount *int64
	var snapshot []byte
	var currency string
	var auctionID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT l.status, l.winner_user_id, l.total_due, l.financial_snapshot, a.currency, a.id
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1
  AND a.tenant_id = $2
FOR UPDATE`, lotID, tenantID).Scan(&status, &winnerUserID, &totalAmount, &snapshot, &currency, &auctionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("lot not found")
		}
		return nil, err
	}
	if status != domain.LotEnded {
		return nil, errors.New("invoice can be generated only for ended lots")
	}
	if winnerUserID == nil || totalAmount == nil || len(snapshot) == 0 {
		return nil, errors.New("lot has no winner settlement snapshot")
	}
	if currency == "" {
		currency = "USD"
	}
	snapshotWithCurrency := append([]byte(nil), snapshot...)
	var snapMap map[string]any
	if err := json.Unmarshal(snapshotWithCurrency, &snapMap); err == nil {
		snapMap["currency"] = currency
		if b, err := json.Marshal(snapMap); err == nil {
			snapshotWithCurrency = b
		}
	}
	hammerMinor, feesMinor, taxMinor, totalDueMinor := snapshotTotals(snapshotWithCurrency, *totalAmount)

	var prefix string
	var counter int64
	if err := tx.QueryRow(ctx, `
UPDATE tenants
SET invoice_counter = invoice_counter + 1
WHERE id = $1
RETURNING invoice_prefix, invoice_counter`, tenantID).Scan(&prefix, &counter); err != nil {
		return nil, err
	}
	if prefix == "" {
		prefix = "INV"
	}
	invoiceNumber := fmt.Sprintf("%s-%04d-%06d", prefix, now.Year(), counter)
	dueDate := now.AddDate(0, 0, 7)
	invoiceID := uuid.New()

	if _, err := tx.Exec(ctx, `
INSERT INTO invoices (
  id, lot_id, auction_id, tenant_id, winner_user_id,
  invoice_number, issued_at, due_date, currency,
  total_amount, hammer_price_minor, fees_total_minor, tax_total_minor, total_due_minor,
  settlement_status, payment_status, snapshot, status, created_at
) VALUES (
  $1,$2,$3,$4,$5,
  $6,$7,$8,$9,
  $10,$11,$12,$13,$14,
  'draft','unpaid',$15,'ISSUED',now()
)`, invoiceID, lotID, auctionID, tenantID, *winnerUserID, invoiceNumber, now, dueDate, currency, totalDueMinor, hammerMinor, feesMinor, taxMinor, totalDueMinor, snapshotWithCurrency); err != nil {
		return nil, err
	}
	if r.Audit != nil {
		base := middleware.AuditContextFromContext(ctx)
		payload, _ := json.Marshal(map[string]any{
			"invoice_id":         invoiceID,
			"invoice_number":     invoiceNumber,
			"total_amount_minor": totalDueMinor,
			"currency":           currency,
			"reason":             "MANUAL_GENERATE",
		})
		if err := r.Audit.AppendTx(ctx, tx, port.AuditEntry{
			TenantID:    &tenantID,
			ActorUserID: base.ActorUserID,
			ActorRole:   base.ActorRole,
			Action:      "INVOICE_ISSUED",
			EntityType:  "INVOICE",
			EntityID:    &invoiceID,
			RequestID:   base.RequestID,
			IP:          base.IP,
			UserAgent:   base.UserAgent,
			Metadata:    payload,
		}); err != nil {
			return nil, err
		}
	}

	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) GetAuctionSettlementSummary(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.AuctionSettlementSummary, bool, bool, error) {
	const q = `
SELECT
  COUNT(*)::bigint AS invoice_count,
  COUNT(*) FILTER (WHERE payment_status = 'paid')::bigint AS paid_count,
  COUNT(*) FILTER (WHERE payment_status IN ('unpaid', 'payment_pending'))::bigint AS unpaid_count,
  COUNT(*) FILTER (WHERE settlement_status = 'cancelled')::bigint AS cancelled_count,
  COALESCE(SUM(hammer_price_minor) FILTER (WHERE settlement_status <> 'cancelled'), 0)::bigint AS total_hammer_minor,
  COALESCE(SUM(fees_total_minor) FILTER (WHERE settlement_status <> 'cancelled'), 0)::bigint AS total_fees_minor,
  COALESCE(SUM(tax_total_minor) FILTER (WHERE settlement_status <> 'cancelled'), 0)::bigint AS total_tax_minor,
  COALESCE(SUM(total_due_minor) FILTER (WHERE settlement_status <> 'cancelled'), 0)::bigint AS total_due_minor,
  COALESCE(SUM(total_due_minor) FILTER (WHERE payment_status = 'paid'), 0)::bigint AS total_paid_minor,
  MIN(currency) AS min_currency,
  MAX(currency) AS max_currency
FROM invoices
WHERE tenant_id = $1
  AND auction_id = $2`

	var s domain.AuctionSettlementSummary
	var minCurrency, maxCurrency *string
	err := r.DB.QueryRow(ctx, q, tenantID, auctionID).Scan(
		&s.InvoiceCount,
		&s.PaidCount,
		&s.UnpaidCount,
		&s.CancelledCount,
		&s.TotalHammerMinor,
		&s.TotalFeesMinor,
		&s.TotalTaxMinor,
		&s.TotalDueMinor,
		&s.TotalPaidMinor,
		&minCurrency,
		&maxCurrency,
	)
	if err != nil {
		return nil, false, false, err
	}
	s.AuctionID = auctionID
	s.OutstandingMinor = s.TotalDueMinor - s.TotalPaidMinor
	if minCurrency != nil {
		s.Currency = *minCurrency
	}
	mixedCurrency := false
	if minCurrency != nil && maxCurrency != nil && *minCurrency != *maxCurrency {
		mixedCurrency = true
	}
	return &s, s.InvoiceCount > 0, mixedCurrency, nil
}

func (r *InvoiceRepo) ListAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	search := strings.TrimSpace(filter.Search)

	where := "WHERE i.tenant_id = $1 AND i.auction_id = $2"
	args := []any{tenantID, auctionID}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where += fmt.Sprintf(" AND i.payment_status = $%d", len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR split_part(u.email, '@', 1) ILIKE $%d)", len(args), len(args))
	}

	countSQL := `SELECT COUNT(*)::bigint FROM invoices i JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id ` + where
	var total int64
	if err := r.DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listSQL := `
SELECT i.id, i.lot_id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.settlement_status,
       i.payment_status,
       i.issued_at,
       i.invoice_sent_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason,
       i.stripe_checkout_url
FROM invoices i
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
JOIN lots l ON l.id = i.lot_id
` + where + fmt.Sprintf(`
ORDER BY i.issued_at DESC, i.id DESC
LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.SettlementInvoiceRow, 0, limit)
	for rows.Next() {
		var row domain.SettlementInvoiceRow
		if err := rows.Scan(
			&row.ID,
			&row.LotID,
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.SettlementStatus,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.InvoiceSentAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
			&row.CheckoutURL,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

func (r *InvoiceRepo) ListTenantInvoices(ctx context.Context, tenantID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	search := strings.TrimSpace(filter.Search)

	where := "WHERE i.tenant_id = $1"
	args := []any{tenantID}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where += fmt.Sprintf(" AND i.payment_status = $%d", len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR i.invoice_number ILIKE $%d OR l.title ILIKE $%d)", len(args), len(args), len(args))
	}

	countSQL := `SELECT COUNT(*)::bigint FROM invoices i JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id JOIN lots l ON l.id = i.lot_id ` + where
	var total int64
	if err := r.DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listSQL := `
SELECT i.id, i.lot_id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.settlement_status,
       i.payment_status,
       i.issued_at,
       i.invoice_sent_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason,
       i.stripe_checkout_url
FROM invoices i
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
JOIN lots l ON l.id = i.lot_id
` + where + fmt.Sprintf(`
ORDER BY i.issued_at DESC, i.id DESC
LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.SettlementInvoiceRow, 0, limit)
	for rows.Next() {
		var row domain.SettlementInvoiceRow
		if err := rows.Scan(
			&row.ID,
			&row.LotID,
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.SettlementStatus,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.InvoiceSentAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
			&row.CheckoutURL,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

func (r *InvoiceRepo) ListWinnerInvoices(ctx context.Context, tenantID, userID uuid.UUID, filter port.ListAuctionInvoicesFilter) ([]domain.SettlementInvoiceRow, int64, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	search := strings.TrimSpace(filter.Search)

	where := "WHERE i.tenant_id = $1 AND i.winner_user_id = $2"
	args := []any{tenantID, userID}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where += fmt.Sprintf(" AND i.payment_status = $%d", len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (i.invoice_number ILIKE $%d OR l.title ILIKE $%d)", len(args), len(args))
	}

	countSQL := `SELECT COUNT(*)::bigint FROM invoices i JOIN lots l ON l.id = i.lot_id ` + where
	var total int64
	if err := r.DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listSQL := `
SELECT i.id, i.lot_id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.settlement_status,
       i.payment_status,
       i.issued_at,
       i.invoice_sent_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason,
       i.stripe_checkout_url
FROM invoices i
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
JOIN lots l ON l.id = i.lot_id
` + where + fmt.Sprintf(`
ORDER BY COALESCE(i.paid_at, i.issued_at) DESC, i.id DESC
LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.SettlementInvoiceRow, 0, limit)
	for rows.Next() {
		var row domain.SettlementInvoiceRow
		if err := rows.Scan(
			&row.ID,
			&row.LotID,
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.SettlementStatus,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.InvoiceSentAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
			&row.CheckoutURL,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

func (r *InvoiceRepo) MarkPaid(ctx context.Context, tenantID, invoiceID uuid.UUID, paymentReference *string, now time.Time) (*domain.Invoice, error) {
	if paymentReference != nil {
		trimmed := strings.TrimSpace(*paymentReference)
		paymentReference = &trimmed
	}

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var current domain.InvoicePaymentStatus
	var settlementStatus domain.InvoiceSettlementStatus
	if err := tx.QueryRow(ctx, `SELECT payment_status, settlement_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current, &settlementStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if err := canMarkPaid(current, settlementStatus); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET settlement_status = 'paid',
    payment_status = 'paid',
    status = 'PAID',
    paid_at = $3,
    payment_received_at = $3,
    payment_reference = $4,
    cancelled_at = NULL,
    cancel_reason = NULL
WHERE tenant_id = $1
  AND id = $2`, tenantID, invoiceID, now, paymentReference); err != nil {
		return nil, err
	}

	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if err != nil {
		return nil, err
	}

	if err := r.appendAuditTx(ctx, tx, tenantID, invoiceID, "INVOICE_MARKED_PAID", map[string]any{
		"payment_reference": paymentReference,
		"paid_at":           now,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) Cancel(ctx context.Context, tenantID, invoiceID uuid.UUID, reason string, now time.Time) (*domain.Invoice, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("cancel reason required")
	}

	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var current domain.InvoicePaymentStatus
	var settlementStatus domain.InvoiceSettlementStatus
	if err := tx.QueryRow(ctx, `SELECT payment_status, settlement_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current, &settlementStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if err := canCancel(current, settlementStatus); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET settlement_status = 'cancelled',
    payment_status = 'refunded',
    status = 'CANCELLED',
    cancelled_at = $3,
    cancel_reason = $4
WHERE tenant_id = $1
  AND id = $2`, tenantID, invoiceID, now, reason); err != nil {
		return nil, err
	}

	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if err != nil {
		return nil, err
	}
	if err := r.appendAuditTx(ctx, tx, tenantID, invoiceID, "INVOICE_CANCELLED", map[string]any{
		"cancel_reason": reason,
		"cancelled_at":  now,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) ApproveSettlement(ctx context.Context, tenantID, invoiceID uuid.UUID, now time.Time) (*domain.Invoice, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var current domain.InvoiceSettlementStatus
	if err := tx.QueryRow(ctx, `SELECT settlement_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if current != domain.InvoiceSettlementDraft {
		return nil, errors.New("settlement is already locked")
	}
	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET settlement_status = 'approved',
    settlement_approved_at = $3
WHERE tenant_id = $1
  AND id = $2`, tenantID, invoiceID, now); err != nil {
		return nil, err
	}
	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if err != nil {
		return nil, err
	}
	if err := r.appendAuditTx(ctx, tx, tenantID, invoiceID, "SETTLEMENT_APPROVED", map[string]any{"approved_at": now}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) RecordInvoiceDispatch(ctx context.Context, tenantID, invoiceID uuid.UUID, session port.CheckoutSession, now time.Time) (*domain.Invoice, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var current domain.InvoiceSettlementStatus
	if err := tx.QueryRow(ctx, `SELECT settlement_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if current != domain.InvoiceSettlementApproved && current != domain.InvoiceSettlementInvoiceSent {
		return nil, errors.New("settlement must be approved before sending an invoice")
	}
	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET settlement_status = 'invoice_sent',
    payment_status = 'payment_pending',
    invoice_sent_at = $3,
    stripe_checkout_session_id = $4,
    stripe_checkout_url = $5,
    stripe_payment_intent_id = NULLIF($6, '')
WHERE tenant_id = $1
  AND id = $2`, tenantID, invoiceID, now, session.ID, session.URL, session.PaymentIntentID); err != nil {
		return nil, err
	}
	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), tenantID, invoiceID))
	if err != nil {
		return nil, err
	}
	if err := r.appendAuditTx(ctx, tx, tenantID, invoiceID, "INVOICE_SENT_MANUAL", map[string]any{
		"checkout_session_id": session.ID,
		"checkout_url":        session.URL,
		"sent_at":             now,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InvoiceRepo) MarkPaidByWebhook(ctx context.Context, event port.PaymentWebhookEvent, now time.Time) (*domain.Invoice, bool, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	var alreadyProcessed bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM payment_webhook_events
  WHERE provider = 'stripe'
    AND event_id = $1
)`, event.EventID).Scan(&alreadyProcessed); err != nil {
		return nil, false, err
	}
	if alreadyProcessed {
		inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), event.TenantID, event.InvoiceID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, true, nil
			}
			return nil, true, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, true, err
		}
		return inv, true, nil
	}

	webhookLogID := uuid.New()
	if _, err := tx.Exec(ctx, `
INSERT INTO payment_webhook_events (
  id, provider, event_id, event_type, tenant_id, invoice_id, payload, processing_status, received_at
) VALUES (
  $1, 'stripe', $2, $3, $4, $5, $6::jsonb, 'received', $7
)`, webhookLogID, event.EventID, event.EventType, event.TenantID, event.InvoiceID, string(event.Payload), now); err != nil {
		return nil, false, err
	}

	var expectedAmount int64
	var expectedCurrency string
	if err := tx.QueryRow(ctx, `
SELECT total_due_minor, currency
FROM invoices
WHERE tenant_id = $1
  AND id = $2
FOR UPDATE`, event.TenantID, event.InvoiceID).Scan(&expectedAmount, &expectedCurrency); err != nil {
		_, _ = tx.Exec(ctx, `UPDATE payment_webhook_events SET processing_status = 'failed', error_message = $2 WHERE id = $1`, webhookLogID, "invoice not found")
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, errors.New("invoice not found")
		}
		return nil, false, err
	}
	if event.AmountMinor != expectedAmount || !strings.EqualFold(strings.TrimSpace(event.Currency), strings.TrimSpace(expectedCurrency)) {
		_, _ = tx.Exec(ctx, `UPDATE payment_webhook_events SET processing_status = 'failed', error_message = $2 WHERE id = $1`, webhookLogID, "payment amount mismatch")
		return nil, false, errors.New("payment amount mismatch")
	}

	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET settlement_status = 'paid',
    payment_status = 'paid',
    status = 'PAID',
    paid_at = $3,
    payment_received_at = $3,
    stripe_checkout_session_id = NULLIF($4, ''),
    stripe_payment_intent_id = NULLIF($5, '')
WHERE tenant_id = $1
  AND id = $2`, event.TenantID, event.InvoiceID, now, event.CheckoutSessionID, event.PaymentIntentID); err != nil {
		_, _ = tx.Exec(ctx, `UPDATE payment_webhook_events SET processing_status = 'failed', error_message = $2 WHERE id = $1`, webhookLogID, "invoice update failed")
		return nil, false, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE payment_webhook_events
SET processing_status = 'processed',
    processed_at = $2,
    error_message = NULL
WHERE id = $1`, webhookLogID, now); err != nil {
		return nil, false, err
	}
	inv, err := scanInvoice(tx.QueryRow(ctx, invoiceSelectByID(), event.TenantID, event.InvoiceID))
	if err != nil {
		return nil, false, err
	}
	if err := r.appendAuditTx(ctx, tx, event.TenantID, event.InvoiceID, "PAYMENT_WEBHOOK_PAID", map[string]any{
		"provider":            "stripe",
		"event_id":            event.EventID,
		"checkout_session_id": event.CheckoutSessionID,
		"payment_intent_id":   event.PaymentIntentID,
		"paid_at":             now,
	}); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return inv, false, nil
}

func (r *InvoiceRepo) StreamAuctionInvoices(ctx context.Context, tenantID, auctionID uuid.UUID, filter port.StreamAuctionInvoicesFilter, fn func(domain.SettlementInvoiceRow) error) error {
	search := strings.TrimSpace(filter.Search)
	where := "WHERE i.tenant_id = $1 AND i.auction_id = $2"
	args := []any{tenantID, auctionID}
	if filter.Status != nil {
		args = append(args, *filter.Status)
		where += fmt.Sprintf(" AND i.payment_status = $%d", len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (u.email ILIKE $%d OR split_part(u.email, '@', 1) ILIKE $%d)", len(args), len(args))
	}

	q := `
SELECT i.id, i.lot_id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.settlement_status,
       i.payment_status,
       i.issued_at,
       i.invoice_sent_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason,
       i.stripe_checkout_url
FROM invoices i
JOIN users u ON u.id = i.winner_user_id AND u.tenant_id = i.tenant_id
JOIN lots l ON l.id = i.lot_id
` + where + `
ORDER BY i.issued_at DESC, i.id DESC`

	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row domain.SettlementInvoiceRow
		if err := rows.Scan(
			&row.ID,
			&row.LotID,
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.SettlementStatus,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.InvoiceSentAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
			&row.CheckoutURL,
		); err != nil {
			return err
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *InvoiceRepo) appendAuditTx(ctx context.Context, tx pgx.Tx, tenantID, invoiceID uuid.UUID, action string, metadata map[string]any) error {
	if r.Audit == nil {
		return nil
	}
	base := middleware.AuditContextFromContext(ctx)
	payload, _ := json.Marshal(metadata)
	return r.Audit.AppendTx(ctx, tx, port.AuditEntry{
		TenantID:    &tenantID,
		ActorUserID: base.ActorUserID,
		ActorRole:   base.ActorRole,
		Action:      action,
		EntityType:  "INVOICE",
		EntityID:    &invoiceID,
		RequestID:   base.RequestID,
		IP:          base.IP,
		UserAgent:   base.UserAgent,
		Metadata:    payload,
	})
}

func snapshotTotals(snapshot []byte, fallbackTotal int64) (hammerMinor, feesMinor, taxMinor, totalDueMinor int64) {
	hammerMinor = 0
	feesMinor = 0
	taxMinor = 0
	totalDueMinor = fallbackTotal
	var payload struct {
		HammerPrice int64 `json:"hammer_price"`
		TotalDue    int64 `json:"total_due"`
		Fees        []struct {
			Name     string `json:"name"`
			Amount   int64  `json:"amount"`
			Taxable  bool   `json:"taxable"`
			Category string `json:"category"`
		} `json:"fees"`
	}
	if err := json.Unmarshal(snapshot, &payload); err != nil {
		return
	}
	hammerMinor = payload.HammerPrice
	if payload.TotalDue > 0 {
		totalDueMinor = payload.TotalDue
	}
	for _, fee := range payload.Fees {
		category := strings.ToUpper(strings.TrimSpace(fee.Category))
		switch {
		case category == "TAX":
			taxMinor += fee.Amount
		case category == "FEE":
			feesMinor += fee.Amount
		case isLegacyTaxLineName(fee.Name):
			taxMinor += fee.Amount
		default:
			feesMinor += fee.Amount
		}
	}
	return
}

func isLegacyTaxLineName(name string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	return strings.Contains(normalized, "VAT") || strings.Contains(normalized, "TAX") || strings.Contains(normalized, "GST")
}

func canMarkPaid(current domain.InvoicePaymentStatus, settlementStatus domain.InvoiceSettlementStatus) error {
	if settlementStatus == domain.InvoiceSettlementDraft {
		return errors.New("settlement must be approved before payment can be recorded")
	}
	if settlementStatus == domain.InvoiceSettlementCancelled {
		return errors.New("cannot mark paid cancelled invoice")
	}
	if current == domain.InvoicePaymentRefunded || current == domain.InvoicePaymentPartiallyRefunded {
		return errors.New("cannot mark paid refunded invoice")
	}
	return nil
}

func canCancel(current domain.InvoicePaymentStatus, settlementStatus domain.InvoiceSettlementStatus) error {
	if current == domain.InvoicePaymentPaid {
		return errors.New("cannot cancel paid invoice")
	}
	if settlementStatus == domain.InvoiceSettlementPaid {
		return errors.New("cannot cancel paid settlement")
	}
	return nil
}
