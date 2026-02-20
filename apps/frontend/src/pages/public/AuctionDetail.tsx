import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Link } from "react-router-dom";
import Countdown from "../../components/Countdown";
import { api, type Auction, type Lot } from "../../api/client";

export default function AuctionDetail() {
  const { id } = useParams();
  const [auction, setAuction] = useState<Auction | null>(null);
  const [lots, setLots] = useState<Lot[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const a = await api.getAuction(id);
      const l = await api.listLotsByAuction(id);
      setAuction(a);
      setLots(l);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load auction");
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  if (error) return <p className="error">{error}</p>;
  if (!auction) return <p className="text-sm text-neutral-500">Loading auction...</p>;

  return (
    <section className="space-y-6">
      <header className="rounded-md border border-neutral-200 bg-white p-6">
        <h1 className="font-heading text-3xl text-neutral-900">{auction.title}</h1>
        <p className="mt-2 text-sm leading-6 text-neutral-600">{auction.description}</p>
        <div className="mt-3 text-sm text-neutral-500">
          Event ends in <Countdown endTime={auction.end_time} />
        </div>
      </header>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {lots.map((lot) => (
          <article key={lot.id} className="rounded-md border border-neutral-200 bg-white p-5">
            <div className="text-xs text-neutral-500">Lot {lot.lot_number}</div>
            <h3 className="mt-1 font-heading text-xl">{lot.title}</h3>
            <div className="mt-2 text-sm text-neutral-600">Status: {lot.status}</div>
            <div className="bid-amount mt-2 text-2xl text-brand-primary">${lot.current_price.toLocaleString()}</div>
            {lot.ended_at ? (
              <div className="mt-2 text-xs text-neutral-500">
                Ends in <Countdown endTime={lot.ended_at} />
              </div>
            ) : null}
            <Link to={`/lots/${lot.id}`} className="btn-link-brand mt-4 inline-block">
              View Lot →
            </Link>
          </article>
        ))}
      </div>
    </section>
  );
}
