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
  const closeMenu = () => setMobileMenuOpen(false);

  return (
    <header className="relative z-40 h-[72px] w-full border-b border-neutral-200 bg-white">
      <div className="mx-auto flex h-full w-[min(1440px,96vw)] items-center justify-between px-4 md:px-8">
        <Link to="/" className="no-underline" onClick={closeMenu}>
          <div className="flex items-center gap-2">
            {tenantLogoUrl ? (
              <img src={tenantLogoUrl} alt="Tenant logo" className="h-8 w-8 rounded-sm object-contain" />
            ) : (
              <div className="flex h-8 w-8 items-center justify-center rounded-sm bg-brand-primary">
                <span className="font-heading font-semibold text-white">E</span>
              </div>
            )}
            <span className="max-w-[58vw] truncate font-heading text-base font-semibold text-neutral-900 sm:max-w-none sm:text-xl">
              {tenantName || "Elevare"}
            </span>
          </div>
        </Link>

        <nav className="hidden items-center gap-8 md:flex">
          <Link to="/auctions" className="text-neutral-900 no-underline hover:text-neutral-500">
            Auctions
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

        <button
          className="inline-flex h-10 w-10 items-center justify-center rounded-md border border-neutral-200 text-neutral-900 md:hidden"
          onClick={() => setMobileMenuOpen((v) => !v)}
          type="button"
          aria-label="Open menu"
          aria-expanded={mobileMenuOpen}
        >
          ☰
        </button>

        {mobileMenuOpen ? (
          <div className="fixed left-0 right-0 top-[72px] z-20 max-h-[calc(100vh-72px)] overflow-y-auto border-b border-neutral-200 bg-white shadow-sm md:hidden">
            <nav className="flex flex-col gap-3 p-4">
              <Link to="/auctions" className="py-1 text-neutral-900 no-underline" onClick={closeMenu}>
                Auctions
              </Link>
              {!isAuthenticated ? (
                <>
                  <Link to="/login" className="py-1 text-neutral-900 no-underline" onClick={closeMenu}>
                    Login
                  </Link>
                  <Link to="/register" className="py-1 text-neutral-900 no-underline" onClick={closeMenu}>
                    Register
                  </Link>
                </>
              ) : (
                <>
                  {userEmail ? <div className="text-xs text-neutral-500">{userEmail}</div> : null}
                  <button
                    type="button"
                    className="text-left text-neutral-900"
                    onClick={() => {
                      closeMenu();
                      onLogout();
                    }}
                  >
                    Logout
                  </button>
                </>
              )}
            </nav>
          </div>
        ) : null}
      </div>
    </header>
  );
}
