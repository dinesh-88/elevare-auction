package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepo struct {
	DB *pgxpool.Pool
}

func scanInvoice(row pgx.Row) (*domain.Invoice, error) {
	var inv domain.Invoice
	err := row.Scan(
		&inv.ID, &inv.LotID, &inv.TenantID, &inv.WinnerUserID,
		&inv.InvoiceNumber, &inv.IssuedAt, &inv.DueDate, &inv.Currency,
		&inv.TotalAmount, &inv.Snapshot, &inv.Status, &inv.CreatedAt,
		&inv.LotNumber, &inv.LotTitle, &inv.AuctionTitle,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	const q = `
SELECT i.id, i.lot_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = l.auction_id
WHERE i.tenant_id = $1
  AND i.id = $2`
	inv, err := scanInvoice(r.DB.QueryRow(ctx, q, tenantID, invoiceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invoice not found")
	}
	return inv, err
}

func (r *InvoiceRepo) GetByLotID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Invoice, error) {
	const q = `
SELECT i.id, i.lot_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = l.auction_id
WHERE i.tenant_id = $1
  AND i.lot_id = $2`
	inv, err := scanInvoice(r.DB.QueryRow(ctx, q, tenantID, lotID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("invoice not found")
	}
	return inv, err
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, tenantID, invoiceID uuid.UUID, status domain.InvoiceStatus) error {
	const q = `UPDATE invoices SET status = $3 WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, invoiceID, status)
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
	if err := tx.QueryRow(ctx, `SELECT id FROM invoices WHERE lot_id = $1`, lotID).Scan(&existingID); err == nil {
		inv, err := scanInvoice(tx.QueryRow(ctx, `
SELECT i.id, i.lot_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = l.auction_id
WHERE i.id = $1`, existingID))
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
	if err := tx.QueryRow(ctx, `
SELECT l.status, l.winner_user_id, l.total_due, l.financial_snapshot
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1
  AND a.tenant_id = $2
FOR UPDATE`, lotID, tenantID).Scan(&status, &winnerUserID, &totalAmount, &snapshot); err != nil {
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
  id, lot_id, tenant_id, winner_user_id,
  invoice_number, issued_at, due_date, currency,
  total_amount, snapshot, status, created_at
) VALUES (
  $1,$2,$3,$4,
  $5,$6,$7,'USD',
  $8,$9,'ISSUED',now()
)`, invoiceID, lotID, tenantID, *winnerUserID, invoiceNumber, now, dueDate, *totalAmount, snapshot); err != nil {
		return nil, err
	}

	inv, err := scanInvoice(tx.QueryRow(ctx, `
SELECT i.id, i.lot_id, i.tenant_id, i.winner_user_id,
       i.invoice_number, i.issued_at, i.due_date, i.currency,
       i.total_amount, i.snapshot, i.status, i.created_at,
       l.lot_number, l.title, a.title
FROM invoices i
JOIN lots l ON l.id = i.lot_id
JOIN auctions a ON a.id = l.auction_id
WHERE i.id = $1`, invoiceID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}
