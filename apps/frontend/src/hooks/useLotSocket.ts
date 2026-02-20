import { useEffect } from "react";

interface LotBidEvent {
  bid_id: string;
  amount: number;
  current_price: number;
  ended_at?: string | null;
  created_at: string;
}

export function useLotSocket(lotId: string | undefined, onBidEvent: (event: LotBidEvent) => void, enabled = true) {
  useEffect(() => {
    if (!lotId || !enabled) return;

    const base =
      import.meta.env.VITE_WS_BASE_URL ??
      `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`;

    const ws = new WebSocket(`${base}/lots/${lotId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & LotBidEvent;
        if (data.type === "lot_bid_placed") onBidEvent(data);
      } catch {
        // ignore invalid socket payload
      }
    };

    return () => ws.close();
  }, [lotId, enabled, onBidEvent]);
}
