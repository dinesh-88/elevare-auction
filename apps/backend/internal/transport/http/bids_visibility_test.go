package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/middleware"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/dinesh-88/elevare-auction/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeBidLotRepo struct {
	bids      []domain.Bid
	placeResp *port.PlaceLotBidTxResult
}

func (f *fakeBidLotRepo) Create(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return nil
}
func (f *fakeBidLotRepo) Update(ctx context.Context, tenantID uuid.UUID, lot *domain.Lot) error {
	return nil
}
func (f *fakeBidLotRepo) Delete(ctx context.Context, tenantID, lotID uuid.UUID) error { return nil }
func (f *fakeBidLotRepo) Reorder(ctx context.Context, tenantID, auctionID uuid.UUID, lotIDs []uuid.UUID) error {
	return nil
}
func (f *fakeBidLotRepo) ListByAuction(ctx context.Context, tenantID, auctionID uuid.UUID) ([]domain.Lot, error) {
	return nil, nil
}
func (f *fakeBidLotRepo) GetByID(ctx context.Context, tenantID, lotID uuid.UUID) (*domain.Lot, error) {
	return &domain.Lot{ID: lotID, AuctionID: uuid.New()}, nil
}
func (f *fakeBidLotRepo) Start(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) error {
	return nil
}
func (f *fakeBidLotRepo) End(ctx context.Context, tenantID, lotID uuid.UUID, now time.Time) (*port.LotSettlementResult, error) {
	return nil, nil
}
func (f *fakeBidLotRepo) OnAuctionStart(ctx context.Context, tenantID, auctionID uuid.UUID, now time.Time) error {
	return nil
}
func (f *fakeBidLotRepo) PlaceBidTx(ctx context.Context, in port.PlaceLotBidTxInput) (*port.PlaceLotBidTxResult, error) {
	if f.placeResp != nil {
		return f.placeResp, nil
	}
	return nil, nil
}
func (f *fakeBidLotRepo) ListBidsByLot(ctx context.Context, tenantID, lotID uuid.UUID, limit int) ([]domain.Bid, error) {
	return f.bids, nil
}
func (f *fakeBidLotRepo) ListWonByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.Lot, error) {
	return nil, nil
}
func (f *fakeBidLotRepo) RunSchedulerTick(ctx context.Context, now time.Time) error { return nil }

type noopBidBroadcaster struct{}

func (n *noopBidBroadcaster) BroadcastLotBid(ctx context.Context, lotID uuid.UUID, bid domain.Bid, currentPrice int64, endedAt *time.Time) {
}
func (n *noopBidBroadcaster) BroadcastLotExtended(ctx context.Context, lotID uuid.UUID, newEndTime time.Time) {}
func (n *noopBidBroadcaster) BroadcastLotStatus(ctx context.Context, lotID uuid.UUID, status domain.LotStatus) {
}
func (n *noopBidBroadcaster) BroadcastAuctionStatus(ctx context.Context, auctionID uuid.UUID, status domain.AuctionStatus) {
}

