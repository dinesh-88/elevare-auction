package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuctionRepo struct {
	DB    *pgxpool.Pool
	Audit port.AuditRepository
}

func (r *AuctionRepo) Create(ctx context.Context, auction *domain.Auction) error {
	const q = `
INSERT INTO auctions (
  id, tenant_id, currency, title, description, metadata,
  start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
  auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds, sniping_window_seconds, sniping_extension_seconds, bid_visibility, require_bidder_approval,
  wizard_step, auction_status, published_at, created_at
) VALUES (
  $1,$2,$3,$4,$5,$6,
  $7,$8,$9,$10,$11,$12,
  $13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,now()
)`
	_, err := r.DB.Exec(ctx, q,
		auction.ID, auction.TenantID, auction.Currency, auction.Title, auction.Description, auction.Metadata,
		auction.StartTime, auction.EndTime, auction.Status,
		auction.AllowManualStart, auction.AllowManualPause, auction.AllowManualEnd,
		auction.AutoExtendEnabled, auction.AuctionMode, auction.AutoProgressEnabled, auction.LotDurationSeconds,
		auction.SnipingWindowSeconds, auction.SnipingExtensionSeconds, auction.BidVisibility, auction.RequireBidderApproval,
		auction.WizardStep, auction.AuctionStatus, auction.PublishedAt,
	)
	return err
}

