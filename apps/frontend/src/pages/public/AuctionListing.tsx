import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Auction } from "../../api/client";
import AuctionCard from "../../components/AuctionCard";
import CountdownTimer from "../../components/CountdownTimer";
import StatusBadge from "../../components/ui/StatusBadge";
import { useAuth } from "../../hooks/useAuth";

export default function AuctionListing() {
  const { isAuthenticated, isAdmin } = useAuth();
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const list = isAuthenticated && isAdmin ? api.listAdminAuctions : api.listAuctions;
    list()
      .then(setAuctions)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"))
      .finally(() => setLoading(false));
  }, [isAuthenticated, isAdmin]);

  const featuredAuction = useMemo(() => {
    if (auctions.length === 0) return null;
    return [...auctions].sort((a, b) => new Date(a.end_time).getTime() - new Date(b.end_time).getTime())[0];
  }, [auctions]);

  return (
    <section className="min-h-screen bg-neutral-50">
      <section className="border-b border-neutral-200 bg-white">
        <div className="container-shell py-12 md:py-16">
          {featuredAuction ? (
            <div className="grid grid-cols-1 items-center gap-8 md:gap-12 lg:grid-cols-2">
              <div className="aspect-[4/3] overflow-hidden rounded-xl bg-neutral-100">
                {featuredAuction.metadata?.image || featuredAuction.metadata?.images?.[0] ? (
                  <img
                    src={featuredAuction.metadata?.image ?? featuredAuction.metadata?.images?.[0]}
                    alt={featuredAuction.title}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="grid h-full place-content-center text-sm text-neutral-500">No image</div>
                )}
              </div>

              <div className="space-y-6">
                <StatusBadge status={featuredAuction.status} />
                <h2 className="font-heading text-3xl md:text-4xl">{featuredAuction.title}</h2>
                <div className="space-y-2">
                  <div className="text-sm text-neutral-500">Current Bid</div>
                  <div className="bid-amount text-3xl font-heading font-bold text-neutral-900 md:text-4xl">
                    $ {featuredAuction.current_price.toLocaleString()}
                  </div>
                </div>
                <CountdownTimer endTime={new Date(featuredAuction.end_time)} variant="large" />
                <Link to={`/auction/${featuredAuction.id}`} className="btn-brand inline-flex items-center gap-2 px-8 py-3">
                  View Auction <span>→</span>
                </Link>
              </div>
            </div>
          ) : (
            <div className="text-sm text-neutral-500">No auctions available.</div>
          )}
        </div>
      </section>

      <section className="container-shell py-12 md:py-16">
        <div className="mb-8">
          <h2 className="font-heading text-2xl md:text-3xl">Live Auctions</h2>
          <p className="text-neutral-500">Browse all available lots</p>
        </div>

        {loading ? <p className="text-sm text-neutral-500">Loading auctions...</p> : null}
        {error ? <p className="error">{error}</p> : null}

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3 md:gap-8">
          {auctions.map((auction) => (
            <AuctionCard key={auction.id} auction={auction} />
          ))}
        </div>
      </section>
    </section>
  );
}
