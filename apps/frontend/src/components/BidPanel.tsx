import { useState } from "react";
import { api, type Auction } from "../api/client";

const errorMap: Record<string, string> = {
  "email not verified": "Your email is not verified yet.",
  "user not approved": "Your account is waiting for admin approval.",
  "auction already ended": "This auction has already ended.",
  "bid amount below minimum increment": "Your bid is below the minimum increment.",
};

export default function BidPanel({ auction, onBidPlaced }: { auction: Auction; onBidPlaced: () => void }) {
  const [amount, setAmount] = useState(auction.current_price + auction.min_increment);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const minBid = auction.current_price + auction.min_increment;

  const placeBid = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await api.placeBid(auction.id, Number(amount));
      onBidPlaced();
      setAmount(Number(amount) + auction.min_increment);
    } catch (err) {
      const raw = err instanceof Error ? err.message : "Unable to place bid";
      setError(errorMap[raw] ?? raw);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className="bid-panel" onSubmit={placeBid}>
      <h3 className="font-heading text-xl text-neutral-900">Place Bid</h3>
      <label htmlFor="amount" className="text-sm text-neutral-500">
        Amount (minimum ${minBid.toLocaleString()})
      </label>
      <input
        id="amount"
        type="number"
        min={minBid}
        step={auction.min_increment}
        value={amount}
        onChange={(e) => setAmount(Number(e.target.value))}
      />
      <button type="submit" className="btn-brand w-full" disabled={submitting}>
        {submitting ? "Placing..." : "Place Bid"}
      </button>
      {error ? <p className="error">{error}</p> : null}
    </form>
  );
}
