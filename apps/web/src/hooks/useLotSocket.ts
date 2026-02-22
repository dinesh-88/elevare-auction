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

    const envBase = process.env.NEXT_PUBLIC_WS_BASE_URL?.trim();
    const base =
      envBase && envBase.length > 0
        ? adaptWsBaseToCurrentHostname(envBase.replace(/\/+$/, ""))
        : `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.hostname}:${process.env.NEXT_PUBLIC_WS_PORT ?? "8080"}`;

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
