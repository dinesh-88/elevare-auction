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
      <div className="hero" style={{ backgroundImage: config?.branding?.hero_image ? `url(${config.branding.hero_image})` : undefined }}>
        <h1>Live & Upcoming Auctions</h1>
      </div>

      {loading ? <p>Loading auctions...</p> : null}
      {error ? <p className="error">{error}</p> : null}

      <div className="auction-grid">
        {auctions.map((a) => (
          <AuctionCard key={a.id} auction={a} />
        ))}
      </div>
    </section>
  );
}
