package ws

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"strconv"
	"sync"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type Hub struct {
	mu       sync.RWMutex
	rooms    map[uuid.UUID]map[*websocket.Conn]struct{}
	lotRooms map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:    make(map[uuid.UUID]map[*websocket.Conn]struct{}),
		lotRooms: make(map[uuid.UUID]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Join(auctionID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[auctionID]; !ok {
		h.rooms[auctionID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[auctionID][conn] = struct{}{}
}

func (h *Hub) JoinLot(lotID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.lotRooms[lotID]; !ok {
		h.lotRooms[lotID] = make(map[*websocket.Conn]struct{})
	}
	h.lotRooms[lotID][conn] = struct{}{}
}

func (h *Hub) LeaveLot(lotID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.lotRooms[lotID]; ok {
		delete(room, conn)
		if len(room) == 0 {
			delete(h.lotRooms, lotID)
		}
	}
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

func (h *Hub) BroadcastLotBid(ctx context.Context, lotID uuid.UUID, bid domain.Bid, currentPrice int64, endedAt *time.Time) {
	h.mu.RLock()
	room := h.lotRooms[lotID]
	conns := make([]*websocket.Conn, 0, len(room))
	for c := range room {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	if len(conns) == 0 {
		return
	}
	payload := map[string]any{
		"type":          "lot_bid_placed",
		"lot_id":        lotID,
		"bid_id":        bid.ID,
		"amount":        bid.Amount,
		"bidder_display": lotBidderAlias(bid.UserID),
		"current_price": currentPrice,
		"ended_at":      endedAt,
		"created_at":    bid.CreatedAt,
	}
	b, _ := json.Marshal(payload)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, b)
	}
}

func lotBidderAlias(userID uuid.UUID) string {
	return "Bidder " + strconv.Itoa(int(crc32.ChecksumIEEE(userID[:])%900)+100)
}

func (h *Hub) BroadcastLotExtended(ctx context.Context, lotID uuid.UUID, newEndTime time.Time) {
	h.mu.RLock()
	room := h.lotRooms[lotID]
	conns := make([]*websocket.Conn, 0, len(room))
	for c := range room {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	if len(conns) == 0 {
		return
	}
	payload := map[string]any{
		"type":         "LOT_EXTENDED",
		"lot_id":       lotID,
		"new_end_time": newEndTime,
	}
	b, _ := json.Marshal(payload)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, b)
	}
}

func (h *Hub) BroadcastLotStatus(ctx context.Context, lotID uuid.UUID, status domain.LotStatus) {
	h.mu.RLock()
	room := h.lotRooms[lotID]
	conns := make([]*websocket.Conn, 0, len(room))
	for c := range room {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	if len(conns) == 0 {
		return
	}
	payload := map[string]any{
		"type":   "lot_status_changed",
		"lot_id": lotID,
		"status": status,
	}
	b, _ := json.Marshal(payload)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, b)
	}
}

func (h *Hub) BroadcastAuctionStatus(ctx context.Context, auctionID uuid.UUID, status domain.AuctionStatus) {
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
		"type":       "auction_status_changed",
		"auction_id": auctionID,
		"status":     status,
	}
	b, _ := json.Marshal(payload)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, b)
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

// BroadcastUserNotification is currently a no-op because user-scoped WS rooms are not yet implemented.
func (h *Hub) BroadcastUserNotification(ctx context.Context, tenantID, userID uuid.UUID, notification domain.Notification) {
}
