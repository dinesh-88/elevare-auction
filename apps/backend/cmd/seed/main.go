package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/adapter/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultDBURL = "postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable"
)

var (
	tenantID      = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	adminID       = uuid.MustParse("21111111-1111-1111-1111-111111111111")
	bidderOneID   = uuid.MustParse("31111111-1111-1111-1111-111111111111")
	bidderTwoID   = uuid.MustParse("41111111-1111-1111-1111-111111111111")
	auctionEnded  = uuid.MustParse("51111111-1111-1111-1111-111111111111")
	auctionLive   = uuid.MustParse("61111111-1111-1111-1111-111111111111")
	auctionFuture = uuid.MustParse("71111111-1111-1111-1111-111111111111")
	lotEndedSold  = uuid.MustParse("81111111-1111-1111-1111-111111111111")
	lotEndedNoRes = uuid.MustParse("91111111-1111-1111-1111-111111111111")
	lotLiveOne    = uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	lotLiveTwo    = uuid.MustParse("b1111111-1111-1111-1111-111111111111")
	lotFutureOne  = uuid.MustParse("c1111111-1111-1111-1111-111111111111")
	feePremium    = uuid.MustParse("d1111111-1111-1111-1111-111111111111")
	feeVAT        = uuid.MustParse("e1111111-1111-1111-1111-111111111111")
	invoiceOneID  = uuid.MustParse("f1111111-1111-1111-1111-111111111111")
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	hasher := &security.Argon2Hasher{}
	adminHash, err := hasher.Hash("admin123")
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}
	bidderHash, err := hasher.Hash("bidder123")
	if err != nil {
		log.Fatalf("failed to hash bidder password: %v", err)
	}

	if err := seedTenant(ctx, tx, now); err != nil {
		log.Fatalf("tenant seed failed: %v", err)
	}
	if err := seedUsers(ctx, tx, now, adminHash, bidderHash); err != nil {
		log.Fatalf("user seed failed: %v", err)
	}
	if err := seedAuctions(ctx, tx, now); err != nil {
		log.Fatalf("auction seed failed: %v", err)
	}
	if err := seedLots(ctx, tx, now); err != nil {
		log.Fatalf("lot seed failed: %v", err)
	}
	if err := seedBids(ctx, tx, now); err != nil {
		log.Fatalf("bid seed failed: %v", err)
	}
	if err := seedFees(ctx, tx, now); err != nil {
		log.Fatalf("fee seed failed: %v", err)
	}
	if err := seedInvoices(ctx, tx, now); err != nil {
		log.Fatalf("invoice seed failed: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit seed: %v", err)
	}

	log.Println("seed complete")
	log.Println("tenant host: demo.localhost")
	log.Println("admin: admin@demo.elevare.test / admin123")
	log.Println("bidder: bidder@demo.elevare.test / bidder123")
	log.Println("bidder2: bidder2@demo.elevare.test / bidder123")
}

func seedTenant(ctx context.Context, tx pgx.Tx, now time.Time) error {
	branding := map[string]any{
		"logo_url":        "https://picsum.photos/seed/elevare-logo/120/120",
		"primary_color":   "#1E2A3A",
		"secondary_color": "#B89A5D",
		"hero_image":      "https://picsum.photos/seed/elevare-hero/1400/700",
		"footer_text":     "Elevare Demo Collection",
	}
	brandingJSON, _ := json.Marshal(branding)
	_, err := tx.Exec(ctx, `
INSERT INTO tenants (
  id, name, subdomain, status, country, estimated_auction_volume,
  branding, require_admin_approval, require_kyc, default_currency, invoice_prefix, invoice_counter, created_at
)
VALUES ($1,$2,$3,'ACTIVE',$4,$5,$6,false,false,'GBP','INV',1,$7)
ON CONFLICT (subdomain) DO UPDATE
SET name = EXCLUDED.name,
    status = 'ACTIVE',
    country = EXCLUDED.country,
    estimated_auction_volume = EXCLUDED.estimated_auction_volume,
    branding = EXCLUDED.branding,
    require_admin_approval = false,
    require_kyc = false,
    default_currency = EXCLUDED.default_currency,
    invoice_prefix = 'INV',
    invoice_counter = GREATEST(tenants.invoice_counter, 1)
`, tenantID, "Elevare Demo House", "demo", "US", 200, brandingJSON, now.Add(-72*time.Hour))
	return err
}

