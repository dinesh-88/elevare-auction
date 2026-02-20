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
  if (!auction) return <p>Loading auction...</p>;

  const images = auction.metadata?.images ?? (auction.metadata?.image ? [auction.metadata.image] : []);

  return (
    <section className="auction-detail">
      <div>
        <h1>{auction.title}</h1>
        <p>{auction.description}</p>
        <p className="price">Current: ${auction.current_price.toLocaleString()}</p>
        <Countdown endTime={auction.end_time} />

        <div className="image-strip">
          {images.length === 0 ? <div className="card-placeholder">No images</div> : null}
          {images.map((image) => (
            <img src={image} alt={auction.title} key={image} className="detail-image" />
          ))}
        </div>
      </div>

      <aside>
        {isAuthenticated ? <BidPanel auction={auction} onBidPlaced={load} /> : <p>Login to place a bid.</p>}

        <div className="history-panel">
          <h3>Bid History</h3>
          {bidHistory.length === 0 ? <p>No bids yet.</p> : null}
          {bidHistory.map((b) => (
            <div key={b.bid_id} className="history-item">
              <span>${b.amount.toLocaleString()}</span>
              <time>{new Date(b.created_at).toLocaleString()}</time>
            </div>
          ))}
        </div>
      </aside>
    </section>
  );
}
