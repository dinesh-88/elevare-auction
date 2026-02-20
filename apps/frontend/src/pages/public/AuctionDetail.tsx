import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import BidPanel from "../../components/BidPanel";
import Countdown from "../../components/Countdown";
import { useAuth } from "../../hooks/useAuth";
import { useAuctionSocket } from "../../hooks/useAuctionSocket";
import { api, type Auction } from "../../api/client";

interface BidHistoryItem {
  bid_id: string;
  amount: number;
  created_at: string;
}

export default function AuctionDetail() {
  const { id } = useParams();
  const { isAuthenticated } = useAuth();
  const [auction, setAuction] = useState<Auction | null>(null);
  const [bidHistory, setBidHistory] = useState<BidHistoryItem[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const a = await api.getAuction(id);
      setAuction(a);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load auction");
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  const handleBidEvent = useCallback(
    (event: { bid_id: string; amount: number; current_price: number; end_time: string; created_at: string }) => {
      setAuction((prev) =>
        prev
          ? {
              ...prev,
              current_price: event.current_price,
              end_time: event.end_time,
            }
          : prev,
      );
      setBidHistory((prev) => [{ bid_id: event.bid_id, amount: event.amount, created_at: event.created_at }, ...prev]);
    },
    [],
  );
  useAuctionSocket(id, handleBidEvent, true);

  if (error) return <p className="error">{error}</p>;
  if (!auction) return <p className="text-sm text-neutral-500">Loading auction...</p>;

  const images = auction.metadata?.images ?? (auction.metadata?.image ? [auction.metadata.image] : []);

  return (
    <section className="grid gap-12 md:grid-cols-5">
      <div className="md:col-span-3">
        <div className="mb-6 overflow-hidden rounded-md border border-neutral-200 bg-neutral-100">
          {images[0] ? (
            <img src={images[0]} alt={auction.title} className="w-full object-cover" />
          ) : (
            <div className="grid h-72 place-content-center text-sm text-neutral-500">No images</div>
          )}
        </div>

        {images.length > 1 ? (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            {images.slice(1).map((image) => (
              <img key={image} src={image} alt={auction.title} className="h-28 w-full rounded-md object-cover" />
            ))}
          </div>
        ) : null}
      </div>

      <aside className="space-y-6 md:col-span-2">
        <h1 className="font-heading text-3xl text-neutral-900">{auction.title}</h1>
        <p className="text-sm leading-6 text-neutral-600">{auction.description}</p>

        <div>
          <p className="text-sm text-neutral-500">Current Bid</p>
          <div className="bid-amount font-heading text-4xl text-brand-primary">${auction.current_price.toLocaleString()}</div>
        </div>

        <div>
          <p className="text-sm text-neutral-500">Ends In</p>
          <div className="text-xl font-medium text-neutral-800">
            <Countdown endTime={auction.end_time} />
          </div>
        </div>

        {isAuthenticated ? <BidPanel auction={auction} onBidPlaced={load} /> : <p className="text-sm">Login to place a bid.</p>}

        <div className="history-panel">
          <h3 className="mb-3 font-heading text-xl">Bid History</h3>
          {bidHistory.length === 0 ? <p className="text-sm text-neutral-500">No bids yet.</p> : null}
          <div className="space-y-2">
            {bidHistory.map((b) => (
              <div key={b.bid_id} className="flex items-center justify-between border-t border-neutral-100 pt-2 text-sm">
                <span className="bid-amount text-brand-primary">${b.amount.toLocaleString()}</span>
                <time className="text-neutral-500">{new Date(b.created_at).toLocaleString()}</time>
              </div>
            ))}
          </div>
        </div>
      </aside>
    </section>
  );
}