func seedUsers(ctx context.Context, tx pgx.Tx, now time.Time, adminHash, bidderHash string) error {
	queries := []struct {
		id    uuid.UUID
		email string
		role  string
		hash  string
	}{
		{adminID, "admin@demo.elevare.test", "ADMIN", adminHash},
		{bidderOneID, "bidder@demo.elevare.test", "BIDDER", bidderHash},
		{bidderTwoID, "bidder2@demo.elevare.test", "BIDDER", bidderHash},
	}
	for _, q := range queries {
		_, err := tx.Exec(ctx, `
INSERT INTO users (
  id, tenant_id, email, password_hash, role, email_verified, approval_status, kyc_status, created_at
)
VALUES ($1,$2,$3,$4,$5,true,'APPROVED','NONE',$6)
ON CONFLICT (tenant_id, email) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    email_verified = true,
    approval_status = 'APPROVED',
    kyc_status = 'NONE'
`, q.id, tenantID, q.email, q.hash, q.role, now.Add(-48*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedAuctions(ctx context.Context, tx pgx.Tx, now time.Time) error {
	endedMeta, _ := json.Marshal(map[string]any{
		"image":  "https://picsum.photos/seed/auction-ended/1200/800",
		"images": []string{"https://picsum.photos/seed/auction-ended/1200/800", "https://picsum.photos/seed/auction-ended-2/1200/800"},
	})
	liveMeta, _ := json.Marshal(map[string]any{
		"image":  "https://picsum.photos/seed/auction-live/1200/800",
		"images": []string{"https://picsum.photos/seed/auction-live/1200/800", "https://picsum.photos/seed/auction-live-2/1200/800"},
	})
	futureMeta, _ := json.Marshal(map[string]any{
		"image":  "https://picsum.photos/seed/auction-future/1200/800",
		"images": []string{"https://picsum.photos/seed/auction-future/1200/800"},
	})

	items := []struct {
		id          uuid.UUID
		currency    string
		title       string
		description string
		metadata    []byte
		startTime   time.Time
		endTime     time.Time
		status      string
		mode        string
	}{
		{auctionEnded, "GBP", "Old Masters Evening Sale", "A completed premium sale with invoice records.", endedMeta, now.Add(-7 * 24 * time.Hour), now.Add(-24 * time.Hour), "ENDED", "TIMED"},
		{auctionLive, "USD", "Modern & Contemporary Live", "Live lots open for bidding now.", liveMeta, now.Add(-2 * time.Hour), now.Add(8 * time.Hour), "LIVE", "TIMED"},
		{auctionFuture, "EUR", "Sculpture Highlights", "Scheduled sale with upcoming lots.", futureMeta, now.Add(24 * time.Hour), now.Add(48 * time.Hour), "SCHEDULED", "TIMED"},
	}
	for _, a := range items {
		_, err := tx.Exec(ctx, `
INSERT INTO auctions (
  id, tenant_id, currency, title, description, metadata,
  start_time, end_time, status, allow_manual_start, allow_manual_pause, allow_manual_end,
  auto_extend_enabled, auction_mode, auto_progress_enabled, lot_duration_seconds,
  sniping_window_seconds, sniping_extension_seconds, bid_visibility, created_at
)
VALUES (
  $1,$2,$3,$4,$5,$6,
  $7,$8,$9,true,true,true,
  true,$10,false,0,
  120,90,'PUBLIC',$11
)
ON CONFLICT (id) DO UPDATE
SET currency = EXCLUDED.currency,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    metadata = EXCLUDED.metadata,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    status = EXCLUDED.status,
    auction_mode = EXCLUDED.auction_mode
`, a.id, tenantID, a.currency, a.title, a.description, a.metadata, a.startTime, a.endTime, a.status, a.mode, now.Add(-72*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedLots(ctx context.Context, tx pgx.Tx, now time.Time) error {
	lot1Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-1a/1200/900",
			"https://picsum.photos/seed/lot-1b/1200/900",
		},
		"details": []map[string]string{
			{"label": "Medium", "value": "Oil on canvas"},
			{"label": "Dimensions", "value": "65 x 80 cm"},
			{"label": "Date", "value": "c. 1905"},
			{"label": "Provenance", "value": "Private Collection"},
		},
	})
	lot2Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-2a/1200/900",
		},
	})
	lot3Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-3a/1200/900",
			"https://picsum.photos/seed/lot-3b/1200/900",
		},
	})
	lot4Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-4a/1200/900",
		},
	})
	lot5Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-5a/1200/900",
		},
	})
	snapshot1, _ := json.Marshal(map[string]any{
		"currency":     "GBP",
		"hammer_price": 12000,
		"fees": []map[string]any{
			{"name": "Buyer's Premium", "type": "PERCENTAGE", "value": 15, "amount": 1800},
			{"name": "VAT", "type": "PERCENTAGE", "value": 10, "amount": 1380},
		},
		"total_due": 15180,
	})

	type lotRow struct {
		id            uuid.UUID
		auctionID     uuid.UUID
		lotNumber     string
		orderIndex    int
		title         string
		description   string
		startingPrice int64
		currentPrice  int64
		reservePrice  *int64
		minIncrement  int64
		metadata      []byte
		status        string
		startedAt     *time.Time
		endedAt       *time.Time
		hammer        *int64
		totalDue      *int64
		winner        *uuid.UUID
		snapshot      []byte
	}

	r11000 := int64(11000)
	r12000 := int64(12000)
	r6500 := int64(6500)
	r7000 := int64(7000)
	h12000 := int64(12000)
	t15180 := int64(15180)
	startEnded := now.Add(-6 * 24 * time.Hour)
	endEnded := now.Add(-24 * time.Hour)
	startLive := now.Add(-30 * time.Minute)
	endLive := now.Add(2 * time.Hour)

	rows := []lotRow{
		{lotEndedSold, auctionEnded, "1", 1, "Sunlit Coastline", "Signed landscape, framed.", 10000, 12000, &r11000, 500, lot1Meta, "ENDED", &startEnded, &endEnded, &h12000, &t15180, &bidderOneID, snapshot1},
		{lotEndedNoRes, auctionEnded, "2", 2, "Portrait Study", "Unsigned portrait study.", 9000, 9000, &r12000, 250, lot2Meta, "ENDED", &startEnded, &endEnded, nil, nil, nil, nil},
		{lotLiveOne, auctionLive, "3", 1, "Abstract Composition", "Bright abstract work.", 5000, 6200, &r6500, 200, lot3Meta, "LIVE", &startLive, &endLive, nil, nil, nil, nil},
		{lotLiveTwo, auctionLive, "4", 2, "Bronze Figure", "Mid-century bronze figure.", 6000, 6800, &r7000, 200, lot4Meta, "LIVE", &startLive, &endLive, nil, nil, nil, nil},
		{lotFutureOne, auctionFuture, "5", 1, "Marble Bust", "Neoclassical marble bust.", 7000, 7000, nil, 300, lot5Meta, "PENDING", nil, nil, nil, nil, nil, nil},
	}

	for _, l := range rows {
		_, err := tx.Exec(ctx, `
INSERT INTO lots (
  id, auction_id, lot_number, order_index, title, description,
  starting_price, current_price, reserve_price, min_increment, metadata, status,
  started_at, ended_at, hammer_price, total_due, winner_user_id, financial_snapshot, created_at, updated_at
)
VALUES (
  $1,$2,$3,$4,$5,$6,
  $7,$8,$9,$10,$11,$12,
  $13,$14,$15,$16,$17,$18,$19,$20
)
ON CONFLICT (id) DO UPDATE
SET auction_id = EXCLUDED.auction_id,
    lot_number = EXCLUDED.lot_number,
    order_index = EXCLUDED.order_index,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    starting_price = EXCLUDED.starting_price,
    current_price = EXCLUDED.current_price,
    reserve_price = EXCLUDED.reserve_price,
    min_increment = EXCLUDED.min_increment,
    metadata = EXCLUDED.metadata,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    ended_at = EXCLUDED.ended_at,
    hammer_price = EXCLUDED.hammer_price,
    total_due = EXCLUDED.total_due,
    winner_user_id = EXCLUDED.winner_user_id,
    financial_snapshot = EXCLUDED.financial_snapshot,
    updated_at = EXCLUDED.updated_at
`, l.id, l.auctionID, l.lotNumber, l.orderIndex, l.title, l.description,
			l.startingPrice, l.currentPrice, l.reservePrice, l.minIncrement, l.metadata, l.status,
			l.startedAt, l.endedAt, l.hammer, l.totalDue, l.winner, l.snapshot, now.Add(-48*time.Hour), now)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedBids(ctx context.Context, tx pgx.Tx, now time.Time) error {
	bids := []struct {
		id      uuid.UUID
		lotID   uuid.UUID
		userID  uuid.UUID
		amount  int64
		created time.Time
	}{
		{uuid.MustParse("11111111-aaaa-1111-1111-111111111111"), lotEndedSold, bidderTwoID, 11000, now.Add(-26 * time.Hour)},
		{uuid.MustParse("22222222-aaaa-1111-1111-111111111111"), lotEndedSold, bidderOneID, 12000, now.Add(-25 * time.Hour)},
		{uuid.MustParse("33333333-aaaa-1111-1111-111111111111"), lotLiveOne, bidderTwoID, 5600, now.Add(-20 * time.Minute)},
		{uuid.MustParse("44444444-aaaa-1111-1111-111111111111"), lotLiveOne, bidderOneID, 6200, now.Add(-10 * time.Minute)},
		{uuid.MustParse("55555555-aaaa-1111-1111-111111111111"), lotLiveTwo, bidderTwoID, 6800, now.Add(-8 * time.Minute)},
	}
	for _, b := range bids {
		_, err := tx.Exec(ctx, `
INSERT INTO lot_bids (id, lot_id, user_id, amount, created_at)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (id) DO NOTHING
`, b.id, b.lotID, b.userID, b.amount, b.created)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedFees(ctx context.Context, tx pgx.Tx, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO fee_components (id, tenant_id, name, calculation_type, value, applies_to, taxable, active, created_at)
VALUES
  ($1,$2,'Buyer''s Premium','PERCENTAGE',15,'HAMMER',false,true,$3),
  ($4,$2,'VAT','PERCENTAGE',10,'SUBTOTAL',true,true,$3)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    calculation_type = EXCLUDED.calculation_type,
    value = EXCLUDED.value,
    applies_to = EXCLUDED.applies_to,
    taxable = EXCLUDED.taxable,
    active = EXCLUDED.active
`, feePremium, tenantID, now.Add(-72*time.Hour), feeVAT)
	if err != nil {
		return err
	}

	assignments := []struct {
		id           uuid.UUID
		scopeType    string
		scopeID      uuid.UUID
		feeComponent uuid.UUID
	}{
		{uuid.MustParse("66666666-aaaa-1111-1111-111111111111"), "TENANT", tenantID, feePremium},
		{uuid.MustParse("77777777-aaaa-1111-1111-111111111111"), "TENANT", tenantID, feeVAT},
	}
	for _, a := range assignments {
		_, err := tx.Exec(ctx, `
INSERT INTO fee_assignments (id, scope_type, scope_id, fee_component_id, created_at)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (scope_type, scope_id, fee_component_id) DO NOTHING
`, a.id, a.scopeType, a.scopeID, a.feeComponent, now.Add(-48*time.Hour))
		if err != nil {
			return err
		}
	}
	return nil
}

func seedInvoices(ctx context.Context, tx pgx.Tx, now time.Time) error {
	snapshot, _ := json.Marshal(map[string]any{
		"currency":     "GBP",
		"hammer_price": 12000,
		"fees": []map[string]any{
			{"name": "Buyer's Premium", "type": "PERCENTAGE", "value": 15, "amount": 1800},
			{"name": "VAT", "type": "PERCENTAGE", "value": 10, "amount": 1380},
		},
		"total_due": 15180,
	})
	issued := now.Add(-20 * time.Hour)
	due := now.Add(7 * 24 * time.Hour)
	_, err := tx.Exec(ctx, `
INSERT INTO invoices (
  id, lot_id, tenant_id, winner_user_id, invoice_number, issued_at, due_date,
  currency, total_amount, snapshot, status, created_at
)
VALUES ($1,$2,$3,$4,'INV-2026-000001',$5,$6,'GBP',15180,$7,'ISSUED',$8)
ON CONFLICT (lot_id) DO UPDATE
SET winner_user_id = EXCLUDED.winner_user_id,
    invoice_number = EXCLUDED.invoice_number,
    issued_at = EXCLUDED.issued_at,
    due_date = EXCLUDED.due_date,
    currency = EXCLUDED.currency,
    total_amount = EXCLUDED.total_amount,
    snapshot = EXCLUDED.snapshot,
    status = EXCLUDED.status
`, invoiceOneID, lotEndedSold, tenantID, bidderOneID, issued, due, snapshot, issued)
	return err
}
