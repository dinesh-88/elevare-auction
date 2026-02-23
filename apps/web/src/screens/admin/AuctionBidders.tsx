import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type AuctionBidderAdminRow } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

type StatusFilter = "ALL" | "PENDING" | "APPROVED" | "REJECTED";

export default function AuctionBidders() {
  const { auctionId } = useParams<{ auctionId: string }>();
  const [rows, setRows] = useState<AuctionBidderAdminRow[]>([]);
  const [status, setStatus] = useState<StatusFilter>("ALL");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const load = async () => {
    if (!auctionId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.listAuctionBidders(auctionId, status === "ALL" ? undefined : status);
      setRows(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load bidders");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [auctionId, status]);

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
                <td>{row.risk_level}</td>
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
                  <Link className="btn-link-brand" to={`/admin/bidders/${row.user_id}`}>View profile</Link>
                </td>
              </tr>
            ))}
            {rows.length === 0 ? (
              <tr>
                <td colSpan={7} className="text-sm text-neutral-500">No bidder requests found.</td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </section>
  );
}
