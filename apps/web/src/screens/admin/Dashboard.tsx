import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Auction, type InstitutionOnboardingState, type User } from "../../api/client";

type HealthRow = {
  label: string;
  status: string;
  tone: "ok" | "warn";
  detail: string;
};

function metricCard(label: string, value: number) {
  return (
    <article className="panel">
      <p className="text-sm text-neutral-500">{label}</p>
      <p className="mt-2 font-heading text-4xl">{value}</p>
    </article>
  );
}

export default function Dashboard() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [bidders, setBidders] = useState<User[]>([]);
  const [onboarding, setOnboarding] = useState<InstitutionOnboardingState | null>(null);
  const [feeCount, setFeeCount] = useState(0);
  const [taxCount, setTaxCount] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    Promise.all([
      api.listAdminAuctions(),
      api.listBidders(),
      api.getInstitutionOnboarding(),
      api.listFeeComponents().catch(() => []),
      api.listTaxRules().catch(() => []),
    ])
      .then(([auctionRows, bidderRows, onboardingState, feeRows, taxRows]) => {
        if (cancelled) return;
        setAuctions(auctionRows);
        setBidders(bidderRows);
        setOnboarding(onboardingState);
        setFeeCount(feeRows.length);
        setTaxCount(taxRows.length);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Failed to load dashboard");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const pendingBidders = bidders.filter((b) => b.approval_status === "PENDING").length;
  const liveAuctions = auctions.filter((auction) => auction.status === "LIVE").length;
  const scheduledAuctions = auctions.filter((auction) => auction.status === "SCHEDULED" || auction.status === "DRAFT").length;
  const hasAuctions = auctions.length > 0;

  const healthRows: HealthRow[] = [
    {
      label: "Email status",
      status: onboarding?.data.support_email ? "Configured" : "Needs review",
      tone: onboarding?.data.support_email ? "ok" : "warn",
      detail: onboarding?.data.support_email || "Add a support email in onboarding.",
    },
    {
      label: "VAT configured",
      status: taxCount > 0 ? "Ready" : "Needs review",
      tone: taxCount > 0 ? "ok" : "warn",
      detail: taxCount > 0 ? `${onboarding?.data.vat_rate_percent ?? 0}% default VAT` : "No VAT rule detected.",
    },
    {
      label: "Fee defaults configured",
      status: feeCount > 0 ? "Ready" : "Needs review",
      tone: feeCount > 0 ? "ok" : "warn",
      detail: feeCount > 0 ? `${feeCount} active fee default${feeCount === 1 ? "" : "s"}` : "No default fee components detected.",
    },
    {
      label: "Legal terms active",
      status: onboarding?.data.legal_terms_markdown ? "Ready" : "Needs review",
      tone: onboarding?.data.legal_terms_markdown ? "ok" : "warn",
      detail: onboarding?.data.legal_terms_markdown ? "Institution legal terms are saved." : "Add legal terms in onboarding.",
    },
    {
      label: "Default currency",
      status: onboarding?.data.default_currency || "Needs review",
      tone: onboarding?.data.default_currency ? "ok" : "warn",
      detail: onboarding?.data.default_currency ? "Used for new auctions and invoices." : "No default currency set.",
    },
  ];

  if (loading) {
    return (
      <section className="page-shell">
        <div className="panel">
          <p>Loading dashboard…</p>
        </div>
      </section>
    );
  }

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Dashboard</h1>
          <p className="section-copy">
            {hasAuctions ? "Snapshot of current auction activity." : "Your institution is activated. The next step is launching the first auction."}
          </p>
        </div>
      </header>

      {error ? <p className="error">{error}</p> : null}

      {!hasAuctions ? (
        <>
          <section
            className="panel mb-6"
            style={{
              background:
                "radial-gradient(circle at top right, rgba(190, 242, 100, 0.35), transparent 40%), linear-gradient(135deg, #f8fafc, #eef2ff 55%, #fefce8)",
            }}
          >
            <p className="mb-2 text-xs font-semibold uppercase tracking-[0.22em] text-neutral-500">Activation Confirmed</p>
            <div className="grid gap-5 lg:grid-cols-[minmax(0,2fr)_minmax(240px,1fr)] lg:items-end">
              <div>
                <h2 className="font-heading text-3xl text-neutral-900">Your institution is ready to go live.</h2>
                <p className="mt-3 max-w-2xl text-sm leading-6 text-neutral-600">
                  Core onboarding is complete. Use the launch checklist below to review defaults, then create the first auction.
                </p>
              </div>
              <div className="flex flex-col gap-3">
                <Link className="btn-brand" to="/admin/create-auction">
                  Create Your First Auction
                </Link>
              </div>
            </div>
          </section>

          <div className="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(320px,1fr)]">
            <section className="panel">
              <h2 className="mb-4 font-heading text-2xl text-neutral-900">Launch Checklist</h2>
              <div className="grid gap-3 sm:grid-cols-3">
                <Link className="rounded-2xl border border-neutral-200 bg-white p-4 text-left transition hover:border-neutral-300" to="/admin/financial-rules">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-500">Review Fees</p>
                  <p className="mt-2 text-sm text-neutral-700">Validate buyer&apos;s premium and tax defaults before publishing the first catalog.</p>
                </Link>
                <Link className="rounded-2xl border border-neutral-200 bg-white p-4 text-left transition hover:border-neutral-300" to="/admin/onboarding?step=5">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-500">Review Legal Terms</p>
                  <p className="mt-2 text-sm text-neutral-700">Confirm the legal wording and support contacts captured during activation.</p>
                </Link>
                <Link className="rounded-2xl border border-neutral-200 bg-white p-4 text-left transition hover:border-neutral-300" to="/admin/staff">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-500">Invite Staff</p>
                  <p className="mt-2 text-sm text-neutral-700">Send role-based invitations and keep a single primary owner in place.</p>
                </Link>
              </div>
            </section>

            <aside className="panel">
              <h2 className="mb-4 font-heading text-2xl text-neutral-900">System Health</h2>
              <div className="space-y-3">
                {healthRows.map((row) => (
                  <div key={row.label} className="rounded-2xl border border-neutral-200 bg-white p-4">
                    <div className="flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-neutral-900">{row.label}</p>
                      <span
                        className="rounded-full px-3 py-1 text-xs font-semibold"
                        style={{
                          backgroundColor: row.tone === "ok" ? "#dcfce7" : "#fef3c7",
                          color: row.tone === "ok" ? "#166534" : "#92400e",
                        }}
                      >
                        {row.status}
                      </span>
                    </div>
                    <p className="mt-2 text-sm text-neutral-600">{row.detail}</p>
                  </div>
                ))}
              </div>
            </aside>
          </div>
        </>
      ) : (
        <>
          <div className="stats-grid">
            {metricCard("Total Auctions", auctions.length)}
            {metricCard("Live Auctions", liveAuctions)}
            {metricCard("Scheduled / Draft", scheduledAuctions)}
            {metricCard("Pending Bidders", pendingBidders)}
          </div>
          <section className="panel mt-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="font-heading text-2xl text-neutral-900">Auction Pipeline</h2>
                <p className="mt-1 text-sm text-neutral-600">Manage active catalogs, bidder approvals, and readiness checks from here.</p>
              </div>
              <div className="flex flex-wrap gap-3">
                <Link className="btn-secondary" to="/admin/auctions">
                  View Auctions
                </Link>
                <Link className="btn-brand" to="/admin/create-auction">
                  Create Auction
                </Link>
              </div>
            </div>
          </section>
        </>
      )}
    </section>
  );
}
