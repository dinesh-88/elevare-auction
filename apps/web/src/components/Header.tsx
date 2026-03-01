import { Menu, X } from "lucide-react";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "react-router-dom";
import ElevareButton from "./ui/ElevareButton";
import NotificationBell from "./NotificationBell";

export default function Header({
  tenantName,
  tenantLogoUrl,
  isAuthenticated,
  isAdmin,
  userEmail,
  onLogout,
}: {
  tenantName?: string;
  tenantLogoUrl?: string;
  isAuthenticated: boolean;
  isAdmin: boolean;
  userEmail?: string;
  onLogout: () => void;
}) {
  const t = useTranslations("header");
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const closeMenu = () => setMobileMenuOpen(false);

  useEffect(() => {
    setMobileMenuOpen(false);
  }, [pathname]);

  useEffect(() => {
    if (!mobileMenuOpen) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [mobileMenuOpen]);

  const mobileLinkClassName = "rounded-xl px-4 py-3 text-base font-medium text-neutral-900 no-underline transition-colors hover:bg-neutral-100";

  return (
    <header className="sticky top-0 z-40 min-h-[72px] w-full border-b border-neutral-200 bg-white/95 backdrop-blur">
      <div className="mx-auto flex h-[72px] w-[min(1440px,96vw)] items-center justify-between px-4 md:px-8">
        <Link to="/" className="no-underline py-4" onClick={closeMenu}>
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
            {t("auctions")}
          </Link>
          {isAuthenticated && !isAdmin ? (
            <Link to="/dashboard" className="text-neutral-900 no-underline hover:text-neutral-500">
              Dashboard
            </Link>
          ) : null}
          {isAuthenticated && !isAdmin ? (
            <Link to="/my-winnings" className="text-neutral-900 no-underline hover:text-neutral-500">
              {t("myWinnings")}
            </Link>
          ) : null}
          {isAuthenticated ? (
            <Link to="/profile" className="text-neutral-900 no-underline hover:text-neutral-500">
              {t("profile")}
            </Link>
          ) : null}
          {!isAuthenticated ? (
            <>
              <Link to="/login" className="text-neutral-900 no-underline hover:text-neutral-500">
                {t("login")}
              </Link>
              <Link to="/register" className="inline-flex items-center rounded-lg bg-accent-500 px-6 py-2 text-white no-underline transition-all hover:bg-accent-600">
                {t("register")}
              </Link>
            </>
          ) : (
            <>
              <NotificationBell isAuthenticated={isAuthenticated} />
              <span className="text-xs text-neutral-500">{userEmail}</span>
              <ElevareButton variant="secondary" onClick={onLogout}>
                {t("logout")}
              </ElevareButton>
            </>
          )}
        </nav>

        <div className="flex items-center gap-2 md:hidden">
          {isAuthenticated ? <NotificationBell isAuthenticated={isAuthenticated} /> : null}
          <button
            className="inline-flex h-11 w-11 items-center justify-center rounded-xl border border-neutral-200 bg-white text-neutral-900 shadow-sm"
            onClick={() => setMobileMenuOpen((v) => !v)}
            type="button"
            aria-label={mobileMenuOpen ? "Close menu" : t("openMenu")}
            aria-expanded={mobileMenuOpen}
            aria-controls="mobile-site-menu"
          >
            {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
      </div>

      <div
        className={`fixed inset-0 top-[73px] z-30 bg-neutral-950/30 transition-opacity duration-200 md:hidden ${
          mobileMenuOpen ? "pointer-events-auto opacity-100" : "pointer-events-none opacity-0"
        }`}
        onClick={closeMenu}
        aria-hidden={!mobileMenuOpen}
      />

      <div
        id="mobile-site-menu"
        className={`fixed inset-x-0 top-[73px] z-40 max-h-[calc(100vh-73px)] overflow-y-auto border-b border-neutral-200 bg-white px-4 pb-6 pt-4 shadow-xl transition-transform duration-200 md:hidden ${
          mobileMenuOpen ? "translate-y-0" : "-translate-y-[110%]"
        }`}
      >
        <nav className="mx-auto flex w-[min(1440px,96vw)] flex-col gap-2">
          <Link to="/auctions" className={mobileLinkClassName} onClick={closeMenu}>
            {t("auctions")}
          </Link>
          {isAuthenticated && !isAdmin ? (
            <Link to="/dashboard" className={mobileLinkClassName} onClick={closeMenu}>
              Dashboard
            </Link>
          ) : null}
          {isAuthenticated && !isAdmin ? (
            <Link to="/my-winnings" className={mobileLinkClassName} onClick={closeMenu}>
              {t("myWinnings")}
            </Link>
          ) : null}
          {isAuthenticated ? (
            <Link to="/profile" className={mobileLinkClassName} onClick={closeMenu}>
              {t("profile")}
            </Link>
          ) : null}

          <div className="my-2 h-px bg-neutral-200" />

          {!isAuthenticated ? (
            <>
              <Link to="/login" className={mobileLinkClassName} onClick={closeMenu}>
                {t("login")}
              </Link>
              <Link
                to="/register"
                className="inline-flex min-h-12 items-center justify-center rounded-xl bg-accent-500 px-4 py-3 text-base font-medium text-white no-underline"
                onClick={closeMenu}
              >
                {t("register")}
              </Link>
            </>
          ) : (
            <>
              {userEmail ? <div className="rounded-xl bg-neutral-50 px-4 py-3 text-sm text-neutral-500">{userEmail}</div> : null}
              <button
                type="button"
                className="min-h-12 rounded-xl border border-neutral-200 px-4 py-3 text-left text-base font-medium text-neutral-900"
                onClick={() => {
                  closeMenu();
                  onLogout();
                }}
              >
                {t("logout")}
              </button>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
