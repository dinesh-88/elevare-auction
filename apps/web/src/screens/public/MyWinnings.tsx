import { useEffect, useMemo, useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { api, type Lot, type SettlementInvoiceRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";
import { useAuth } from "../../hooks/useAuth";
import { useTheme } from "../../theme/ThemeProvider";
import { currencySymbol } from "../../utils/currency";

export default function MyWinnings() {
  const { loading: authLoading, isAuthenticated, isAdmin } = useAuth();
  const { config } = useTheme();
  const [lots, setLots] = useState<Lot[]>([]);
  const [settlements, setSettlements] = useState<SettlementInvoiceRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (authLoading || !isAuthenticated || isAdmin) return;
    Promise.all([
      api.listMyWinningLots(),
      api.listMySettlements("current", { limit: 200 }),
    ])
      .then(([lots, settlementRes]) => {
        setLots(lots);
        setSettlements(settlementRes.items);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load winning lots"))
      .finally(() => setLoading(false));
  }, [authLoading, isAuthenticated, isAdmin]);

  const settlementsByLotID = useMemo(() => {
    const map = new Map<string, SettlementInvoiceRow>();
    settlements.forEach((item) => {
      if (item.lot_id) {
        map.set(item.lot_id, item);
      }
    });
    return map;
  }, [settlements]);

  if (authLoading) {
    return (
      <section className="mx-auto max-w-[1440px] px-4 py-10 md:px-8 md:py-12">
        <p className="text-sm text-neutral-500">Loading your account...</p>
      </section>
    );
  }
  if (!isAuthenticated || isAdmin) return <Navigate to="/" replace />;
  const symbol = currencySymbol(config?.default_currency);

  return (
    <section className="mx-auto max-w-[1440px] px-4 py-10 md:px-8 md:py-12">
      <header className="mb-8">
        <h1 className="font-heading text-3xl text-neutral-900">My Winning Lots</h1>
        <p className="mt-2 text-sm text-neutral-500">Review lots you have won, even after auctions have ended.</p>
      </header>

      {loading ? <p className="text-sm text-neutral-500">Loading winning lots...</p> : null}
      {error ? <p className="error">{error}</p> : null}

      {!loading && !error && lots.length === 0 ? <p className="text-sm text-neutral-500">You do not have winning lots yet.</p> : null}

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {lots.map((lot) => {
          const images = lot.metadata?.images?.length ? lot.metadata.images : lot.metadata?.image ? [lot.metadata.image] : [];
          const image = images[0];
          const settlement = settlementsByLotID.get(lot.id);
          const canPay =
            settlement &&
            settlement.settlement_status === "invoice_sent" &&
            (settlement.payment_status === "unpaid" || settlement.payment_status === "payment_pending") &&
            settlement.checkout_url;
          return (
            <Link
              key={lot.id}
              to={`/lots/${lot.id}`}
              className="group block overflow-hidden rounded-lg border border-neutral-200 bg-white outline-none transition hover:shadow-md focus-visible:border-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-300/70"
            >
              <div className="aspect-[4/3] overflow-hidden bg-neutral-100">
                {image ? (
                  <img src={image} alt={lot.title} className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
                ) : (
                  <div className="grid h-full place-content-center text-sm text-neutral-500">Image coming soon</div>
                )}
              </div>
              <div className="space-y-3 p-5">
                <div className="text-sm text-neutral-500">Lot {lot.lot_number}</div>
                <h2 className="font-heading text-xl text-neutral-900">{lot.title}</h2>
                <div className="inline-flex">
                  <StatusBadge status={lot.status} />
                </div>
                {lot.hammer_price != null ? (
                  <p className="text-sm text-neutral-600">
                    Winning Bid:{" "}
                    <span className="bid-amount font-medium text-neutral-900">
                      {symbol}
                      {lot.hammer_price.toLocaleString()}
                    </span>
                  </p>
                ) : null}
                {settlement ? (
                  <div className="space-y-2 rounded-md border border-neutral-200 bg-neutral-50 p-3">
                    <div className="flex flex-wrap items-center gap-2">
                      <StatusBadge status={settlement.settlement_status} />
                      <StatusBadge status={settlement.payment_status} />
                    </div>
                    <p className="text-sm text-neutral-600">
                      Invoice {settlement.invoice_number}:{" "}
                      <span className="font-medium text-neutral-900">
                        {settlement.currency}
                        {" "}
                        {(settlement.total_due_minor / 100).toFixed(2)}
                      </span>
                    </p>
                    {canPay ? (
                      <div
                        onClick={(event) => {
                          event.preventDefault();
                          event.stopPropagation();
                        }}
                      >
                        <ElevareButton onClick={() => window.open(settlement.checkout_url ?? "", "_blank", "noopener,noreferrer")}>
                          Pay Now
                        </ElevareButton>
                      </div>
                    ) : null}
                  </div>
                ) : null}
                <div className="flex items-center gap-2 pt-1 text-brand-primary">
                  <span className="font-medium">View Details</span>
                  <span className="transition-transform group-hover:translate-x-1">→</span>
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </section>
  );
}
