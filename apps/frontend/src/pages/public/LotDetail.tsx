import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import Countdown from "../../components/Countdown";
import { api, type Lot } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { useLotSocket } from "../../hooks/useLotSocket";

interface LotBidHistory {
  bid_id: string;
  amount: number;
  created_at: string;
}

export default function LotDetail() {
  const { lotId } = useParams();
  const { isAuthenticated } = useAuth();
  const [lot, setLot] = useState<Lot | null>(null);
  const [amount, setAmount] = useState(0);
  const [history, setHistory] = useState<LotBidHistory[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!lotId) return;
    try {
      const data = await api.getLot(lotId);
      setLot(data);
      setAmount(data.current_price + data.min_increment);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load lot");
    }
  }, [lotId]);

  useEffect(() => {
    load();
  }, [load]);

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
      setAmount((prev) => prev + (lot?.min_increment ?? 1));
    },
    true,
  );

  const placeBid = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!lotId) return;
    try {
      setError(null);
      await api.placeLotBid(lotId, amount);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to place bid");
    }
  };

  if (error) return <p className="error">{error}</p>;
  if (!lot) return <p className="text-sm text-neutral-500">Loading lot...</p>;

  const image = lot.metadata?.image ?? lot.metadata?.images?.[0];

  return (
    <section className="grid gap-12 md:grid-cols-5">
      <div className="md:col-span-3">
        <div className="mb-6 overflow-hidden rounded-md border border-neutral-200 bg-neutral-100">
          {image ? <img src={image} alt={lot.title} className="w-full object-cover" /> : <div className="grid h-72 place-content-center">No image</div>}
        </div>
      </div>

      <aside className="space-y-6 md:col-span-2">
        <h1 className="font-heading text-3xl">{lot.title}</h1>
        <p className="text-sm text-neutral-600">{lot.description}</p>
        <p className="text-sm text-neutral-500">Lot {lot.lot_number} • {lot.status}</p>

        <div>
          <p className="text-sm text-neutral-500">Current Bid</p>
          <div className="bid-amount text-4xl font-heading text-brand-primary">${lot.current_price.toLocaleString()}</div>
        </div>

        {lot.ended_at ? (
          <div>
            <p className="text-sm text-neutral-500">Ends In</p>
            <div className="text-xl"><Countdown endTime={lot.ended_at} /></div>
          </div>
        ) : null}

        {isAuthenticated ? (
          <form className="bid-panel" onSubmit={placeBid}>
            <label>Bid Amount</label>
            <input
              type="number"
              min={lot.current_price + lot.min_increment}
              step={lot.min_increment}
              value={amount}
              onChange={(e) => setAmount(Number(e.target.value))}
            />
            <button className="btn-brand" type="submit">Place Bid</button>
          </form>
        ) : (
          <p className="text-sm">Login to place a bid.</p>
        )}

        <div className="history-panel">
          <h3 className="font-heading text-xl">Bid History</h3>
          {history.length === 0 ? <p className="text-sm text-neutral-500">No bids yet.</p> : null}
          {history.map((b) => (
            <div key={b.bid_id} className="mt-2 flex justify-between border-t border-neutral-100 pt-2 text-sm">
              <span className="bid-amount text-brand-primary">${b.amount.toLocaleString()}</span>
              <span className="text-neutral-500">{new Date(b.created_at).toLocaleString()}</span>
            </div>
          ))}
        </div>
      </aside>
    </section>
  );
}
