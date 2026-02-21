import { Link } from 'react-router';
import { Menu } from 'lucide-react';
import { useState } from 'react';

interface HeaderProps {
  variant?: 'marketing' | 'tenant';
  tenantName?: string;
}

export function Header({ variant = 'marketing', tenantName }: HeaderProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="w-full h-[72px] bg-white border-b border-border">
      <div className="max-w-[1440px] mx-auto px-4 md:px-8 h-full flex items-center justify-between">
        <Link to="/" className="no-underline">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary rounded-sm flex items-center justify-center">
              <span className="text-white font-serif font-semibold">E</span>
            </div>
            <span className="font-serif text-xl font-semibold text-foreground">
              {variant === 'tenant' && tenantName ? tenantName : 'Elevare'}
            </span>
          </div>
        </Link>
        
        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center gap-8">
          {variant === 'marketing' ? (
            <>
              <Link to="/auctions" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                Features
              </Link>
              <Link to="/auctions" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                Institutions
              </Link>
              <a href="#" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                Login
              </a>
              <button className="px-6 py-2 bg-primary text-primary-foreground rounded-md hover:opacity-90 transition-opacity">
                Get Started
              </button>
            </>
          ) : (
            <>
              <Link to="/auctions" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                Auctions
              </Link>
              <a href="#" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                About
              </a>
              <a href="#" className="text-foreground no-underline hover:text-muted-foreground transition-colors">
                Login
              </a>
            </>
          )}
        </nav>

        {/* Mobile Menu Button */}
        <button 
          className="md:hidden p-2"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
        >
          <Menu className="w-6 h-6" />
        </button>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="absolute top-[72px] left-0 right-0 bg-white border-b border-border md:hidden">
            <nav className="flex flex-col p-4 gap-4">
              {variant === 'marketing' ? (
                <>
                  <Link to="/auctions" className="text-foreground no-underline py-2">Features</Link>
                  <Link to="/auctions" className="text-foreground no-underline py-2">Institutions</Link>
                  <a href="#" className="text-foreground no-underline py-2">Login</a>
                  <button className="px-6 py-2 bg-primary text-primary-foreground rounded-md">
                    Get Started
                  </button>
                </>
              ) : (
                <>
                  <Link to="/auctions" className="text-foreground no-underline py-2">Auctions</Link>
                  <a href="#" className="text-foreground no-underline py-2">About</a>
                  <a href="#" className="text-foreground no-underline py-2">Login</a>
                </>
              )}
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}