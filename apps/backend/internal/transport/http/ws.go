package httptransport

import (
	"context"
	"errors"
	"net/http"
	"time"

	wsadapter "github.com/dinesh-88/elevare-auction/internal/adapter/broadcaster/ws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

func AuctionWS(hub *wsadapter.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auctionID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid auction id", http.StatusBadRequest)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "bye")
		hub.Join(auctionID, conn)
		defer hub.Leave(auctionID, conn)

		ctx := r.Context()
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure || errors.Is(err, context.Canceled) {
					return
				}
				return
			}
			_ = conn.Ping(ctx)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func LotWS(hub *wsadapter.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lotID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid lot id", http.StatusBadRequest)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "bye")
		hub.JoinLot(lotID, conn)
		defer hub.LeaveLot(lotID, conn)

		ctx := r.Context()
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure || errors.Is(err, context.Canceled) {
					return
				}
				return
			}
			_ = conn.Ping(ctx)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
