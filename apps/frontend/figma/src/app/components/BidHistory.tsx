interface Bid {
  id: string;
  amount: number;
  bidder: string;
  timestamp: Date;
}

interface BidHistoryProps {
  bids: Bid[];
  currency: string;
}

export function BidHistory({ bids, currency }: BidHistoryProps) {
  const formatTime = (date: Date) => {
    return new Intl.DateTimeFormat('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      day: '2-digit',
      month: 'short',
    }).format(date);
  };

  return (
    <div className="space-y-3">
      <h3 className="font-medium text-foreground">Bid History</h3>
      <div className="max-h-[300px] overflow-y-auto space-y-2">
        {bids.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center">No bids yet</p>
        ) : (
          bids.map((bid) => (
            <div 
              key={bid.id} 
              className="flex items-center justify-between p-3 bg-neutral-50 rounded-md"
            >
              <div>
                <div className="font-medium text-foreground">
                  {currency} {bid.amount.toLocaleString()}
                </div>
                <div className="text-xs text-muted-foreground">{bid.bidder}</div>
              </div>
              <div className="text-xs text-muted-foreground">
                {formatTime(bid.timestamp)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
