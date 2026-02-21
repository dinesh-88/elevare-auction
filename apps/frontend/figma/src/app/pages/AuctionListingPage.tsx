import { Header } from '../components/Header';
import { AuctionCard } from '../components/AuctionCard';
import { StatusBadge } from '../components/StatusBadge';
import { CountdownTimer } from '../components/CountdownTimer';
import { ArrowRight } from 'lucide-react';
import { Link } from 'react-router';

// Mock data
const featuredAuction = {
  id: '1',
  lotNumber: 1,
  title: 'Important Old Masters Collection',
  currentBid: 145000,
  currency: '£',
  endTime: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000), // 2 days from now
  imageUrl: 'https://images.unsplash.com/photo-1694727504199-44bebbe72ad2?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxyZW5haXNzYW5jZSUyMHBhaW50aW5nJTIwYXJ0fGVufDF8fHx8MTc3MTYwNzI2NHww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
};

const auctions = [
  {
    id: '1',
    lotNumber: 12,
    title: 'Italian Renaissance Oil Painting',
    currentBid: 12500,
    currency: '£',
    endTime: new Date(Date.now() + 2 * 60 * 60 * 1000 + 14 * 60 * 1000 + 33 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1766169776604-e916937f1f02?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjbGFzc2ljYWwlMjBvaWwlMjBwYWludGluZyUyMGFydHdvcmt8ZW58MXx8fHwxNzcxNjcyNjI2fDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
  {
    id: '2',
    lotNumber: 13,
    title: 'Hellenistic Bronze Sculpture',
    currentBid: 28750,
    currency: '£',
    endTime: new Date(Date.now() + 3 * 60 * 60 * 1000 + 22 * 60 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1763336318550-8618c9ac2509?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxhbnRpcXVlJTIwc2N1bHB0dXJlJTIwYnJvbnplfGVufDF8fHx8MTc3MTY3MjYyNnww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
  {
    id: '3',
    lotNumber: 14,
    title: 'Ming Dynasty Porcelain Vase',
    currentBid: 45000,
    currency: '£',
    endTime: new Date(Date.now() + 1 * 60 * 60 * 1000 + 45 * 60 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1596381302286-65d0324b7a31?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHx2aW50YWdlJTIwcG9yY2VsYWluJTIwdmFzZXxlbnwxfHx8fDE3NzE1ODU3Mjh8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
  {
    id: '4',
    lotNumber: 15,
    title: 'Baroque Oil on Canvas',
    currentBid: 18500,
    currency: '£',
    endTime: new Date(Date.now() + 4 * 60 * 60 * 1000 + 10 * 60 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1694727504199-44bebbe72ad2?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxyZW5haXNzYW5jZSUyMHBhaW50aW5nJTIwYXJ0fGVufDF8fHx8MTc3MTYwNzI2NHww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
  {
    id: '5',
    lotNumber: 16,
    title: 'Louis XVI Gilt Chair',
    currentBid: 8200,
    currency: '£',
    endTime: new Date(Date.now() + 5 * 60 * 60 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1763337925283-9ed80c77bb6e?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxhbnRpcXVlJTIwZnVybml0dXJlJTIwY2hhaXJ8ZW58MXx8fHwxNzcxNjcyNjI3fDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
  {
    id: '6',
    lotNumber: 17,
    title: 'Victorian Portrait Painting',
    currentBid: 15750,
    currency: '£',
    endTime: new Date(Date.now() + 6 * 60 * 60 * 1000 + 30 * 60 * 1000),
    imageUrl: 'https://images.unsplash.com/photo-1763070605752-fe349372a651?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjbGFzc2ljYWwlMjBhcnQlMjBwb3J0cmFpdHxlbnwxfHx8fDE3NzE2NzI2Mjh8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral',
  },
];

export function AuctionListingPage() {
  return (
    <div className="min-h-screen bg-neutral-50">
      <Header variant="tenant" tenantName="Christie's" />
      
      {/* Featured Auction Section */}
      <section className="bg-white border-b border-border">
        <div className="max-w-[1440px] mx-auto px-4 md:px-8 py-12 md:py-16">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 md:gap-12 items-center">
            <div className="aspect-[4/3] bg-neutral-50 rounded-xl overflow-hidden">
              <img 
                src={featuredAuction.imageUrl}
                alt={featuredAuction.title}
                className="w-full h-full object-cover"
              />
            </div>
            
            <div className="space-y-6">
              <StatusBadge status="LIVE" />
              <h2 className="text-3xl md:text-4xl">{featuredAuction.title}</h2>
              <div className="space-y-2">
                <div className="text-sm text-muted-foreground">Current Bid</div>
                <div className="text-3xl md:text-4xl font-serif font-bold text-foreground">
                  {featuredAuction.currency} {featuredAuction.currentBid.toLocaleString()}
                </div>
              </div>
              <CountdownTimer endTime={featuredAuction.endTime} variant="large" />
              <Link to={`/auction/${featuredAuction.id}`}>
                <button className="flex items-center gap-2 px-8 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity">
                  View Auction
                  <ArrowRight className="w-5 h-5" />
                </button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Lots Grid Section */}
      <section className="max-w-[1440px] mx-auto px-4 md:px-8 py-12 md:py-16">
        <div className="mb-8">
          <h2 className="text-2xl md:text-3xl mb-2">Live Auctions</h2>
          <p className="text-muted-foreground">Browse all available lots</p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 md:gap-8">
          {auctions.map((auction) => (
            <AuctionCard key={auction.id} {...auction} />
          ))}
        </div>
      </section>
    </div>
  );
}