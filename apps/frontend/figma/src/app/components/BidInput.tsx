import { useState } from 'react';

interface BidInputProps {
  currentBid: number;
  currency: string;
  minIncrement?: number;
  onPlaceBid: (amount: number) => void;
}

export function BidInput({ currentBid, currency, minIncrement = 100, onPlaceBid }: BidInputProps) {
  const [bidAmount, setBidAmount] = useState(currentBid + minIncrement);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (bidAmount > currentBid) {
      onPlaceBid(bidAmount);
    }
  };

  const quickBid = (increment: number) => {
    setBidAmount(currentBid + increment);
  };

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="bid-amount" className="block text-sm font-medium mb-2">
            Your Bid
          </label>
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground">
              {currency}
            </span>
            <input
              id="bid-amount"
              type="number"
              value={bidAmount}
              onChange={(e) => setBidAmount(Number(e.target.value))}
              min={currentBid + minIncrement}
              step={minIncrement}
              className="w-full pl-12 pr-4 py-3 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            Minimum bid: {currency} {(currentBid + minIncrement).toLocaleString()}
          </p>
        </div>

        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => quickBid(minIncrement)}
            className="flex-1 px-3 py-2 text-sm border border-border rounded-md hover:bg-neutral-50 transition-colors"
          >
            +{currency}{minIncrement}
          </button>
          <button
            type="button"
            onClick={() => quickBid(minIncrement * 5)}
            className="flex-1 px-3 py-2 text-sm border border-border rounded-md hover:bg-neutral-50 transition-colors"
          >
            +{currency}{minIncrement * 5}
          </button>
          <button
            type="button"
            onClick={() => quickBid(minIncrement * 10)}
            className="flex-1 px-3 py-2 text-sm border border-border rounded-md hover:bg-neutral-50 transition-colors"
          >
            +{currency}{minIncrement * 10}
          </button>
        </div>

        <button
          type="submit"
          className="w-full px-6 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity"
        >
          Place Bid
        </button>
      </form>
    </div>
  );
}
