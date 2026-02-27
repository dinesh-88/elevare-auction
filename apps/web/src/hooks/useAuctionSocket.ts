import { useEffect } from "react";
import { resolveWsBase } from "./wsBase";

interface AuctionStatusEvent {
  auction_id: string;
  status: "DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
}

export interface AuctionBidEvent {
  auction_id: string;
  lot_id: string;
  lot_number: string;
  bid_id: string;
  amount: number;
  current_price: number;
  ended_at?: string | null;
  created_at: string;
}

export function useAuctionSocket(
  auctionId: string | undefined,
  onStatusEvent: (event: AuctionStatusEvent) => void,
  enabled = true,
  onBidEvent?: (event: AuctionBidEvent) => void,
) {
  useEffect(() => {
    if (!auctionId || !enabled) return;

    const base = resolveWsBase();

    const ws = new WebSocket(`${base}/auctions/${auctionId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & AuctionStatusEvent & AuctionBidEvent;
        if (data.type === "auction_status_changed") onStatusEvent(data);
        if (data.type === "bid.created" && onBidEvent) onBidEvent(data);
      } catch {
        // ignore invalid socket payloads
      }
    };

    return () => ws.close();
  }, [auctionId, enabled, onBidEvent, onStatusEvent]);
}
