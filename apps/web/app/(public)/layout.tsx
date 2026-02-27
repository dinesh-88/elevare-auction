"use client";

import Header from "../../src/components/Header";
import { Link } from "../../src/react-router-dom";
import { useAuth } from "../../src/hooks/useAuth";
import { useTheme } from "../../src/theme/ThemeProvider";
import { Facebook, Globe, Instagram, Linkedin, Twitter, Youtube } from "lucide-react";

function socialIcon(label: string, url: string) {
  const key = `${label} ${url}`.toLowerCase();
  if (key.includes("facebook")) return <Facebook size={16} />;
  if (key.includes("instagram")) return <Instagram size={16} />;
  if (key.includes("linkedin")) return <Linkedin size={16} />;
  if (key.includes("twitter") || key.includes("x.com")) return <Twitter size={16} />;
  if (key.includes("youtube")) return <Youtube size={16} />;
  return <Globe size={16} />;
}

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  const { config } = useTheme();
  const { isAuthenticated, isAdmin, user, logout } = useAuth();
  const content = (config?.content ?? {}) as Record<string, unknown>;
  const footer = (content.footer ?? {}) as Record<string, unknown>;
  const footerText = typeof footer.text === "string" ? footer.text : "";
  const footerContactEmail = typeof footer.contact_email === "string" ? footer.contact_email : "";
  const socialLinks = footer.social_links && typeof footer.social_links === "object" ? (footer.social_links as Record<string, unknown>) : {};
  const year = new Date().getFullYear();

  return (
    <div className="min-h-screen bg-white font-sans text-neutral-800">
      <Header
        tenantName={config?.name ?? "Elevare"}
        tenantLogoUrl={config?.branding?.logo_url}
        isAuthenticated={isAuthenticated}
        isAdmin={isAdmin}
        userEmail={user?.email}
        onLogout={logout}
      />
      <main>{children}</main>
      <footer className="border-t border-neutral-200 bg-neutral-50">
        <div className="mx-auto w-[min(1440px,96vw)] px-4 py-10 md:px-8 md:py-12">
          <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
            <div className="space-y-2">
              <p className="font-serif text-lg text-neutral-900">{config?.name ?? "Elevare"}</p>
              <p className="text-sm text-neutral-600">Institutional Auction Platform</p>
              {footerText ? <p className="text-sm leading-relaxed text-neutral-600">{footerText}</p> : null}
            </div>

            <div>
              <p className="mb-3 text-xs font-semibold uppercase tracking-[0.08em] text-neutral-500">Company</p>
              <nav className="flex flex-col gap-2 text-sm">
                <Link to="/about" className="text-neutral-700 no-underline hover:text-brand-primary">
                  About
                </Link>
                <Link to="/contact" className="text-neutral-700 no-underline hover:text-brand-primary">
                  Contact
                </Link>
              </nav>
            </div>

            <div className="space-y-3">
              <p className="text-xs font-semibold uppercase tracking-[0.08em] text-neutral-500">Connect</p>
              {footerContactEmail ? (
                <p className="text-sm">
                  <a href={`mailto:${footerContactEmail}`} className="text-brand-primary hover:underline">
                    {footerContactEmail}
                  </a>
                </p>
              ) : null}
              {Object.keys(socialLinks).length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {Object.entries(socialLinks).map(([label, href]) => {
                    const url = typeof href === "string" ? href : "";
                    if (!url) return null;
                    return (
                      <a
                        key={label}
                        href={url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-2 rounded-md border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-700 hover:border-brand-primary/40 hover:text-brand-primary"
                      >
                        {socialIcon(label, url)}
                        <span>{label}</span>
                      </a>
                    );
                  })}
                </div>
              ) : null}
            </div>
          </div>
          <div className="mt-8 border-t border-neutral-200 pt-4 text-xs text-neutral-500">
            © {year} {config?.name ?? "Elevare"}. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  );
}
