package httptransport

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

func TestAuctionWebSocketBroadcastsBidCreated(t *testing.T) {
	hub := wsadapter.NewHub()
	router := chi.NewRouter()
	router.Get("/auctions/{id}/ws", AuctionWS(hub))
	server := httptest.NewServer(router)
	defer server.Close()

	auctionID := uuid.New()
	lotID := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/auctions/"+auctionID.String()+"/ws", nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	endedAt := time.Now().UTC().Add(time.Minute)
	bid := domain.Bid{ID: uuid.New(), UserID: uuid.New(), Amount: 2500, CreatedAt: time.Now().UTC()}
	hub.BroadcastAuctionBid(ctx, auctionID, domain.Lot{ID: lotID, LotNumber: "14"}, bid, 2500, &endedAt)

	_, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("websocket read failed: %v", err)
	}

	var message map[string]any
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if got := message["type"]; got != "bid.created" {
		t.Fatalf("expected bid.created event, got %v", got)
	}
	if got := message["lot_number"]; got != "14" {
		t.Fatalf("expected lot number in payload, got %v", got)
	}
}