func scanAuction(row pgx.Row) (*domain.Auction, error) {
	var a domain.Auction
	err := row.Scan(
		&a.ID, &a.TenantID, &a.Currency, &a.Title, &a.Description, &a.Metadata,
		&a.StartingPrice, &a.CurrentPrice, &a.ReservePrice, &a.MinIncrement,
		&a.StartTime, &a.EndTime, &a.Status,
		&a.AllowManualStart, &a.AllowManualPause, &a.AllowManualEnd,
		&a.AutoExtendEnabled, &a.AuctionMode, &a.AutoProgressEnabled, &a.LotDurationSeconds,
		&a.SnipingWindowSeconds, &a.SnipingExtensionSeconds, &a.BidVisibility, &a.RequireBidderApproval, &a.WizardStep, &a.AuctionStatus, &a.PublishedAt, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func summaryLotsLateral() string {
	return `
LEFT JOIN LATERAL (
  SELECT MIN(l.starting_price) AS min_starting_price,
         MAX(l.current_price) AS max_current_price,
         MAX(l.reserve_price) AS max_reserve_price,
         MIN(l.min_increment) AS min_increment
  FROM lots l
  WHERE l.auction_id = a.id
) la ON true`
}

func (r *AuctionRepo) ListPublic(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Auction, error) {
	q := `
SELECT a.id, a.tenant_id, a.currency, a.title, a.description, a.metadata,
       COALESCE(la.min_starting_price, 0) AS starting_price,
       COALESCE(la.max_current_price, 0) AS current_price,
       la.max_reserve_price AS reserve_price,
       COALESCE(la.min_increment, 1) AS min_increment,
       a.start_time, a.end_time, a.status, a.allow_manual_start, a.allow_manual_pause, a.allow_manual_end,
       a.auto_extend_enabled, a.auction_mode, a.auto_progress_enabled, a.lot_duration_seconds, a.sniping_window_seconds, a.sniping_extension_seconds, a.bid_visibility, a.require_bidder_approval, a.wizard_step, a.auction_status, a.published_at, a.created_at
FROM auctions a
` + summaryLotsLateral() + `
WHERE a.tenant_id = $1 AND a.status IN ('LIVE','SCHEDULED') AND a.end_time > $2
ORDER BY a.start_time ASC`
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
	q := `
SELECT a.id, a.tenant_id, a.currency, a.title, a.description, a.metadata,
       COALESCE(la.min_starting_price, 0) AS starting_price,
       COALESCE(la.max_current_price, 0) AS current_price,
       la.max_reserve_price AS reserve_price,
       COALESCE(la.min_increment, 1) AS min_increment,
       a.start_time, a.end_time, a.status, a.allow_manual_start, a.allow_manual_pause, a.allow_manual_end,
       a.auto_extend_enabled, a.auction_mode, a.auto_progress_enabled, a.lot_duration_seconds, a.sniping_window_seconds, a.sniping_extension_seconds, a.bid_visibility, a.require_bidder_approval, a.wizard_step, a.auction_status, a.published_at, a.created_at
FROM auctions a
` + summaryLotsLateral() + `
WHERE a.tenant_id = $1
ORDER BY a.created_at DESC`
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
	q := `
SELECT a.id, a.tenant_id, a.currency, a.title, a.description, a.metadata,
       COALESCE(la.min_starting_price, 0) AS starting_price,
       COALESCE(la.max_current_price, 0) AS current_price,
       la.max_reserve_price AS reserve_price,
       COALESCE(la.min_increment, 1) AS min_increment,
       a.start_time, a.end_time, a.status, a.allow_manual_start, a.allow_manual_pause, a.allow_manual_end,
       a.auto_extend_enabled, a.auction_mode, a.auto_progress_enabled, a.lot_duration_seconds, a.sniping_window_seconds, a.sniping_extension_seconds, a.bid_visibility, a.require_bidder_approval, a.wizard_step, a.auction_status, a.published_at, a.created_at
FROM auctions a
` + summaryLotsLateral() + `
WHERE a.tenant_id = $1 AND a.id = $2`
	a, err := scanAuction(r.DB.QueryRow(ctx, q, tenantID, auctionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("auction not found")
	}
	return a, err
}

func (r *AuctionRepo) Update(ctx context.Context, auction *domain.Auction) error {
	const q = `
UPDATE auctions
SET currency = $3,
    title = $4,
    description = $5,
    metadata = $6,
    start_time = $7,
    end_time = $8,
    allow_manual_start = $9,
    allow_manual_pause = $10,
    allow_manual_end = $11,
    auto_extend_enabled = $12,
    auction_mode = $13,
    auto_progress_enabled = $14,
    lot_duration_seconds = $15,
    sniping_window_seconds = $16,
    sniping_extension_seconds = $17,
    bid_visibility = $18,
    require_bidder_approval = $19,
    wizard_step = $20,
    auction_status = $21,
    published_at = $22
WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q,
		auction.TenantID, auction.ID,
		auction.Currency, auction.Title, auction.Description,
		auction.Metadata,
		auction.StartTime, auction.EndTime,
		auction.AllowManualStart, auction.AllowManualPause, auction.AllowManualEnd,
		auction.AutoExtendEnabled, auction.AuctionMode, auction.AutoProgressEnabled, auction.LotDurationSeconds,
		auction.SnipingWindowSeconds, auction.SnipingExtensionSeconds, auction.BidVisibility, auction.RequireBidderApproval,
		auction.WizardStep, auction.AuctionStatus, auction.PublishedAt,
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
	publication := ""
	switch status {
	case domain.AuctionEnded:
		publication = string(domain.AuctionPublicationClosed)
	case domain.AuctionCancelled:
		publication = string(domain.AuctionPublicationCancelled)
	}
	q := `UPDATE auctions SET status = $3`
	args := []any{tenantID, auctionID, status}
	if publication != "" {
		q += `, auction_status = $4`
		args = append(args, publication)
	}
	q += ` WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found")
	}
	return nil
}

func (r *AuctionRepo) UpdateRequireBidderApproval(ctx context.Context, tenantID, auctionID uuid.UUID, required bool) error {
	const q = `UPDATE auctions SET require_bidder_approval = $3 WHERE tenant_id = $1 AND id = $2`
	ct, err := r.DB.Exec(ctx, q, tenantID, auctionID, required)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("auction not found")
	}
	return nil
}

func (r *AuctionRepo) RunSchedulerTick(ctx context.Context, now time.Time) error {
	const qStart = `
UPDATE auctions
SET status = 'LIVE'
WHERE status = 'SCHEDULED'
  AND start_time <= $1
  AND end_time > $1
RETURNING id, tenant_id`
	startedRows, err := r.DB.Query(ctx, qStart, now)
	if err != nil {
		return err
	}
	started := make([][2]uuid.UUID, 0)
	for startedRows.Next() {
		var auctionID, tenantID uuid.UUID
		if err := startedRows.Scan(&auctionID, &tenantID); err != nil {
			startedRows.Close()
			return err
		}
		started = append(started, [2]uuid.UUID{auctionID, tenantID})
	}
	if err := startedRows.Err(); err != nil {
		startedRows.Close()
		return err
	}
	startedRows.Close()

	const qEnd = `
UPDATE auctions
SET status = 'ENDED'
WHERE status IN ('LIVE','PAUSED','SCHEDULED')
  AND end_time <= $1
RETURNING id, tenant_id`
	endedRows, err := r.DB.Query(ctx, qEnd, now)
	if err != nil {
		return err
	}
	ended := make([][2]uuid.UUID, 0)
	for endedRows.Next() {
		var auctionID, tenantID uuid.UUID
		if err := endedRows.Scan(&auctionID, &tenantID); err != nil {
			endedRows.Close()
			return err
		}
		ended = append(ended, [2]uuid.UUID{auctionID, tenantID})
	}
	if err := endedRows.Err(); err != nil {
		endedRows.Close()
		return err
	}
	endedRows.Close()

	if r.Audit != nil {
		base := middleware.AuditContextFromContext(ctx)
		for _, item := range started {
			tenantID := item[1]
			auctionID := item[0]
			payload, _ := json.Marshal(map[string]any{"status": domain.AuctionLive, "reason": "SCHEDULED_AUTO_START"})
			_ = r.Audit.Append(ctx, port.AuditEntry{
				TenantID:    &tenantID,
				ActorUserID: base.ActorUserID,
				ActorRole:   "SYSTEM",
				Action:      "AUCTION_STARTED",
				EntityType:  "AUCTION",
				EntityID:    &auctionID,
				RequestID:   base.RequestID,
				IP:          base.IP,
				UserAgent:   base.UserAgent,
				Metadata:    payload,
			})
		}
		for _, item := range ended {
			tenantID := item[1]
			auctionID := item[0]
			payload, _ := json.Marshal(map[string]any{"status": domain.AuctionEnded, "reason": "SCHEDULED_AUTO_END"})
			_ = r.Audit.Append(ctx, port.AuditEntry{
				TenantID:    &tenantID,
				ActorUserID: base.ActorUserID,
				ActorRole:   "SYSTEM",
				Action:      "AUCTION_ENDED",
				EntityType:  "AUCTION",
				EntityID:    &auctionID,
				RequestID:   base.RequestID,
				IP:          base.IP,
				UserAgent:   base.UserAgent,
				Metadata:    payload,
			})
		}
	}
	return nil
}
