"use client";

import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Countdown from "../../components/Countdown";
import BidHistory from "../../components/BidHistory";
import BidInput from "../../components/BidInput";
import LotImageGallery from "../../components/LotImageGallery";
import { api, type Lot, type LotFinancials } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { useAuctionSocket } from "../../hooks/useAuctionSocket";
import { useLotSocket } from "../../hooks/useLotSocket";
import StatusBadge from "../../components/ui/StatusBadge";
import { currencySymbol } from "../../utils/currency";

interface LotBidHistory {
  bid_id: string;
  amount_minor: number;
  created_at: string;
  bidder_display: string;
}

export default function LotDetail({
  lotID,
  initialLot,
  initialAuction,
  initialRelatedLots,
  initialBids,
}: {
  lotID?: string;
  initialLot?: Lot | null;
  initialAuction?: { id: string; status: "DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED"; currency: string } | null;
  initialRelatedLots?: Lot[];
  initialBids?: LotBidHistory[];
}) {
  const { lotId } = useParams();
  const routeLotID = lotID ?? lotId;
  const { isAuthenticated, isAdmin, user } = useAuth();
  const [lot, setLot] = useState<Lot | null>(initialLot ?? null);
  const [auctionCurrency, setAuctionCurrency] = useState(initialAuction?.currency ?? "USD");
  const [auctionStatus, setAuctionStatus] = useState<"DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED" | null>(
    initialAuction?.status ?? null,
  );
  const [relatedLots, setRelatedLots] = useState<Lot[]>(initialRelatedLots ?? []);
  const [history, setHistory] = useState<LotBidHistory[]>(initialBids ?? []);
  const [financials, setFinancials] = useState<LotFinancials | null>(null);
  const [canViewFinancials, setCanViewFinancials] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [bidMessage, setBidMessage] = useState<string | null>(null);
  const [requestToBidStatus, setRequestToBidStatus] = useState<"PENDING" | "APPROVED" | "REJECTED" | null>(null);
  const [extendedBadge, setExtendedBadge] = useState(false);

  const applyBidResponse = (raw: unknown) => {
    const record = (raw ?? {}) as Record<string, unknown>;
    const currentPrice = Number(record.current_price ?? 0);
    const endedAt = (record.ended_at ?? null) as string | null;
    const bid = (record.bid ?? {}) as Record<string, unknown>;
    const bidID = String(bid.id ?? "");
    const amountMinor = Number(bid.amount_minor ?? 0);
    const createdAt = String(bid.created_at ?? new Date().toISOString());

    if (Number.isFinite(currentPrice) && currentPrice > 0) {
      setLot((prev) =>
        prev
          ? {
              ...prev,
              current_price: currentPrice,
              minimum_next_bid_minor: currentPrice + (prev.increment_minor ?? prev.min_increment),
              ended_at: endedAt ?? prev.ended_at,
            }
          : prev,
      );
    }
    if (bidID) {
      setHistory((prev) => {
        if (prev.some((item) => item.bid_id === bidID)) return prev;
        return [{ bid_id: bidID, amount_minor: amountMinor, created_at: createdAt, bidder_display: "Bidder" }, ...prev];
      });
    }
  };

  const load = useCallback(async () => {
    if (!routeLotID) return;
    try {
      const data = await api.getLot(routeLotID);
      setLot(data);
      const auction = await api.getAuction(data.auction_id);
      setAuctionStatus(auction.status);
      setAuctionCurrency(auction.currency || "USD");
      const lots = await api.listLotsByAuction(data.auction_id);
      setRelatedLots(lots.filter((l) => l.id !== data.id));
      const bids = await api.listLotBids(routeLotID);
      setHistory(
        bids.map((b) => ({
          bid_id: b.id,
          amount_minor: b.amount_minor,
          created_at: b.created_at,
          bidder_display: b.bidder_display,
        })),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load lot");
    }
  }, [routeLotID]);

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

  useEffect(() => {
    if (!routeLotID || !lot || lot.status !== "ENDED" || !isAuthenticated) {
      setFinancials(null);
      setCanViewFinancials(false);
      return;
    }
    let cancelled = false;
    api
      .getLotFinancials(routeLotID)
      .then((data) => {
        if (cancelled) return;
        setFinancials(data);
        setCanViewFinancials(true);
      })
      .catch(() => {
        if (cancelled) return;
        setFinancials(null);
        setCanViewFinancials(false);
      });
    return () => {
      cancelled = true;
    };
  }, [routeLotID, lot, isAuthenticated]);

  useLotSocket(
    routeLotID,
    (event) => {
      setLot((prev) =>
        prev
          ? {
              ...prev,
              current_price: event.current_price,
              minimum_next_bid_minor: event.current_price + (prev.increment_minor ?? prev.min_increment),
              ended_at: event.ended_at ?? prev.ended_at,
            }
          : prev,
      );
      setHistory((prev) => {
        if (prev.some((b) => b.bid_id === event.bid_id)) return prev;
        return [
          {
            bid_id: event.bid_id,
            amount_minor: event.amount,
            created_at: event.created_at,
            bidder_display: event.bidder_display || "Bidder",
          },
          ...prev,
        ];
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
    (event) => {
      setLot((prev) => (prev ? { ...prev, ended_at: event.new_end_time } : prev));
      setExtendedBadge(true);
      window.setTimeout(() => setExtendedBadge(false), 5000);
    },
    true,
  );

  useAuctionSocket(
    lot?.auction_id,
    (event) => {
      setAuctionStatus(event.status);
      void load();
    },
    Boolean(lot?.auction_id),
  );

  const placeQuickBid = async (nextAmount: number) => {
    if (!routeLotID) return;
    try {
      const res = await api.placeLotBid(routeLotID, { bid_amount_minor: nextAmount, max_bid_minor: null });
      applyBidResponse(res);
      setBidMessage("Bid placed successfully");
      window.setTimeout(() => setBidMessage(null), 2500);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to place bid";
      if (message.toLowerCase().includes("approval")) {
        setRequestToBidStatus("PENDING");
      }
      throw new Error(message);
    }
  };

  const placeMaxBid = async (maxBidMinor: number) => {
    if (!routeLotID) return;
    try {
      const res = await api.placeLotBid(routeLotID, { bid_amount_minor: null, max_bid_minor: maxBidMinor });
      applyBidResponse(res);
      setBidMessage("Bid placed successfully");
      window.setTimeout(() => setBidMessage(null), 2500);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to place bid";
      if (message.toLowerCase().includes("approval")) {
        setRequestToBidStatus("PENDING");
      }
      throw new Error(message);
    }
  };

  const requestToBid = async () => {
    if (!lot) return;
    try {
      const res = await api.requestToBid(lot.auction_id);
      setRequestToBidStatus(res.status);
      setBidMessage(res.status === "APPROVED" ? "You are approved to bid." : "Bid request submitted.");
      window.setTimeout(() => setBidMessage(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to request approval");
    }
  };

  const downloadInvoice = async () => {
    if (!financials?.invoice_id) return;
    try {
      const blob = await api.downloadInvoicePDF(financials.invoice_id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `invoice-${financials.invoice_id}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to download invoice");
    }
  };

  const generateInvoiceManually = async () => {
    if (!routeLotID || !isAdmin) return;
    try {
      setError(null);
      const inv = await api.generateLotInvoice(routeLotID);
      setFinancials((prev) => (prev ? { ...prev, invoice_id: inv.id } : prev));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate invoice");
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
    amount: b.amount_minor,
    bidder: b.bidder_display || "Bidder",
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
  const isEnded = lot.status === "ENDED";
  const hammerPrice = lot.hammer_price ?? null;
  const isWinner = Boolean(user && financials?.winner_user && user.id === financials.winner_user);
  const symbol = currencySymbol(auctionCurrency);
  const minimumNextBid = lot.current_price + (lot.increment_minor ?? lot.min_increment);

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
                {symbol}
                {lot.current_price.toLocaleString()}
              </div>
              {isEnded && hammerPrice !== null ? (
                <div className="mt-3 space-y-1 text-sm">
                  <p className="font-medium text-neutral-900">
                    Winning Bid (Hammer): {symbol}
                    {hammerPrice.toLocaleString()}
                  </p>
                  {lot.sale_result ? <p className="text-neutral-600">Result: {lot.sale_result === "SOLD" ? "Sold" : "Reserve Not Met"}</p> : null}
                  <p className="text-neutral-500">Buyer&apos;s premium may apply.</p>
                </div>
              ) : null}
            </div>

            {lot.status === "LIVE" && auctionStatus === "LIVE" && lot.ended_at ? (
              <div className="border-t border-neutral-200 pt-6">
                <p className="mb-1 text-sm text-neutral-500">Ends In</p>
                <div className="text-xl"><Countdown endTime={lot.ended_at} /></div>
                {extendedBadge ? <p className="mt-2 text-sm font-medium text-amber-600">Auction Extended</p> : null}
              </div>
            ) : null}

            <div className="border-t border-neutral-200 pt-6">
              {isAuthenticated && !isAdmin && lot.status === "LIVE" && auctionStatus === "LIVE" ? (
                <>
                  <p className="mb-1 text-sm text-neutral-500">
                    Minimum Next Bid: {symbol}{minimumNextBid.toLocaleString()}
                  </p>
                  <BidInput minimumNextBid={minimumNextBid} currency={symbol} onQuickBid={placeQuickBid} onMaxBid={placeMaxBid} />
                  {requestToBidStatus !== "APPROVED" ? (
                    <div className="mt-3">
                      <button type="button" className="btn-secondary" onClick={requestToBid}>
                        {requestToBidStatus === "PENDING" ? "Pending approval" : "Request to bid"}
                      </button>
                    </div>
                  ) : null}
                </>
              ) : isAuthenticated && !isAdmin && auctionStatus !== "LIVE" ? (
                <p className="text-sm text-neutral-500">Bidding will open when the auction goes live.</p>
              ) : isAuthenticated && !isAdmin && lot.status !== "LIVE" ? (
                <p className="text-sm text-neutral-500">Bidding will open when this lot goes live.</p>
              ) : isAuthenticated && isAdmin ? (
                <p className="text-sm text-neutral-500">Bidding is hidden while you are signed in as organizer.</p>
              ) : (
                <p className="text-sm">Sign in to place a bid.</p>
              )}
              {bidMessage ? <p className="mt-2 text-sm text-green-700">{bidMessage}</p> : null}
            </div>

            <div className="border-t border-neutral-200 pt-6">
              <BidHistory bids={formattedHistory} currency={symbol} />
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

        {isEnded && canViewFinancials && financials ? (
          <div>
            <h2 className="mb-4 font-heading text-2xl">{isAdmin ? "Settlement Breakdown" : isWinner ? "Your Settlement Breakdown" : "Financial Breakdown"}</h2>
            <div className="rounded-lg bg-neutral-50 p-4 md:p-6">
              {financials.hammer_price === null || financials.hammer_price === undefined ? (
                <p className="text-sm text-neutral-600">No winning bid was recorded for this lot.</p>
              ) : (
                <>
                  <div className="mb-3 text-neutral-700">
                    Winning Bid:{" "}
                    <span className="bid-amount font-medium text-neutral-900">
                      {symbol}
                      {financials.hammer_price.toLocaleString()}
                    </span>
                  </div>
                  {financials.fees?.length ? (
                    <dl className="space-y-2">
                      {financials.fees.map((fee, idx) => (
                        <div key={`${fee.name}-${idx}`} className="flex items-center justify-between gap-4 text-sm">
                          <dt className="text-neutral-600">
                            {fee.name}
                            {fee.type === "PERCENTAGE" ? ` (${fee.value}%)` : ""}
                          </dt>
                          <dd className="bid-amount text-neutral-900">
                            {symbol}
                            {fee.amount.toLocaleString()}
                          </dd>
                        </div>
                      ))}
                    </dl>
                  ) : (
                    <p className="text-sm text-neutral-600">No additional fees.</p>
                  )}
                  <div className="mt-4 border-t border-neutral-200 pt-4 text-base font-medium text-neutral-900">
                    Total Due:{" "}
                    <span className="bid-amount">
                      {symbol}
                      {(financials.total_due ?? 0).toLocaleString()}
                    </span>
                  </div>
                  <div className="mt-4 flex flex-wrap gap-2">
                    {financials.invoice_id ? (
                      <button type="button" className="btn-brand" onClick={downloadInvoice}>
                        Download Invoice (PDF)
                      </button>
                    ) : null}
                    {isAdmin && !financials.invoice_id && financials.hammer_price != null ? (
                      <button type="button" className="btn-brand" onClick={generateInvoiceManually}>
                        Generate Invoice
                      </button>
                    ) : null}
                  </div>
                </>
              )}
            </div>
          </div>
        ) : null}

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
                    className="group block overflow-hidden rounded-lg border border-neutral-200 bg-white outline-none transition-shadow hover:shadow-md focus-visible:border-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-300/70"
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
                      <div className="bid-amount text-xl font-heading text-brand-primary md:text-2xl">
                        {symbol}
                        {l.current_price.toLocaleString()}
                      </div>
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
