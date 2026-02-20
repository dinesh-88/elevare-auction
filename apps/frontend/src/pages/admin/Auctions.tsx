import { useEffect, useState } from "react";
import { api, type Auction } from "../../api/client";

export default function Auctions() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    api
      .listAuctions()
      .then(setAuctions)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"));
  };

  useEffect(() => {
    load();
  }, []);

  const setStatus = async (action: "start" | "pause" | "end", id: string) => {
    try {
      if (action === "start") await api.startAuction(id);
      if (action === "pause") await api.pauseAuction(id);
      if (action === "end") await api.endAuction(id);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction");
    }
  };

  return (
    <section>
      <h1>Manage Auctions</h1>
      {error ? <p className="error">{error}</p> : null}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Title</th>
              <th>Current Price</th>
              <th>Ends</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {auctions.map((a) => (
              <tr key={a.id}>
                <td>{a.title}</td>
                <td>${a.current_price.toLocaleString()}</td>
                <td>{new Date(a.end_time).toLocaleString()}</td>
                <td className="actions">
                  <button onClick={() => setStatus("start", a.id)}>Start</button>
                  <button onClick={() => setStatus("pause", a.id)}>Pause</button>
                  <button onClick={() => setStatus("end", a.id)}>End</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
