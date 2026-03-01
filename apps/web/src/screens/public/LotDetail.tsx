"use client";

import { Suspense, lazy, useCallback, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import Countdown from "../../components/Countdown";
import BidInput from "../../components/BidInput";
import LotImageGallery from "../../components/LotImageGallery";
import { api, type Lot, type LotFinancials } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { useAuctionSocket } from "../../hooks/useAuctionSocket";
import { useLotSocket } from "../../hooks/useLotSocket";
import StatusBadge from "../../components/ui/StatusBadge";
import { currencySymbol } from "../../utils/currency";

const LazyBidHistory = lazy(() => import("../../components/BidHistory"));

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
  const params = useParams<{ id?: string; lotId?: string }>();
  const navigate = useNavigate();
  const routeLotID = lotID ?? params.lotId;
  const routeAuctionSlug = params.id ?? initialAuction?.id ?? "";
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
  const [termsAccepted, setTermsAccepted] = useState(true);
  const [lastPlacedBidAmount, setLastPlacedBidAmount] = useState<number | null>(null);
  const [outbidNotice, setOutbidNotice] = useState<string | null>(null);
  const [openSections, setOpenSections] = useState({
    description: true,
    details: false,
    condition: false,
    provenance: false,
    history: false,
  });

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
    void load();
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

  useEffect(() => {
    if (!lot || !isAuthenticated || isAdmin) {
      setTermsAccepted(true);
      return;
    }
    let cancelled = false;
    api
      .getAuctionTermsStatus(lot.auction_id)
      .then((status) => {
        if (cancelled) return;
        if (!status.has_active_terms) {
          setTermsAccepted(false);
          return;
        }
        setTermsAccepted(status.accepted);
      })
      .catch(() => {
        if (cancelled) return;
        setTermsAccepted(true);
      });
    return () => {
      cancelled = true;
    };
  }, [lot, isAuthenticated, isAdmin]);

  const redirectToTerms = () => {
    const slug = routeAuctionSlug || lot?.auction_id || "";
    const returnTo = typeof window !== "undefined" ? window.location.pathname : `/auction/${slug}`;
    navigate(`/auction/${slug}/terms?returnTo=${encodeURIComponent(returnTo)}`);
  };

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

  const flashBidMessage = (message: string, timeout = 2500) => {
    setBidMessage(message);
    window.setTimeout(() => setBidMessage(null), timeout);
  };

  const placeQuickBid = async (nextAmount: number) => {
    if (!routeLotID) return;
    try {
      const res = await api.placeLotBid(routeLotID, { bid_amount_minor: nextAmount, max_bid_minor: null });
      applyBidResponse(res);
      setLastPlacedBidAmount(nextAmount);
      setOutbidNotice(null);
      flashBidMessage("Bid placed successfully");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to place bid";
      if (message.toLowerCase().includes("approval")) {
        setRequestToBidStatus("PENDING");
      }
      if (message.toLowerCase().includes("terms")) {
        redirectToTerms();
      }
      throw new Error(message);
    }
  };

  const placeMaxBid = async (maxBidMinor: number) => {
    if (!routeLotID) return;
    try {
      const res = await api.placeLotBid(routeLotID, { bid_amount_minor: null, max_bid_minor: maxBidMinor });
      applyBidResponse(res);
      setLastPlacedBidAmount(maxBidMinor);
      setOutbidNotice(null);
      flashBidMessage("Bid placed successfully");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to place bid";
      if (message.toLowerCase().includes("approval")) {
        setRequestToBidStatus("PENDING");
      }
      if (message.toLowerCase().includes("terms")) {
        redirectToTerms();
      }
      throw new Error(message);
    }
  };

  const requestToBid = async () => {
    if (!lot) return;
    try {
      const res = await api.requestToBid(lot.auction_id);
      setRequestToBidStatus(res.status);
      flashBidMessage(res.status === "APPROVED" ? "You are approved to bid." : "Bid request submitted.", 3000);
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

  useEffect(() => {
    if (!lot || lastPlacedBidAmount === null) return;
    const nextMinimum = lot.current_price + (lot.increment_minor ?? lot.min_increment);
    if (lot.current_price > lastPlacedBidAmount) {
      setOutbidNotice(`Another bidder moved ahead. The next minimum is ${currencySymbol(auctionCurrency)}${nextMinimum.toLocaleString()}.`);
      return;
    }
    setOutbidNotice(null);
  }, [lot, lastPlacedBidAmount, auctionCurrency]);

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
  ].filter((d) => d.value !== "");
  const details = dynamicDetails.length > 0 ? dynamicDetails : fallbackDetails;
  const conditionReport = typeof meta.condition_report === "string" ? meta.condition_report : "Condition report will be added soon.";
  const provenance = typeof meta.provenance === "string" && meta.provenance.trim() ? meta.provenance : "Provenance details will be added soon.";
  const isEnded = lot.status === "ENDED";
  const hammerPrice = lot.hammer_price ?? null;
  const isWinner = Boolean(user && financials?.winner_user && user.id === financials.winner_user);
  const symbol = currencySymbol(auctionCurrency);
  const minimumNextBid = lot.current_price + (lot.increment_minor ?? lot.min_increment);
  const isLiveForBidding = lot.status === "LIVE" && auctionStatus === "LIVE";
  const showBidControls = isAuthenticated && !isAdmin && isLiveForBidding;
  const userBidStatus =
    !isAuthenticated
      ? { label: "Sign In to Bid", className: "bg-neutral-100 text-neutral-700" }
      : isAdmin
        ? { label: "Organizer View", className: "bg-neutral-100 text-neutral-700" }
        : isEnded
          ? { label: isWinner ? "Won Lot" : "Auction Closed", className: isWinner ? "bg-emerald-100 text-emerald-800" : "bg-neutral-100 text-neutral-700" }
          : !termsAccepted
            ? { label: "Terms Required", className: "bg-amber-100 text-amber-900" }
            : requestToBidStatus === "PENDING"
              ? { label: "Approval Pending", className: "bg-amber-100 text-amber-900" }
              : outbidNotice
                ? { label: "Outbid", className: "bg-rose-100 text-rose-800" }
                : lastPlacedBidAmount !== null && lot.current_price <= lastPlacedBidAmount
                  ? { label: "Leading", className: "bg-emerald-100 text-emerald-800" }
                  : isLiveForBidding
                    ? { label: "Ready to Bid", className: "bg-sky-100 text-sky-800" }
                    : { label: "Watching", className: "bg-neutral-100 text-neutral-700" };

  const toggleSection = (key: keyof typeof openSections) => {
    setOpenSections((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  return (
    <section className="mx-auto max-w-[1440px] space-y-8 px-4 py-6 pb-32 md:space-y-12 md:px-8 md:py-12 md:pb-12">
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
          <div className="space-y-5 lg:sticky lg:top-8">
            <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="mb-2 text-sm text-neutral-500">Lot {lot.lot_number}</div>
                  <h1 className="font-heading text-2xl leading-tight sm:text-3xl">{lot.title}</h1>
                </div>
                <StatusBadge status={lot.status} />
              </div>
              <div className="mt-4 flex flex-wrap items-center gap-2">
                <span className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold ${userBidStatus.className}`}>{userBidStatus.label}</span>
                {extendedBadge ? <span className="inline-flex rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-900">Extended</span> : null}
              </div>
            </div>

            <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
              <div className="grid gap-4 sm:grid-cols-2 sm:gap-6 lg:grid-cols-1">
                <div>
                  <p className="mb-1 text-xs font-semibold uppercase tracking-[0.12em] text-neutral-500">Current Bid</p>
                  <div className="bid-amount text-[34px] font-heading font-bold leading-none text-brand-primary sm:text-[40px]">
                    {symbol}
                    {lot.current_price.toLocaleString()}
                  </div>
                </div>
                {isLiveForBidding && lot.ended_at ? (
                  <div>
                    <p className="mb-1 text-xs font-semibold uppercase tracking-[0.12em] text-neutral-500">Time Remaining</p>
                    <div className="text-xl font-semibold text-neutral-900"><Countdown endTime={lot.ended_at} /></div>
                  </div>
                ) : null}
              </div>
              {outbidNotice ? (
                <div className="mt-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-medium text-amber-900">
                  {outbidNotice}
                </div>
              ) : null}
              {bidMessage ? <p className="mt-3 text-sm text-green-700">{bidMessage}</p> : null}
              {isEnded && hammerPrice !== null ? (
                <div className="mt-4 space-y-1 border-t border-neutral-200 pt-4 text-sm">
                  <p className="font-medium text-neutral-900">
                    Winning Bid (Hammer): {symbol}
                    {hammerPrice.toLocaleString()}
                  </p>
                  {lot.sale_result ? <p className="text-neutral-600">Result: {lot.sale_result === "SOLD" ? "Sold" : "Reserve Not Met"}</p> : null}
                  <p className="text-neutral-500">Buyer&apos;s premium may apply.</p>
                </div>
              ) : null}
            </div>

            <div className={`hidden rounded-3xl border bg-white p-5 shadow-sm md:block ${outbidNotice ? "border-amber-300 ring-2 ring-amber-200" : "border-neutral-200"}`}>
              {showBidControls ? (
                <>
                  {termsAccepted ? (
                    <>
                      <BidInput
                        minimumNextBid={minimumNextBid}
                        currency={symbol}
                        onQuickBid={placeQuickBid}
                        onMaxBid={placeMaxBid}
                        highlight={Boolean(outbidNotice)}
                        outbidNotice={outbidNotice}
                      />
                      {requestToBidStatus !== "APPROVED" ? (
                        <div className="mt-3">
                          <button type="button" className="btn-secondary" onClick={requestToBid}>
                            {requestToBidStatus === "PENDING" ? "Pending approval" : "Request to bid"}
                          </button>
                        </div>
                      ) : null}
                    </>
                  ) : (
                    <div className="space-y-3 rounded-lg border border-amber-200 bg-amber-50 p-4">
                      <p className="text-sm text-amber-900">You must accept the active auction terms before placing a bid.</p>
                      <button type="button" className="btn-brand" onClick={redirectToTerms}>
                        Review Terms
                      </button>
                    </div>
                  )}
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
            </div>

            <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
              <AccordionSection title="Bid History" subtitle={`${formattedHistory.length} bids`} isOpen={openSections.history} onToggle={() => toggleSection("history")}>
                <Suspense fallback={<p className="text-sm text-neutral-500">Loading bid history...</p>}>
                  <LazyBidHistory bids={formattedHistory} currency={symbol} />
                </Suspense>
              </AccordionSection>
            </div>
          </div>
        </aside>
      </div>

      <div className="max-w-4xl space-y-4 md:space-y-5">
        <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
          <AccordionSection title="Description" isOpen={openSections.description} onToggle={() => toggleSection("description")}>
            <p className="leading-relaxed text-neutral-700">{lot.description}</p>
          </AccordionSection>
        </div>

        <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
          <AccordionSection title="Details" isOpen={openSections.details} onToggle={() => toggleSection("details")}>
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
          </AccordionSection>
        </div>

        <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
          <AccordionSection title="Condition" isOpen={openSections.condition} onToggle={() => toggleSection("condition")}>
            <p className="leading-relaxed text-neutral-700">{conditionReport}</p>
          </AccordionSection>
        </div>

        <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
          <AccordionSection title="Provenance" isOpen={openSections.provenance} onToggle={() => toggleSection("provenance")}>
            <p className="leading-relaxed text-neutral-700">{provenance}</p>
          </AccordionSection>
        </div>

        {isEnded && canViewFinancials && financials ? (
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
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
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="mb-4 font-heading text-2xl">Other Lots in This Auction</h2>
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
              {relatedLots.map((relatedLot) => {
                const lotImages = relatedLot.metadata?.images?.length
                  ? relatedLot.metadata.images
                  : relatedLot.metadata?.image
                    ? [relatedLot.metadata.image]
                    : [];
                const image = lotImages[0];

                return (
                  <Link
                    key={relatedLot.id}
                    to={`/lots/${relatedLot.id}`}
                    className="group block overflow-hidden rounded-2xl border border-neutral-200 bg-white outline-none transition-shadow hover:shadow-md focus-visible:border-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-300/70"
                  >
                    <div className="aspect-[4/3] overflow-hidden bg-neutral-100">
                      {image ? (
                        <img
                          src={image}
                          alt={relatedLot.title}
                          loading="lazy"
                          decoding="async"
                          className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                        />
                      ) : (
                        <div className="grid h-full place-content-center text-sm text-neutral-500">Image coming soon</div>
                      )}
                    </div>

                    <div className="space-y-3 p-5 md:p-6">
                      <div className="text-sm text-neutral-500">Lot {relatedLot.lot_number}</div>
                      <h3 className="font-heading text-lg font-medium text-neutral-900 md:text-xl">{relatedLot.title}</h3>
                      <div className="inline-flex"><StatusBadge status={relatedLot.status} /></div>
                      <div className="bid-amount text-xl font-heading text-brand-primary md:text-2xl">
                        {symbol}
                        {relatedLot.current_price.toLocaleString()}
                      </div>
                      {relatedLot.ended_at ? (
                        <div className="text-sm text-neutral-500">
                          Ends in <Countdown endTime={relatedLot.ended_at} />
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

      {showBidControls ? (
        <div className="fixed inset-x-0 bottom-0 z-[60] md:hidden">
          {termsAccepted ? (
            <BidInput
              minimumNextBid={minimumNextBid}
              currency={symbol}
              onQuickBid={placeQuickBid}
              onMaxBid={placeMaxBid}
              mobileSticky
              highlight={Boolean(outbidNotice)}
              outbidNotice={outbidNotice}
            />
          ) : (
            <div className="rounded-t-3xl border border-neutral-900 bg-neutral-950 px-4 pb-[max(1rem,env(safe-area-inset-bottom))] pt-4 text-white shadow-[0_-10px_40px_rgba(0,0,0,0.35)]">
              <p className="text-sm text-white/80">Review the current auction terms before you place a bid.</p>
              <button type="button" className="mt-3 min-h-12 w-full rounded-xl bg-amber-300 px-4 font-semibold text-neutral-950" onClick={redirectToTerms}>
                Review Terms
              </button>
            </div>
          )}
        </div>
      ) : null}
    </section>
  );
}

function AccordionSection({
  title,
  subtitle,
  isOpen,
  onToggle,
  children,
}: {
  title: string;
  subtitle?: string;
  isOpen: boolean;
  onToggle: () => void;
  children: React.ReactNode;
}) {
  return (
    <div>
      <button type="button" className="flex w-full items-center justify-between gap-4 text-left" onClick={onToggle} aria-expanded={isOpen}>
        <div>
          <h2 className="font-heading text-xl text-neutral-900">{title}</h2>
          {subtitle ? <p className="mt-1 text-sm text-neutral-500">{subtitle}</p> : null}
        </div>
        <span className="text-lg text-neutral-500">{isOpen ? "−" : "+"}</span>
      </button>
      {isOpen ? <div className="mt-4">{children}</div> : null}
    </div>
  );
}
