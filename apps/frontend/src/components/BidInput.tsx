import { useEffect, useState } from "react";
import ElevareButton from "./ui/ElevareButton";

export default function BidInput({
  currentBid,
  currency,
  minIncrement = 100,
  onPlaceBid,
}: {
  currentBid: number;
  currency: string;
  minIncrement?: number;
  onPlaceBid: (amount: number) => void | Promise<void>;
}) {
  const [bidAmount, setBidAmount] = useState(currentBid + minIncrement);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setBidAmount(currentBid + minIncrement);
  }, [currentBid, minIncrement]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (bidAmount < currentBid + minIncrement) return;
    setSubmitting(true);
    try {
      await onPlaceBid(bidAmount);
    } finally {
      setSubmitting(false);
    }
  };

  const quickBid = (increment: number) => setBidAmount(currentBid + increment);

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="bid-amount" className="mb-2 block text-sm font-medium">
            Your Bid
          </label>
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500">{currency}</span>
            <input
              id="bid-amount"
              type="number"
              value={bidAmount}
              onChange={(e) => setBidAmount(Number(e.target.value))}
              min={currentBid + minIncrement}
              step={minIncrement}
              className="w-full rounded-lg border border-neutral-200 py-3 pl-12 pr-4 focus:outline-none focus:ring-2 focus:ring-brand-primary/20"
            />
          </div>
          <p className="mt-1 text-xs text-neutral-500">
            Minimum bid: {currency} {(currentBid + minIncrement).toLocaleString()}
          </p>
        </div>

        <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <ElevareButton
            type="button"
            variant="secondary"
            className="flex-1 px-3 py-2 text-sm"
            onClick={() => quickBid(minIncrement)}
          >
            +{currency}
            {minIncrement}
          </ElevareButton>
          <ElevareButton
            type="button"
            variant="secondary"
            className="flex-1 px-3 py-2 text-sm"
            onClick={() => quickBid(minIncrement * 5)}
          >
            +{currency}
            {minIncrement * 5}
          </ElevareButton>
          <ElevareButton
            type="button"
            variant="secondary"
            className="flex-1 px-3 py-2 text-sm"
            onClick={() => quickBid(minIncrement * 10)}
          >
            +{currency}
            {minIncrement * 10}
          </ElevareButton>
        </div>

        <ElevareButton type="submit" size="large" className="w-full" disabled={submitting}>
          {submitting ? "Placing..." : "Place Bid"}
        </ElevareButton>
      </form>
    </div>
  );
}
