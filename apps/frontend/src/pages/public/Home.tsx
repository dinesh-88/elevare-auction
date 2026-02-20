import { useEffect, useState } from "react";
import { api, type Auction } from "../../api/client";
import AuctionCard from "../../components/AuctionCard";
import { useTheme } from "../../theme/ThemeProvider";

export default function Home() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { config } = useTheme();

  useEffect(() => {
    api
      .listAuctions()
      .then(setAuctions)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <section>
      <div
        className="mb-8 flex min-h-[220px] items-end rounded-md border border-neutral-200 bg-neutral-100 bg-cover bg-center p-6"
        style={{ backgroundImage: config?.branding?.hero_image ? `url(${config.branding.hero_image})` : undefined }}
      >
        <h1 className="font-heading text-4xl text-neutral-900">Live & Upcoming Auctions</h1>
      </div>

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
