import { useEffect, useRef } from "react";
import { resolveWsBase } from "./wsBase";

interface LotBidEvent {
  bid_id: string;
  amount: number;
  bidder_display?: string;
  current_price: number;
  ended_at?: string | null;
  created_at: string;
}

interface LotStatusEvent {
  lot_id: string;
  status: "PENDING" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
}

interface LotExtendedEvent {
  lot_id: string;
  new_end_time: string;
}

export function useLotSocket(
  lotId: string | undefined,
  onBidEvent: (event: LotBidEvent) => void,
  onStatusEvent?: (event: LotStatusEvent) => void,
  onExtendedEvent?: (event: LotExtendedEvent) => void,
  enabled = true,
) {
  const bidHandlerRef = useRef(onBidEvent);
  const statusHandlerRef = useRef(onStatusEvent);
  const extendedHandlerRef = useRef(onExtendedEvent);

  useEffect(() => {
    bidHandlerRef.current = onBidEvent;
    statusHandlerRef.current = onStatusEvent;
    extendedHandlerRef.current = onExtendedEvent;
  }, [onBidEvent, onStatusEvent, onExtendedEvent]);

  useEffect(() => {
    if (!lotId || !enabled) return;

    const base = resolveWsBase();
    let ws: WebSocket | null = null;
    let cancelled = false;
    let idleId: number | null = null;
    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    const connect = () => {
      if (cancelled) return;
      ws = new WebSocket(`${base}/lots/${lotId}/ws`);

      ws.onmessage = (evt) => {
        try {
          const data = JSON.parse(evt.data) as { type?: string } & LotBidEvent & LotStatusEvent & LotExtendedEvent;
          if (data.type === "lot_bid_placed") bidHandlerRef.current(data);
          if (data.type === "lot_status_changed" && statusHandlerRef.current) statusHandlerRef.current(data);
          if (data.type === "LOT_EXTENDED" && extendedHandlerRef.current) extendedHandlerRef.current(data);
        } catch {
          // ignore invalid socket payload
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
  }, [lotId, enabled]);
}
