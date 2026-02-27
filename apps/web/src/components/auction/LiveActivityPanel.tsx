"use client";

import { useCallback, useEffect, useState } from "react";
import { type AuctionBidEvent, useAuctionSocket } from "../../hooks/useAuctionSocket";

type ActivityItem = Pick<AuctionBidEvent, "bid_id" | "lot_number" | "amount" | "created_at">;

const EMPTY_STATUS_HANDLER = () => {};

function formatRelativeTime(timestamp: string, nowMs: number) {
  const deltaSeconds = Math.max(0, Math.floor((nowMs - new Date(timestamp).getTime()) / 1000));
  if (deltaSeconds < 60) return `${deltaSeconds}s ago`;
  const minutes = Math.floor(deltaSeconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export default function LiveActivityPanel({ auctionId, currencySymbol }: { auctionId: string | undefined; currencySymbol: string }) {
  const [items, setItems] = useState<ActivityItem[]>([]);
  const [nowMs, setNowMs] = useState(() => Date.now());

  const handleBidEvent = useCallback((event: AuctionBidEvent) => {
    setItems((prev) => {
      const next = [{ bid_id: event.bid_id, lot_number: event.lot_number, amount: event.amount, created_at: event.created_at }, ...prev];
      return next.slice(0, 10);
    });
  }, []);

  useAuctionSocket(auctionId, EMPTY_STATUS_HANDLER, Boolean(auctionId), handleBidEvent);

  useEffect(() => {
    const timer = window.setInterval(() => setNowMs(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  return (
    <aside className="rounded-[24px] border border-neutral-200 bg-neutral-50 px-5 py-5 sm:px-6">
      <div className="flex items-center justify-between gap-3 border-b border-neutral-200 pb-3">
        <h3 className="font-heading text-xl text-neutral-900">Live Activity</h3>
        <span className="text-[11px] uppercase tracking-[0.18em] text-neutral-500">Recent bids</span>
      </div>

      {items.length === 0 ? (
        <p className="pt-4 text-sm leading-6 text-neutral-500">No recent bids yet.</p>
      ) : (
        <div className="space-y-3 pt-4">
          {items.map((item) => (
            <div key={item.bid_id} className="rounded-2xl bg-white px-4 py-3 shadow-sm ring-1 ring-neutral-200/80">
              <div className="flex items-center justify-between gap-3 text-sm text-neutral-600">
                <span>Lot {item.lot_number}</span>
                <span>{formatRelativeTime(item.created_at, nowMs)}</span>
              </div>
              <div className="mt-1 font-heading text-lg text-neutral-900">
                {currencySymbol}
                {item.amount.toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}
