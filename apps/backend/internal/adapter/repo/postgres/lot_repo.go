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

type LotRepo struct {
	DB *pgxpool.Pool
}

func scanLot(row pgx.Row) (*domain.Lot, error) {
	var l domain.Lot
	err := row.Scan(
		&l.ID, &l.AuctionID, &l.LotNumber, &l.OrderIndex,
		&l.Title, &l.Description,
		&l.StartingPrice, &l.CurrentPrice, &l.ReservePrice, &l.MinIncrement,
		&l.Metadata, &l.Status,
		&l.StartedAt, &l.EndedAt, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	const q = `
INSERT INTO lots (
  id, auction_id, lot_number, order_index, title, description,
  starting_price, current_price, reserve_price, min_increment,
  metadata, status, created_at, updated_at
)
SELECT $1, $2, $3, $4, $5, $6,
       $7, $8, $9, $10,
       $11, $12, now(), now()
WHERE EXISTS (
  SELECT 1 FROM auctions WHERE id = $2 AND tenant_id = $13
)`
	ct, err := r.DB.Exec(ctx, q,
		lot.ID, lot.AuctionID, lot.LotNumber, lot.OrderIndex, lot.Title, lot.Description,
		lot.StartingPrice, lot.CurrentPrice, lot.ReservePrice, lot.MinIncrement,
		lot.Metadata, lot.Status, tenantID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found")
	}
	return nil
}

func (r *LotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	const q = `
UPDATE lots l
SET lot_number = $3,
    order_index = $4,
    title = $5,
    description = $6,
    starting_price = $7,
    reserve_price = $8,
    min_increment = $9,
    metadata = $10,
    updated_at = now()
FROM auctions a
WHERE l.auction_id = a.id
  AND a.tenant_id = $1
  AND l.id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, lot.ID,
		lot.LotNumber, lot.OrderIndex, lot.Title, lot.Description, lot.StartingPrice, lot.ReservePrice, lot.MinIncrement, lot.Metadata,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("lot not found")
	}
	return nil
}

func (r *LotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error {
	const q = `
DELETE FROM lots l
USING auctions a
WHERE l.auction_id = a.id
  AND a.tenant_id = $1
  AND l.id = $2
  AND l.status IN ('PENDING','CANCELLED','ENDED')`
	ct, err := r.DB.Exec(ctx, q, tenantID, lotID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("lot not found or cannot be deleted")
	}
	return nil
}

func (r *LotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const auctionCheck = `SELECT status FROM auctions WHERE id = $1 AND tenant_id = $2 FOR UPDATE`
	var status domain.AuctionStatus
	if err := tx.QueryRow(ctx, auctionCheck, auctionID, tenantID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("auction not found")
		}
		return err
	}
	if status == domain.AuctionLive {
		return errors.New("cannot reorder while auction is live")
	}

	const upd = `UPDATE lots SET order_index = $3, updated_at = now() WHERE id = $1 AND auction_id = $2`
	for i, id := range lotIDs {
		ct, err := tx.Exec(ctx, upd, id, auctionID, i+1)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return errors.New("invalid lot in reorder request")
		}
	}
	return tx.Commit(ctx)
}

func (r *LotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	const q = `
SELECT l.id, l.auction_id, l.lot_number, l.order_index,
       l.title, l.description,
       l.starting_price, l.current_price, l.reserve_price, l.min_increment,
       l.metadata, l.status,
       l.started_at, l.ended_at, l.created_at, l.updated_at
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE a.tenant_id = $1 AND l.auction_id = $2
ORDER BY l.order_index ASC`
	rows, err := r.DB.Query(ctx, q, tenantID, auctionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Lot, 0)
	for rows.Next() {
		l, err := scanLot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *LotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	const q = `
SELECT l.id, l.auction_id, l.lot_number, l.order_index,
       l.title, l.description,
       l.starting_price, l.current_price, l.reserve_price, l.min_increment,
       l.metadata, l.status,
       l.started_at, l.ended_at, l.created_at, l.updated_at
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE a.tenant_id = $1 AND l.id = $2`
	lot, err := scanLot(r.DB.QueryRow(ctx, q, tenantID, lotID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("lot not found")
	}
	return lot, err
}

func (r *LotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const lockQ = `
SELECT l.auction_id, a.status, a.end_time, a.auction_mode, a.lot_duration_seconds, l.status
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1 AND a.tenant_id = $2
FOR UPDATE`
	var auctionID uuid.UUID
	var auctionStatus domain.AuctionStatus
	var auctionEnd time.Time
	var mode domain.AuctionMode
	var lotDuration int
	var status domain.LotStatus
	if err := tx.QueryRow(ctx, lockQ, lotID, tenantID).Scan(&auctionID, &auctionStatus, &auctionEnd, &mode, &lotDuration, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("lot not found")
		}
		return err
	}
	if auctionStatus != domain.AuctionLive {
		return errors.New("auction is not live")
	}
	if status == domain.LotLive {
		return errors.New("lot already live")
	}
	if status == domain.LotCancelled {
		return errors.New("cancelled lot cannot be started")
	}

	if mode == domain.AuctionModeSequential {
		const qOtherLive = `SELECT 1 FROM lots WHERE auction_id = $1 AND status = 'LIVE' AND id <> $2 LIMIT 1`
		var one int
		if err := tx.QueryRow(ctx, qOtherLive, auctionID, lotID).Scan(&one); err == nil {
			return errors.New("another lot is already live")
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	var endedAt *time.Time
	if mode == domain.AuctionModeSequential {
		if lotDuration > 0 {
			t := now.Add(time.Duration(lotDuration) * time.Second)
			endedAt = &t
		}
	} else {
		if !now.Before(auctionEnd) {
			return errors.New("auction already ended")
		}
		t := auctionEnd
		endedAt = &t
	}
	const upd = `UPDATE lots SET status = 'LIVE', started_at = $3, ended_at = $4, updated_at = now() WHERE id = $1 AND auction_id = $2`
	if _, err := tx.Exec(ctx, upd, lotID, auctionID, now, endedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const lockQ = `
SELECT l.auction_id, a.auction_mode
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1 AND a.tenant_id = $2
FOR UPDATE`
	var auctionID uuid.UUID
	var mode domain.AuctionMode
	if err := tx.QueryRow(ctx, lockQ, lotID, tenantID).Scan(&auctionID, &mode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("lot not found")
		}
		return err
	}

	const endQ = `UPDATE lots SET status = 'ENDED', ended_at = $2, updated_at = now() WHERE id = $1`
	if _, err := tx.Exec(ctx, endQ, lotID, now); err != nil {
		return err
	}

	if mode == domain.AuctionModeSequential {
		const nextQ = `
SELECT id
FROM lots
WHERE auction_id = $1 AND status = 'PENDING'
ORDER BY order_index ASC
LIMIT 1
FOR UPDATE`
		var nextID uuid.UUID
		err = tx.QueryRow(ctx, nextQ, auctionID).Scan(&nextID)
		if err == nil {
			const auctionCfgQ = `SELECT lot_duration_seconds FROM auctions WHERE id = $1`
			var lotDuration int
			if err := tx.QueryRow(ctx, auctionCfgQ, auctionID).Scan(&lotDuration); err != nil {
				return err
			}
			var endedAt *time.Time
			if lotDuration > 0 {
				t := now.Add(time.Duration(lotDuration) * time.Second)
				endedAt = &t
			}
			const activateQ = `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1`
			if _, err := tx.Exec(ctx, activateQ, nextID, now, endedAt); err != nil {
				return err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *LotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const cfgQ = `
SELECT auction_mode, lot_duration_seconds, end_time
FROM auctions
WHERE id = $1 AND tenant_id = $2
FOR UPDATE`
	var mode domain.AuctionMode
	var lotDuration int
	var auctionEnd time.Time
	if err := tx.QueryRow(ctx, cfgQ, auctionID, tenantID).Scan(&mode, &lotDuration, &auctionEnd); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("auction not found")
		}
		return err
	}

	if mode == "" {
		mode = domain.AuctionModeTimed
	}

	if mode == domain.AuctionModeTimed {
		const q = `
UPDATE lots
SET status = 'LIVE', started_at = COALESCE(started_at, $2), ended_at = $3, updated_at = now()
WHERE auction_id = $1 AND status IN ('PENDING','PAUSED')`
		if _, err := tx.Exec(ctx, q, auctionID, now, auctionEnd); err != nil {
			return err
		}
	} else {
		const qDeactivate = `UPDATE lots SET status = 'PENDING', started_at = NULL, ended_at = NULL, updated_at = now() WHERE auction_id = $1 AND status <> 'ENDED'`
		if _, err := tx.Exec(ctx, qDeactivate, auctionID); err != nil {
			return err
		}
		const qNext = `SELECT id FROM lots WHERE auction_id = $1 AND status = 'PENDING' ORDER BY order_index ASC LIMIT 1 FOR UPDATE`
		var nextID uuid.UUID
		if err := tx.QueryRow(ctx, qNext, auctionID).Scan(&nextID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return tx.Commit(ctx)
			}
			return err
		}
		var endedAt *time.Time
		if lotDuration > 0 {
			t := now.Add(time.Duration(lotDuration) * time.Second)
			endedAt = &t
		}
		const qActivate = `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1`
		if _, err := tx.Exec(ctx, qActivate, nextID, now, endedAt); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *LotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const lockQ = `
SELECT l.id, l.auction_id, l.lot_number, l.order_index,
       l.title, l.description,
       l.starting_price, l.current_price, l.reserve_price, l.min_increment,
       l.metadata, l.status,
       l.started_at, l.ended_at, l.created_at, l.updated_at,
       a.status,
       a.auto_extend_enabled, a.sniping_window_seconds, a.sniping_extension_seconds
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1 AND a.tenant_id = $2
FOR UPDATE`
	var lot domain.Lot
	var auctionStatus domain.AuctionStatus
	var autoExtend bool
	var snipingWindow, snipingExtension int
	err = tx.QueryRow(ctx, lockQ, in.LotID, in.TenantID).Scan(
		&lot.ID, &lot.AuctionID, &lot.LotNumber, &lot.OrderIndex,
		&lot.Title, &lot.Description,
		&lot.StartingPrice, &lot.CurrentPrice, &lot.ReservePrice, &lot.MinIncrement,
		&lot.Metadata, &lot.Status,
		&lot.StartedAt, &lot.EndedAt, &lot.CreatedAt, &lot.UpdatedAt,
		&auctionStatus,
		&autoExtend, &snipingWindow, &snipingExtension,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("lot not found")
	}
	if err != nil {
		return nil, err
	}

	if auctionStatus != domain.AuctionLive {
		return nil, errors.New("auction is not live")
	}
	if lot.Status != domain.LotLive {
		return nil, errors.New("lot is not live")
	}
	if lot.EndedAt != nil && !in.Now.Before(*lot.EndedAt) {
		return nil, errors.New("lot already ended")
	}
	minBid := lot.CurrentPrice + lot.MinIncrement
	if in.Amount < minBid {
		return nil, errors.New("bid amount below minimum increment")
	}

	bidID := uuid.New()
	const insBid = `INSERT INTO lot_bids (id, lot_id, user_id, amount, created_at) VALUES ($1,$2,$3,$4,$5)`
	if _, err := tx.Exec(ctx, insBid, bidID, lot.ID, in.UserID, in.Amount, in.Now); err != nil {
		return nil, err
	}

	newEndedAt := lot.EndedAt
	if autoExtend && lot.EndedAt != nil {
		windowStart := lot.EndedAt.Add(-time.Duration(snipingWindow) * time.Second)
		if in.Now.After(windowStart) {
			t := lot.EndedAt.Add(time.Duration(snipingExtension) * time.Second)
			newEndedAt = &t
		}
	}

	const updLot = `UPDATE lots SET current_price = $2, ended_at = $3, updated_at = now() WHERE id = $1`
	if _, err := tx.Exec(ctx, updLot, lot.ID, in.Amount, newEndedAt); err != nil {
		return nil, err
	}

	lot.CurrentPrice = in.Amount
	lot.EndedAt = newEndedAt
	bid := domain.Bid{ID: bidID, LotID: &lot.ID, UserID: in.UserID, Amount: in.Amount, CreatedAt: in.Now}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &port.PlaceLotBidTxResult{Lot: lot, Bid: bid}, nil
}

func (r *LotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	const qTimedEnd = `
UPDATE lots l
SET status = 'ENDED', updated_at = now()
FROM auctions a
WHERE l.auction_id = a.id
  AND a.auction_mode = 'TIMED'
  AND l.status = 'LIVE'
  AND l.ended_at IS NOT NULL
  AND l.ended_at <= $1`
	if _, err := r.DB.Exec(ctx, qTimedEnd, now); err != nil {
		return err
	}

	const qSeqAuctions = `
SELECT id, lot_duration_seconds
FROM auctions
WHERE status = 'LIVE'
  AND auction_mode = 'SEQUENTIAL'
  AND auto_progress_enabled = true`
	rows, err := r.DB.Query(ctx, qSeqAuctions)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var auctionID uuid.UUID
		var lotDuration int
		if err := rows.Scan(&auctionID, &lotDuration); err != nil {
			return err
		}

		tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}

		var liveLotID uuid.UUID
		var lotEndedAt *time.Time
		err = tx.QueryRow(ctx, `SELECT id, ended_at FROM lots WHERE auction_id = $1 AND status = 'LIVE' ORDER BY order_index ASC LIMIT 1 FOR UPDATE`, auctionID).Scan(&liveLotID, &lotEndedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			tx.Rollback(ctx)
			continue
		}
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		if lotEndedAt == nil || lotEndedAt.After(now) {
			tx.Rollback(ctx)
			continue
		}

		if _, err := tx.Exec(ctx, `UPDATE lots SET status = 'ENDED', updated_at = now() WHERE id = $1`, liveLotID); err != nil {
			tx.Rollback(ctx)
			return err
		}

		var nextID uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id FROM lots WHERE auction_id = $1 AND status = 'PENDING' ORDER BY order_index ASC LIMIT 1 FOR UPDATE`, auctionID).Scan(&nextID)
		if err == nil {
			var nextEndedAt *time.Time
			if lotDuration > 0 {
				t := now.Add(time.Duration(lotDuration) * time.Second)
				nextEndedAt = &t
			}
			if _, err := tx.Exec(ctx, `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1`, nextID, now, nextEndedAt); err != nil {
				tx.Rollback(ctx)
				return err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			tx.Rollback(ctx)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (r *LotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	if limit <= 0 {
		limit = 100
	}
	const q = `
SELECT b.id, b.user_id, b.amount, b.created_at
FROM lot_bids b
JOIN lots l ON l.id = b.lot_id
JOIN auctions a ON a.id = l.auction_id
WHERE a.tenant_id = $1 AND b.lot_id = $2
ORDER BY b.created_at DESC
LIMIT $3`
	rows, err := r.DB.Query(ctx, q, tenantID, lotID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Bid, 0)
	for rows.Next() {
		var b domain.Bid
		if err := rows.Scan(&b.ID, &b.UserID, &b.Amount, &b.CreatedAt); err != nil {
			return nil, err
		}
		id := lotID
		b.LotID = &id
		out = append(out, b)
	}
	return out, rows.Err()
}
