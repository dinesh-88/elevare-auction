import { useEffect, useMemo, useState } from "react";
import { api, type Auction } from "../../api/client";

type FormState = {
  title: string;
  description: string;
  starting_price: number;
  reserve_price: string;
  min_increment: number;
  start_time: string;
  end_time: string;
  allow_manual_start: boolean;
  allow_manual_pause: boolean;
  allow_manual_end: boolean;
  auto_extend_enabled: boolean;
  sniping_window_seconds: number;
  sniping_extension_seconds: number;
  bid_visibility: "PUBLIC" | "ANONYMOUS" | "SEALED";
};

const toInputDateTime = (iso: string) => {
  const d = new Date(iso);
  const offset = d.getTimezoneOffset() * 60000;
  return new Date(d.getTime() - offset).toISOString().slice(0, 16);
};

const fromAuction = (a: Auction): FormState => ({
  title: a.title,
  description: a.description,
  starting_price: a.starting_price,
  reserve_price: a.reserve_price == null ? "" : String(a.reserve_price),
  min_increment: a.min_increment,
  start_time: toInputDateTime(a.start_time),
  end_time: toInputDateTime(a.end_time),
  allow_manual_start: a.allow_manual_start,
  allow_manual_pause: a.allow_manual_pause,
  allow_manual_end: a.allow_manual_end,
  auto_extend_enabled: a.auto_extend_enabled,
  sniping_window_seconds: a.sniping_window_seconds,
  sniping_extension_seconds: a.sniping_extension_seconds,
  bid_visibility: a.bid_visibility,
});

export default function Auctions() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState | null>(null);
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
      await api.updateAuction(selectedAuction.id, {
        ...form,
        reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
        start_time: new Date(form.start_time).toISOString(),
        end_time: new Date(form.end_time).toISOString(),
      });
      setMessage("Auction updated.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction");
    }
  };

  return (
    <section>
      <h1 className="mb-4 font-heading text-2xl">Manage Auctions</h1>
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
                <td>{a.status}</td>
                <td>${a.current_price.toLocaleString()}</td>
                <td>{new Date(a.end_time).toLocaleString()}</td>
                <td className="actions">
                  <button className="btn-brand" onClick={() => startEdit(a)}>
                    Edit
                  </button>
                  <button className="btn-brand" onClick={() => setStatus("start", a.id)}>
                    Start
                  </button>
                  <button className="btn-brand" onClick={() => setStatus("pause", a.id)}>
                    Pause
                  </button>
                  <button className="btn-brand" onClick={() => setStatus("end", a.id)}>
                    End
                  </button>
                  <button className="btn-brand" onClick={() => removeAuction(a.id)}>
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {selectedAuction && form ? (
        <form className="form-grid" onSubmit={saveEdit}>
          <h2 className="font-heading text-xl">Edit Auction</h2>
          <label>Title</label>
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />

          <label>Description</label>
          <textarea value={form.description} rows={4} onChange={(e) => setForm({ ...form, description: e.target.value })} />

          <label>Starting Price</label>
          <input
            type="number"
            value={form.starting_price}
            onChange={(e) => setForm({ ...form, starting_price: Number(e.target.value) })}
            required
          />

          <label>Reserve Price</label>
          <input value={form.reserve_price} onChange={(e) => setForm({ ...form, reserve_price: e.target.value })} />

          <label>Min Increment</label>
          <input
            type="number"
            value={form.min_increment}
            onChange={(e) => setForm({ ...form, min_increment: Number(e.target.value) })}
            required
          />

          <label>Start Time</label>
          <input
            type="datetime-local"
            value={form.start_time}
            onChange={(e) => setForm({ ...form, start_time: e.target.value })}
            required
          />

          <label>End Time</label>
          <input
            type="datetime-local"
            value={form.end_time}
            onChange={(e) => setForm({ ...form, end_time: e.target.value })}
            required
          />

          <label>Bid Visibility</label>
          <select
            value={form.bid_visibility}
            onChange={(e) => setForm({ ...form, bid_visibility: e.target.value as FormState["bid_visibility"] })}
          >
            <option value="PUBLIC">PUBLIC</option>
            <option value="ANONYMOUS">ANONYMOUS</option>
            <option value="SEALED">SEALED</option>
          </select>

          <label>Auto Extend</label>
          <input
            type="checkbox"
            checked={form.auto_extend_enabled}
            onChange={(e) => setForm({ ...form, auto_extend_enabled: e.target.checked })}
          />

          <label>Sniping Window (sec)</label>
          <input
            type="number"
            value={form.sniping_window_seconds}
            onChange={(e) => setForm({ ...form, sniping_window_seconds: Number(e.target.value) })}
          />

          <label>Sniping Extension (sec)</label>
          <input
            type="number"
            value={form.sniping_extension_seconds}
            onChange={(e) => setForm({ ...form, sniping_extension_seconds: Number(e.target.value) })}
          />

          <div className="actions">
            <button type="submit" className="btn-brand">
              Save changes
            </button>
            <button
              type="button"
              className="btn-brand"
              onClick={() => {
                setSelectedId(null);
                setForm(null);
              }}
            >
              Cancel
            </button>
          </div>
        </form>
      ) : null}
    </section>
  );
}
