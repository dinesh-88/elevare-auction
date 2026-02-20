import { useEffect } from "react";

interface BidEvent {
  bid_id: string;
  amount: number;
  current_price: number;
  end_time: string;
  created_at: string;
}

export function useAuctionSocket(
  auctionId: string | undefined,
  onBidEvent: (event: BidEvent) => void,
  enabled = true,
) {
  useEffect(() => {
    if (!auctionId || !enabled) return;

    const base =
      import.meta.env.VITE_WS_BASE_URL ??
      `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`;

    const ws = new WebSocket(`${base}/auctions/${auctionId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & BidEvent;
        if (data.type === "bid_placed") onBidEvent(data);
      } catch {
        // ignore invalid socket payloads
      }
    };

    return () => ws.close();
  }, [auctionId, enabled, onBidEvent]);
}
