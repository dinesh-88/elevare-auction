package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uuid.UUID]map[*websocket.Conn]struct{})}
}

func (h *Hub) Join(auctionID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[auctionID]; !ok {
		h.rooms[auctionID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[auctionID][conn] = struct{}{}
}

func (h *Hub) Leave(auctionID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[auctionID]; ok {
		delete(room, conn)
		if len(room) == 0 {
			delete(h.rooms, auctionID)
		}
	}
}

func (h *Hub) BroadcastBid(ctx context.Context, auctionID uuid.UUID, bid domain.Bid, currentPrice int64, endTime time.Time) {
	h.mu.RLock()
	room := h.rooms[auctionID]
	conns := make([]*websocket.Conn, 0, len(room))
	for c := range room {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	if len(conns) == 0 {
		return
	}
	payload := map[string]any{
		"type":          "bid_placed",
		"auction_id":    auctionID,
		"bid_id":        bid.ID,
		"amount":        bid.Amount,
		"current_price": currentPrice,
		"end_time":      endTime,
		"created_at":    bid.CreatedAt,
	}
	b, _ := json.Marshal(payload)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, b)
	}
}
