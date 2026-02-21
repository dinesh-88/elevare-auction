package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
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
		&l.HammerPrice, &l.TotalDue, &l.WinnerUserID, &l.FinancialData,
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
       l.hammer_price, l.total_due, l.winner_user_id, l.financial_snapshot,
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
       l.hammer_price, l.total_due, l.winner_user_id, l.financial_snapshot,
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
	if status == domain.LotEnded {
		return errors.New("ended lot cannot be restarted")
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
SELECT l.auction_id, l.status, a.auction_mode
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1 AND a.tenant_id = $2
FOR UPDATE`
	var auctionID uuid.UUID
	var lotStatus domain.LotStatus
	var mode domain.AuctionMode
	if err := tx.QueryRow(ctx, lockQ, lotID, tenantID).Scan(&auctionID, &lotStatus, &mode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("lot not found")
		}
		return err
	}
	if lotStatus == domain.LotEnded || lotStatus == domain.LotCancelled {
		return errors.New("lot already closed")
	}

	if err := r.finalizeLotSettlementTx(ctx, tx, lotID, now); err != nil {
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
			const activateQ = `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1 AND status = 'PENDING'`
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
       l.hammer_price, l.total_due, l.winner_user_id, l.financial_snapshot,
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
		&lot.HammerPrice, &lot.TotalDue, &lot.WinnerUserID, &lot.FinancialData,
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
	const qTimedCandidates = `
SELECT l.id
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE a.auction_mode = 'TIMED'
  AND a.status = 'LIVE'
  AND l.status = 'LIVE'
  AND l.ended_at IS NOT NULL
  AND l.ended_at <= $1`
	rows, err := r.DB.Query(ctx, qTimedCandidates, now)
	if err != nil {
		return err
	}
	candidates := make([]uuid.UUID, 0)
	for rows.Next() {
		var lotID uuid.UUID
		if err := rows.Scan(&lotID); err != nil {
			rows.Close()
			return err
		}
		candidates = append(candidates, lotID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, lotID := range candidates {
		tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		if err := r.finalizeLotSettlementTx(ctx, tx, lotID, now); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	const qSeqAuctions = `
SELECT a.id, a.lot_duration_seconds
FROM auctions a
WHERE a.status = 'LIVE'
  AND a.auction_mode = 'SEQUENTIAL'
  AND a.auto_progress_enabled = true`
	seqRows, err := r.DB.Query(ctx, qSeqAuctions)
	if err != nil {
		return err
	}
	type seqAuction struct {
		auctionID   uuid.UUID
		lotDuration int
	}
	seqAuctions := make([]seqAuction, 0)
	for seqRows.Next() {
		var a seqAuction
		if err := seqRows.Scan(&a.auctionID, &a.lotDuration); err != nil {
			seqRows.Close()
			return err
		}
		seqAuctions = append(seqAuctions, a)
	}
	if err := seqRows.Err(); err != nil {
		seqRows.Close()
		return err
	}
	seqRows.Close()

	for _, a := range seqAuctions {
		tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}

		var liveLotID uuid.UUID
		var lotEndedAt *time.Time
		err = tx.QueryRow(ctx, `SELECT id, ended_at FROM lots WHERE auction_id = $1 AND status = 'LIVE' ORDER BY order_index ASC LIMIT 1 FOR UPDATE`, a.auctionID).Scan(&liveLotID, &lotEndedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			continue
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if lotEndedAt == nil || lotEndedAt.After(now) {
			_ = tx.Rollback(ctx)
			continue
		}

		if err := r.finalizeLotSettlementTx(ctx, tx, liveLotID, now); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}

		var nextID uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id FROM lots WHERE auction_id = $1 AND status = 'PENDING' ORDER BY order_index ASC LIMIT 1 FOR UPDATE`, a.auctionID).Scan(&nextID)
		if err == nil {
			var nextEndedAt *time.Time
			if a.lotDuration > 0 {
				t := now.Add(time.Duration(a.lotDuration) * time.Second)
				nextEndedAt = &t
			}
			if _, err := tx.Exec(ctx, `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1 AND status = 'PENDING'`, nextID, now, nextEndedAt); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *LotRepo) finalizeLotSettlementTx(
	ctx context.Context,
	tx pgx.Tx,
	lotID uuid.UUID,
	now time.Time,
) error {
	var currentStatus domain.LotStatus
	var auctionID, tenantID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT l.status, l.auction_id, a.tenant_id
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1
FOR UPDATE`, lotID).Scan(&currentStatus, &auctionID, &tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("lot not found")
		}
		return err
	}
	if currentStatus == domain.LotEnded || currentStatus == domain.LotCancelled {
		return nil
	}

	var winnerUserID *uuid.UUID
	var hammerPrice *int64
	var winnerID uuid.UUID
	var winningAmount int64
	err := tx.QueryRow(ctx, `
SELECT user_id, amount
FROM lot_bids
WHERE lot_id = $1
ORDER BY amount DESC, created_at ASC
LIMIT 1`, lotID).Scan(&winnerID, &winningAmount)
	if err == nil {
		winnerUserID = &winnerID
		hammerPrice = &winningAmount
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	var totalDue *int64
	var snapshot []byte
	if hammerPrice != nil {
		fees, err := r.resolveFeeComponentsTx(ctx, tx, lotID, auctionID, tenantID)
		if err != nil {
			return err
		}
		components, computedTotal := calculateFinancials(*hammerPrice, fees)
		totalDue = &computedTotal
		payload := domain.LotFinancialSnapshot{
			HammerPrice: *hammerPrice,
			Fees:        components,
			TotalDue:    computedTotal,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		snapshot = b
	}

	const q = `
UPDATE lots
SET status = 'ENDED',
    ended_at = $2,
    hammer_price = $3,
    winner_user_id = $4,
    total_due = $5,
    financial_snapshot = $6,
    updated_at = now()
WHERE id = $1
  AND status NOT IN ('ENDED','CANCELLED')`
	if _, err := tx.Exec(ctx, q, lotID, now, hammerPrice, winnerUserID, totalDue, snapshot); err != nil {
		return err
	}
	return nil
}

type rawFeeRule struct {
	ID        uuid.UUID
	Name      string
	Type      string
	Value     float64
	AppliesTo string
	Taxable   bool
	CreatedAt time.Time
}

func (r *LotRepo) resolveFeeComponentsTx(ctx context.Context, tx pgx.Tx, lotID, auctionID, tenantID uuid.UUID) ([]rawFeeRule, error) {
	byScope := func(scopeType string, scopeID uuid.UUID) ([]rawFeeRule, error) {
		const q = `
SELECT fc.id, fc.name, fc.calculation_type, fc.value::double precision, fc.applies_to, fc.taxable, fc.created_at
FROM fee_assignments fa
JOIN fee_components fc ON fc.id = fa.fee_component_id
WHERE fa.scope_type = $1
  AND fa.scope_id = $2
  AND fc.active = true
ORDER BY fc.created_at ASC, fc.id ASC`
		rows, err := tx.Query(ctx, q, scopeType, scopeID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := make([]rawFeeRule, 0)
		for rows.Next() {
			var fee rawFeeRule
			if err := rows.Scan(&fee.ID, &fee.Name, &fee.Type, &fee.Value, &fee.AppliesTo, &fee.Taxable, &fee.CreatedAt); err != nil {
				return nil, err
			}
			out = append(out, fee)
		}
		return out, rows.Err()
	}

	if fees, err := byScope("LOT", lotID); err != nil {
		return nil, err
	} else if len(fees) > 0 {
		return fees, nil
	}
	if fees, err := byScope("AUCTION", auctionID); err != nil {
		return nil, err
	} else if len(fees) > 0 {
		return fees, nil
	}
	return byScope("TENANT", tenantID)
}

func calculateFinancials(hammer int64, rules []rawFeeRule) ([]domain.FinancialFee, int64) {
	if hammer < 0 {
		hammer = 0
	}
	percentageRules := make([]rawFeeRule, 0)
	fixedRules := make([]rawFeeRule, 0)
	for _, r := range rules {
		if r.Value < 0 || r.Name == "" {
			continue
		}
		if r.Type == string(domain.FeeTypeFixed) {
			fixedRules = append(fixedRules, r)
		} else if r.Type == string(domain.FeeTypePercentage) {
			percentageRules = append(percentageRules, r)
		}
	}
	sort.SliceStable(percentageRules, func(i, j int) bool {
		ri, rj := appliesToRank(percentageRules[i].AppliesTo), appliesToRank(percentageRules[j].AppliesTo)
		if ri != rj {
			return ri < rj
		}
		if percentageRules[i].CreatedAt != percentageRules[j].CreatedAt {
			return percentageRules[i].CreatedAt.Before(percentageRules[j].CreatedAt)
		}
		return percentageRules[i].ID.String() < percentageRules[j].ID.String()
	})
	sort.SliceStable(fixedRules, func(i, j int) bool {
		if fixedRules[i].CreatedAt != fixedRules[j].CreatedAt {
			return fixedRules[i].CreatedAt.Before(fixedRules[j].CreatedAt)
		}
		return fixedRules[i].ID.String() < fixedRules[j].ID.String()
	})

	hammerBase := float64(hammer)
	premiumBase := 0.0
	subtotalBase := hammerBase
	fees := make([]domain.FinancialFee, 0, len(rules))
	totalFees := int64(0)
	for _, rule := range percentageRules {
		var base float64
		switch rule.AppliesTo {
		case "HAMMER":
			base = hammerBase
		case "PREMIUM":
			base = premiumBase
		case "SUBTOTAL":
			base = subtotalBase
		default:
			base = hammerBase
		}
		amount := int64(math.Round(base * rule.Value / 100))
		if amount < 0 {
			amount = 0
		}
		totalFees += amount
		subtotalBase += float64(amount)
		if rule.AppliesTo == "HAMMER" || rule.AppliesTo == "PREMIUM" {
			premiumBase += float64(amount)
		}
		fees = append(fees, domain.FinancialFee{
			Name:      rule.Name,
			Type:      domain.FeeTypePercentage,
			Value:     rule.Value,
			Amount:    amount,
			AppliesTo: rule.AppliesTo,
			Taxable:   rule.Taxable,
		})
	}
	for _, rule := range fixedRules {
		amount := int64(0)
		amount = int64(math.Round(rule.Value))
		if amount < 0 {
			amount = 0
		}
		totalFees += amount
		subtotalBase += float64(amount)
		fees = append(fees, domain.FinancialFee{
			Name:      rule.Name,
			Type:      domain.FeeTypeFixed,
			Value:     rule.Value,
			Amount:    amount,
			AppliesTo: rule.AppliesTo,
			Taxable:   rule.Taxable,
		})
	}
	return fees, hammer + totalFees
}

func appliesToRank(appliesTo string) int {
	switch appliesTo {
	case "HAMMER":
		return 1
	case "PREMIUM":
		return 2
	case "SUBTOTAL":
		return 3
	default:
		return 4
	}
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
