import { useEffect } from "react";

const adaptWsBaseToCurrentHostname = (base: string) => {
  if (!base) return base;
  try {
    const url = new URL(base);
    const currentHost = window.location.hostname;
    const baseHost = url.hostname;
    const isLocalBase = baseHost === "localhost" || baseHost === "127.0.0.1";
    const isTenantHost = currentHost.endsWith(".localhost") && currentHost !== "localhost";
    if (isLocalBase && isTenantHost) {
      url.hostname = currentHost;
      return url.toString().replace(/\/+$/, "");
    }
    return base;
  } catch {
    return base;
  }
};

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

    const envBase = process.env.NEXT_PUBLIC_WS_BASE_URL?.trim();
    const base =
      envBase && envBase.length > 0
        ? adaptWsBaseToCurrentHostname(envBase.replace(/\/+$/, ""))
        : `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.hostname}:${process.env.NEXT_PUBLIC_WS_PORT ?? "8080"}`;

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
