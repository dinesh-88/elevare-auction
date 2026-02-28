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
	tenantID         = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantTwoID      = uuid.MustParse("12111111-1111-1111-1111-111111111111")
	adminID          = uuid.MustParse("21111111-1111-1111-1111-111111111111")
	bidderOneID      = uuid.MustParse("31111111-1111-1111-1111-111111111111")
	bidderTwoID      = uuid.MustParse("41111111-1111-1111-1111-111111111111")
	auctionEnded     = uuid.MustParse("51111111-1111-1111-1111-111111111111")
	auctionLive      = uuid.MustParse("61111111-1111-1111-1111-111111111111")
	auctionFuture    = uuid.MustParse("71111111-1111-1111-1111-111111111111")
	auctionTwoEnded  = uuid.MustParse("14111111-1111-1111-1111-111111111111")
	auctionTwoFuture = uuid.MustParse("13111111-1111-1111-1111-111111111111")
	lotEndedSold     = uuid.MustParse("81111111-1111-1111-1111-111111111111")
	lotEndedNoRes    = uuid.MustParse("91111111-1111-1111-1111-111111111111")
	lotLiveOne       = uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	lotLiveTwo       = uuid.MustParse("b1111111-1111-1111-1111-111111111111")
	lotFutureOne     = uuid.MustParse("c1111111-1111-1111-1111-111111111111")
	lotFutureTwo     = uuid.MustParse("15111111-1111-1111-1111-111111111111")
	lotTwoEndedOne   = uuid.MustParse("17111111-1111-1111-1111-111111111111")
	lotTwoFutureOne  = uuid.MustParse("16111111-1111-1111-1111-111111111111")
	feePremium       = uuid.MustParse("d1111111-1111-1111-1111-111111111111")
	taxRuleVAT       = uuid.MustParse("e1111111-1111-1111-1111-111111111111")
	invoiceOneID     = uuid.MustParse("f1111111-1111-1111-1111-111111111111")
	assetLogoID      = uuid.MustParse("f2111111-1111-1111-1111-111111111111")
	assetBannerID    = uuid.MustParse("f3111111-1111-1111-1111-111111111111")
	assetImageID     = uuid.MustParse("f4111111-1111-1111-1111-111111111111")
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}
	subdomain := os.Getenv("DEMO_SUBDOMAIN")
	if subdomain == "" {
		subdomain = "demo"
	}
	customDomain := os.Getenv("DEMO_CUSTOM_DOMAIN")
	if customDomain == "" {
		customDomain = "demo.elevareauctions.com"
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

	if err := seedTenant(ctx, tx, now, subdomain, customDomain); err != nil {
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
	if err := seedAuctionApprovals(ctx, tx, now); err != nil {
		log.Fatalf("auction approval seed failed: %v", err)
	}
	if err := seedFees(ctx, tx, now); err != nil {
		log.Fatalf("fee seed failed: %v", err)
	}
	if err := seedInvoices(ctx, tx, now); err != nil {
		log.Fatalf("invoice seed failed: %v", err)
	}
	if err := seedAssets(ctx, tx, now); err != nil {
		log.Fatalf("asset seed failed: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit seed: %v", err)
	}

	log.Println("seed complete")
	log.Printf("tenant host (prod): %s", customDomain)
	log.Println("api host (prod): api.elevareauctions.com")
	log.Printf("tenant host (local): %s.localhost", subdomain)
	log.Println("admin: admin@elevareauctions.com / admin123")
	log.Println("bidder: bidder@elevareauctions.com / bidder123")
	log.Println("bidder2: bidder2@elevareauctions.com / bidder123")
	log.Printf("proxy scenario lot: %s (LIVE)", lotLiveOne)
	log.Printf("soft-close scenario lot: %s (LIVE, near end)", lotLiveTwo)
}

func seedTenant(ctx context.Context, tx pgx.Tx, now time.Time, subdomain, customDomain string) error {
	type tenantSeed struct {
		id           uuid.UUID
		name         string
		subdomain    string
		customDomain string
		country      string
		volume       int
		branding     map[string]any
		siteBranding map[string]any
		siteContent  map[string]any
		createdAt    time.Time
	}

	items := []tenantSeed{
		{
			id:           tenantID,
			name:         "Elevare Demo House",
			subdomain:    subdomain,
			customDomain: customDomain,
			country:      "US",
			volume:       200,
			branding: map[string]any{
				"logo_url":        "https://picsum.photos/seed/elevare-logo/120/120",
				"primary_color":   "#1E2A3A",
				"secondary_color": "#B89A5D",
				"hero_image":      "https://picsum.photos/seed/elevare-hero/1400/700",
				"footer_text":     "Elevare Demo Collection",
			},
			siteBranding: map[string]any{
				"logo_url":        "https://picsum.photos/seed/elevare-logo/120/120",
				"primary_color":   "#1E2A3A",
				"secondary_color": "#B89A5D",
				"favicon_url":     "https://picsum.photos/seed/elevare-favicon/64/64",
				"font_family":     "Playfair Display",
			},
			siteContent: map[string]any{
				"homepage": map[string]any{
					"hero": map[string]any{
						"headline":         "Elevare Demo House",
						"subheadline":      "Curated collections and timed auctions",
						"banner_image_url": "https://picsum.photos/seed/elevare-hero/1400/700",
						"cta": map[string]any{
							"label": "View auctions",
							"href":  "/auctions",
						},
					},
					"featuredAuction": map[string]any{
						"mode": "AUTO",
					},
					"highlights": []map[string]any{
						{"title": "Institution-grade", "body": "Audit-safe records and immutable settlements.", "icon": "shield"},
						{"title": "Global reach", "body": "Run events for private collectors and galleries.", "icon": "globe"},
						{"title": "Trusted workflow", "body": "Draft, review, and publish with version control.", "icon": "file-check"},
					},
				},
				"footer": map[string]any{
					"text":          "© 2026 Elevare Demo House",
					"contact_email": "info@elevareauctions.com",
					"social_links": map[string]any{
						"instagram": "https://instagram.com/elevare",
						"x":         "https://x.com/elevare",
					},
				},
				"about": map[string]any{
					"markdown": "Elevare Demo House presents catalogued works from private and institutional collections.",
				},
				"contact": map[string]any{
					"address": "21 Regent Street, London",
					"phone":   "+44 20 7000 0000",
					"email":   "info@elevareauctions.com",
				},
			},
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			id:           tenantTwoID,
			name:         "Regency Archive House",
			subdomain:    "regency",
			customDomain: "regency.elevareauctions.com",
			country:      "GB",
			volume:       120,
			branding: map[string]any{
				"logo_url":        "https://picsum.photos/seed/regency-logo/120/120",
				"primary_color":   "#2A2A2A",
				"secondary_color": "#A88C55",
				"hero_image":      "https://picsum.photos/seed/regency-hero/1400/700",
				"footer_text":     "Regency Archive House",
			},
			siteBranding: map[string]any{
				"logo_url":        "https://picsum.photos/seed/regency-logo/120/120",
				"primary_color":   "#2A2A2A",
				"secondary_color": "#A88C55",
				"favicon_url":     "https://picsum.photos/seed/regency-favicon/64/64",
				"font_family":     "Cormorant Garamond",
			},
			siteContent: map[string]any{
				"homepage": map[string]any{
					"hero": map[string]any{
						"headline":         "Regency Archive House",
						"subheadline":      "Specialist catalogues for collections, libraries, and estates",
						"banner_image_url": "https://picsum.photos/seed/regency-hero/1400/700",
					},
				},
				"footer": map[string]any{
					"text":          "© 2026 Regency Archive House",
					"contact_email": "catalogue@regencyarchive.co.uk",
				},
				"about": map[string]any{
					"markdown": "Regency Archive House conducts scholarly online sales for manuscripts, decorative arts, and estate property.",
				},
				"contact": map[string]any{
					"address": "14 St James's Square, London",
					"phone":   "+44 20 7123 4567",
					"email":   "catalogue@regencyarchive.co.uk",
				},
			},
			createdAt: now.Add(-60 * time.Hour),
		},
	}

	for _, item := range items {
		brandingJSON, _ := json.Marshal(item.branding)
		siteBrandingJSON, _ := json.Marshal(item.siteBranding)
		siteContentJSON, _ := json.Marshal(item.siteContent)
		_, err := tx.Exec(ctx, `
INSERT INTO tenants (
  id, name, subdomain, custom_domain, status, country, estimated_auction_volume,
  branding, branding_published, branding_draft, content_published, content_draft,
  content_version, content_schema_version,
  require_admin_approval, require_kyc, default_currency, invoice_prefix, invoice_counter, created_at
)
VALUES ($1,$2,$3,$4,'ACTIVE',$5,$6,$7,$8,$9,$10,$11,1,1,false,false,'GBP','INV',1,$12)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    subdomain = EXCLUDED.subdomain,
    custom_domain = EXCLUDED.custom_domain,
    status = 'ACTIVE',
    country = EXCLUDED.country,
    estimated_auction_volume = EXCLUDED.estimated_auction_volume,
    branding = EXCLUDED.branding,
    branding_published = EXCLUDED.branding_published,
    branding_draft = EXCLUDED.branding_draft,
    content_published = EXCLUDED.content_published,
    content_draft = EXCLUDED.content_draft,
    content_version = GREATEST(tenants.content_version, 1),
    content_schema_version = 1,
    require_admin_approval = false,
    require_kyc = false,
    default_currency = EXCLUDED.default_currency,
    invoice_prefix = 'INV',
    invoice_counter = GREATEST(tenants.invoice_counter, 1)
`, item.id, item.name, item.subdomain, item.customDomain, item.country, item.volume, brandingJSON, siteBrandingJSON, siteBrandingJSON, siteContentJSON, siteContentJSON, item.createdAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedUsers(ctx context.Context, tx pgx.Tx, now time.Time, adminHash, bidderHash string) error {
	queries := []struct {
		id    uuid.UUID
		email string
		role  string
		hash  string
	}{
		{adminID, "admin@elevareauctions.com", "ADMIN", adminHash},
		{bidderOneID, "bidder@elevareauctions.com", "BIDDER", bidderHash},
		{bidderTwoID, "bidder2@elevareauctions.com", "BIDDER", bidderHash},
	}
	for _, q := range queries {
		_, err := tx.Exec(ctx, `
INSERT INTO users (
  id, tenant_id, email, password_hash, role, status, email_verified, approval_status, kyc_status, created_at
)
VALUES ($1,$2,$3,$4,$5,'ACTIVE',true,'APPROVED','NONE',$6)
ON CONFLICT (id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    tenant_id = EXCLUDED.tenant_id,
    email = EXCLUDED.email,
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
		tenantID    uuid.UUID
		currency    string
		title       string
		description string
		metadata    []byte
		startTime   time.Time
		endTime     time.Time
		status      string
		mode        string
	}{
		{auctionEnded, tenantID, "GBP", "Old Masters Evening Sale", "A completed premium sale with invoice records.", endedMeta, now.Add(-7 * 24 * time.Hour), now.Add(-24 * time.Hour), "ENDED", "TIMED"},
		{auctionLive, tenantID, "USD", "Modern & Contemporary Live", "Live lots open for bidding now.", liveMeta, now.Add(-2 * time.Hour), now.Add(8 * time.Hour), "LIVE", "TIMED"},
		{auctionFuture, tenantID, "EUR", "Sculpture Highlights", "Scheduled sale with upcoming lots.", futureMeta, now.Add(24 * time.Hour), now.Add(48 * time.Hour), "SCHEDULED", "TIMED"},
		{auctionTwoFuture, tenantTwoID, "GBP", "Fine Books & Manuscripts", "A forthcoming specialist sale of rare books and manuscripts.", futureMeta, now.Add(36 * time.Hour), now.Add(60 * time.Hour), "SCHEDULED", "TIMED"},
		{auctionTwoEnded, tenantTwoID, "GBP", "Decorative Arts Cabinet", "A completed cabinet sale with published results.", endedMeta, now.Add(-10 * 24 * time.Hour), now.Add(-4 * 24 * time.Hour), "ENDED", "TIMED"},
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
`, a.id, a.tenantID, a.currency, a.title, a.description, a.metadata, a.startTime, a.endTime, a.status, a.mode, now.Add(-72*time.Hour))
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
	lot6Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-6a/1200/900",
		},
	})
	lot7Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-7a/1200/900",
		},
	})
	lot8Meta, _ := json.Marshal(map[string]any{
		"images": []string{
			"https://picsum.photos/seed/lot-8a/1200/900",
		},
	})
	snapshot1, _ := json.Marshal(map[string]any{
		"currency":     "GBP",
		"hammer_price": 12000,
		"fees": []map[string]any{
			{"name": "Buyer's Premium", "type": "PERCENTAGE", "value": 15, "amount": 1800, "taxable": true, "category": "FEE"},
			{"name": "VAT", "type": "PERCENTAGE", "value": 10, "amount": 180, "category": "TAX"},
		},
		"total_due": 13980,
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
	r5000 := int64(5000)
	h12000 := int64(12000)
	h5400 := int64(5400)
	t13980 := int64(13980)
	t5994 := int64(5994)
	startEnded := now.Add(-6 * 24 * time.Hour)
	endEnded := now.Add(-24 * time.Hour)
	startEndedTwo := now.Add(-9 * 24 * time.Hour)
	endEndedTwo := now.Add(-4 * 24 * time.Hour)
	startLive := now.Add(-30 * time.Minute)
	endLive := now.Add(2 * time.Hour)
	endLiveSoftClose := now.Add(75 * time.Second)
	snapshot2, _ := json.Marshal(map[string]any{
		"currency":     "GBP",
		"hammer_price": 5400,
		"fees": []map[string]any{
			{"name": "Buyer's Premium", "type": "PERCENTAGE", "value": 10, "amount": 540, "taxable": true, "category": "FEE"},
			{"name": "VAT", "type": "PERCENTAGE", "value": 10, "amount": 54, "category": "TAX"},
		},
		"total_due": 5994,
	})

	rows := []lotRow{
		{lotEndedSold, auctionEnded, "1", 1, "Sunlit Coastline", "Signed landscape, framed.", 10000, 12000, &r11000, 500, lot1Meta, "ENDED", &startEnded, &endEnded, &h12000, &t13980, &bidderOneID, snapshot1},
		{lotEndedNoRes, auctionEnded, "2", 2, "Portrait Study", "Unsigned portrait study.", 9000, 9000, &r12000, 250, lot2Meta, "ENDED", &startEnded, &endEnded, nil, nil, nil, nil},
		{lotLiveOne, auctionLive, "3", 1, "Abstract Composition", "Bright abstract work.", 5000, 8600, &r6500, 200, lot3Meta, "LIVE", &startLive, &endLive, nil, nil, nil, nil},
		{lotLiveTwo, auctionLive, "4", 2, "Bronze Figure", "Mid-century bronze figure.", 6000, 9200, &r7000, 200, lot4Meta, "LIVE", &startLive, &endLiveSoftClose, nil, nil, nil, nil},
		{lotFutureOne, auctionFuture, "5", 1, "Marble Bust", "Neoclassical marble bust.", 7000, 7000, nil, 300, lot5Meta, "PENDING", nil, nil, nil, nil, nil, nil},
		{lotFutureTwo, auctionFuture, "6", 2, "Giltwood Mirror", "Late Georgian giltwood wall mirror.", 4200, 4200, nil, 200, lot6Meta, "PENDING", nil, nil, nil, nil, nil, nil},
		{lotTwoFutureOne, auctionTwoFuture, "1", 1, "Annotated Travel Journal", "A nineteenth-century journal with hand-coloured plates.", 3800, 3800, nil, 150, lot7Meta, "PENDING", nil, nil, nil, nil, nil, nil},
		{lotTwoEndedOne, auctionTwoEnded, "1", 1, "Mahogany Collectors Cabinet", "A fitted collectors cabinet with brass pulls.", 4800, 5400, &r5000, 150, lot8Meta, "ENDED", &startEndedTwo, &endEndedTwo, &h5400, &t5994, &bidderTwoID, snapshot2},
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
		maxBid  *int64
		created time.Time
	}{
		{uuid.MustParse("11111111-aaaa-1111-1111-111111111111"), lotEndedSold, bidderTwoID, 11000, nil, now.Add(-26 * time.Hour)},
		{uuid.MustParse("22222222-aaaa-1111-1111-111111111111"), lotEndedSold, bidderOneID, 12000, nil, now.Add(-25 * time.Hour)},
		{uuid.MustParse("33333333-aaaa-1111-1111-111111111111"), lotLiveOne, bidderOneID, 6200, int64Ptr(10000), now.Add(-20 * time.Minute)},
		{uuid.MustParse("44444444-aaaa-1111-1111-111111111111"), lotLiveOne, bidderTwoID, 8500, nil, now.Add(-18 * time.Minute)},
		{uuid.MustParse("55555555-aaaa-1111-1111-111111111111"), lotLiveOne, bidderOneID, 8600, int64Ptr(10000), now.Add(-18*time.Minute + time.Millisecond)},
		{uuid.MustParse("66666666-bbbb-2222-2222-222222222222"), lotLiveTwo, bidderOneID, 7000, int64Ptr(10000), now.Add(-6 * time.Minute)},
		{uuid.MustParse("77777777-bbbb-2222-2222-222222222222"), lotLiveTwo, bidderTwoID, 7200, int64Ptr(9000), now.Add(-5 * time.Minute)},
		{uuid.MustParse("88888888-bbbb-2222-2222-222222222222"), lotLiveTwo, bidderOneID, 9200, int64Ptr(10000), now.Add(-5*time.Minute + time.Millisecond)},
	}
	for _, b := range bids {
		_, err := tx.Exec(ctx, `
INSERT INTO lot_bids (id, tenant_id, lot_id, user_id, amount, max_bid_minor, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    lot_id = EXCLUDED.lot_id,
    user_id = EXCLUDED.user_id,
    amount = EXCLUDED.amount,
    max_bid_minor = EXCLUDED.max_bid_minor,
    created_at = EXCLUDED.created_at
`, b.id, tenantID, b.lotID, b.userID, b.amount, b.maxBid, b.created)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedAuctionApprovals(ctx context.Context, tx pgx.Tx, now time.Time) error {
	rows := []struct {
		id               uuid.UUID
		auctionID        uuid.UUID
		userID           uuid.UUID
		status           string
		riskLevel        string
		maxBidLimitMinor *int64
		depositRequired  bool
		depositAmount    *int64
		depositStatus    string
		internalNotes    *string
		decidedBy        *uuid.UUID
		decidedAt        *time.Time
		approvedBy       *uuid.UUID
		approvedAt       *time.Time
		createdAt        time.Time
		updatedAt        time.Time
	}{
		{
			id:              uuid.MustParse("99999999-aaaa-1111-1111-111111111111"),
			auctionID:       auctionLive,
			userID:          bidderOneID,
			status:          "APPROVED",
			riskLevel:       "low",
			depositRequired: false,
			depositStatus:   "not_required",
			decidedBy:       &adminID,
			decidedAt:       timePtr(now.Add(-3 * time.Hour)),
			approvedBy:      &adminID,
			approvedAt:      timePtr(now.Add(-3 * time.Hour)),
			createdAt:       now.Add(-20 * time.Hour),
			updatedAt:       now.Add(-3 * time.Hour),
		},
		{
			id:               uuid.MustParse("aaaaaaaa-bbbb-1111-1111-111111111111"),
			auctionID:        auctionLive,
			userID:           bidderTwoID,
			status:           "APPROVED",
			riskLevel:        "medium",
			maxBidLimitMinor: int64Ptr(9000),
			depositRequired:  false,
			depositStatus:    "not_required",
			internalNotes:    stringPtr("Approved with a capped bidding limit."),
			decidedBy:        &adminID,
			decidedAt:        timePtr(now.Add(-2 * time.Hour)),
			approvedBy:       &adminID,
			approvedAt:       timePtr(now.Add(-2 * time.Hour)),
			createdAt:        now.Add(-18 * time.Hour),
			updatedAt:        now.Add(-2 * time.Hour),
		},
		{
			id:              uuid.MustParse("bbbbbbbb-cccc-1111-1111-111111111111"),
			auctionID:       auctionFuture,
			userID:          bidderTwoID,
			status:          "PENDING",
			riskLevel:       "high",
			depositRequired: true,
			depositAmount:   int64Ptr(1500),
			depositStatus:   "required",
			internalNotes:   stringPtr("High-risk bidder pending deposit."),
			createdAt:       now.Add(-4 * time.Hour),
			updatedAt:       now.Add(-30 * time.Minute),
		},
	}

	for _, row := range rows {
		_, err := tx.Exec(ctx, `
INSERT INTO auction_bidder_approvals (
  id, tenant_id, auction_id, user_id, status, risk_level, max_bid_limit_minor,
  deposit_required, deposit_amount_minor, deposit_status, internal_notes,
  decided_by, decided_at, approved_by_user_id, approved_at, created_at, updated_at
)
VALUES (
  $1,$2,$3,$4,$5,$6,$7,
  $8,$9,$10,$11,
  $12,$13,$14,$15,$16,$17
)
ON CONFLICT (auction_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    risk_level = EXCLUDED.risk_level,
    max_bid_limit_minor = EXCLUDED.max_bid_limit_minor,
    deposit_required = EXCLUDED.deposit_required,
    deposit_amount_minor = EXCLUDED.deposit_amount_minor,
    deposit_status = EXCLUDED.deposit_status,
    internal_notes = EXCLUDED.internal_notes,
    decided_by = EXCLUDED.decided_by,
    decided_at = EXCLUDED.decided_at,
    approved_by_user_id = EXCLUDED.approved_by_user_id,
    approved_at = EXCLUDED.approved_at,
    updated_at = EXCLUDED.updated_at
`, row.id, tenantID, row.auctionID, row.userID, row.status, row.riskLevel, row.maxBidLimitMinor,
			row.depositRequired, row.depositAmount, row.depositStatus, row.internalNotes,
			row.decidedBy, row.decidedAt, row.approvedBy, row.approvedAt, row.createdAt, row.updatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}

func seedFees(ctx context.Context, tx pgx.Tx, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO fee_components (id, tenant_id, name, calculation_type, value, applies_to, taxable, active, created_at)
VALUES
  ($1,$2,'Buyer''s Premium','PERCENTAGE',15,'HAMMER',true,true,$3)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    calculation_type = EXCLUDED.calculation_type,
    value = EXCLUDED.value,
    applies_to = EXCLUDED.applies_to,
    taxable = EXCLUDED.taxable,
    active = EXCLUDED.active
`, feePremium, tenantID, now.Add(-72*time.Hour))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO tax_rules (id, tenant_id, name, rate_percent, active, created_at)
VALUES
  ($1,$2,'VAT',10,true,$3)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    rate_percent = EXCLUDED.rate_percent,
    active = EXCLUDED.active
`, taxRuleVAT, tenantID, now.Add(-72*time.Hour))
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

	_, err = tx.Exec(ctx, `
INSERT INTO tax_rule_assignments (id, scope_type, scope_id, tax_rule_id, created_at)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (scope_type, scope_id, tax_rule_id) DO NOTHING
`, uuid.MustParse("77777777-aaaa-1111-1111-111111111111"), "TENANT", tenantID, taxRuleVAT, now.Add(-48*time.Hour))
	if err != nil {
		return err
	}
	return nil
}

func seedInvoices(ctx context.Context, tx pgx.Tx, now time.Time) error {
	snapshot, _ := json.Marshal(map[string]any{
		"currency":     "GBP",
		"hammer_price": 12000,
		"fees": []map[string]any{
			{"name": "Buyer's Premium", "type": "PERCENTAGE", "value": 15, "amount": 1800, "taxable": true, "category": "FEE"},
			{"name": "VAT", "type": "PERCENTAGE", "value": 10, "amount": 180, "category": "TAX"},
		},
		"total_due": 13980,
	})
	issued := now.Add(-20 * time.Hour)
	due := now.Add(7 * 24 * time.Hour)
	_, err := tx.Exec(ctx, `
INSERT INTO invoices (
  id, lot_id, tenant_id, auction_id, winner_user_id, invoice_number, issued_at, due_date,
  currency, total_amount, hammer_price_minor, fees_total_minor, tax_total_minor, total_due_minor,
  payment_status, snapshot, status, created_at
)
VALUES ($1,$2,$3,$4,$5,'INV-2026-000001',$6,$7,'GBP',13980,12000,1800,180,13980,'UNPAID',$8,'ISSUED',$9)
ON CONFLICT (lot_id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    auction_id = EXCLUDED.auction_id,
    winner_user_id = EXCLUDED.winner_user_id,
    invoice_number = EXCLUDED.invoice_number,
    issued_at = EXCLUDED.issued_at,
    due_date = EXCLUDED.due_date,
    currency = EXCLUDED.currency,
    total_amount = EXCLUDED.total_amount,
    hammer_price_minor = EXCLUDED.hammer_price_minor,
    fees_total_minor = EXCLUDED.fees_total_minor,
    tax_total_minor = EXCLUDED.tax_total_minor,
    total_due_minor = EXCLUDED.total_due_minor,
    payment_status = EXCLUDED.payment_status,
    snapshot = EXCLUDED.snapshot,
    status = EXCLUDED.status
`, invoiceOneID, lotEndedSold, tenantID, auctionEnded, bidderOneID, issued, due, snapshot, issued)
	return err
}

func seedAssets(ctx context.Context, tx pgx.Tx, now time.Time) error {
	items := []struct {
		id        uuid.UUID
		kind      string
		url       string
		width     int
		height    int
		sizeBytes int64
		mime      string
		createdAt time.Time
	}{
		{
			id:        assetLogoID,
			kind:      "logo",
			url:       "https://picsum.photos/seed/elevare-logo/512/512",
			width:     512,
			height:    512,
			sizeBytes: 128000,
			mime:      "image/png",
			createdAt: now.Add(-72 * time.Hour),
		},
		{
			id:        assetBannerID,
			kind:      "banner",
			url:       "https://picsum.photos/seed/elevare-hero/1600/900",
			width:     1600,
			height:    900,
			sizeBytes: 420000,
			mime:      "image/jpeg",
			createdAt: now.Add(-48 * time.Hour),
		},
		{
			id:        assetImageID,
			kind:      "image",
			url:       "https://picsum.photos/seed/elevare-gallery-1/1200/900",
			width:     1200,
			height:    900,
			sizeBytes: 310000,
			mime:      "image/jpeg",
			createdAt: now.Add(-36 * time.Hour),
		},
	}

	for _, item := range items {
		_, err := tx.Exec(ctx, `
INSERT INTO tenant_assets (id, tenant_id, kind, url, width, height, size_bytes, mime, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (id) DO UPDATE
SET kind = EXCLUDED.kind,
    url = EXCLUDED.url,
    width = EXCLUDED.width,
    height = EXCLUDED.height,
    size_bytes = EXCLUDED.size_bytes,
    mime = EXCLUDED.mime,
    created_at = EXCLUDED.created_at
`, item.id, tenantID, item.kind, item.url, item.width, item.height, item.sizeBytes, item.mime, item.createdAt)
		if err != nil {
			return err
		}
	}
	return nil
}
