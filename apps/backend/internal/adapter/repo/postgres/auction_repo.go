package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuctionRepo struct {
	DB *pgxpool.Pool
}

func (r *AuctionRepo) Create(ctx context.Context, auction *domain.Auction) error {
	const q = `
INSERT INTO auctions (
  id, tenant_id, title, description, metadata, starting_price, current_price, reserve_price, min_increment,
  start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
  auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds, sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,
  $10,$11,$12,$13,$14,$15,
  $16,$17,$18,$19,$20,$21,$22,now()
)`
	_, err := r.DB.Exec(ctx, q,
		auction.ID, auction.TenantID, auction.Title, auction.Description, auction.Metadata,
		auction.StartingPrice, auction.CurrentPrice, auction.ReservePrice, auction.MinIncrement,
		auction.StartTime, auction.EndTime, auction.Status,
		auction.AllowManualStart, auction.AllowManualPause, auction.AllowManualEnd,
		auction.AutoExtendEnabled, auction.AuctionMode, auction.AutoProgressEnabled, auction.LotDurationSeconds,
		auction.SnipingWindowSeconds, auction.SnipingExtensionSeconds, auction.BidVisibility,
	)
	return err
}

func scanAuction(row pgx.Row) (*domain.Auction, error) {
	var a domain.Auction
	err := row.Scan(
		&a.ID, &a.TenantID, &a.Title, &a.Description, &a.Metadata,
		&a.StartingPrice, &a.CurrentPrice, &a.ReservePrice, &a.MinIncrement,
		&a.StartTime, &a.EndTime, &a.Status,
		&a.AllowManualStart, &a.AllowManualPause, &a.AllowManualEnd,
		&a.AutoExtendEnabled, &a.AuctionMode, &a.AutoProgressEnabled, &a.LotDurationSeconds,
		&a.SnipingWindowSeconds, &a.SnipingExtensionSeconds, &a.BidVisibility, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AuctionRepo) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	const q = `
SELECT id, tenant_id, title, description, metadata, starting_price, current_price, reserve_price, min_increment,
       start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
       auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds, sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
FROM auctions
WHERE tenant_id = $1 AND status IN ('LIVE','SCHEDULED') AND end_time > $2
ORDER BY start_time ASC`
	rows, err := r.DB.Query(ctx, q, tenantID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Auction, 0)
	for rows.Next() {
		a, err := scanAuction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *AuctionRepo) ListAdmin(ctx context.Context, tenantID uuid.UUID) ([]domain.Auction, error) {
	const q = `
SELECT id, tenant_id, title, description, metadata, starting_price, current_price, reserve_price, min_increment,
       start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
       auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds, sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
FROM auctions
WHERE tenant_id = $1
ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Auction, 0)
	for rows.Next() {
		a, err := scanAuction(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *AuctionRepo) GetByID(ctx context.Context, tenantID, auctionID uuid.UUID) (*domain.Auction, error) {
	const q = `
SELECT id, tenant_id, title, description, metadata, starting_price, current_price, reserve_price, min_increment,
       start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
       auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds, sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
FROM auctions
WHERE tenant_id = $1 AND id = $2`
	a, err := scanAuction(r.DB.QueryRow(ctx, q, tenantID, auctionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("auction not found")
	}
	return a, err
}

func (r *AuctionRepo) Update(ctx context.Context, auction *domain.Auction) error {
	const q = `
UPDATE auctions
SET title = $3,
    description = $4,
    starting_price = $5,
    reserve_price = $6,
    min_increment = $7,
    start_time = $8,
    end_time = $9,
    allow_manual_start = $10,
    allow_manual_pause = $11,
    allow_manual_end = $12,
    auto_extend_enabled = $13,
    auction_mode = $14,
    auto_progress_enabled = $15,
    lot_duration_seconds = $16,
    sniping_window_seconds = $17,
    sniping_extension_seconds = $18,
    bid_visibility = $19
WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q,
		auction.TenantID, auction.ID,
		auction.Title, auction.Description, auction.StartingPrice, auction.ReservePrice, auction.MinIncrement,
		auction.StartTime, auction.EndTime,
		auction.AllowManualStart, auction.AllowManualPause, auction.AllowManualEnd,
		auction.AutoExtendEnabled, auction.AuctionMode, auction.AutoProgressEnabled, auction.LotDurationSeconds,
		auction.SnipingWindowSeconds, auction.SnipingExtensionSeconds, auction.BidVisibility,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found")
	}
	return nil
}

