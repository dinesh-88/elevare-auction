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
		&inv.TotalDueMinor, &inv.PaymentStatus, &inv.PaidAt, &inv.CancelledAt,
		&inv.PaymentRef, &inv.CancelReason,
		&inv.Snapshot, &inv.Status, &inv.CreatedAt,
		&inv.LotNumber, &inv.LotTitle, &inv.AuctionTitle,
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
       i.total_due_minor, i.payment_status, i.paid_at, i.cancelled_at,
       i.payment_reference, i.cancel_reason,
       i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = i.auction_id
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
       i.total_due_minor, i.payment_status, i.paid_at, i.cancelled_at,
       i.payment_reference, i.cancel_reason,
       i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = i.auction_id
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
	switch status {
	case domain.InvoicePaid:
		paymentStatus = domain.InvoicePaymentPaid
	case domain.InvoiceCancelled:
		paymentStatus = domain.InvoicePaymentCancelled
	case domain.InvoiceIssued:
		paymentStatus = domain.InvoicePaymentUnpaid
	default:
		return errors.New("invalid invoice status")
	}
	const q = `
UPDATE invoices
SET status = $3,
    payment_status = $4,
    paid_at = CASE WHEN $4 = 'PAID' THEN NOW() ELSE NULL END,
    cancelled_at = CASE WHEN $4 = 'CANCELLED' THEN NOW() ELSE NULL END
WHERE tenant_id = $1
  AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, invoiceID, status, paymentStatus)
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
  payment_status, snapshot, status, created_at
) VALUES (
  $1,$2,$3,$4,$5,
  $6,$7,$8,$9,
  $10,$11,$12,$13,$14,
  'UNPAID',$15,'ISSUED',now()
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
  COUNT(*) FILTER (WHERE payment_status = 'PAID')::bigint AS paid_count,
  COUNT(*) FILTER (WHERE payment_status = 'UNPAID')::bigint AS unpaid_count,
  COUNT(*) FILTER (WHERE payment_status = 'CANCELLED')::bigint AS cancelled_count,
  COALESCE(SUM(hammer_price_minor) FILTER (WHERE payment_status <> 'CANCELLED'), 0)::bigint AS total_hammer_minor,
  COALESCE(SUM(fees_total_minor) FILTER (WHERE payment_status <> 'CANCELLED'), 0)::bigint AS total_fees_minor,
  COALESCE(SUM(tax_total_minor) FILTER (WHERE payment_status <> 'CANCELLED'), 0)::bigint AS total_tax_minor,
  COALESCE(SUM(total_due_minor) FILTER (WHERE payment_status <> 'CANCELLED'), 0)::bigint AS total_due_minor,
  COALESCE(SUM(total_due_minor) FILTER (WHERE payment_status = 'PAID'), 0)::bigint AS total_paid_minor,
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
SELECT i.id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.payment_status,
       i.issued_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason
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
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
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
	if err := tx.QueryRow(ctx, `SELECT payment_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if err := canMarkPaid(current); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET payment_status = 'PAID',
    status = 'PAID',
    paid_at = $3,
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
	if err := tx.QueryRow(ctx, `SELECT payment_status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE`, tenantID, invoiceID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if err := canCancel(current); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE invoices
SET payment_status = 'CANCELLED',
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
SELECT i.id, i.invoice_number,
       split_part(u.email, '@', 1) AS bidder_name,
       u.email AS bidder_email,
       l.title AS lot_summary,
       i.hammer_price_minor,
       i.fees_total_minor,
       i.tax_total_minor,
       i.total_due_minor,
       i.currency,
       i.payment_status,
       i.issued_at,
       i.paid_at,
       i.payment_reference,
       i.cancel_reason
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
			&row.InvoiceNumber,
			&row.BidderName,
			&row.BidderEmail,
			&row.LotSummary,
			&row.HammerMinor,
			&row.FeesMinor,
			&row.TaxMinor,
			&row.TotalDueMinor,
			&row.Currency,
			&row.PaymentStatus,
			&row.IssuedAt,
			&row.PaidAt,
			&row.PaymentReference,
			&row.CancelReason,
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
			Name    string `json:"name"`
			Amount  int64  `json:"amount"`
			Taxable bool   `json:"taxable"`
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
		feesMinor += fee.Amount
		name := strings.ToUpper(strings.TrimSpace(fee.Name))
		if fee.Taxable || strings.Contains(name, "VAT") || strings.Contains(name, "TAX") || strings.Contains(name, "GST") {
			taxMinor += fee.Amount
		}
	}
	return
}

func canMarkPaid(current domain.InvoicePaymentStatus) error {
	if current == domain.InvoicePaymentCancelled {
		return errors.New("cannot mark paid cancelled invoice")
	}
	return nil
}

func canCancel(current domain.InvoicePaymentStatus) error {
	if current == domain.InvoicePaymentPaid {
		return errors.New("cannot cancel paid invoice")
	}
	return nil
}
