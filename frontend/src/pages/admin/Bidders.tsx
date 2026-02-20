import { useEffect, useState } from "react";
import { api, type User } from "../../api/client";

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
    <section>
      <h1>Bidders</h1>
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
                <td>{b.approval_status}</td>
                <td>
                  {b.approval_status === "PENDING" ? <button onClick={() => approve(b.id)}>Approve</button> : "-"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