func (r *AuctionRepo) Delete(ctx context.Context, tenantID, auctionID uuid.UUID) error {
	const q = `
DELETE FROM auctions
WHERE tenant_id = $1
  AND id = $2
  AND status IN ('DRAFT','SCHEDULED','PAUSED','ENDED','CANCELLED')`
	ct, err := r.DB.Exec(ctx, q, tenantID, auctionID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found or cannot be deleted while live")
	}
	return nil
}

func (r *AuctionRepo) UpdateStatus(ctx context.Context, tenantID, auctionID uuid.UUID, status domain.AuctionStatus) error {
	const q = `UPDATE auctions SET status = $3 WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, auctionID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found")
	}
	return nil
}

func (r *AuctionRepo) PlaceBidTx(ctx context.Context, in port.PlaceBidTxInput) (*port.PlaceBidTxResult, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const lockQ = `
SELECT id, tenant_id, title, description, metadata, starting_price, current_price, reserve_price, min_increment,
       start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
       auto_extend_enabled, sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
FROM auctions
WHERE tenant_id = $1 AND id = $2
FOR UPDATE`
	a, err := scanAuction(tx.QueryRow(ctx, lockQ, in.TenantID, in.AuctionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("auction not found")
	}
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AuctionLive {
		return nil, errors.New("auction is not live")
	}
	if !in.Now.Before(a.EndTime) {
		return nil, errors.New("auction already ended")
	}
	minBid := a.CurrentPrice + a.MinIncrement
	if in.Amount < minBid {
		return nil, errors.New("bid amount below minimum increment")
	}

	bid := domain.Bid{
		ID:        uuid.New(),
		AuctionID: a.ID,
		UserID:    in.UserID,
		Amount:    in.Amount,
		CreatedAt: in.Now,
	}
	const insBid = `INSERT INTO bids (id, auction_id, user_id, amount, created_at) VALUES ($1,$2,$3,$4,$5)`
	if _, err := tx.Exec(ctx, insBid, bid.ID, bid.AuctionID, bid.UserID, bid.Amount, bid.CreatedAt); err != nil {
		return nil, err
	}

	newEnd := a.EndTime
	if a.AutoExtendEnabled {
		windowStart := a.EndTime.Add(-time.Duration(a.SnipingWindowSeconds) * time.Second)
		if in.Now.After(windowStart) {
			newEnd = a.EndTime.Add(time.Duration(a.SnipingExtensionSeconds) * time.Second)
		}
	}
	const updAuction = `UPDATE auctions SET current_price = $3, end_time = $4 WHERE tenant_id = $1 AND id = $2`
	if _, err := tx.Exec(ctx, updAuction, in.TenantID, in.AuctionID, in.Amount, newEnd); err != nil {
		return nil, err
	}

	a.CurrentPrice = in.Amount
	a.EndTime = newEnd
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &port.PlaceBidTxResult{Auction: *a, Bid: bid}, nil
}

func (r *AuctionRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	const qStart = `
UPDATE auctions
SET status = 'LIVE'
WHERE status = 'SCHEDULED'
  AND start_time <= $1
  AND end_time > $1`
	if _, err := r.DB.Exec(ctx, qStart, now); err != nil {
		return err
	}
	const qEnd = `
UPDATE auctions
SET status = 'ENDED'
WHERE status IN ('LIVE','PAUSED','SCHEDULED')
  AND end_time <= $1`
	_, err := r.DB.Exec(ctx, qEnd, now)
	return err
}
