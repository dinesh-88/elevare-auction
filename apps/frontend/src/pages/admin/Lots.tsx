import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type Lot } from "../../api/client";
import { useLotsSockets } from "../../hooks/useLotsSockets";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";

export default function Lots() {
  const { auctionId } = useParams();
  const [lots, setLots] = useState<Lot[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [form, setForm] = useState({
    lot_number: "",
    order_index: 1,
    title: "",
    description: "",
    starting_price: 100,
    reserve_price: "",
    min_increment: 10,
  });

  const load = () => {
    if (!auctionId) return;
    api
      .listLotsByAuction(auctionId)
      .then(setLots)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load lots"));
  };

  useEffect(() => {
    load();
  }, [auctionId]);

  useLotsSockets(
    lots.map((l) => l.id),
    (lotId, event) => {
      setLots((prev) =>
        prev.map((l) =>
          l.id === lotId
            ? {
                ...l,
                current_price: event.current_price,
                ended_at: event.ended_at ?? l.ended_at,
              }
            : l,
        ),
      );
    },
    (lotId, event) => {
      setLots((prev) => prev.map((l) => (l.id === lotId ? { ...l, status: event.status } : l)));
    },
    lots.length > 0,
  );

  const createLot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!auctionId) return;
    try {
      setError(null);
      await api.createLot(auctionId, {
        ...form,
        reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
      });
      setMessage("Lot created.");
      setForm({ ...form, lot_number: "", title: "", description: "" });
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create lot");
    }
  };

  const move = async (index: number, direction: -1 | 1) => {
    if (!auctionId) return;
    const next = [...lots];
    const target = index + direction;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    try {
      await api.reorderLots(
        auctionId,
        next.map((l) => l.id),
      );
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reorder lots");
    }
  };

  const removeLot = async (lotId: string) => {
    if (!window.confirm("Delete lot?")) return;
    try {
      await api.deleteLot(lotId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete lot");
    }
  };

  const startLot = async (lotId: string) => {
    try {
      await api.startLot(lotId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start lot");
    }
  };

  const endLot = async (lotId: string) => {
    try {
      await api.endLot(lotId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to end lot");
    }
  };

  return (
    <section className="page-shell">
      <div className="section-head">
        <div>
          <h1 className="section-title">Manage Lots</h1>
          <p className="section-copy">Create, reorder and control lot status.</p>
        </div>
        <Link className="btn-link-brand" to="/admin/auctions">← Back to Auctions</Link>
      </div>

      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <form className="form-grid mb-6" onSubmit={createLot}>
        <h2 className="font-heading text-xl">Create Lot</h2>
        <label>Lot Number</label>
        <input value={form.lot_number} onChange={(e) => setForm({ ...form, lot_number: e.target.value })} required />
        <label>Order Index</label>
        <input type="number" value={form.order_index} onChange={(e) => setForm({ ...form, order_index: Number(e.target.value) })} required />
        <label>Title</label>
        <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
        <label>Description</label>
        <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={3} />
        <label>Starting Price</label>
        <input type="number" value={form.starting_price} onChange={(e) => setForm({ ...form, starting_price: Number(e.target.value) })} required />
        <label>Reserve Price</label>
        <input value={form.reserve_price} onChange={(e) => setForm({ ...form, reserve_price: e.target.value })} />
        <label>Min Increment</label>
        <input type="number" value={form.min_increment} onChange={(e) => setForm({ ...form, min_increment: Number(e.target.value) })} required />
        <ElevareButton type="submit">Create Lot</ElevareButton>
      </form>

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>#</th>
              <th>Lot</th>
              <th>Status</th>
              <th>Current Price</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {lots.map((lot, idx) => (
              <tr key={lot.id}>
                <td>{lot.order_index}</td>
                <td>{lot.lot_number} - {lot.title}</td>
                <td><StatusBadge status={lot.status} /></td>
                <td>${lot.current_price.toLocaleString()}</td>
                <td className="actions">
                  <ElevareButton variant="secondary" onClick={() => move(idx, -1)}>↑</ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => move(idx, 1)}>↓</ElevareButton>
                  <ElevareButton onClick={() => startLot(lot.id)}>Start</ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => endLot(lot.id)}>End</ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => removeLot(lot.id)}>Delete</ElevareButton>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
