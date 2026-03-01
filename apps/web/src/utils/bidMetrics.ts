export type BidMetricName =
  | "lot_viewed"
  | "watcher_to_bid_conversion"
  | "rebid"
  | "last_minute_bid"
  | "outbid_seen"
  | "quick_increment_selected";

type BidMetricEvent = {
  name: BidMetricName;
  lotId?: string;
  auctionId?: string;
  amount?: number;
  secondsRemaining?: number;
  source?: string;
  timestamp: string;
};

const STORAGE_KEY = "elevare_bid_metrics";
const MAX_EVENTS = 200;

export function trackBidMetric(event: Omit<BidMetricEvent, "timestamp">) {
  if (typeof window === "undefined") return;

  const record: BidMetricEvent = {
    ...event,
    timestamp: new Date().toISOString(),
  };

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    const existing = raw ? ((JSON.parse(raw) as unknown[]) || []) : [];
    const next = [...existing, record].slice(-MAX_EVENTS);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  } catch {
    // Ignore storage errors; the in-memory event still fires below.
  }

  window.dispatchEvent(new CustomEvent("elevare:bid-metric", { detail: record }));
}
