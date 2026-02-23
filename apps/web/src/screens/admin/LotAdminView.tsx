import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ChevronDown, ChevronRight } from "lucide-react";
import { api, type Lot, type LotAuditItem, type LotBid } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";
import { currencySymbol } from "../../utils/currency";
import { useAuth } from "../../hooks/useAuth";

type TabKey = "details" | "bids" | "settlement" | "audit";
type Role = "OWNER" | "ADMIN" | "ACCOUNTING";

const EVENT_LABELS: Record<string, string> = {
  BID_PLACED: "Bid placed",
  LOT_EXTENDED: "Lot extended",
  LOT_OPENED: "Lot opened",
  LOT_CLOSED: "Lot closed",
  LOT_MANUALLY_CLOSED: "Lot manually closed",
  LOT_UPDATED: "Lot updated",
  RESERVE_UPDATED: "Reserve updated",
  WINNER_DETERMINED: "Winner determined",
  INVOICE_GENERATED: "Invoice generated",
  INVOICE_MARKED_PAID: "Invoice marked paid",
  INVOICE_CANCELLED: "Invoice cancelled",
  NOTIFICATION_SENT: "Notification sent",
  NOTIFICATION_FAILED: "Notification failed",
};

const EVENT_OPTIONS = ["ALL", ...Object.keys(EVENT_LABELS)];
const PAGE_LIMIT = 20;

