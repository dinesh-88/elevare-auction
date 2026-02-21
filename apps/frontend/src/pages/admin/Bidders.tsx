import { useEffect, useState } from "react";
import { api, type User } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

export default function Bidders() {
  const [bidders, setBidders] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);

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
      await api.approveBidder(id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Approval failed");
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
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Email</th>
              <th>Email Verified</th>
              <th>Approval</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {bidders.map((b) => (
              <tr key={b.id}>
                <td>{b.email}</td>
                <td>{b.email_verified ? "Yes" : "No"}</td>
                <td><StatusBadge status={b.approval_status} /></td>
                <td>
                  {b.approval_status === "PENDING" ? (
                    <ElevareButton onClick={() => approve(b.id)}>Approve</ElevareButton>
                  ) : (
                    "-"
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
