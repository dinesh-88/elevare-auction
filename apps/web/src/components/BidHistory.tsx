type BidEntry = {
  id: string;
  amount: number;
  bidder: string;
  timestamp: Date;
};

export default function BidHistory({ bids, currency }: { bids: BidEntry[]; currency: string }) {
  const formatTime = (date: Date) =>
    new Intl.DateTimeFormat("en-GB", {
      hour: "2-digit",
      minute: "2-digit",
      day: "2-digit",
      month: "short",
    }).format(date);

  return (
    <div className="space-y-3">
      <h3 className="font-medium text-neutral-900">Bid History</h3>
      <div className="max-h-[300px] space-y-2 overflow-y-auto">
        {bids.length === 0 ? (
          <p className="py-4 text-center text-sm text-neutral-500">No bids yet</p>
        ) : (
          bids.map((bid) => (
            <div key={bid.id} className="flex items-center justify-between rounded-md bg-neutral-50 p-3">
              <div>
                <div className="font-medium text-neutral-900">
                  {currency} {bid.amount.toLocaleString()}
                </div>
                <div className="text-xs text-neutral-500">{bid.bidder}</div>
              </div>
              <div className="text-xs text-neutral-500">{formatTime(bid.timestamp)}</div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
