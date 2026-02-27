import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type AuctionBidderAdminRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

type StatusFilter = "ALL" | "PENDING" | "APPROVED" | "REJECTED";
type RiskFilter = "ALL" | "low" | "medium" | "high";
type DepositStatusFilter = "ALL" | "not_required" | "required" | "pending" | "received" | "waived";

type EditorState = {
  risk_level: AuctionBidderAdminRow["risk_level"];
  max_bid_limit_minor: string;
  deposit_required: boolean;
  deposit_amount_minor: string;
  deposit_status: AuctionBidderAdminRow["deposit_status"];
  internal_notes: string;
};

export default function AuctionBidders() {
  const { auctionId } = useParams<{ auctionId: string }>();
  const [rows, setRows] = useState<AuctionBidderAdminRow[]>([]);
  const [status, setStatus] = useState<StatusFilter>("ALL");
  const [riskLevel, setRiskLevel] = useState<RiskFilter>("ALL");
  const [depositRequired, setDepositRequired] = useState<"ALL" | "true" | "false">("ALL");
  const [depositStatus, setDepositStatus] = useState<DepositStatusFilter>("ALL");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [editors, setEditors] = useState<Record<string, EditorState>>({});

  const load = async () => {
    if (!auctionId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.listAuctionBidders(auctionId, {
        status: status === "ALL" ? undefined : status,
        risk_level: riskLevel === "ALL" ? undefined : riskLevel,
        deposit_required: depositRequired === "ALL" ? undefined : depositRequired === "true",
        deposit_status: depositStatus === "ALL" ? undefined : depositStatus,
      });
      setRows(data);
      setEditors(
        Object.fromEntries(
          data.map((row) => [
            row.user_id,
            {
              risk_level: row.risk_level,
              max_bid_limit_minor: row.max_bid_limit_minor != null ? String(row.max_bid_limit_minor) : "",
              deposit_required: row.deposit_required,
              deposit_amount_minor: row.deposit_amount_minor != null ? String(row.deposit_amount_minor) : "",
              deposit_status: row.deposit_status,
              internal_notes: row.internal_notes ?? "",
            },
          ]),
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load bidders");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [auctionId, status, riskLevel, depositRequired, depositStatus]);

  const approve = async (userID: string) => {
    if (!auctionId) return;
    try {
      await api.approveAuctionBidder(auctionId, userID);
      setMessage("Bidder approved.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to approve bidder");
    }
  };

  const reject = async (userID: string) => {
    if (!auctionId) return;
    try {
      await api.rejectAuctionBidder(auctionId, userID);
      setMessage("Bidder rejected.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reject bidder");
    }
  };

  const toggleSuspend = async (row: AuctionBidderAdminRow) => {
    try {
      await api.updateBidderProfile(row.user_id, { status: row.tenant_status === "SUSPENDED" ? "ACTIVE" : "SUSPENDED" });
      setMessage(row.tenant_status === "SUSPENDED" ? "Bidder unsuspended." : "Bidder suspended.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update bidder status");
    }
  };

  const updateEditor = (userId: string, patch: Partial<EditorState>) => {
    setEditors((prev) => ({ ...prev, [userId]: { ...prev[userId], ...patch } }));
  };

  const saveControls = async (row: AuctionBidderAdminRow) => {
    if (!auctionId) return;
    const editor = editors[row.user_id];
    if (!editor) return;
    try {
      await api.updateAuctionBidderApproval(auctionId, row.user_id, {
        risk_level: editor.risk_level,
        max_bid_limit_minor: editor.max_bid_limit_minor.trim() === "" ? null : Number(editor.max_bid_limit_minor),
        deposit_required: editor.deposit_required,
        deposit_amount_minor: editor.deposit_amount_minor.trim() === "" ? null : Number(editor.deposit_amount_minor),
        deposit_status: editor.deposit_status,
        internal_notes: editor.internal_notes.trim() === "" ? null : editor.internal_notes,
      });
      setMessage("Bidder controls updated.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update bidder controls");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Auction Bidders</h1>
          <p className="section-copy">Review and decide bidder approvals for this auction.</p>
        </div>
        <div className="flex gap-2">
          <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/lots`}>← Back to Lots</Link>
        </div>
      </header>

      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <div className="panel mb-4 flex items-center gap-3">
        <label className="text-sm text-neutral-700">Status</label>
        <select value={status} onChange={(e) => setStatus(e.target.value as StatusFilter)}>
          <option value="ALL">All</option>
          <option value="PENDING">Pending</option>
          <option value="APPROVED">Approved</option>
          <option value="REJECTED">Rejected</option>
        </select>
        <label className="text-sm text-neutral-700">Risk</label>
        <select value={riskLevel} onChange={(e) => setRiskLevel(e.target.value as RiskFilter)}>
          <option value="ALL">All</option>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
        <label className="text-sm text-neutral-700">Deposit Required</label>
        <select value={depositRequired} onChange={(e) => setDepositRequired(e.target.value as "ALL" | "true" | "false")}>
          <option value="ALL">All</option>
          <option value="true">Yes</option>
          <option value="false">No</option>
        </select>
        <label className="text-sm text-neutral-700">Deposit Status</label>
        <select value={depositStatus} onChange={(e) => setDepositStatus(e.target.value as DepositStatusFilter)}>
          <option value="ALL">All</option>
          <option value="not_required">Not required</option>
          <option value="required">Required</option>
          <option value="pending">Pending</option>
          <option value="received">Received</option>
          <option value="waived">Waived</option>
        </select>
      </div>

      {loading ? <p className="text-sm text-neutral-500">Loading bidders...</p> : null}

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Bidder</th>
              <th>Approval</th>
              <th>Tenant Status</th>
              <th>Risk</th>
              <th>Limit</th>
              <th>Deposit</th>
              <th>Notes</th>
              <th>Requested</th>
              <th>Decided</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.user_id}>
                <td>{row.email}</td>
                <td><StatusBadge status={row.approval_status} /></td>
                <td><StatusBadge status={row.tenant_status} /></td>
                <td>
                  <select value={editors[row.user_id]?.risk_level ?? row.risk_level} onChange={(e) => updateEditor(row.user_id, { risk_level: e.target.value as EditorState["risk_level"] })}>
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                  </select>
                </td>
                <td>
                  <input
                    type="number"
                    min="0"
                    value={editors[row.user_id]?.max_bid_limit_minor ?? ""}
                    onChange={(e) => updateEditor(row.user_id, { max_bid_limit_minor: e.target.value })}
                    placeholder="No limit"
                  />
                </td>
                <td>
                  <div className="flex flex-col gap-2">
                    <label className="text-xs text-neutral-600">
                      <input
                        type="checkbox"
                        className="mr-2"
                        checked={editors[row.user_id]?.deposit_required ?? row.deposit_required}
                        onChange={(e) => updateEditor(row.user_id, { deposit_required: e.target.checked })}
                      />
                      Required
                    </label>
                    <input
                      type="number"
                      min="0"
                      value={editors[row.user_id]?.deposit_amount_minor ?? ""}
                      onChange={(e) => updateEditor(row.user_id, { deposit_amount_minor: e.target.value })}
                      placeholder="Amount"
                    />
                    <select
                      value={editors[row.user_id]?.deposit_status ?? row.deposit_status}
                      onChange={(e) => updateEditor(row.user_id, { deposit_status: e.target.value as EditorState["deposit_status"] })}
                    >
                      <option value="not_required">Not required</option>
                      <option value="required">Required</option>
                      <option value="pending">Pending</option>
                      <option value="received">Received</option>
                      <option value="waived">Waived</option>
                    </select>
                  </div>
                </td>
                <td>
                  <textarea
                    rows={3}
                    value={editors[row.user_id]?.internal_notes ?? ""}
                    onChange={(e) => updateEditor(row.user_id, { internal_notes: e.target.value })}
                    placeholder="Internal notes"
                  />
                </td>
                <td>{new Date(row.requested_at).toLocaleString()}</td>
                <td>{row.decided_at ? new Date(row.decided_at).toLocaleString() : "-"}</td>
                <td className="actions">
                  {row.approval_status !== "APPROVED" ? <ElevareButton onClick={() => approve(row.user_id)}>Approve</ElevareButton> : null}
                  {row.approval_status !== "REJECTED" ? (
                    <ElevareButton variant="secondary" onClick={() => reject(row.user_id)}>Reject</ElevareButton>
                  ) : null}
                  <ElevareButton variant="secondary" onClick={() => toggleSuspend(row)}>
                    {row.tenant_status === "SUSPENDED" ? "Unsuspend" : "Suspend"}
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => saveControls(row)}>Save Controls</ElevareButton>
                  <Link className="btn-link-brand" to={`/admin/bidders/${row.user_id}`}>View profile</Link>
                </td>
              </tr>
            ))}
            {rows.length === 0 ? (
              <tr>
                <td colSpan={10} className="text-sm text-neutral-500">No bidder requests found.</td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </section>
  );
}
