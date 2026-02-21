import { useState } from "react";
import { Link } from "react-router-dom";
import ElevareButton from "./ui/ElevareButton";

export default function Header({
  tenantName,
  tenantLogoUrl,
  isAuthenticated,
  userEmail,
  onLogout,
}: {
  tenantName?: string;
  tenantLogoUrl?: string;
  isAuthenticated: boolean;
  userEmail?: string;
  onLogout: () => void;
}) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="h-[72px] w-full border-b border-neutral-200 bg-white">
      <div className="mx-auto flex h-full w-[min(1440px,96vw)] items-center justify-between px-4 md:px-8">
        <Link to="/" className="no-underline">
          <div className="flex items-center gap-2">
            {tenantLogoUrl ? (
              <img src={tenantLogoUrl} alt="Tenant logo" className="h-8 w-8 rounded-sm object-contain" />
            ) : (
              <div className="flex h-8 w-8 items-center justify-center rounded-sm bg-brand-primary">
                <span className="font-heading font-semibold text-white">E</span>
              </div>
            )}
            <span className="font-heading text-xl font-semibold text-neutral-900">{tenantName || "Elevare"}</span>
          </div>
        </Link>

        <nav className="hidden items-center gap-6 md:flex">
          <Link to="/auctions" className="text-neutral-900 no-underline hover:text-neutral-500">
            Auctions
          </Link>
          <Link to="/signup" className="text-neutral-900 no-underline hover:text-neutral-500">
            Organization Signup
          </Link>
          {!isAuthenticated ? (
            <>
              <Link to="/login" className="text-neutral-900 no-underline hover:text-neutral-500">
                Login
              </Link>
              <ElevareButton size="default">
                <Link to="/register" className="text-inherit no-underline">
                  Register
                </Link>
              </ElevareButton>
            </>
          ) : (
            <>
              <span className="text-xs text-neutral-500">{userEmail}</span>
              <ElevareButton variant="secondary" onClick={onLogout}>
                Logout
              </ElevareButton>
            </>
          )}
        </nav>

        <button className="p-2 md:hidden" onClick={() => setMobileMenuOpen((v) => !v)} type="button" aria-label="Open menu">
          ☰
        </button>

        {mobileMenuOpen ? (
          <div className="absolute left-0 right-0 top-[72px] z-20 border-b border-neutral-200 bg-white md:hidden">
            <nav className="flex flex-col gap-3 p-4">
              <Link to="/auctions" className="py-1 text-neutral-900 no-underline">
                Auctions
              </Link>
              <Link to="/signup" className="py-1 text-neutral-900 no-underline">
                Organization Signup
              </Link>
              {!isAuthenticated ? (
                <>
                  <Link to="/login" className="py-1 text-neutral-900 no-underline">
                    Login
                  </Link>
                  <Link to="/register" className="py-1 text-neutral-900 no-underline">
                    Register
                  </Link>
                </>
              ) : (
                <button type="button" className="text-left text-neutral-900" onClick={onLogout}>
                  Logout
                </button>
              )}
            </nav>
          </div>
        ) : null}
      </div>
    </header>
  );
}
