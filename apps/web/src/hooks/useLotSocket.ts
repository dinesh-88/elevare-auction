import { useEffect } from "react";
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
  useEffect(() => {
    if (!lotId || !enabled) return;

    const base = resolveWsBase();

    const ws = new WebSocket(`${base}/lots/${lotId}/ws`);

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data) as { type?: string } & LotBidEvent & LotStatusEvent & LotExtendedEvent;
        if (data.type === "lot_bid_placed") onBidEvent(data);
        if (data.type === "lot_status_changed" && onStatusEvent) onStatusEvent(data);
        if (data.type === "LOT_EXTENDED" && onExtendedEvent) onExtendedEvent(data);
      } catch {
        // ignore invalid socket payload
      }
    };

    return () => ws.close();
  }, [lotId, enabled, onBidEvent, onStatusEvent, onExtendedEvent]);
}
