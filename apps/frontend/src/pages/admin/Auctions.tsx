import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { api, type Auction } from "../../api/client";
import AuctionForm, { fromAuction, toPayload, type AuctionFormState } from "../../components/admin/AuctionForm";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

export default function Auctions() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [form, setForm] = useState<AuctionFormState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const selectedAuction = useMemo(() => auctions.find((a) => a.id === selectedId) ?? null, [auctions, selectedId]);

  const load = () => {
    api
      .listAdminAuctions()
      .then((rows) => {
        setAuctions(rows);
        if (selectedId) {
          const fresh = rows.find((a) => a.id === selectedId);
          if (!fresh) {
            setSelectedId(null);
            setForm(null);
          }
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"));
  };

  useEffect(() => {
    load();
  }, []);

  const setStatus = async (action: "start" | "pause" | "end", id: string) => {
    try {
      setError(null);
      setMessage(null);
      if (action === "start") await api.startAuction(id);
      if (action === "pause") await api.pauseAuction(id);
      if (action === "end") await api.endAuction(id);
      setMessage(`Auction ${action}ed.`);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction status");
    }
  };

  const removeAuction = async (id: string) => {
    if (!window.confirm("Delete this auction?")) return;
    try {
      setError(null);
      setMessage(null);
      await api.deleteAuction(id);
      setMessage("Auction deleted.");
      if (selectedId === id) {
        setSelectedId(null);
        setForm(null);
      }
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete auction");
    }
  };

  const startEdit = (auction: Auction) => {
    setSelectedId(auction.id);
    setForm(fromAuction(auction));
    setError(null);
    setMessage(null);
  };

  const saveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAuction || !form) return;
    try {
      setError(null);
      setMessage(null);
      await api.updateAuction(selectedAuction.id, toPayload(form));
      setMessage("Auction updated.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Manage Auctions</h1>
          <p className="section-copy">Control auction lifecycle and configuration.</p>
        </div>
        <Link className="btn-brand" to="/admin/create-auction">
          New Auction
        </Link>
      </header>
      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <div className="table-wrap mb-6">
        <table>
          <thead>
            <tr>
              <th>Title</th>
              <th>Status</th>
              <th>Current Price</th>
              <th>Ends</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {auctions.map((a) => (
              <tr key={a.id}>
                <td>{a.title}</td>
                <td><StatusBadge status={a.status} /></td>
                <td>${a.current_price.toLocaleString()}</td>
                <td>{new Date(a.end_time).toLocaleString()}</td>
                <td className="actions">
                  <Link className="btn-brand" to={`/admin/auctions/${a.id}/lots`}>
                    Lots
                  </Link>
                  <ElevareButton variant="secondary" onClick={() => startEdit(a)}>
                    Edit
                  </ElevareButton>
                  <ElevareButton onClick={() => setStatus("start", a.id)}>
                    Start
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => setStatus("pause", a.id)}>
                    Pause
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => setStatus("end", a.id)}>
                    End
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => removeAuction(a.id)}>
                    Delete
                  </ElevareButton>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {selectedAuction && form ? (
        <div className="panel">
          <h2 className="font-heading text-xl">Edit Auction</h2>
          <AuctionForm form={form} onChange={setForm} onSubmit={saveEdit} submitLabel="Save changes" />
          <div className="actions">
            <ElevareButton
              type="button"
              variant="secondary"
              onClick={() => {
                setSelectedId(null);
                setForm(null);
              }}
            >
              Cancel
            </ElevareButton>
          </div>
        </div>
      ) : null}
    </section>
  );
}
