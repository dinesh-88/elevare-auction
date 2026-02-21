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

export function useLotSocket(
  lotId: string | undefined,
  onBidEvent: (event: LotBidEvent) => void,
  onStatusEvent?: (event: LotStatusEvent) => void,
  enabled = true,
) {
  useEffect(() => {
    if (!lotId || !enabled) return;

    const base =
      import.meta.env.VITE_WS_BASE_URL ??
      `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`;

    const ws = new WebSocket(`${base}/lots/${lotId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & LotBidEvent & LotStatusEvent;
        if (data.type === "lot_bid_placed") onBidEvent(data);
        if (data.type === "lot_status_changed" && onStatusEvent) onStatusEvent(data);
      } catch {
        // ignore invalid socket payload
      }
    };

    return () => ws.close();
  }, [lotId, enabled, onBidEvent, onStatusEvent]);
}
