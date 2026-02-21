import { useEffect } from "react";

interface AuctionStatusEvent {
  auction_id: string;
  status: "DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
}

export function useAuctionSocket(auctionId: string | undefined, onStatusEvent: (event: AuctionStatusEvent) => void, enabled = true) {
  useEffect(() => {
    if (!auctionId || !enabled) return;

    const base =
      import.meta.env.VITE_WS_BASE_URL ??
      `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`;

    const ws = new WebSocket(`${base}/auctions/${auctionId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & AuctionStatusEvent;
        if (data.type === "auction_status_changed") onStatusEvent(data);
      } catch {
        // ignore invalid socket payloads
      }
    };

    return () => ws.close();
  }, [auctionId, enabled, onStatusEvent]);
}
