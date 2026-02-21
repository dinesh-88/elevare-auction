import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Link } from "react-router-dom";
import Countdown from "../../components/Countdown";
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

  const selectedImages = selectedLot.metadata?.images?.length
    ? selectedLot.metadata.images
    : selectedLot.metadata?.image
      ? [selectedLot.metadata.image]
      : [];

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">{auction.title}</h1>
          <p className="section-copy">{auction.description}</p>
          <div className="mt-2 text-sm text-neutral-500">
            Event ends in <Countdown endTime={auction.end_time} />
          </div>
        </div>
        {auctionStatus ? <StatusBadge status={auctionStatus} /> : null}
        {isAuthenticated && isAdmin ? (
          <div className="text-xs text-neutral-500">Admin view: public bidding controls are disabled.</div>
        ) : null}
      </header>

      <div className="grid grid-cols-1 gap-8 lg:grid-cols-8 lg:gap-12">
        <div className="lg:col-span-5">
          <div className="space-y-4">
            <div className="aspect-square overflow-hidden rounded-lg border border-neutral-200 bg-neutral-100">
              {selectedImages.length > 0 ? (
                <img src={selectedImages[0]} alt={selectedLot.title} className="h-full w-full object-cover" />
              ) : (
                <div className="grid h-full place-content-center text-sm text-neutral-500">No image</div>
              )}
            </div>
            {selectedImages.length > 1 ? (
              <div className="grid grid-cols-5 gap-2">
                {selectedImages.map((img, idx) => (
                  <div key={`${img}-${idx}`} className="aspect-square overflow-hidden rounded-md border border-neutral-200 bg-neutral-100">
                    <img src={img} alt={`${selectedLot.title} ${idx + 1}`} className="h-full w-full object-cover" />
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        </div>

        <aside className="lg:col-span-3">
          <div className="panel space-y-5 lg:sticky lg:top-8">
            <div>
              <div className="text-sm text-neutral-500">Lot {selectedLot.lot_number}</div>
              <h2 className="mt-1 font-heading text-2xl text-neutral-900">{selectedLot.title}</h2>
              <div className="mt-2"><StatusBadge status={selectedLot.status} /></div>
            </div>

            <div className="border-t border-neutral-200 pt-5">
              <p className="text-sm text-neutral-500">Current Bid</p>
              <div className="bid-amount mt-1 text-[40px] font-heading leading-none text-brand-primary">
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
              <Link to={`/lots/${selectedLot.id}`} className="btn-brand mt-4 inline-flex">
                Open Lot Detail
              </Link>
            </div>
          </div>
        </aside>
      </div>

      <div className="panel">
        <h3 className="mb-4 font-heading text-2xl">Lots in This Auction</h3>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {lots.map((lot) => (
            <button
              key={lot.id}
              type="button"
              onClick={() => setSelectedLotId(lot.id)}
              className={`rounded-lg border p-4 text-left transition ${selectedLot.id === lot.id ? "border-brand-primary bg-neutral-50" : "border-neutral-200 bg-white hover:bg-neutral-50"}`}
            >
              <div className="text-xs text-neutral-500">Lot {lot.lot_number}</div>
              <h4 className="mt-1 font-heading text-lg">{lot.title}</h4>
              <div className="mt-2"><StatusBadge status={lot.status} /></div>
              <div className="bid-amount mt-2 text-xl text-brand-primary">${lot.current_price.toLocaleString()}</div>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}