export default function LotAdminView() {
  const { lotId } = useParams<{ lotId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const canViewAudit = (["OWNER", "ADMIN", "ACCOUNTING"] as Role[]).includes((user?.role ?? "BIDDER") as Role);

  const [activeTab, setActiveTab] = useState<TabKey>("audit");
  const [lot, setLot] = useState<Lot | null>(null);
  const [bids, setBids] = useState<LotBid[]>([]);
  const [loadingLot, setLoadingLot] = useState(true);
  const [loadingBids, setLoadingBids] = useState(false);

  const [items, setItems] = useState<LotAuditItem[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [loadingAudit, setLoadingAudit] = useState(false);
  const [auditError, setAuditError] = useState<string | null>(null);
  const [eventType, setEventType] = useState("ALL");
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [filterError, setFilterError] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [showAllMetadata, setShowAllMetadata] = useState<Record<string, boolean>>({});

  useEffect(() => {
    if (!lotId) return;
    setLoadingLot(true);
    api
      .getLot(lotId)
      .then(setLot)
      .catch(() => setLot(null))
      .finally(() => setLoadingLot(false));
  }, [lotId]);

  useEffect(() => {
    if (activeTab !== "bids" || !lotId) return;
    setLoadingBids(true);
    api
      .listLotBids(lotId)
      .then(setBids)
      .catch(() => setBids([]))
      .finally(() => setLoadingBids(false));
  }, [activeTab, lotId]);

  const loadAudit = async (
    cursor?: string,
    append?: boolean,
    overrides?: { eventType?: string; fromDate?: string; toDate?: string },
  ) => {
    if (!lotId) return;
    setLoadingAudit(true);
    if (!append) setAuditError(null);
    try {
      const eventFilter = overrides?.eventType ?? eventType;
      const fromFilter = overrides?.fromDate ?? fromDate;
      const toFilter = overrides?.toDate ?? toDate;
      const result = await api.listLotAuditLogs(lotId, {
        limit: PAGE_LIMIT,
        cursor,
        event_type: eventFilter === "ALL" ? undefined : eventFilter,
        from: fromFilter || undefined,
        to: toFilter || undefined,
      });
      setItems((prev) => (append ? [...prev, ...result.items] : result.items));
      setNextCursor(result.next_cursor);
    } catch (err) {
      const status = (err as Error & { status?: number }).status;
      if (status === 401) {
        navigate("/login");
        return;
      }
      if (status === 403) {
        setAuditError("You do not have permission to view audit logs.");
        return;
      }
      setAuditError("Failed to load audit log.");
    } finally {
      setLoadingAudit(false);
    }
  };

  useEffect(() => {
    if (activeTab !== "audit") return;
    loadAudit();
  }, [activeTab, lotId]);

  const applyFilters = () => {
    if (fromDate && toDate && fromDate > toDate) {
      setFilterError("From date must be before To date.");
      return;
    }
    setFilterError(null);
    loadAudit(undefined, false);
  };

  const resetFilters = () => {
    setEventType("ALL");
    setFromDate("");
    setToDate("");
    setFilterError(null);
    loadAudit(undefined, false, { eventType: "ALL", fromDate: "", toDate: "" });
  };

  const currency = lot?.metadata?.currency && typeof lot.metadata.currency === "string" ? lot.metadata.currency : "USD";
  const symbol = currencySymbol(currency);
  const header = useMemo(() => {
    if (!lot) return "Lot";
    return `Lot ${lot.lot_number} · ${lot.title}`;
  }, [lot]);

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">{header}</h1>
          <p className="section-copy">Admin lot operations and audit timeline.</p>
        </div>
        <div className="flex gap-2">
          {lot ? (
            <Link className="btn-link-brand" to={`/admin/auctions/${lot.auction_id}/lots`}>
              ← Back to Lots
            </Link>
          ) : null}
        </div>
      </header>

      <div className="panel mb-4">
        <div className="flex flex-wrap gap-2">
          <TabButton active={activeTab === "details"} onClick={() => setActiveTab("details")} label="Details" />
          <TabButton active={activeTab === "bids"} onClick={() => setActiveTab("bids")} label="Bids" />
          <TabButton active={activeTab === "settlement"} onClick={() => setActiveTab("settlement")} label="Settlement" />
          <TabButton active={activeTab === "audit"} onClick={() => setActiveTab("audit")} label="Audit Log" />
        </div>
      </div>

      {activeTab === "details" ? (
        <div className="panel">
          {loadingLot ? <p className="text-sm text-neutral-500">Loading lot details...</p> : null}
          {!loadingLot && !lot ? <p className="error">Failed to load lot details.</p> : null}
          {lot ? (
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <DetailRow label="Status" value={<StatusBadge status={lot.status} />} />
              <DetailRow label="Current Bid" value={`${symbol}${lot.current_price.toLocaleString()}`} />
              <DetailRow label="Minimum Next Bid" value={`${symbol}${(lot.minimum_next_bid_minor ?? lot.current_price + lot.min_increment).toLocaleString()}`} />
              <DetailRow label="Ended At" value={lot.ended_at ? formatDateTime(lot.ended_at) : "Live"} />
            </div>
          ) : null}
        </div>
      ) : null}

      {activeTab === "bids" ? (
        <div className="panel">
          {loadingBids ? <p className="text-sm text-neutral-500">Loading bid history...</p> : null}
          {!loadingBids && bids.length === 0 ? <p className="text-sm text-neutral-500">No bids yet.</p> : null}
          {bids.length > 0 ? (
            <div className="space-y-2">
              {bids.map((bid) => (
                <div key={bid.id} className="rounded-md border border-neutral-200 bg-white px-3 py-2 text-sm">
                  <span className="font-semibold text-neutral-900">
                    {symbol}
                    {bid.amount_minor.toLocaleString()}
                  </span>
                  <span className="mx-2 text-neutral-500">•</span>
                  <span className="text-neutral-700">{bid.bidder_display}</span>
                  <span className="mx-2 text-neutral-500">•</span>
                  <span className="text-neutral-500">{formatDateTime(bid.created_at)}</span>
                </div>
              ))}
            </div>
          ) : null}
        </div>
      ) : null}

      {activeTab === "settlement" ? (
        <div className="panel">
          {lot ? (
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm text-neutral-600">Use auction settlement for invoice actions and CSV export.</p>
              <Link className="btn-brand" to={`/admin/auctions/${lot.auction_id}/settlement`}>
                Open Settlement
              </Link>
            </div>
          ) : null}
        </div>
      ) : null}

      {activeTab === "audit" ? (
        <div className="space-y-4">
          <div className="panel">
            <div className="grid grid-cols-1 gap-3 md:grid-cols-5">
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">Event Type</label>
                <select value={eventType} onChange={(e) => setEventType(e.target.value)}>
                  {EVENT_OPTIONS.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt === "ALL" ? "All" : (EVENT_LABELS[opt] ?? opt)}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">From</label>
                <input type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)} />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-neutral-600">To</label>
                <input type="date" value={toDate} onChange={(e) => setToDate(e.target.value)} />
              </div>
              <div className="md:col-span-2 flex items-end gap-2">
                <ElevareButton onClick={applyFilters}>Apply Filters</ElevareButton>
                <ElevareButton variant="secondary" onClick={resetFilters}>
                  Reset
                </ElevareButton>
              </div>
            </div>
            {filterError ? <p className="mt-2 text-sm text-red-600">{filterError}</p> : null}
          </div>

          {auditError ? (
            <div className="panel border border-red-200 bg-red-50">
              <p className="text-sm text-red-700">{auditError}</p>
              <div className="mt-2">
                <ElevareButton variant="secondary" onClick={() => loadAudit(undefined, false)}>
                  Retry
                </ElevareButton>
              </div>
            </div>
          ) : null}

          {loadingAudit && items.length === 0 ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="h-20 animate-pulse rounded-lg border border-neutral-200 bg-neutral-100" />
              ))}
            </div>
          ) : null}

          {!loadingAudit && !auditError && items.length === 0 ? (
            <div className="panel">
              <p className="text-sm text-neutral-600">No audit events found for this lot.</p>
            </div>
          ) : null}

          {items.map((item) => {
            const isOpen = Boolean(expanded[item.id]);
            const metadataText = JSON.stringify(item.metadata ?? {}, null, 2);
            const truncated = metadataText.length > 900;
            const shouldTruncate = truncated && !showAllMetadata[item.id];
            const metadataValue = shouldTruncate ? `${metadataText.slice(0, 900)}…` : metadataText;
            const title = item.summary || EVENT_LABELS[item.event_type] || item.event_type;
            const actor = item.actor?.display_name || "System";
            return (
              <article key={item.id} className="rounded-lg border border-neutral-200 bg-white p-4">
                <div className="grid grid-cols-1 gap-3 md:grid-cols-[180px_1fr_auto] md:items-start">
                  <div className="text-sm text-neutral-600">{formatDateTime(item.created_at)}</div>
                  <div>
                    <div className="mb-1 flex flex-wrap items-center gap-2">
                      <span className="rounded-full bg-neutral-100 px-2 py-0.5 text-xs font-medium text-neutral-800">{item.event_type}</span>
                    </div>
                    <p className="font-medium text-neutral-900">{title}</p>
                    <p className="text-sm text-neutral-600">Actor: {actor}</p>
                  </div>
                  <button
                    type="button"
                    aria-expanded={isOpen}
                    onClick={() => setExpanded((prev) => ({ ...prev, [item.id]: !prev[item.id] }))}
                    className="inline-flex items-center gap-1 rounded-md border border-neutral-200 px-2 py-1 text-sm text-neutral-700 hover:bg-neutral-50"
                  >
                    {isOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                    Details
                  </button>
                </div>
                {isOpen ? (
                  <div className="mt-3 rounded-md bg-neutral-50 p-3">
                    <pre className="max-h-72 overflow-auto text-xs text-neutral-700">{metadataValue}</pre>
                    {truncated ? (
                      <button
                        type="button"
                        onClick={() => setShowAllMetadata((prev) => ({ ...prev, [item.id]: !prev[item.id] }))}
                        className="mt-2 text-xs font-medium text-brand-primary"
                      >
                        {showAllMetadata[item.id] ? "Show less" : "Show more"}
                      </button>
                    ) : null}
                  </div>
                ) : null}
              </article>
            );
          })}

          {nextCursor && items.length < 200 ? (
            <div className="flex justify-center pt-2">
              <ElevareButton variant="secondary" onClick={() => loadAudit(nextCursor, true)} disabled={loadingAudit}>
                {loadingAudit ? "Loading..." : "Load More"}
              </ElevareButton>
            </div>
          ) : null}

          {!canViewAudit && !loadingAudit ? (
            <div className="panel border border-amber-200 bg-amber-50">
              <p className="text-sm text-amber-800">You do not have permission to view audit logs.</p>
            </div>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

function TabButton({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-md border px-3 py-1.5 text-sm ${active ? "border-neutral-900 bg-neutral-900 text-white" : "border-neutral-300 bg-white text-neutral-700"}`}
    >
      {label}
    </button>
  );
}

function DetailRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-md border border-neutral-200 bg-neutral-50 p-3">
      <p className="text-xs uppercase tracking-[0.06em] text-neutral-500">{label}</p>
      <div className="mt-1 text-sm text-neutral-900">{value}</div>
    </div>
  );
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
    timeZoneName: "short",
  }).format(date);
}
