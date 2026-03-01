import { useEffect, useState } from "react";
import { api, type SettlementInvoiceRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

type PaymentStatus = "unpaid" | "payment_pending" | "paid" | "refunded" | "partially_refunded";

const PAGE_SIZE = 25;

export default function Reconciliation() {
  const [rows, setRows] = useState<SettlementInvoiceRow[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState<PaymentStatus | "ALL">("ALL");
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async (nextOffset = 0) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.listReconciliation({
        status: status === "ALL" ? undefined : status,
        search: search.trim() || undefined,
        limit: PAGE_SIZE,
        offset: nextOffset,
      });
      setRows(response.items);
      setTotal(response.total);
      setOffset(nextOffset);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load reconciliation");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(0);
  }, [status]);

  const pageStart = total === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + PAGE_SIZE, total);

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Reconciliation</h1>
          <p className="section-copy">Tenant-wide settlement and payment reconciliation.</p>
        </div>
      </header>

      {error ? <p className="error mb-3">{error}</p> : null}
      {loading ? <p className="text-sm text-neutral-500">Loading reconciliation...</p> : null}

      <div className="panel mb-4">
        <div className="flex flex-wrap items-center gap-2">
          <select value={status} onChange={(e) => setStatus(e.target.value as PaymentStatus | "ALL")}>
            <option value="ALL">All Statuses</option>
            <option value="unpaid">Unpaid</option>
            <option value="payment_pending">Pending</option>
            <option value="paid">Paid</option>
            <option value="refunded">Refunded</option>
            <option value="partially_refunded">Partially Refunded</option>
          </select>
          <input
            placeholder="Search invoice, bidder, or lot"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="min-w-[240px]"
          />
          <ElevareButton variant="secondary" onClick={() => load(0)}>
            Search
          </ElevareButton>
        </div>
      </div>

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Invoice #</th>
              <th>Bidder</th>
              <th>Lot</th>
              <th>Settlement</th>
              <th>Payment</th>
              <th>Total Due</th>
              <th>Issued</th>
              <th>Paid</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>{row.invoice_number}</td>
                <td>
                  <div>{row.bidder_name}</div>
                  <div className="text-xs text-neutral-500">{row.bidder_email}</div>
                </td>
                <td>{row.lot_summary}</td>
                <td><StatusBadge status={row.settlement_status} /></td>
                <td><StatusBadge status={row.payment_status} /></td>
                <td>{row.currency} {(row.total_due_minor / 100).toFixed(2)}</td>
                <td>{new Date(row.issued_at).toLocaleString()}</td>
                <td>{row.paid_at ? new Date(row.paid_at).toLocaleString() : "-"}</td>
                <td className="actions">
                  {row.checkout_url ? (
                    <ElevareButton variant="secondary" onClick={() => window.open(row.checkout_url ?? "", "_blank", "noopener,noreferrer")}>
                      Checkout
                    </ElevareButton>
                  ) : (
                    <span className="text-xs text-neutral-500">No checkout</span>
                  )}
                </td>
              </tr>
            ))}
            {rows.length === 0 ? (
              <tr>
                <td colSpan={9} className="text-center text-sm text-neutral-500">No invoices found.</td>
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
    </section>
  );
}
