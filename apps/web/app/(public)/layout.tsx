"use client";

import Header from "../../src/components/Header";
import { Link } from "../../src/react-router-dom";
import { useAuth } from "../../src/hooks/useAuth";
import { useTheme } from "../../src/theme/ThemeProvider";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  const { config } = useTheme();
  const { isAuthenticated, isAdmin, user, logout } = useAuth();
  const content = (config?.content ?? {}) as Record<string, unknown>;
  const footer = (content.footer ?? {}) as Record<string, unknown>;
  const footerText = typeof footer.text === "string" ? footer.text : "";
  const footerContactEmail = typeof footer.contact_email === "string" ? footer.contact_email : "";
  const socialLinks = footer.social_links && typeof footer.social_links === "object" ? (footer.social_links as Record<string, unknown>) : {};

  return (
    <div className="min-h-screen bg-white text-neutral-800 font-body">
      <Header
        tenantName={config?.name ?? "Elevare"}
        tenantLogoUrl={config?.branding?.logo_url}
        isAuthenticated={isAuthenticated}
        isAdmin={isAdmin}
        userEmail={user?.email}
        onLogout={logout}
      />
      <main>{children}</main>
      <footer className="border-t border-neutral-200 bg-white">
        <div className="mx-auto w-[min(1440px,96vw)] px-4 py-8 md:px-8">
          <div className="mb-3 flex flex-wrap gap-5 text-sm">
            <Link to="/about" className="text-neutral-600 no-underline hover:text-brand-primary">
              About
            </Link>
            <Link to="/contact" className="text-neutral-600 no-underline hover:text-brand-primary">
              Contact
            </Link>
          </div>
          {footerText ? <p className="text-sm text-neutral-600">{footerText}</p> : null}
          {footerContactEmail ? (
            <p className="mt-2 text-sm">
              <a href={`mailto:${footerContactEmail}`} className="text-brand-primary hover:underline">
                {footerContactEmail}
              </a>
            </p>
          ) : null}
          {Object.keys(socialLinks).length > 0 ? (
            <div className="mt-3 flex flex-wrap gap-4">
              {Object.entries(socialLinks).map(([label, href]) => {
                const url = typeof href === "string" ? href : "";
                if (!url) return null;
                return (
                  <a key={label} href={url} target="_blank" rel="noreferrer" className="text-sm text-neutral-600 hover:text-brand-primary">
                    {label}
                  </a>
                );
              })}
            </div>
          ) : null}
        </div>
      </footer>
    </div>
  );
}
