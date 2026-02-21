import { Header } from '../components/Header';
import { Building2, Zap, Shield } from 'lucide-react';
import { Link } from 'react-router';

export function HomePage() {
  return (
    <div className="min-h-screen bg-white">
      <Header variant="marketing" />
      
      {/* Hero Section */}
      <section className="max-w-[1440px] mx-auto px-4 md:px-8 py-16 md:py-[120px]">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-8 md:gap-16 items-center">
          <div className="lg:col-span-3 space-y-6 md:space-y-8">
            <h1 className="text-[42px] md:text-[56px] leading-tight">Elevare</h1>
            <p className="text-lg md:text-xl text-muted-foreground max-w-xl">
              Auction Infrastructure for Institutions
            </p>
            <div className="flex flex-col sm:flex-row gap-4">
              <button className="px-8 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity">
                Request Access
              </button>
              <Link to="/auctions">
                <button className="w-full sm:w-auto px-8 py-3 bg-secondary text-secondary-foreground rounded-lg font-medium border border-border hover:bg-neutral-200 transition-colors">
                  View Demo
                </button>
              </Link>
            </div>
          </div>
          
          <div className="lg:col-span-2">
            <div className="aspect-square rounded-2xl overflow-hidden shadow-sm">
              <img 
                src="https://images.unsplash.com/photo-1551554781-c46200ea959d?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVnYW50JTIwc3RvbmUlMjB0ZXh0dXJlJTIwbWFyYmxlfGVufDF8fHx8MTc3MTY3MjYwNHww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral"
                alt="Elegant stone texture"
                className="w-full h-full object-cover"
              />
            </div>
          </div>
        </div>
      </section>

      {/* Feature Section */}
      <section className="max-w-[1440px] mx-auto px-4 md:px-8 py-16 md:py-24 border-t border-border">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-12 md:gap-16">
          <div className="space-y-4">
            <div className="w-12 h-12 bg-neutral-50 rounded-lg flex items-center justify-center">
              <Building2 className="w-6 h-6 text-primary" />
            </div>
            <h3 className="text-xl">Multi-Lot Auctions</h3>
            <p className="text-muted-foreground leading-relaxed">
              Run sophisticated multi-lot auctions with advanced lot management and flexible bidding structures.
            </p>
          </div>

          <div className="space-y-4">
            <div className="w-12 h-12 bg-neutral-50 rounded-lg flex items-center justify-center">
              <Zap className="w-6 h-6 text-primary" />
            </div>
            <h3 className="text-xl">Real-Time Bidding</h3>
            <p className="text-muted-foreground leading-relaxed">
              Live bidding with instant updates, ensuring seamless experiences for all participants.
            </p>
          </div>

          <div className="space-y-4">
            <div className="w-12 h-12 bg-neutral-50 rounded-lg flex items-center justify-center">
              <Shield className="w-6 h-6 text-primary" />
            </div>
            <h3 className="text-xl">Institutional-Grade Security</h3>
            <p className="text-muted-foreground leading-relaxed">
              Enterprise-level security and compliance built for the most demanding institutional requirements.
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border mt-16 md:mt-24">
        <div className="max-w-[1440px] mx-auto px-4 md:px-8 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="text-sm text-muted-foreground text-center md:text-left">
              © 2026 Elevare. Auction Infrastructure for Institutions.
            </div>
            <div className="flex gap-8">
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Privacy
              </a>
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Terms
              </a>
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Contact
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}