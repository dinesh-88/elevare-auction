import { useEffect } from "react";

interface LotBidEvent {
  bid_id: string;
  amount: number;
  current_price: number;
  ended_at?: string | null;
  created_at: string;
}

interface LotStatusEvent {
  lot_id: string;
  status: "PENDING" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
}

export function useLotsSockets(
  lotIds: string[],
  onBidEvent: (lotId: string, event: LotBidEvent) => void,
  onStatusEvent?: (lotId: string, event: LotStatusEvent) => void,
  enabled = true,
) {
  useEffect(() => {
    if (!enabled || lotIds.length === 0) return;

    const base =
      process.env.NEXT_PUBLIC_WS_BASE_URL ??
      `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.hostname}:${process.env.NEXT_PUBLIC_WS_PORT ?? "8080"}`;

    const sockets = lotIds.map((lotId) => {
      const ws = new WebSocket(`${base}/lots/${lotId}/ws`);
      ws.onmessage = (evt) => {
        try {
          const data = JSON.parse(evt.data) as { type?: string } & LotBidEvent & LotStatusEvent;
          if (data.type === "lot_bid_placed") onBidEvent(lotId, data);
          if (data.type === "lot_status_changed" && onStatusEvent) onStatusEvent(lotId, data);
        } catch {
          // ignore invalid payload
        }
      };
      return ws;
    });

    return () => {
      sockets.forEach((ws) => ws.close());
    };
  }, [enabled, lotIds, onBidEvent, onStatusEvent]);
}
