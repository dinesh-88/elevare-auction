import { useEffect, useRef } from "react";
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
  const statusHandlerRef = useRef(onStatusEvent);
  const bidHandlerRef = useRef(onBidEvent);

  useEffect(() => {
    statusHandlerRef.current = onStatusEvent;
    bidHandlerRef.current = onBidEvent;
  }, [onStatusEvent, onBidEvent]);

  useEffect(() => {
    if (!auctionId || !enabled) return;

    const base = resolveWsBase();
    let ws: WebSocket | null = null;
    let cancelled = false;
    let idleId: number | null = null;
    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    const connect = () => {
      if (cancelled) return;
      ws = new WebSocket(`${base}/auctions/${auctionId}/ws`);

      ws.onmessage = (evt) => {
        try {
          const data = JSON.parse(evt.data) as { type?: string } & AuctionStatusEvent & AuctionBidEvent;
          if (data.type === "auction_status_changed") statusHandlerRef.current(data);
          if (data.type === "bid.created" && bidHandlerRef.current) bidHandlerRef.current(data);
        } catch {
          // ignore invalid socket payloads
        }
      };
    };

    if (typeof window !== "undefined" && "requestIdleCallback" in window) {
      idleId = window.requestIdleCallback(connect, { timeout: 1200 });
    } else {
      timeoutId = globalThis.setTimeout(connect, 0);
    }

    return () => {
      cancelled = true;
      if (idleId !== null && typeof window !== "undefined" && "cancelIdleCallback" in window) {
        window.cancelIdleCallback(idleId);
      }
      if (timeoutId !== null) {
        globalThis.clearTimeout(timeoutId);
      }
      ws?.close();
    };
  }, [auctionId, enabled]);
}
