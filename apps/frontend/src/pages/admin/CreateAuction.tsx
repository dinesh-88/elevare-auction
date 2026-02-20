import { useState } from "react";
import { api } from "../../api/client";

const now = new Date();
const defaultStart = new Date(now.getTime() + 10 * 60 * 1000).toISOString().slice(0, 16);
const defaultEnd = new Date(now.getTime() + 70 * 60 * 1000).toISOString().slice(0, 16);

export default function CreateAuction() {
  const [form, setForm] = useState({
    title: "",
    description: "",
    starting_price: 100,
    reserve_price: "",
    min_increment: 10,
    start_time: defaultStart,
    end_time: defaultEnd,
    auto_extend_enabled: true,
    sniping_window_seconds: 60,
    sniping_extension_seconds: 120,
    allow_manual_start: true,
    allow_manual_pause: true,
    allow_manual_end: true,
    bid_visibility: "PUBLIC",
  });
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setMessage(null);

    try {
      await api.createAuction({
        ...form,
        reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
        start_time: new Date(form.start_time).toISOString(),
        end_time: new Date(form.end_time).toISOString(),
      });
      setMessage("Auction created.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create auction");
    }
  };

  return (
    <section>
      <h1>Create Auction</h1>
      <form className="form-grid" onSubmit={submit}>
        <label>Title</label>
        <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />

        <label>Description</label>
        <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={4} />

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

        <button type="submit">Create</button>
      </form>

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
