import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, type User } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

export default function Bidders() {
  const [bidders, setBidders] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const load = () => {
    api
      .listBidders()
      .then(setBidders)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load bidders"));
  };

  useEffect(() => {
    load();
  }, []);

  const approve = async (id: string) => {
    try {
      setError(null);
      await api.approveBidder(id);
      setMessage("Bidder approved.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Approval failed");
    }
  };

  const reject = async (id: string) => {
    try {
      setError(null);
      await api.rejectBidder(id);
      setMessage("Bidder declined.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Decline failed");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Bidders</h1>
          <p className="section-copy">Approve or review bidder access.</p>
        </div>
      </header>
      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Email</th>
              <th>Role</th>
              <th>Email Verified</th>
              <th>Approval</th>
              <th>KYC</th>
              <th>Registered</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {bidders.map((b) => (
              <tr key={b.id}>
                <td>{b.email}</td>
                <td>{b.role}</td>
                <td>{b.email_verified ? "Yes" : "No"}</td>
                <td><StatusBadge status={b.approval_status} /></td>
                <td><StatusBadge status={b.kyc_status} /></td>
                <td>{b.created_at ? new Date(b.created_at).toLocaleString() : "-"}</td>
                <td>
                  <div className="flex gap-2">
                    <Link className="btn-link-brand" to={`/admin/bidders/${b.id}`}>Profile</Link>
                    {b.approval_status !== "APPROVED" ? (
                      <ElevareButton onClick={() => approve(b.id)}>Approve</ElevareButton>
                    ) : null}
                    {b.approval_status !== "REJECTED" ? (
                      <ElevareButton variant="secondary" onClick={() => reject(b.id)}>Decline</ElevareButton>
                    ) : null}
                  </div>
                </td>
              </tr>
            ))}
            {bidders.length === 0 ? (
              <tr>
                <td colSpan={7} className="text-sm text-neutral-500">
                  No bidders found.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </section>
  );
}
