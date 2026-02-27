import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type AuctionSettlementSummary, type SettlementInvoiceRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";
import { useAuth } from "../../hooks/useAuth";

type PaymentStatus = "UNPAID" | "PAID" | "CANCELLED";

const PAGE_SIZE = 25;

export default function Settlement() {
  const { auctionId } = useParams<{ auctionId: string }>();
  const { user } = useAuth();
  const canSettle = user?.role === "OWNER" || user?.role === "ADMIN" || user?.role === "ACCOUNTING";

  const [summary, setSummary] = useState<AuctionSettlementSummary | null>(null);
  const [rows, setRows] = useState<SettlementInvoiceRow[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState<PaymentStatus | "ALL">("ALL");
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [paidModalInvoice, setPaidModalInvoice] = useState<SettlementInvoiceRow | null>(null);
  const [paidReference, setPaidReference] = useState("");
  const [cancelModalInvoice, setCancelModalInvoice] = useState<SettlementInvoiceRow | null>(null);
  const [cancelReason, setCancelReason] = useState("");

  const settlementStatus = useMemo(() => {
    if (!summary) return "OPEN";
    if (summary.invoice_count === 0) return "OPEN";
    if (summary.outstanding_minor <= 0 && summary.unpaid_count === 0) return "FULLY_SETTLED";
    if (summary.paid_count > 0) return "PARTIAL";
    return "OPEN";
  }, [summary]);

  const load = async (nextOffset = offset) => {
    if (!auctionId) return;
    setLoading(true);
    setError(null);
    try {
      const [summaryData, listData] = await Promise.all([
        api.getAuctionSettlementSummary(auctionId),
        api.listAuctionInvoices(auctionId, {
          status: status === "ALL" ? undefined : status,
          search: search.trim() || undefined,
          limit: PAGE_SIZE,
          offset: nextOffset,
        }),
      ]);
      setSummary(summaryData);
      setRows(listData.items);
      setTotal(listData.total);
      setOffset(nextOffset);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load settlement");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [auctionId, status]);

  const applySearch = () => load(0);

  const refresh = () => load(offset);

  const markPaid = async () => {
    if (!paidModalInvoice) return;
    try {
      await api.markInvoicePaid(paidModalInvoice.id, paidReference.trim() || undefined);
      setPaidModalInvoice(null);
      setPaidReference("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to mark paid");
    }
  };

  const cancelInvoice = async () => {
    if (!cancelModalInvoice) return;
    if (!cancelReason.trim()) {
      setError("Cancel reason is required");
      return;
    }
    try {
      await api.cancelInvoice(cancelModalInvoice.id, cancelReason.trim());
      setCancelModalInvoice(null);
      setCancelReason("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to cancel invoice");
    }
  };

  const exportCSV = async () => {
    if (!auctionId) return;
    try {
      const blob = await api.downloadSettlementCSV(auctionId, {
        status: status === "ALL" ? undefined : status,
        search: search.trim() || undefined,
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `settlement-${auctionId}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to export CSV");
    }
  };

  const pageStart = total === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + PAGE_SIZE, total);

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Settlement</h1>
          <p className="section-copy">Auction invoice settlement summary and actions.</p>
        </div>
        <div className="flex items-center gap-2">
          <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/lots`}>
            ← Back to Lots
          </Link>
          <ElevareButton onClick={exportCSV} disabled={!canSettle}>
            Export CSV
          </ElevareButton>
        </div>
      </header>

      {error ? <p className="error mb-3">{error}</p> : null}
      {loading ? <p className="text-sm text-neutral-500">Loading settlement...</p> : null}

      <div className="grid grid-cols-1 gap-3 md:grid-cols-3 xl:grid-cols-6 mb-6">
        <SummaryCard label="Total Hammer" value={formatMoney(summary?.currency, summary?.total_hammer_minor ?? 0)} />
        <SummaryCard label="Total Fees" value={formatMoney(summary?.currency, summary?.total_fees_minor ?? 0)} />
        <SummaryCard label="Total Tax" value={formatMoney(summary?.currency, summary?.total_tax_minor ?? 0)} />
        <SummaryCard label="Total Due" value={formatMoney(summary?.currency, summary?.total_due_minor ?? 0)} />
        <SummaryCard label="Total Paid" value={formatMoney(summary?.currency, summary?.total_paid_minor ?? 0)} />
        <SummaryCard label="Outstanding" value={formatMoney(summary?.currency, summary?.outstanding_minor ?? 0)} />
      </div>

      <div className="panel mb-6 flex items-center justify-between">
        <div>
          <h2 className="font-heading text-lg text-neutral-900">Settlement Status</h2>
          <p className="text-sm text-neutral-500">Derived from paid and outstanding totals.</p>
        </div>
        <StatusBadge status={settlementStatus} />
      </div>

      <div className="panel mb-4">
        <div className="flex flex-wrap items-center gap-2">
          <select value={status} onChange={(e) => setStatus(e.target.value as PaymentStatus | "ALL")}> 
            <option value="ALL">All Statuses</option>
            <option value="UNPAID">Unpaid</option>
            <option value="PAID">Paid</option>
            <option value="CANCELLED">Cancelled</option>
          </select>
          <input
            placeholder="Search bidder email"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="min-w-[220px]"
          />
          <ElevareButton variant="secondary" onClick={applySearch}>Search</ElevareButton>
        </div>
      </div>

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Invoice #</th>
              <th>Bidder</th>
              <th>Lot</th>
              <th>Total Due</th>
              <th>Status</th>
              <th>Issued</th>
              <th>Paid</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id || `${row.invoice_number}-${row.issued_at}`}>
                <td>{row.invoice_number}</td>
                <td>
                  <div>{row.bidder_name}</div>
                  <div className="text-xs text-neutral-500">{row.bidder_email}</div>
                </td>
                <td>{row.lot_summary}</td>
                <td>{formatMoney(row.currency, row.total_due_minor)}</td>
                <td><StatusBadge status={row.payment_status} /></td>
                <td>{new Date(row.issued_at).toLocaleString()}</td>
                <td>{row.paid_at ? new Date(row.paid_at).toLocaleString() : "-"}</td>
                <td className="actions">
                  <ElevareButton variant="secondary" onClick={async () => {
                    const blob = await api.downloadInvoicePDF(row.id);
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement("a");
                    a.href = url;
                    a.download = `invoice-${row.invoice_number}.pdf`;
                    a.click();
                    URL.revokeObjectURL(url);
                  }}>
                    View Invoice
                  </ElevareButton>
                  {canSettle && row.payment_status !== "PAID" && row.payment_status !== "CANCELLED" ? (
                    <ElevareButton onClick={() => setPaidModalInvoice(row)}>Mark Paid</ElevareButton>
                  ) : null}
                  {canSettle && row.payment_status === "UNPAID" ? (
                    <ElevareButton variant="secondary" onClick={() => setCancelModalInvoice(row)}>Cancel</ElevareButton>
                  ) : null}
                </td>
              </tr>
            ))}
            {rows.length === 0 ? (
              <tr>
                <td colSpan={8} className="text-center text-sm text-neutral-500">No invoices found.</td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>

      <div className="mt-4 flex items-center justify-between text-sm text-neutral-600">
        <span>
          Showing {pageStart}-{pageEnd} of {total}
        </span>
        <div className="flex gap-2">
          <ElevareButton variant="secondary" disabled={offset <= 0} onClick={() => load(Math.max(0, offset - PAGE_SIZE))}>
            Previous
          </ElevareButton>
          <ElevareButton variant="secondary" disabled={offset + PAGE_SIZE >= total} onClick={() => load(offset + PAGE_SIZE)}>
            Next
          </ElevareButton>
        </div>
      </div>

      {paidModalInvoice ? (
        <Modal
          title={`Mark ${paidModalInvoice.invoice_number} as paid`}
          onClose={() => {
            setPaidModalInvoice(null);
            setPaidReference("");
          }}
          onConfirm={markPaid}
          confirmLabel="Confirm"
        >
          <label className="text-sm">Payment reference (optional)</label>
          <input value={paidReference} onChange={(e) => setPaidReference(e.target.value)} maxLength={128} />
        </Modal>
      ) : null}

      {cancelModalInvoice ? (
        <Modal
          title={`Cancel ${cancelModalInvoice.invoice_number}`}
          onClose={() => {
            setCancelModalInvoice(null);
            setCancelReason("");
          }}
          onConfirm={cancelInvoice}
          confirmLabel="Cancel Invoice"
        >
          <label className="text-sm">Reason</label>
          <textarea value={cancelReason} onChange={(e) => setCancelReason(e.target.value)} rows={3} maxLength={1024} />
        </Modal>
      ) : null}
    </section>
  );
}

function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4">
      <div className="text-xs uppercase tracking-[0.06em] text-neutral-500">{label}</div>
      <div className="mt-1 font-heading text-xl text-neutral-900">{value}</div>
    </div>
  );
}

function Modal({
  title,
  onClose,
  onConfirm,
  confirmLabel,
  children,
}: {
  title: string;
  onClose: () => void;
  onConfirm: () => void;
  confirmLabel: string;
  children: React.ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/30 p-4">
      <div className="w-full max-w-lg rounded-lg bg-white p-5 shadow-xl">
        <h3 className="mb-3 font-heading text-xl text-neutral-900">{title}</h3>
        <div className="space-y-2">{children}</div>
        <div className="mt-5 flex justify-end gap-2">
          <ElevareButton variant="secondary" onClick={onClose}>Close</ElevareButton>
          <ElevareButton onClick={onConfirm}>{confirmLabel}</ElevareButton>
        </div>
      </div>
    </div>
  );
}

function formatMoney(currency = "USD", minor = 0): string {
  const safeCurrency = normalizeCurrencyCode(currency);
  const decimals = currencyDecimals(safeCurrency);
  const amount = minor / 10 ** decimals;
  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: safeCurrency,
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    }).format(amount);
  } catch {
    return `${safeCurrency} ${amount.toFixed(decimals)}`;
  }
}

function currencyDecimals(currency: string): number {
  switch ((currency || "").toUpperCase()) {
    case "JPY":
    case "KRW":
    case "VND":
      return 0;
    default:
      return 2;
  }
}

function normalizeCurrencyCode(currency: string): string {
  const normalized = (currency || "").trim().toUpperCase();
  if (/^[A-Z]{3}$/.test(normalized)) {
    return normalized;
  }
  return "USD";
}
