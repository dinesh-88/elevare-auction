import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Link } from "react-router-dom";
import Countdown from "../../components/Countdown";
import LotImageGallery from "../../components/LotImageGallery";
import { api, type Auction, type Lot } from "../../api/client";
import { useAuctionSocket } from "../../hooks/useAuctionSocket";
import { useAuth } from "../../hooks/useAuth";
import { useLotsSockets } from "../../hooks/useLotsSockets";
import StatusBadge from "../../components/ui/StatusBadge";

export default function AuctionDetail() {
  const { id } = useParams();
  const { isAuthenticated, isAdmin } = useAuth();
  const [auction, setAuction] = useState<Auction | null>(null);
  const [auctionStatus, setAuctionStatus] = useState<"DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED" | null>(null);
  const [lots, setLots] = useState<Lot[]>([]);
  const [selectedLotId, setSelectedLotId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      let a: Auction | null = null;
      try {
        a = await api.getAuction(id);
      } catch {
        if (isAuthenticated && isAdmin) {
          const adminRows = await api.listAdminAuctions();
          a = adminRows.find((x) => x.id === id) ?? null;
        } else {
          throw new Error("Failed to load auction");
        }
      }
      if (!a) throw new Error("Auction not found");
      const l = await api.listLotsByAuction(id);
      setAuction(a);
      setAuctionStatus(a.status);
      setLots(l);
      if (!selectedLotId && l.length > 0) {
        const live = l.find((x) => x.status === "LIVE");
        setSelectedLotId((live ?? l[0]).id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load auction");
    }
  }, [id, isAuthenticated, isAdmin, selectedLotId]);

  useEffect(() => {
    load();
  }, [load]);

  useLotsSockets(
    lots.map((l) => l.id),
    (lotId, event) => {
      setLots((prev) =>
        prev.map((l) =>
          l.id === lotId
            ? {
                ...l,
                current_price: event.current_price,
                ended_at: event.ended_at ?? l.ended_at,
              }
            : l,
        ),
      );
    },
    (lotId, event) => {
      setLots((prev) => prev.map((l) => (l.id === lotId ? { ...l, status: event.status } : l)));
    },
    lots.length > 0,
  );

  useAuctionSocket(
    auction?.id,
    (event) => {
      setAuctionStatus(event.status);
      setAuction((prev) => (prev ? { ...prev, status: event.status } : prev));
    },
    Boolean(auction?.id),
  );

  useEffect(() => {
    if (lots.length === 0) {
      setSelectedLotId(null);
      return;
    }
    if (!selectedLotId || !lots.some((l) => l.id === selectedLotId)) {
      const live = lots.find((l) => l.status === "LIVE");
      setSelectedLotId((live ?? lots[0]).id);
    }
  }, [lots, selectedLotId]);

  if (error) return <p className="error">{error}</p>;
  if (!auction) return <p className="text-sm text-neutral-500">Loading auction...</p>;
  if (lots.length === 0) {
    return (
      <section className="page-shell">
        <header className="panel">
          <h1 className="font-heading text-3xl text-neutral-900">{auction.title}</h1>
          <p className="mt-2 text-sm leading-6 text-neutral-600">{auction.description}</p>
          <p className="mt-4 text-sm text-neutral-500">No lots available for this auction yet.</p>
        </header>
      </section>
    );
  }

  const selectedLot = lots.find((l) => l.id === selectedLotId) ?? lots[0];
  const lotImages =
    selectedLot.metadata?.images?.length
      ? selectedLot.metadata.images
      : selectedLot.metadata?.image
        ? [selectedLot.metadata.image]
        : [];
  const auctionImages =
    auction.metadata?.images?.length
      ? auction.metadata.images
      : auction.metadata?.image
        ? [auction.metadata.image]
        : [];
  const selectedImages = lotImages.length > 0 ? lotImages : auctionImages;

  return (
    <section className="mx-auto max-w-[1440px] px-4 py-6 md:px-8 md:py-12">
      <header className="mb-8 flex flex-col items-start justify-between gap-6 sm:flex-row">
        <div>
          <h1 className="font-heading text-2xl sm:text-3xl md:text-4xl">{auction.title}</h1>
          <p className="mt-2 text-neutral-500">{auction.description}</p>
          <div className="mt-2 text-sm text-neutral-500">
            Event ends in <Countdown endTime={auction.end_time} />
          </div>
        </div>
        <div className="space-y-2 sm:text-right">
          {auctionStatus ? <StatusBadge status={auctionStatus} /> : null}
          {isAuthenticated && isAdmin ? <div className="text-xs text-neutral-500">Signed in as organizer</div> : null}
        </div>
      </header>

      <div className="mb-12 grid grid-cols-1 gap-8 md:gap-12 lg:grid-cols-8">
        <div className="lg:col-span-5">
          <LotImageGallery images={selectedImages} alt={selectedLot.title} />
        </div>

        <aside className="lg:col-span-3">
          <div className="space-y-6 lg:sticky lg:top-8">
            <div>
              <div className="text-sm text-neutral-500">Lot {selectedLot.lot_number}</div>
              <h2 className="mt-1 font-heading text-2xl md:text-3xl">{selectedLot.title}</h2>
              <div className="mt-2"><StatusBadge status={selectedLot.status} /></div>
            </div>

            <div className="border-t border-neutral-200 pt-5">
              <p className="text-sm text-neutral-500">Current Bid</p>
              <div className="bid-amount mt-1 text-[34px] font-heading leading-none text-brand-primary md:text-[40px]">
                ${selectedLot.current_price.toLocaleString()}
              </div>
              {selectedLot.ended_at ? (
                <div className="mt-3 text-sm text-neutral-500">
                  Ends in <Countdown endTime={selectedLot.ended_at} />
                </div>
              ) : null}
            </div>

            <div className="border-t border-neutral-200 pt-5">
              <p className="text-sm text-neutral-600">{selectedLot.description}</p>
              {!isAuthenticated ? (
                <div className="mt-4 space-y-2">
                  <p className="text-sm text-neutral-500">Sign in to place a bid on this lot.</p>
                  <Link to="/login" className="inline-flex rounded-lg bg-brand-primary px-8 py-3 font-medium text-white hover:opacity-90">
                    Sign In
                  </Link>
                </div>
              ) : isAdmin ? (
                <div className="mt-4 space-y-2">
                  <p className="text-sm text-neutral-500">Bidding is hidden while you are signed in as organizer.</p>
                  <Link to={`/lots/${selectedLot.id}`} className="inline-flex rounded-lg bg-brand-primary px-8 py-3 font-medium text-white hover:opacity-90">
                    View Lot Details
                  </Link>
                </div>
              ) : (
                <div className="mt-4 space-y-2">
                
                  <Link to={`/lots/${selectedLot.id}`} className="inline-flex rounded-lg bg-brand-primary px-8 py-3 font-medium text-white hover:opacity-90">
                    View Lot Details
                  </Link>
                </div>
              )}
            </div>
          </div>
        </aside>
      </div>

      <div>
        <h3 className="mb-4 font-heading text-2xl md:text-3xl">Lots in This Auction</h3>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {lots.map((lot) => {
            const lotCardImages =
              lot.metadata?.images?.length
                ? lot.metadata.images
                : lot.metadata?.image
                  ? [lot.metadata.image]
                  : [];
            const image = lotCardImages[0];

            return (
              <button
                key={lot.id}
                type="button"
                onClick={() => setSelectedLotId(lot.id)}
                className={`group overflow-hidden rounded-lg border bg-white text-left transition duration-200 hover:-translate-y-0.5 hover:shadow-md ${
                  selectedLot.id === lot.id ? "border-brand-primary ring-1 ring-brand-primary/30" : "border-neutral-200"
                }`}
              >
                <div className="aspect-[4/3] overflow-hidden bg-neutral-100">
                  {image ? (
                    <img src={image} alt={lot.title} className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
                  ) : (
                    <div className="grid h-full place-content-center text-sm text-neutral-500">Image coming soon</div>
                  )}
                </div>

                <div className="space-y-3 p-5">
                  <div className="text-xs text-neutral-500">Lot {lot.lot_number}</div>
                  <h4 className="font-heading text-lg text-neutral-900">{lot.title}</h4>
                  <div className="inline-flex"><StatusBadge status={lot.status} /></div>
                  <div className="bid-amount text-xl text-brand-primary">${lot.current_price.toLocaleString()}</div>
                  <div className="flex items-center gap-2 pt-1 text-brand-primary">
                    <span className="font-medium">{selectedLot.id === lot.id ? "Selected" : "View"}</span>
                    <span className="transition-transform group-hover:translate-x-1">→</span>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </section>
  );
}
