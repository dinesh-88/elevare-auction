import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Auction } from "../../api/client";
import AuctionCard from "../../components/AuctionCard";
import { useAuth } from "../../hooks/useAuth";
import { useTheme } from "../../theme/ThemeProvider";

export default function Home() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { config } = useTheme();
  const { isAuthenticated, isAdmin } = useAuth();

  useEffect(() => {
    const list = isAuthenticated && isAdmin ? api.listAdminAuctions : api.listAuctions;
    list()
      .then(setAuctions)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"))
      .finally(() => setLoading(false));
  }, [isAuthenticated, isAdmin]);

  return (
    <section className="page-shell">
      <section className="py-8 md:py-14">
        <div className="grid grid-cols-1 items-center gap-8 md:gap-16 lg:grid-cols-5">
          <div className="space-y-6 lg:col-span-3">
            <h1 className="font-heading text-[42px] leading-tight md:text-[56px]">{config?.name ?? "Elevare"}</h1>
            <p className="max-w-xl text-lg text-neutral-600 md:text-xl">Auction Infrastructure for Institutions</p>
            <div className="flex flex-col gap-4 sm:flex-row">
              <Link to="/signup" className="btn-brand inline-flex px-8 py-3">
                Request Access
              </Link>
              <Link to="/auctions" className="inline-flex items-center justify-center rounded-lg border border-neutral-200 bg-neutral-50 px-8 py-3 font-medium text-neutral-900 hover:bg-neutral-100">
                View Demo
              </Link>
            </div>
          </div>

          <div className="lg:col-span-2">
            <div
              className="aspect-square overflow-hidden rounded-2xl border border-neutral-200 shadow-sm"
              style={{ backgroundImage: config?.branding?.hero_image ? `url(${config.branding.hero_image})` : undefined }}
            >
              {!config?.branding?.hero_image ? (
                <div className="grid h-full place-content-center bg-neutral-100 text-sm text-neutral-500">Hero image</div>
              ) : null}
            </div>
          </div>
        </div>
      </section>

      <section className="grid grid-cols-1 gap-12 border-t border-neutral-200 py-12 md:grid-cols-3 md:gap-16">
        <article className="space-y-4">
          <div className="grid h-12 w-12 place-content-center rounded-lg bg-neutral-100 text-brand-primary">LOT</div>
          <h3 className="font-heading text-xl">Multi-Lot Auctions</h3>
          <p className="leading-relaxed text-neutral-600">Run sophisticated multi-lot auctions with flexible controls and clear sequencing.</p>
        </article>
        <article className="space-y-4">
          <div className="grid h-12 w-12 place-content-center rounded-lg bg-neutral-100 text-brand-primary">RT</div>
          <h3 className="font-heading text-xl">Real-Time Bidding</h3>
          <p className="leading-relaxed text-neutral-600">Live updates keep every participant synced during active bidding windows.</p>
        </article>
        <article className="space-y-4">
          <div className="grid h-12 w-12 place-content-center rounded-lg bg-neutral-100 text-brand-primary">SEC</div>
          <h3 className="font-heading text-xl">Institutional Security</h3>
          <p className="leading-relaxed text-neutral-600">Tenant isolation and role-based controls built for institutional operations.</p>
        </article>
      </section>

      <section className="page-shell">
        <div className="section-head">
          <div>
            <h2 className="section-title">Live & Upcoming Auctions</h2>
            <p className="section-copy">Browse current auction events.</p>
          </div>
        </div>
      {isAuthenticated && isAdmin ? (
        <p className="mb-4 text-sm text-neutral-500">Admin view: showing all tenant auctions (bidding disabled in public view).</p>
      ) : null}

      {loading ? <p className="text-sm text-neutral-500">Loading auctions...</p> : null}
      {error ? <p className="error">{error}</p> : null}

      <div className="grid gap-6 sm:grid-cols-2 xl:grid-cols-3">
        {auctions.map((a) => (
          <AuctionCard key={a.id} auction={a} />
        ))}
      </div>
      </section>
    </section>
  );
}
