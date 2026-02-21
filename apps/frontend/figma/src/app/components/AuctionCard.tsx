import { Link } from 'react-router';
import { CountdownTimer } from './CountdownTimer';
import { ArrowRight } from 'lucide-react';

interface AuctionCardProps {
  id: string;
  lotNumber: number;
  title: string;
  currentBid: number;
  currency: string;
  endTime: Date;
  imageUrl: string;
}

export function AuctionCard({ id, lotNumber, title, currentBid, currency, endTime, imageUrl }: AuctionCardProps) {
  return (
    <Link to={`/auction/${id}`} className="block no-underline group">
      <div className="bg-white rounded-lg overflow-hidden border border-border hover:shadow-md transition-shadow">
        <div className="aspect-[4/3] overflow-hidden bg-neutral-50">
          <img 
            src={imageUrl} 
            alt={title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
        </div>
        <div className="p-6 space-y-3">
          <div className="text-sm text-muted-foreground">Lot {lotNumber}</div>
          <h3 className="text-xl font-serif font-medium text-foreground">{title}</h3>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-serif font-semibold text-foreground">{currency} {currentBid.toLocaleString()}</span>
          </div>
          <CountdownTimer endTime={endTime} />
          <div className="flex items-center gap-2 text-primary pt-2">
            <span className="font-medium">View</span>
            <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
          </div>
        </div>
      </div>
    </Link>
  );
}
