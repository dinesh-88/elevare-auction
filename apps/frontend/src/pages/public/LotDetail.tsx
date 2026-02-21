import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import Countdown from "../../components/Countdown";
import BidHistory from "../../components/BidHistory";
import BidInput from "../../components/BidInput";
import LotImageGallery from "../../components/LotImageGallery";
import { api, type Lot } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { useAuctionSocket } from "../../hooks/useAuctionSocket";
import { useLotSocket } from "../../hooks/useLotSocket";
import StatusBadge from "../../components/ui/StatusBadge";

interface LotBidHistory {
  bid_id: string;
  amount: number;
  created_at: string;
}

export default function LotDetail() {
  const { lotId } = useParams();
  const { isAuthenticated, isAdmin } = useAuth();
  const [lot, setLot] = useState<Lot | null>(null);
  const [auctionStatus, setAuctionStatus] = useState<"DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED" | null>(null);
  const [history, setHistory] = useState<LotBidHistory[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!lotId) return;
    try {
      const data = await api.getLot(lotId);
      setLot(data);
      const auction = await api.getAuction(data.auction_id);
      setAuctionStatus(auction.status);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load lot");
    }
  }, [lotId]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!lot || auctionStatus !== "LIVE" || lot.status !== "LIVE") return;
    const id = window.setInterval(() => {
      void load();
    }, 5000);
    return () => window.clearInterval(id);
  }, [lot, auctionStatus, load]);

  useLotSocket(
    lotId,
    (event) => {
      setLot((prev) =>
        prev
          ? {
              ...prev,
              current_price: event.current_price,
              ended_at: event.ended_at ?? prev.ended_at,
            }
          : prev,
      );
      setHistory((prev) => [{ bid_id: event.bid_id, amount: event.amount, created_at: event.created_at }, ...prev]);
    },
    (event) => {
      setLot((prev) => {
        if (!prev) return prev;
        if (event.status === "ENDED" || event.status === "CANCELLED") {
          return { ...prev, status: event.status, ended_at: new Date().toISOString() };
        }
        return { ...prev, status: event.status };
      });
      if (event.status === "LIVE") {
        void load();
      }
    },
    true,
  );

  useAuctionSocket(
    lot?.auction_id,
    (event) => {
      setAuctionStatus(event.status);
      if (event.status === "LIVE") {
        void load();
      }
    },
    Boolean(lot?.auction_id),
  );

  const placeBid = async (nextAmount: number) => {
    if (!lotId) return;
    try {
      setError(null);
      await api.placeLotBid(lotId, nextAmount);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to place bid");
      throw err;
    }
  };

  if (error) return <p className="error">{error}</p>;
  if (!lot) return <p className="text-sm text-neutral-500">Loading lot...</p>;

  const images = lot.metadata?.images?.length
    ? lot.metadata.images
    : lot.metadata?.image
      ? [lot.metadata.image]
      : [];
  const formattedHistory = history.map((b) => ({
    id: b.bid_id,
    amount: b.amount,
    bidder: "Bidder",
    timestamp: new Date(b.created_at),
  }));
  const meta = lot.metadata ?? {};
  const medium = typeof meta.medium === "string" ? meta.medium : "Not specified";
  const dimensions = typeof meta.dimensions === "string" ? meta.dimensions : "Not specified";
  const date = typeof meta.date === "string" ? meta.date : "Not specified";
  const provenance = typeof meta.provenance === "string" ? meta.provenance : "Not specified";
  const conditionReport = typeof meta.condition_report === "string" ? meta.condition_report : "No condition report available.";

  return (
    <section className="space-y-12">
      <div className="grid gap-12 md:grid-cols-5">
        <div className="md:col-span-3">
          <LotImageGallery images={images} alt={lot.title} />
        </div>

        <aside className="space-y-6 md:col-span-2">
          <h1 className="font-heading text-3xl">{lot.title}</h1>
          <p className="text-sm text-neutral-600">{lot.description}</p>
          <p className="text-sm text-neutral-500">
            Lot {lot.lot_number} • <StatusBadge status={lot.status} />
          </p>

          <div>
            <p className="text-sm text-neutral-500">Current Bid</p>
            <div className="bid-amount text-4xl font-heading text-brand-primary">${lot.current_price.toLocaleString()}</div>
          </div>

          {lot.status === "LIVE" && auctionStatus === "LIVE" && lot.ended_at ? (
            <div>
              <p className="text-sm text-neutral-500">Ends In</p>
              <div className="text-xl"><Countdown endTime={lot.ended_at} /></div>
            </div>
          ) : null}

          {isAuthenticated && !isAdmin && lot.status === "LIVE" && auctionStatus === "LIVE" ? (
            <BidInput currentBid={lot.current_price} currency="$" minIncrement={lot.min_increment} onPlaceBid={placeBid} />
          ) : isAuthenticated && !isAdmin && auctionStatus !== "LIVE" ? (
            <p className="text-sm text-neutral-500">Bidding is disabled because auction is not live.</p>
          ) : isAuthenticated && !isAdmin && lot.status !== "LIVE" ? (
            <p className="text-sm text-neutral-500">Bidding is disabled because this lot is not live.</p>
          ) : isAuthenticated && isAdmin ? (
            <p className="text-sm text-neutral-500">Admin view: bidding is disabled on public frontend.</p>
          ) : (
            <p className="text-sm">Login to place a bid.</p>
          )}

          <div className="history-panel">
            <BidHistory bids={formattedHistory} currency="$" />
          </div>
        </aside>
      </div>

      <div className="max-w-4xl space-y-8 md:space-y-12">
        <div>
          <h2 className="mb-4 font-heading text-2xl">Description</h2>
          <p className="leading-relaxed text-neutral-700">{lot.description}</p>
        </div>

        <div className="grid grid-cols-1 gap-6 rounded-lg bg-neutral-50 p-4 md:grid-cols-2 md:gap-8 md:p-6">
          <div>
            <h3 className="mb-3 text-base font-medium md:text-lg">Details</h3>
            <dl className="space-y-2">
              <div>
                <dt className="text-sm text-neutral-500">Medium</dt>
                <dd className="text-neutral-900">{medium}</dd>
              </div>
              <div>
                <dt className="text-sm text-neutral-500">Dimensions</dt>
                <dd className="text-neutral-900">{dimensions}</dd>
              </div>
              <div>
                <dt className="text-sm text-neutral-500">Date</dt>
                <dd className="text-neutral-900">{date}</dd>
              </div>
            </dl>
          </div>

          <div>
            <h3 className="mb-3 text-base font-medium md:text-lg">Provenance</h3>
            <p className="text-neutral-900">{provenance}</p>
          </div>
        </div>

        <div>
          <h2 className="mb-4 font-heading text-2xl">Condition Report</h2>
          <div className="rounded-lg bg-neutral-50 p-4 md:p-6">
            <p className="leading-relaxed text-neutral-700">{conditionReport}</p>
          </div>
        </div>
      </div>
    </section>
  );
}
