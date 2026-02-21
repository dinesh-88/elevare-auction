import { useEffect, useState } from "react";
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
      <div
        className="mb-8 flex min-h-[260px] items-end rounded-lg border border-neutral-200 bg-neutral-100 bg-cover bg-center p-8"
        style={{ backgroundImage: config?.branding?.hero_image ? `url(${config.branding.hero_image})` : undefined }}
      >
        <div className="max-w-3xl">
          <p className="mb-2 text-xs uppercase tracking-wide text-neutral-600">Curated Collections</p>
          <h1 className="font-heading text-4xl text-neutral-900">Live & Upcoming Auctions</h1>
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
  );
}
