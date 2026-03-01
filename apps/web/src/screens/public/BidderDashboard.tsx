"use client";

import { useEffect, useMemo, useState } from "react";
import { Link, Navigate, useParams } from "../../react-router-dom";
import { api, type BidderAuctionParticipation, type SettlementInvoiceRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";
import { useAuth } from "../../hooks/useAuth";

type Props = {
  invoicesOnly?: boolean;
};

type PaymentFilter = "ALL" | "unpaid" | "paid";

export default function BidderDashboard({ invoicesOnly = false }: Props) {
  const { tenant } = useParams<{ tenant?: string }>();
  const tenantAlias = tenant || "current";
  const { loading: authLoading, isAuthenticated, isAdmin } = useAuth();
  const [activeAuctions, setActiveAuctions] = useState<BidderAuctionParticipation[]>([]);
  const [watchlistCount, setWatchlistCount] = useState(0);
  const [settlements, setSettlements] = useState<SettlementInvoiceRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<PaymentFilter>("ALL");

  useEffect(() => {
    if (authLoading || !isAuthenticated || isAdmin) return;
    setLoading(true);
    setError(null);
    Promise.all([
      api.listMyActiveAuctions(tenantAlias),
      api.listMyWatchlist(tenantAlias),
      api.listMySettlements(tenantAlias, { limit: 200 }),
    ])
      .then(([auctions, watchlist, settlementsRes]) => {
        setActiveAuctions(auctions);
        setWatchlistCount(Array.isArray(watchlist) ? watchlist.length : 0);
        setSettlements(settlementsRes.items);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load dashboard"))
      .finally(() => setLoading(false));
  }, [authLoading, isAuthenticated, isAdmin, tenantAlias]);

  const visibleActiveAuctions = useMemo(
    () => activeAuctions.filter((item) => item.status === "SCHEDULED" || item.status === "LIVE" || item.status === "PAUSED"),
    [activeAuctions],
  );
  const winningLots = useMemo(
    () => settlements.filter((item) => item.settlement_status !== "cancelled"),
    [settlements],
  );
  const visibleInvoices = useMemo(() => {
    switch (filter) {
      case "unpaid":
        return settlements.filter((item) => item.payment_status === "unpaid" || item.payment_status === "payment_pending");
      case "paid":
        return settlements.filter((item) => item.payment_status === "paid");
      default:
        return settlements;
    }
  }, [filter, settlements]);
  const purchaseHistory = useMemo(
    () =>
      settlements
        .filter((item) => item.payment_status === "paid")
        .slice()
        .sort((a, b) => new Date(b.paid_at ?? b.issued_at).getTime() - new Date(a.paid_at ?? a.issued_at).getTime()),
    [settlements],
  );

  if (authLoading) {
    return <section className="mx-auto max-w-[1440px] px-4 py-10 md:px-8 md:py-12"><p className="text-sm text-neutral-500">Loading your dashboard...</p></section>;
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (isAdmin) return <Navigate to="/admin/dashboard" replace />;

  return (
    <section className="mx-auto max-w-[1440px] px-4 py-10 md:px-8 md:py-12">
      <header className="mb-8 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-heading text-3xl text-neutral-900">Bidder Dashboard</h1>
          <p className="mt-2 text-sm text-neutral-500">Track auctions, invoices, and completed purchases for this institution.</p>
        </div>
        <div className="flex gap-2">
          {!invoicesOnly ? (
            <Link className="btn-link-brand" to={`/dashboard/${tenantAlias}/invoices`}>
              View Invoices
            </Link>
          ) : (
            <Link className="btn-link-brand" to={`/dashboard/${tenantAlias}`}>
              Back to Overview
            </Link>
          )}
        </div>
      </header>

      {error ? <p className="error mb-4">{error}</p> : null}
      {loading ? <p className="text-sm text-neutral-500">Loading dashboard data...</p> : null}

      {!invoicesOnly ? (
        <>
          <section className="panel mb-6">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="font-heading text-xl text-neutral-900">Active Auctions</h2>
              <span className="text-sm text-neutral-500">{visibleActiveAuctions.length} active</span>
            </div>
            {visibleActiveAuctions.length === 0 ? <p className="text-sm text-neutral-500">You are not registered in active auctions right now.</p> : null}
            <div className="space-y-3">
              {visibleActiveAuctions.map((item) => (
                <div key={item.auction_id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-neutral-200 p-4">
                  <div>
                    <div className="font-medium text-neutral-900">{item.title}</div>
                    <div className="text-sm text-neutral-500">
                      {new Date(item.start_time).toLocaleString()} to {new Date(item.end_time).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <StatusBadge status={item.approval_status} />
                    <StatusBadge status={item.status} />
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel mb-6">
            <div className="mb-2 flex items-center justify-between">
              <h2 className="font-heading text-xl text-neutral-900">Watching Lots</h2>
              <span className="text-sm text-neutral-500">{watchlistCount} items</span>
            </div>
            <p className="text-sm text-neutral-500">
              {watchlistCount === 0
                ? "No watched lots yet. Watchlist tracking will appear here when watchlist data is available."
                : "Watched lots are available in your watchlist."}
            </p>
          </section>

          <section className="panel mb-6">
            <h2 className="mb-3 font-heading text-xl text-neutral-900">Winning Lots</h2>
            {winningLots.length === 0 ? <p className="text-sm text-neutral-500">You do not have winning lots yet.</p> : null}
            <div className="space-y-3">
              {winningLots.map((item) => (
                <div key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-neutral-200 p-4">
                  <div>
                    <div className="font-medium text-neutral-900">{item.lot_summary}</div>
                    <div className="text-sm text-neutral-500">
                      Invoice {item.invoice_number} • {item.currency} {(item.total_due_minor / 100).toFixed(2)}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <StatusBadge status={item.settlement_status} />
                    <StatusBadge status={item.payment_status} />
                  </div>
                </div>
              ))}
            </div>
          </section>
        </>
      ) : null}

      <section className="panel mb-6">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-heading text-xl text-neutral-900">Invoices</h2>
          <div className="flex items-center gap-2">
            <select value={filter} onChange={(e) => setFilter(e.target.value as PaymentFilter)}>
              <option value="ALL">All</option>
              <option value="unpaid">Unpaid</option>
              <option value="paid">Paid</option>
            </select>
          </div>
        </div>
        {visibleInvoices.length === 0 ? <p className="text-sm text-neutral-500">No invoices match this filter.</p> : null}
        <div className="space-y-3">
          {visibleInvoices.map((item) => (
            <div key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-neutral-200 p-4">
              <div>
                <div className="font-medium text-neutral-900">{item.invoice_number}</div>
                <div className="text-sm text-neutral-500">{item.lot_summary}</div>
                <div className="text-sm text-neutral-500">
                  {item.currency} {(item.total_due_minor / 100).toFixed(2)}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={item.payment_status} />
                {item.settlement_status === "invoice_sent" && (item.payment_status === "unpaid" || item.payment_status === "payment_pending") && item.checkout_url ? (
                  <ElevareButton onClick={() => window.open(item.checkout_url ?? "", "_blank", "noopener,noreferrer")}>
                    Pay Now
                  </ElevareButton>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="panel">
        <h2 className="mb-3 font-heading text-xl text-neutral-900">Purchase History</h2>
        {purchaseHistory.length === 0 ? <p className="text-sm text-neutral-500">Paid purchases will appear here.</p> : null}
        <div className="space-y-3">
          {purchaseHistory.map((item) => (
            <div key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-neutral-200 p-4">
              <div>
                <div className="font-medium text-neutral-900">{item.lot_summary}</div>
                <div className="text-sm text-neutral-500">
                  Paid {item.paid_at ? new Date(item.paid_at).toLocaleString() : "-"}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-neutral-900">
                  {item.currency} {(item.total_due_minor / 100).toFixed(2)}
                </span>
                <StatusBadge status={item.payment_status} />
              </div>
            </div>
          ))}
        </div>
      </section>
    </section>
  );
}
