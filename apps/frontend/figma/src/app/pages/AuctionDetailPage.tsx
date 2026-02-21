import { useState } from 'react';
import { useParams } from 'react-router';
import { Header } from '../components/Header';
import { StatusBadge } from '../components/StatusBadge';
import { CountdownTimer } from '../components/CountdownTimer';
import { LotImageGallery } from '../components/LotImageGallery';
import { BidInput } from '../components/BidInput';
import { BidHistory } from '../components/BidHistory';

// Mock data
const lotData = {
  id: '1',
  lotNumber: 12,
  title: 'Italian Renaissance Oil Painting',
  artist: 'Giovanni Bellini (Attributed)',
  date: 'circa 1480-1490',
  dimensions: '76 × 63 cm',
  medium: 'Oil on wood panel',
  provenance: 'Private European Collection',
  currentBid: 12500,
  currency: '£',
  reserveMet: true,
  endTime: new Date(Date.now() + 2 * 60 * 60 * 1000 + 14 * 60 * 1000 + 33 * 1000),
  images: [
    'https://images.unsplash.com/photo-1766169776604-e916937f1f02?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjbGFzc2ljYWwlMjBvaWwlMjBwYWludGluZyUyMGFydHdvcmt8ZW58MXx8fHwxNzcxNjcyNjI2fDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
    'https://images.unsplash.com/photo-1694727504199-44bebbe72ad2?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxyZW5haXNzYW5jZSUyMHBhaW50aW5nJTIwYXJ0fGVufDF8fHx8MTc3MTYwNzI2NHww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
    'https://images.unsplash.com/photo-1763070605752-fe349372a651?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjbGFzc2ljYWwlMjBhcnQlMjBwb3J0cmFpdHxlbnwxfHx8fDE3NzE2NzI2Mjh8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  ],
  description: `This exceptional panel painting, attributed to the Venetian master Giovanni Bellini, exemplifies the artistic innovations of the Italian Renaissance. The work demonstrates the artist's masterful use of oil glazing techniques and attention to atmospheric perspective.

The composition reveals the influence of both Byzantine traditions and emerging Renaissance naturalism, characteristic of Venetian painting in the late 15th century. The subtle modeling of form and the luminous quality of the colors showcase the technical sophistication that made Bellini one of the most influential painters of his generation.`,
  condition: `The panel is in good overall condition with minor age-related wear consistent with a work of this period. There is evidence of historical restoration, particularly in the upper left quadrant. The original wooden support remains stable with no signs of structural issues. Surface examination reveals the original paint layer is well-preserved beneath later varnish applications.`,
};

const initialBids = [
  { id: '1', amount: 12500, bidder: 'Bidder #247', timestamp: new Date(Date.now() - 5 * 60 * 1000) },
  { id: '2', amount: 12000, bidder: 'Bidder #103', timestamp: new Date(Date.now() - 15 * 60 * 1000) },
  { id: '3', amount: 11500, bidder: 'Bidder #247', timestamp: new Date(Date.now() - 25 * 60 * 1000) },
  { id: '4', amount: 11000, bidder: 'Bidder #089', timestamp: new Date(Date.now() - 45 * 60 * 1000) },
  { id: '5', amount: 10500, bidder: 'Bidder #247', timestamp: new Date(Date.now() - 60 * 60 * 1000) },
];

export function AuctionDetailPage() {
  const { id } = useParams();
  const [currentBid, setCurrentBid] = useState(lotData.currentBid);
  const [bids, setBids] = useState(initialBids);

  const handlePlaceBid = (amount: number) => {
    const newBid = {
      id: String(bids.length + 1),
      amount,
      bidder: 'You',
      timestamp: new Date(),
    };
    setBids([newBid, ...bids]);
    setCurrentBid(amount);
    // In a real app, this would call an API
  };

  return (
    <div className="min-h-screen bg-white">
      <Header variant="tenant" tenantName="Christie's" />
      
      <div className="max-w-[1440px] mx-auto px-4 md:px-8 py-8 md:py-12">
        {/* Main Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-8 gap-8 md:gap-12 mb-12 md:mb-16">
          {/* Left Side - Image Gallery */}
          <div className="lg:col-span-5">
            <LotImageGallery images={lotData.images} alt={lotData.title} />
          </div>

          {/* Right Side - Bidding Panel */}
          <div className="lg:col-span-3">
            <div className="lg:sticky lg:top-8 space-y-6">
              <div>
                <div className="text-sm text-muted-foreground mb-2">Lot {lotData.lotNumber}</div>
                <h1 className="text-2xl md:text-3xl mb-4">{lotData.title}</h1>
                <div className="text-base md:text-lg text-muted-foreground mb-1">{lotData.artist}</div>
                <div className="text-muted-foreground">{lotData.date}</div>
              </div>

              <div className="border-t border-border pt-6">
                <div className="space-y-4">
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                    <div>
                      <div className="text-sm text-muted-foreground mb-1">Current Bid</div>
                      <div className="text-[32px] md:text-[40px] font-serif font-bold text-foreground tabular-nums leading-none">
                        {lotData.currency} {currentBid.toLocaleString()}
                      </div>
                    </div>
                    {lotData.reserveMet && <StatusBadge status="RESERVE_MET" />}
                  </div>
                  
                  <CountdownTimer endTime={lotData.endTime} variant="large" />
                </div>
              </div>

              <div className="border-t border-border pt-6">
                <BidInput 
                  currentBid={currentBid} 
                  currency={lotData.currency}
                  minIncrement={100}
                  onPlaceBid={handlePlaceBid}
                />
              </div>

              <div className="border-t border-border pt-6">
                <BidHistory bids={bids} currency={lotData.currency} />
              </div>
            </div>
          </div>
        </div>

        {/* Bottom Section - Details */}
        <div className="max-w-4xl space-y-8 md:space-y-12">
          <div>
            <h2 className="text-xl md:text-2xl mb-4">Description</h2>
            <div className="prose prose-neutral max-w-none">
              {lotData.description.split('\n\n').map((paragraph, index) => (
                <p key={index} className="text-foreground mb-4">{paragraph}</p>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 md:gap-8 p-4 md:p-6 bg-neutral-50 rounded-lg">
            <div>
              <h3 className="text-base md:text-lg font-medium mb-3">Details</h3>
              <dl className="space-y-2">
                <div>
                  <dt className="text-sm text-muted-foreground">Medium</dt>
                  <dd className="text-foreground">{lotData.medium}</dd>
                </div>
                <div>
                  <dt className="text-sm text-muted-foreground">Dimensions</dt>
                  <dd className="text-foreground">{lotData.dimensions}</dd>
                </div>
                <div>
                  <dt className="text-sm text-muted-foreground">Date</dt>
                  <dd className="text-foreground">{lotData.date}</dd>
                </div>
              </dl>
            </div>

            <div>
              <h3 className="text-base md:text-lg font-medium mb-3">Provenance</h3>
              <p className="text-foreground">{lotData.provenance}</p>
            </div>
          </div>

          <div>
            <h2 className="text-xl md:text-2xl mb-4">Condition Report</h2>
            <div className="p-4 md:p-6 bg-neutral-50 rounded-lg">
              <p className="text-foreground leading-relaxed">{lotData.condition}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}