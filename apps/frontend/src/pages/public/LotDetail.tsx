import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
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
  const [relatedLots, setRelatedLots] = useState<Lot[]>([]);
  const [history, setHistory] = useState<LotBidHistory[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!lotId) return;
    try {
      const data = await api.getLot(lotId);
      setLot(data);
      const auction = await api.getAuction(data.auction_id);
      setAuctionStatus(auction.status);
      const lots = await api.listLotsByAuction(data.auction_id);
      setRelatedLots(lots.filter((l) => l.id !== data.id));
      const bids = await api.listLotBids(lotId);
      setHistory(
        bids.map((b) => ({
          bid_id: b.id,
          amount: b.amount,
          created_at: b.created_at,
        })),
      );
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
      setHistory((prev) => {
        if (prev.some((b) => b.bid_id === event.bid_id)) return prev;
        return [{ bid_id: event.bid_id, amount: event.amount, created_at: event.created_at }, ...prev];
      });
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
  const dynamicDetails = Array.isArray(meta.details)
    ? meta.details
        .map((d) => {
          if (!d || typeof d !== "object") return { label: "", value: "" };
          const rec = d as Record<string, unknown>;
          return {
            label: typeof rec.label === "string" ? rec.label : "",
            value: typeof rec.value === "string" ? rec.value : "",
          };
        })
        .filter((d) => d.label !== "" && d.value !== "")
    : [];
  const fallbackDetails = [
    { label: "Medium", value: typeof meta.medium === "string" ? meta.medium : "" },
    { label: "Dimensions", value: typeof meta.dimensions === "string" ? meta.dimensions : "" },
    { label: "Date", value: typeof meta.date === "string" ? meta.date : "" },
    { label: "Provenance", value: typeof meta.provenance === "string" ? meta.provenance : "" },
  ].filter((d) => d.value !== "");
  const details = dynamicDetails.length > 0 ? dynamicDetails : fallbackDetails;
  const conditionReport = typeof meta.condition_report === "string" ? meta.condition_report : "Condition report will be added soon.";

  return (
    <section className="mx-auto max-w-[1440px] space-y-10 px-4 py-6 md:space-y-12 md:px-8 md:py-12">
      <header className="mb-2">
        <Link to={lot ? `/auction/${lot.auction_id}` : "/auctions"} className="btn-link-brand">
          ← Back to Auction
        </Link>
      </header>

      <div className="grid grid-cols-1 gap-8 md:gap-12 lg:grid-cols-8">
        <div className="lg:col-span-5">
          <LotImageGallery images={images} alt={lot.title} />
        </div>

        <aside className="lg:col-span-3">
          <div className="space-y-6 lg:sticky lg:top-8">
            <div>
              <div className="mb-2 text-sm text-neutral-500">Lot {lot.lot_number}</div>
              <h1 className="font-heading text-xl sm:text-2xl md:text-3xl">{lot.title}</h1>
              <div className="mt-2"><StatusBadge status={lot.status} /></div>
            </div>

            <div className="border-t border-neutral-200 pt-6">
              <p className="mb-1 text-sm text-neutral-500">Current Bid</p>
              <div className="bid-amount text-[30px] font-heading font-bold leading-none text-brand-primary sm:text-[32px] md:text-[40px]">
                ${lot.current_price.toLocaleString()}
              </div>
            </div>

            {lot.status === "LIVE" && auctionStatus === "LIVE" && lot.ended_at ? (
              <div className="border-t border-neutral-200 pt-6">
                <p className="mb-1 text-sm text-neutral-500">Ends In</p>
                <div className="text-xl"><Countdown endTime={lot.ended_at} /></div>
              </div>
            ) : null}

            <div className="border-t border-neutral-200 pt-6">
              {isAuthenticated && !isAdmin && lot.status === "LIVE" && auctionStatus === "LIVE" ? (
                <BidInput currentBid={lot.current_price} currency="$" minIncrement={lot.min_increment} onPlaceBid={placeBid} />
              ) : isAuthenticated && !isAdmin && auctionStatus !== "LIVE" ? (
                <p className="text-sm text-neutral-500">Bidding will open when the auction goes live.</p>
              ) : isAuthenticated && !isAdmin && lot.status !== "LIVE" ? (
                <p className="text-sm text-neutral-500">Bidding will open when this lot goes live.</p>
              ) : isAuthenticated && isAdmin ? (
                <p className="text-sm text-neutral-500">Bidding is hidden while you are signed in as organizer.</p>
              ) : (
                <p className="text-sm">Sign in to place a bid.</p>
              )}
            </div>

            <div className="border-t border-neutral-200 pt-6">
              <BidHistory bids={formattedHistory} currency="$" />
            </div>
          </div>
        </aside>
      </div>

      <div className="max-w-4xl space-y-8 md:space-y-12">
        <div>
          <h2 className="mb-4 font-heading text-2xl">Description</h2>
          <p className="leading-relaxed text-neutral-700">{lot.description}</p>
        </div>

        <div className="rounded-lg bg-neutral-50 p-4 md:p-6">
          <h3 className="mb-3 text-base font-medium md:text-lg">Details</h3>
          {details.length > 0 ? (
            <dl className="grid grid-cols-1 gap-3 md:grid-cols-2">
              {details.map((detail, idx) => (
                <div key={`${detail.label}-${idx}`}>
                  <dt className="text-sm text-neutral-500">{detail.label}</dt>
                  <dd className="text-neutral-900">{detail.value}</dd>
                </div>
              ))}
            </dl>
          ) : (
            <p className="text-neutral-700">Details will be added soon.</p>
          )}
        </div>

        <div>
          <h2 className="mb-4 font-heading text-2xl">Condition Report</h2>
          <div className="rounded-lg bg-neutral-50 p-4 md:p-6">
            <p className="leading-relaxed text-neutral-700">{conditionReport}</p>
          </div>
        </div>

        {relatedLots.length > 0 ? (
          <div>
            <h2 className="mb-4 font-heading text-2xl">Other Lots in This Auction</h2>
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
              {relatedLots.map((l) => {
                const lotImages = l.metadata?.images?.length
                  ? l.metadata.images
                  : l.metadata?.image
                    ? [l.metadata.image]
                    : [];
                const image = lotImages[0];

                return (
                  <Link
                    key={l.id}
                    to={`/lots/${l.id}`}
                    className="group block overflow-hidden rounded-lg border border-neutral-200 bg-white transition-shadow hover:shadow-md"
                  >
                    <div className="aspect-[4/3] overflow-hidden bg-neutral-100">
                      {image ? (
                        <img src={image} alt={l.title} className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
                      ) : (
                        <div className="grid h-full place-content-center text-sm text-neutral-500">Image coming soon</div>
                      )}
                    </div>

                    <div className="space-y-3 p-5 md:p-6">
                      <div className="text-sm text-neutral-500">Lot {l.lot_number}</div>
                      <h3 className="font-heading text-lg font-medium text-neutral-900 md:text-xl">{l.title}</h3>
                      <div className="inline-flex"><StatusBadge status={l.status} /></div>
                      <div className="bid-amount text-xl font-heading text-brand-primary md:text-2xl">${l.current_price.toLocaleString()}</div>
                      {l.ended_at ? (
                        <div className="text-sm text-neutral-500">
                          Ends in <Countdown endTime={l.ended_at} />
                        </div>
                      ) : null}
                      <div className="flex items-center gap-2 pt-1 text-brand-primary">
                        <span className="font-medium">View</span>
                        <span className="transition-transform group-hover:translate-x-1">→</span>
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}