func TestPublicBidHistoryDoesNotExposeProxyFields(t *testing.T) {
	tenantID := uuid.New()
	lotID := uuid.New()
	repo := &fakeBidLotRepo{bids: []domain.Bid{
		{ID: uuid.New(), UserID: uuid.New(), Amount: 1000, CreatedAt: time.Now().UTC()},
		{ID: uuid.New(), UserID: uuid.New(), Amount: 1100, CreatedAt: time.Now().UTC().Add(2 * time.Second)},
	}}
	h := &Handler{Lot: &usecase.LotUseCase{Lots: repo, Broadcaster: &noopBidBroadcaster{}}}

	req := httptest.NewRequest(http.MethodGet, "/lots/"+lotID.String()+"/bids", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("lotId", lotID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = middleware.WithTenant(ctx, &domain.Tenant{ID: tenantID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ListLotBids(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.Bytes()
	if bytes.Contains(body, []byte("max_bid_minor")) {
		t.Fatal("response must not include max_bid_minor")
	}
	if bytes.Contains(body, []byte("bid_type")) {
		t.Fatal("response must not include bid_type")
	}
	if bytes.Contains(body, []byte("user_id")) {
		t.Fatal("response must not include user_id")
	}

	var rows []map[string]any
	if err := json.Unmarshal(body, &rows); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	for _, row := range rows {
		if _, ok := row["amount_minor"]; !ok {
			t.Fatal("missing amount_minor")
		}
		if _, ok := row["created_at"]; !ok {
			t.Fatal("missing created_at")
		}
		if _, ok := row["bidder_display"]; !ok {
			t.Fatal("missing bidder_display")
		}
	}
}

func TestCompetingProxyLikeBidsHaveNoProxyIndication(t *testing.T) {
	bids := []domain.Bid{
		{ID: uuid.New(), UserID: uuid.New(), Amount: 1000, CreatedAt: time.Now().UTC()},
		{ID: uuid.New(), UserID: uuid.New(), Amount: 1050, CreatedAt: time.Now().UTC().Add(time.Second)},
		{ID: uuid.New(), UserID: uuid.New(), Amount: 1100, CreatedAt: time.Now().UTC().Add(2 * time.Second)},
	}
	resp := toPublicBidHistory(bids)
	for _, row := range resp {
		if row.BidderDisplay == "" {
			t.Fatal("expected bidder_display")
		}
		display := strings.ToLower(row.BidderDisplay)
		if strings.Contains(display, "proxy") || strings.Contains(display, "auto") {
			t.Fatal("bidder display should not indicate proxy")
		}
	}
}

func TestPublicBidHistoryUsesStableAliasPerUser(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	base := time.Now().UTC()
	first := toPublicBidHistory([]domain.Bid{
		{ID: uuid.New(), UserID: userA, Amount: 1000, CreatedAt: base},
		{ID: uuid.New(), UserID: userB, Amount: 1100, CreatedAt: base.Add(time.Second)},
	})
	second := toPublicBidHistory([]domain.Bid{
		{ID: uuid.New(), UserID: userB, Amount: 1200, CreatedAt: base.Add(2 * time.Second)},
		{ID: uuid.New(), UserID: userA, Amount: 1300, CreatedAt: base.Add(3 * time.Second)},
	})
	if first[0].BidderDisplay == first[1].BidderDisplay {
		t.Fatal("expected distinct aliases per user")
	}
	if first[0].BidderDisplay != second[1].BidderDisplay {
		t.Fatal("expected stable alias for user A")
	}
	if first[1].BidderDisplay != second[0].BidderDisplay {
		t.Fatal("expected stable alias for user B")
	}
}

func TestPlaceBidResponseDoesNotExposeMaxBidMinor(t *testing.T) {
	tenantID := uuid.New()
	lotID := uuid.New()
	userID := uuid.New()
	max := int64(1500)
	repo := &fakeBidLotRepo{
		placeResp: &port.PlaceLotBidTxResult{
			Lot: domain.Lot{ID: lotID, CurrentPrice: 1100},
			Bid: domain.Bid{
				ID:        uuid.New(),
				LotID:     &lotID,
				UserID:    userID,
				Amount:    1100,
				MaxBid:    &max,
				CreatedAt: time.Now().UTC(),
			},
		},
	}
	h := &Handler{Lot: &usecase.LotUseCase{Lots: repo, Broadcaster: &noopBidBroadcaster{}}}

	body := strings.NewReader(`{"max_bid_minor":1500}`)
	req := httptest.NewRequest(http.MethodPost, "/lots/"+lotID.String()+"/bid", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("lotId", lotID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = middleware.WithTenant(ctx, &domain.Tenant{ID: tenantID})
	ctx = middleware.WithClaims(ctx, &port.TokenClaims{TenantID: tenantID, UserID: userID, Role: domain.RoleBidder})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.PlaceLotBid(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "max_bid_minor") || strings.Contains(rr.Body.String(), "MaxBid") {
		t.Fatalf("place bid response leaked max bid: %s", rr.Body.String())
	}
}
