package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LotRepo struct {
	DB    *pgxpool.Pool
	Audit port.AuditRepository
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
	if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
		TenantID:   &tenantID,
		Action:     "LOT_STARTED",
		EntityType: "LOT",
		EntityID:   &lotID,
		Metadata: mustMarshalJSON(map[string]any{
			"auction_id": auctionID,
			"reason":     "MANUAL_START",
			"started_at": now,
			"ended_at":   endedAt,
		}),
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
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
			return nil, errors.New("lot not found")
		}
		return nil, err
	}
	if lotStatus == domain.LotEnded || lotStatus == domain.LotCancelled {
		return nil, errors.New("lot already closed")
	}

	settlement, err := r.finalizeLotSettlementTx(ctx, tx, lotID, now, "MANUAL_END")
	if err != nil {
		return nil, err
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
				return nil, err
			}
			var endedAt *time.Time
			if lotDuration > 0 {
				t := now.Add(time.Duration(lotDuration) * time.Second)
				endedAt = &t
			}
			const activateQ = `UPDATE lots SET status = 'LIVE', started_at = $2, ended_at = $3, updated_at = now() WHERE id = $1 AND status = 'PENDING'`
			if _, err := tx.Exec(ctx, activateQ, nextID, now, endedAt); err != nil {
				return nil, err
			}
			if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
				TenantID:   &tenantID,
				Action:     "LOT_STARTED",
				EntityType: "LOT",
				EntityID:   &nextID,
				Metadata: mustMarshalJSON(map[string]any{
					"auction_id": auctionID,
					"reason":     "SEQUENTIAL_NEXT",
					"started_at": now,
					"ended_at":   endedAt,
				}),
			}); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return settlement, nil
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
		if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
			TenantID:   &tenantID,
			Action:     "LOT_STARTED",
			EntityType: "AUCTION",
			EntityID:   &auctionID,
			Metadata: mustMarshalJSON(map[string]any{
				"reason":      "AUCTION_TIMED_START",
				"started_at":  now,
				"default_end": auctionEnd,
			}),
		}); err != nil {
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
		if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
			TenantID:   &tenantID,
			Action:     "LOT_STARTED",
			EntityType: "LOT",
			EntityID:   &nextID,
			Metadata: mustMarshalJSON(map[string]any{
				"auction_id": auctionID,
				"reason":     "AUCTION_SEQUENTIAL_START",
				"started_at": now,
				"ended_at":   endedAt,
			}),
		}); err != nil {
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
       a.title,
       a.currency,
       a.auto_extend_enabled, a.sniping_window_seconds, a.sniping_extension_seconds
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1 AND a.tenant_id = $2
FOR UPDATE`
	var lot domain.Lot
	var auctionStatus domain.AuctionStatus
	var auctionTitle string
	var currency string
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
		&auctionTitle,
		&currency,
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
	if lot.EndedAt == nil || !in.Now.Before(*lot.EndedAt) {
		return nil, errors.New("lot already ended")
	}
	if (in.Amount == nil && in.MaxBid == nil) || (in.Amount != nil && in.MaxBid != nil) {
		return nil, errors.New("provide either bid_amount_minor or max_bid_minor")
	}
	minBid := lot.CurrentPrice + lot.MinIncrement
	if in.Amount != nil && *in.Amount < minBid {
		return nil, errors.New("bid amount below minimum increment")
	}
	if in.MaxBid != nil && *in.MaxBid < minBid {
		return nil, errors.New("bid amount below minimum increment")
	}

	const lastBidQ = `
SELECT created_at
FROM lot_bids
WHERE tenant_id = $1 AND lot_id = $2 AND user_id = $3
ORDER BY created_at DESC
LIMIT 1`
	var lastBidAt time.Time
	if err := tx.QueryRow(ctx, lastBidQ, in.TenantID, lot.ID, in.UserID).Scan(&lastBidAt); err == nil {
		if in.Now.Sub(lastBidAt) < time.Second {
			return nil, errors.New("bidding too quickly, please wait a second")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	previousLeader, err := fetchLotBidLeaderTx(ctx, tx, in.TenantID, lot.ID)
	if err != nil {
		return nil, err
	}

	seedAmount := minBid
	seedMax := minBid
	if in.Amount != nil {
		seedAmount = *in.Amount
		seedMax = *in.Amount
	}
	if in.MaxBid != nil {
		seedMax = *in.MaxBid
	}
	seedMaxForInsert := (*int64)(nil)
	if in.MaxBid != nil {
		seedMaxForInsert = &seedMax
	}
	seedBid, err := insertLotBidTx(ctx, tx, in.TenantID, lot.ID, in.UserID, seedAmount, seedMaxForInsert, in.Now)
	if err != nil {
		return nil, err
	}

	participants, err := fetchLotBidParticipantsTx(ctx, tx, in.TenantID, lot.ID)
	if err != nil {
		return nil, err
	}
	outcome := resolveProxyOutcome(participants, lot.MinIncrement, minBid, seedAmount)

	acceptedBid := seedBid
	if outcome.VisiblePrice > seedBid.Amount {
		autoAt := in.Now.Add(time.Millisecond)
		winnerMax := outcome.WinnerMax
		autoBid, err := insertLotBidTx(ctx, tx, in.TenantID, lot.ID, outcome.WinnerUserID, outcome.VisiblePrice, &winnerMax, autoAt)
		if err != nil {
			return nil, err
		}
		acceptedBid = autoBid
	}

	previousHighestBidder := (*uuid.UUID)(nil)
	if previousLeader != nil && previousLeader.UserID != outcome.WinnerUserID {
		prev := previousLeader.UserID
		previousHighestBidder = &prev
	}

	oldEndedAt := lot.EndedAt
	newEndedAt, wasExtended := resolveSoftCloseEndTime(lot.EndedAt, in.Now, autoExtend, snipingWindow, snipingExtension)

	const updLot = `UPDATE lots SET current_price = $2, ended_at = $3, updated_at = now() WHERE id = $1`
	if _, err := tx.Exec(ctx, updLot, lot.ID, outcome.VisiblePrice, newEndedAt); err != nil {
		return nil, err
	}

	if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
		TenantID:   &in.TenantID,
		Action:     "BID_PLACED",
		EntityType: "LOT",
		EntityID:   &lot.ID,
		Metadata: mustMarshalJSON(map[string]any{
			"lot_id":       lot.ID,
			"auction_id":   lot.AuctionID,
			"amount_minor": outcome.VisiblePrice,
			"currency":     currency,
			"result":       "ACCEPTED",
		}),
	}); err != nil {
		return nil, err
	}
	if wasExtended {
		if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
			TenantID:   &in.TenantID,
			Action:     "LOT_EXTENDED",
			EntityType: "LOT",
			EntityID:   &lot.ID,
			Metadata: mustMarshalJSON(map[string]any{
				"lot_id":       lot.ID,
				"auction_id":   lot.AuctionID,
				"previous_end": *oldEndedAt,
				"new_end":      *newEndedAt,
			}),
		}); err != nil {
			return nil, err
		}
	}

	lot.CurrentPrice = outcome.VisiblePrice
	lot.EndedAt = newEndedAt
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &port.PlaceLotBidTxResult{
		Lot:                   lot,
		Bid:                   acceptedBid,
		AuctionTitle:          auctionTitle,
		PreviousHighestBidder: previousHighestBidder,
		WasExtended:           wasExtended,
		ExtensionEndTime:      newEndedAt,
	}, nil
}

type lotBidLeader struct {
	UserID       uuid.UUID
	EffectiveMax int64
	FirstAt      time.Time
}

type lotBidParticipant struct {
	UserID       uuid.UUID
	EffectiveMax int64
	FirstAt      time.Time
}

type proxyResolutionOutcome struct {
	WinnerUserID uuid.UUID
	WinnerMax    int64
	VisiblePrice int64
}

func fetchLotBidLeaderTx(ctx context.Context, tx pgx.Tx, tenantID, lotID uuid.UUID) (*lotBidLeader, error) {
	const q = `
SELECT user_id,
       MAX(COALESCE(max_bid_minor, amount)) AS effective_max,
       MIN(created_at) AS first_at
FROM lot_bids
WHERE tenant_id = $1 AND lot_id = $2
GROUP BY user_id
ORDER BY effective_max DESC, first_at ASC, user_id ASC
LIMIT 1`
	var out lotBidLeader
	err := tx.QueryRow(ctx, q, tenantID, lotID).Scan(&out.UserID, &out.EffectiveMax, &out.FirstAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func fetchLotBidParticipantsTx(ctx context.Context, tx pgx.Tx, tenantID, lotID uuid.UUID) ([]lotBidParticipant, error) {
	const q = `
SELECT user_id,
       MAX(COALESCE(max_bid_minor, amount)) AS effective_max,
       MIN(created_at) AS first_at
FROM lot_bids
WHERE tenant_id = $1 AND lot_id = $2
GROUP BY user_id
ORDER BY effective_max DESC, first_at ASC, user_id ASC`
	rows, err := tx.Query(ctx, q, tenantID, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]lotBidParticipant, 0, 4)
	for rows.Next() {
		var p lotBidParticipant
		if err := rows.Scan(&p.UserID, &p.EffectiveMax, &p.FirstAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func insertLotBidTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, lotID, userID uuid.UUID,
	amount int64,
	maxBidMinor *int64,
	createdAt time.Time,
) (domain.Bid, error) {
	bidID := uuid.New()
	const q = `INSERT INTO lot_bids (id, tenant_id, lot_id, user_id, amount, max_bid_minor, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	if _, err := tx.Exec(ctx, q, bidID, tenantID, lotID, userID, amount, maxBidMinor, createdAt); err != nil {
		return domain.Bid{}, err
	}
	return domain.Bid{
		ID:        bidID,
		LotID:     &lotID,
		UserID:    userID,
		Amount:    amount,
		MaxBid:    maxBidMinor,
		CreatedAt: createdAt,
	}, nil
}

func resolveProxyOutcome(
	participants []lotBidParticipant,
	incrementMinor int64,
	minimumNextBid int64,
	submittedAmount int64,
) proxyResolutionOutcome {
	if len(participants) == 0 {
		return proxyResolutionOutcome{}
	}
	highest := participants[0]
	visible := submittedAmount
	if len(participants) > 1 {
		second := participants[1]
		visible = second.EffectiveMax + incrementMinor
		if visible > highest.EffectiveMax {
			visible = highest.EffectiveMax
		}
		if visible < submittedAmount {
			visible = submittedAmount
		}
	}
	if visible < minimumNextBid {
		visible = minimumNextBid
	}
	return proxyResolutionOutcome{
		WinnerUserID: highest.UserID,
		WinnerMax:    highest.EffectiveMax,
		VisiblePrice: visible,
	}
}

func resolveSoftCloseEndTime(
	currentEnd *time.Time,
	now time.Time,
	enabled bool,
	windowSeconds int,
	extensionSeconds int,
) (*time.Time, bool) {
	if !enabled || currentEnd == nil || windowSeconds <= 0 || extensionSeconds <= 0 {
		return currentEnd, false
	}
	windowStart := currentEnd.Add(-time.Duration(windowSeconds) * time.Second)
	if now.Before(windowStart) {
		return currentEnd, false
	}
	candidate := now.Add(time.Duration(extensionSeconds) * time.Second)
	if !candidate.After(*currentEnd) {
		return currentEnd, false
	}
	return &candidate, true
}

func (r *LotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	const qAuctionEndedCandidates = `
SELECT l.id
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE a.status = 'ENDED'
  AND l.status NOT IN ('ENDED','CANCELLED')`
	endedRows, err := r.DB.Query(ctx, qAuctionEndedCandidates)
	if err != nil {
		return err
	}
	endedCandidates := make([]uuid.UUID, 0)
	for endedRows.Next() {
		var lotID uuid.UUID
		if err := endedRows.Scan(&lotID); err != nil {
			endedRows.Close()
			return err
		}
		endedCandidates = append(endedCandidates, lotID)
	}
	if err := endedRows.Err(); err != nil {
		endedRows.Close()
		return err
	}
	endedRows.Close()
	for _, lotID := range endedCandidates {
		tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		if _, err := r.finalizeLotSettlementTx(ctx, tx, lotID, now, "AUCTION_ENDED_CLOSE"); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

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
		if _, err := r.finalizeLotSettlementTx(ctx, tx, lotID, now, "AUTO_CLOSE_TIMED"); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	const qSeqAuctions = `
SELECT a.id, a.tenant_id, a.lot_duration_seconds
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
		tenantID    uuid.UUID
		lotDuration int
	}
	seqAuctions := make([]seqAuction, 0)
	for seqRows.Next() {
		var a seqAuction
		if err := seqRows.Scan(&a.auctionID, &a.tenantID, &a.lotDuration); err != nil {
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

		if _, err := r.finalizeLotSettlementTx(ctx, tx, liveLotID, now, "AUTO_CLOSE_SEQUENTIAL"); err != nil {
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
			aid := a.auctionID
			tid := a.tenantID
			if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
				TenantID:   &tid,
				Action:     "LOT_STARTED",
				EntityType: "LOT",
				EntityID:   &nextID,
				Metadata: mustMarshalJSON(map[string]any{
					"auction_id": aid,
					"reason":     "AUTO_PROGRESS",
					"started_at": now,
					"ended_at":   nextEndedAt,
				}),
			}); err != nil {
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
	reason string,
) (*port.LotSettlementResult, error) {
	var currentStatus domain.LotStatus
	var auctionID, tenantID uuid.UUID
	var auctionTitle, lotTitle string
	if err := tx.QueryRow(ctx, `
SELECT l.status, l.auction_id, a.tenant_id, a.title, l.title
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1
FOR UPDATE`, lotID).Scan(&currentStatus, &auctionID, &tenantID, &auctionTitle, &lotTitle); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("lot not found")
		}
		return nil, err
	}
	if currentStatus == domain.LotEnded || currentStatus == domain.LotCancelled {
		return nil, nil
	}

	var winnerUserID *uuid.UUID
	var hammerPrice *int64
	var winnerID uuid.UUID
	var winningAmount int64
	err := tx.QueryRow(ctx, `
SELECT user_id, amount
FROM lot_bids
WHERE lot_id = $1 AND tenant_id = $2
ORDER BY amount DESC, created_at ASC
LIMIT 1`, lotID, tenantID).Scan(&winnerID, &winningAmount)
	if err == nil {
		winnerUserID = &winnerID
		hammerPrice = &winningAmount
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var totalDue *int64
	var snapshot []byte
	if hammerPrice != nil {
		fees, err := r.resolveFeeComponentsTx(ctx, tx, lotID, auctionID, tenantID)
		if err != nil {
			return nil, err
		}
		taxes, err := r.resolveTaxRulesTx(ctx, tx, lotID, auctionID, tenantID)
		if err != nil {
			return nil, err
		}
		computation := calculateFinancials(*hammerPrice, fees, taxes)
		computedTotal := computation.TotalDue
		totalDue = &computedTotal
		payload := domain.LotFinancialSnapshot{
			HammerPrice: *hammerPrice,
			Fees:        computation.Lines,
			TotalDue:    computation.TotalDue,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		snapshot = b
		if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
			TenantID:   &tenantID,
			Action:     "FINANCIAL_SNAPSHOT_CREATED",
			EntityType: "LOT",
			EntityID:   &lotID,
			Metadata: mustMarshalJSON(map[string]any{
				"winning_bid_id":     nil,
				"winner_user_id":     winnerUserID,
				"hammer_price_minor": hammerPrice,
				"total_due_minor":    totalDue,
			}),
		}); err != nil {
			return nil, err
		}

		if winnerUserID != nil {
			if err := r.createInvoiceForSettlementTx(ctx, tx, lotID, auctionID, tenantID, *winnerUserID, computation.TotalDue, snapshot, now); err != nil {
				return nil, err
			}
		}
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
		return nil, err
	}
	if err := r.appendAuditTx(ctx, tx, port.AuditEntry{
		TenantID:   &tenantID,
		Action:     "LOT_ENDED",
		EntityType: "LOT",
		EntityID:   &lotID,
		Metadata: mustMarshalJSON(map[string]any{
			"reason":             reason,
			"winner_user_id":     winnerUserID,
			"hammer_price_minor": hammerPrice,
			"total_due_minor":    totalDue,
		}),
	}); err != nil {
		return nil, err
	}
	return &port.LotSettlementResult{
		LotID:        lotID,
		AuctionID:    auctionID,
		AuctionTitle: auctionTitle,
		LotTitle:     lotTitle,
		WinnerUserID: winnerUserID,
		HammerPrice:  hammerPrice,
	}, nil
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

type rawTaxRule struct {
	ID          uuid.UUID
	Name        string
	RatePercent float64
	CreatedAt   time.Time
}

type financialComputation struct {
	Lines       []domain.FinancialFee
	FeeSubtotal int64
	TaxSubtotal int64
	TotalDue    int64
	TaxableBase int64
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

func (r *LotRepo) resolveTaxRulesTx(ctx context.Context, tx pgx.Tx, lotID, auctionID, tenantID uuid.UUID) ([]rawTaxRule, error) {
	byScope := func(scopeType string, scopeID uuid.UUID) ([]rawTaxRule, error) {
		const q = `
SELECT tr.id, tr.name, tr.rate_percent::double precision, tr.created_at
FROM tax_rule_assignments ta
JOIN tax_rules tr ON tr.id = ta.tax_rule_id
WHERE ta.scope_type = $1
  AND ta.scope_id = $2
  AND tr.active = true
ORDER BY tr.created_at ASC, tr.id ASC`
		rows, err := tx.Query(ctx, q, scopeType, scopeID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := make([]rawTaxRule, 0)
		for rows.Next() {
			var rule rawTaxRule
			if err := rows.Scan(&rule.ID, &rule.Name, &rule.RatePercent, &rule.CreatedAt); err != nil {
				return nil, err
			}
			out = append(out, rule)
		}
		return out, rows.Err()
	}

	if rules, err := byScope("LOT", lotID); err != nil {
		return nil, err
	} else if len(rules) > 0 {
		return rules, nil
	}
	if rules, err := byScope("AUCTION", auctionID); err != nil {
		return nil, err
	} else if len(rules) > 0 {
		return rules, nil
	}
	return byScope("TENANT", tenantID)
}

func calculateFinancials(hammer int64, feeRules []rawFeeRule, taxRules []rawTaxRule) financialComputation {
	if hammer < 0 {
		hammer = 0
	}
	percentageRules := make([]rawFeeRule, 0)
	fixedRules := make([]rawFeeRule, 0)
	for _, rule := range feeRules {
		if rule.Value < 0 || rule.Name == "" {
			continue
		}
		if rule.Type == string(domain.FeeTypeFixed) {
			fixedRules = append(fixedRules, rule)
		} else if rule.Type == string(domain.FeeTypePercentage) {
			percentageRules = append(percentageRules, rule)
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
	lines := make([]domain.FinancialFee, 0, len(feeRules)+len(taxRules))
	totalFees := int64(0)
	taxableBase := int64(0)
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
		if rule.Taxable {
			taxableBase += amount
		}
		subtotalBase += float64(amount)
		if rule.AppliesTo == "HAMMER" || rule.AppliesTo == "PREMIUM" {
			premiumBase += float64(amount)
		}
		lines = append(lines, domain.FinancialFee{
			Name:      rule.Name,
			Type:      domain.FeeTypePercentage,
			Value:     rule.Value,
			Amount:    amount,
			AppliesTo: rule.AppliesTo,
			Taxable:   rule.Taxable,
			Category:  "FEE",
		})
	}
	for _, rule := range fixedRules {
		amount := int64(math.Round(rule.Value))
		if amount < 0 {
			amount = 0
		}
		totalFees += amount
		if rule.Taxable {
			taxableBase += amount
		}
		subtotalBase += float64(amount)
		lines = append(lines, domain.FinancialFee{
			Name:      rule.Name,
			Type:      domain.FeeTypeFixed,
			Value:     rule.Value,
			Amount:    amount,
			AppliesTo: rule.AppliesTo,
			Taxable:   rule.Taxable,
			Category:  "FEE",
		})
	}

	totalTax := int64(0)
	for _, rule := range taxRules {
		if rule.Name == "" || rule.RatePercent < 0 {
			continue
		}
		amount := int64(math.Round(float64(taxableBase) * rule.RatePercent / 100))
		if amount < 0 {
			amount = 0
		}
		totalTax += amount
		lines = append(lines, domain.FinancialFee{
			Name:     rule.Name,
			Type:     domain.FeeTypePercentage,
			Value:    rule.RatePercent,
			Amount:   amount,
			Category: "TAX",
		})
	}

	return financialComputation{
		Lines:       lines,
		FeeSubtotal: totalFees,
		TaxSubtotal: totalTax,
		TotalDue:    hammer + totalFees + totalTax,
		TaxableBase: taxableBase,
	}
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

func (r *LotRepo) createInvoiceForSettlementTx(
	ctx context.Context,
	tx pgx.Tx,
	lotID uuid.UUID,
	auctionID uuid.UUID,
	tenantID uuid.UUID,
	winnerUserID uuid.UUID,
	totalAmount int64,
	snapshot []byte,
	now time.Time,
) error {
	var existing uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM invoices WHERE lot_id = $1 AND tenant_id = $2`, lotID, tenantID).Scan(&existing); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	var prefix string
	var counter int64
	if err := tx.QueryRow(ctx, `
UPDATE tenants
SET invoice_counter = invoice_counter + 1
WHERE id = $1
RETURNING invoice_prefix, invoice_counter
`, tenantID).Scan(&prefix, &counter); err != nil {
		return err
	}
	if prefix == "" {
		prefix = "INV"
	}
	invoiceNumber := fmt.Sprintf("%s-%04d-%06d", prefix, now.Year(), counter)
	dueDate := now.AddDate(0, 0, 7)
	var currency string
	if err := tx.QueryRow(ctx, `
SELECT a.currency
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE l.id = $1
  AND a.tenant_id = $2
`, lotID, tenantID).Scan(&currency); err != nil {
		return err
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

	hammerMinor, feesMinor, taxMinor, totalDueMinor := snapshotTotals(snapshotWithCurrency, totalAmount)
	const q = `
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
)`
	_, err := tx.Exec(ctx, q,
		uuid.New(), lotID, auctionID, tenantID, winnerUserID,
		invoiceNumber, now, dueDate, currency,
		totalDueMinor, hammerMinor, feesMinor, taxMinor, totalDueMinor,
		snapshotWithCurrency,
	)
	if err != nil {
		return err
	}
	return r.appendAuditTx(ctx, tx, port.AuditEntry{
		TenantID:   &tenantID,
		Action:     "INVOICE_ISSUED",
		EntityType: "LOT",
		EntityID:   &lotID,
		Metadata: mustMarshalJSON(map[string]any{
			"winner_user_id":     winnerUserID,
			"invoice_number":     invoiceNumber,
			"total_amount_minor": totalDueMinor,
			"currency":           currency,
			"issued_at":          now,
			"reason":             "AUTO_ON_LOT_END",
		}),
	})
}

func (r *LotRepo) appendAuditTx(ctx context.Context, tx pgx.Tx, entry port.AuditEntry) error {
	if r.Audit == nil {
		return nil
	}
	base := middleware.AuditContextFromContext(ctx)
	if entry.TenantID == nil {
		entry.TenantID = base.TenantID
	}
	if entry.ActorUserID == nil {
		entry.ActorUserID = base.ActorUserID
	}
	if entry.ActorRole == "" {
		entry.ActorRole = base.ActorRole
	}
	if entry.RequestID == "" {
		entry.RequestID = base.RequestID
	}
	if entry.IP == "" {
		entry.IP = base.IP
	}
	if entry.UserAgent == "" {
		entry.UserAgent = base.UserAgent
	}
	if len(entry.Metadata) == 0 {
		entry.Metadata = []byte(`{}`)
	}
	return r.Audit.AppendTx(ctx, tx, entry)
}

func mustMarshalJSON(v map[string]any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

func (r *LotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	if limit <= 0 {
		limit = 100
	}
	const q = `
SELECT b.id, b.user_id, b.amount, b.max_bid_minor, b.created_at
FROM lot_bids b
WHERE b.tenant_id = $1 AND b.lot_id = $2
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
		if err := rows.Scan(&b.ID, &b.UserID, &b.Amount, &b.MaxBid, &b.CreatedAt); err != nil {
			return nil, err
		}
		id := lotID
		b.LotID = &id
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *LotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	if limit <= 0 {
		limit = 100
	}
	const q = `
SELECT l.id, l.auction_id, l.lot_number, l.order_index,
       l.title, l.description,
       l.starting_price, l.current_price, l.reserve_price, l.min_increment,
       l.metadata, l.status,
       l.hammer_price, l.total_due, l.winner_user_id, l.financial_snapshot,
       l.started_at, l.ended_at, l.created_at, l.updated_at
FROM lots l
JOIN auctions a ON a.id = l.auction_id
WHERE a.tenant_id = $1
  AND l.winner_user_id = $2
  AND l.status = 'ENDED'
ORDER BY l.ended_at DESC NULLS LAST, l.created_at DESC
LIMIT $3`
	rows, err := r.DB.Query(ctx, q, tenantID, userID, limit)
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
