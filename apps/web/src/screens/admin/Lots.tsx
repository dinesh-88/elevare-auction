import { useEffect, useState } from "react";
import { ArrowDown, ArrowUp, Pencil, Play, Square, Trash2 } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { api, type Lot } from "../../api/client";
import { useLotsSockets } from "../../hooks/useLotsSockets";
import StatusBadge from "../../components/ui/StatusBadge";
import IconButton from "../../components/ui/IconButton";
import { currencySymbol } from "../../utils/currency";

export default function Lots() {
  const { auctionId } = useParams();
  const [lots, setLots] = useState<Lot[]>([]);
  const [auctionCurrency, setAuctionCurrency] = useState("USD");
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const load = () => {
    if (!auctionId) return;
    Promise.all([api.listLotsByAuction(auctionId), api.getAuction(auctionId)])
      .then(([rows, auction]) => {
        setLots(rows);
        setAuctionCurrency(auction.currency || "USD");
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load lots"));
  };

  useEffect(() => {
    load();
  }, [auctionId]);

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

  const move = async (index: number, direction: -1 | 1) => {
    if (!auctionId) return;
    const next = [...lots];
    const target = index + direction;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    try {
      await api.reorderLots(
        auctionId,
        next.map((l) => l.id),
      );
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reorder lots");
    }
  };

  const removeLot = async (lotId: string) => {
    if (!window.confirm("Delete lot?")) return;
    try {
      await api.deleteLot(lotId);
      setMessage("Lot deleted.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete lot");
    }
  };

  const startLot = async (lotId: string) => {
    try {
      await api.startLot(lotId);
      setMessage("Lot started.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start lot");
    }
  };

  const endLot = async (lotId: string) => {
    try {
      await api.endLot(lotId);
      setMessage("Lot ended.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to end lot");
    }
  };

  const lotImages = (lot: Lot): string[] => {
    const fromImages = Array.isArray(lot.metadata?.images) ? lot.metadata.images.filter((x): x is string => typeof x === "string") : [];
    const fromImage = typeof lot.metadata?.image === "string" && lot.metadata.image ? [lot.metadata.image] : [];
    if (fromImages.length > 0) return fromImages;
    return fromImage;
  };
  const symbol = currencySymbol(auctionCurrency);

  return (
    <section className="page-shell">
      <div className="section-head">
        <div>
          <h1 className="section-title">Manage Lots</h1>
          <p className="section-copy">View and manage all lots for this auction.</p>
        </div>
        <div className="flex items-center gap-2">
          <Link className="btn-link-brand" to="/admin/auctions">← Back to Auctions</Link>
          <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/settlement`}>Settlement</Link>
          <Link className="btn-brand" to={`/admin/auctions/${auctionId}/lots/create`}>
            Create Lot
          </Link>
        </div>
      </div>

      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <div className="space-y-3">
        <h2 className="font-heading text-2xl text-neutral-900">Lots</h2>
        <p className="text-sm text-neutral-500">Open a lot to edit details, images, and pricing.</p>
      </div>

      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
        {lots.map((lot, idx) => {
          const images = lotImages(lot);
          const image = images[0];
          return (
            <article
              key={lot.id}
              className="group overflow-hidden rounded-lg border border-neutral-200 bg-white transition duration-200 hover:-translate-y-0.5 hover:shadow-md"
            >
              <div className="relative aspect-[4/3] overflow-hidden bg-neutral-100">
                {image ? (
                  <img
                    src={image}
                    alt={lot.title}
                    className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                  />
                ) : (
                  <div className="grid h-full place-content-center text-sm text-neutral-500">Image coming soon</div>
                )}
                <div className="absolute left-3 top-3 rounded-md bg-white/90 px-2 py-1 text-xs text-neutral-700 shadow-sm">
                  Order #{lot.order_index}
                </div>
              </div>

              <div className="space-y-4 p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-xs uppercase tracking-[0.08em] text-neutral-500">Lot {lot.lot_number}</p>
                    <h3 className="mt-1 font-heading text-xl text-neutral-900">{lot.title}</h3>
                  </div>
                  <StatusBadge status={lot.status} />
                </div>

                <p className="line-clamp-2 text-sm text-neutral-600">{lot.description || "Description coming soon."}</p>

                <div className="flex items-end justify-between gap-3 border-t border-neutral-100 pt-3">
                  <div>
                    <p className="text-xs text-neutral-500">Current Price</p>
                    <p className="bid-amount text-2xl font-heading text-brand-primary">
                      {symbol}
                      {lot.current_price.toLocaleString()}
                    </p>
                  </div>
                  <span className="text-brand-primary transition-transform duration-300 group-hover:translate-x-1">→</span>
                </div>

                <div className="flex flex-wrap gap-2 border-t border-neutral-100 pt-3">
                  <IconButton onClick={() => move(idx, -1)} aria-label="Move up" title="Move up" tooltip="Move up">
                    <ArrowUp size={16} strokeWidth={2.2} />
                  </IconButton>
                  <IconButton onClick={() => move(idx, 1)} aria-label="Move down" title="Move down" tooltip="Move down">
                    <ArrowDown size={16} strokeWidth={2.2} />
                  </IconButton>
                  <Link
                    to={`/admin/auctions/${auctionId}/lots/${lot.id}/edit`}
                    className="group/edit-button relative inline-flex h-9 w-9 items-center justify-center rounded-md border border-neutral-200 bg-neutral-50 text-neutral-700 outline-none transition hover:bg-neutral-100 focus-visible:border-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-300/70"
                    aria-label="Edit lot"
                    title="Edit lot"
                  >
                    <Pencil size={16} strokeWidth={2.2} />
                    <span className="pointer-events-none absolute -top-8 left-1/2 z-20 hidden -translate-x-1/2 whitespace-nowrap rounded bg-neutral-900 px-2 py-1 text-[11px] text-white shadow-sm group-hover/edit-button:block">
                      Edit
                    </span>
                  </Link>
                  <IconButton onClick={() => startLot(lot.id)} aria-label="Start lot" title="Start lot" tooltip="Start">
                    <Play size={16} strokeWidth={2.2} />
                  </IconButton>
                  <IconButton onClick={() => endLot(lot.id)} aria-label="End lot" title="End lot" tooltip="End">
                    <Square size={16} strokeWidth={2.2} />
                  </IconButton>
                  <IconButton onClick={() => removeLot(lot.id)} aria-label="Delete lot" title="Delete lot" tooltip="Delete">
                    <Trash2 size={16} strokeWidth={2.2} />
                  </IconButton>
                </div>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
