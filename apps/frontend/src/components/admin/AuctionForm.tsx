import type { Auction } from "../../api/client";

export type AuctionFormState = {
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

export const now = new Date();

export const defaultAuctionFormState = (): AuctionFormState => ({
  title: "",
  description: "",
  starting_price: 100,
  reserve_price: "",
  min_increment: 10,
  start_time: new Date(now.getTime() + 10 * 60 * 1000).toISOString().slice(0, 16),
  end_time: new Date(now.getTime() + 70 * 60 * 1000).toISOString().slice(0, 16),
  auto_extend_enabled: true,
  sniping_window_seconds: 60,
  sniping_extension_seconds: 120,
  allow_manual_start: true,
  allow_manual_pause: true,
  allow_manual_end: true,
  bid_visibility: "PUBLIC",
});

export const toInputDateTime = (iso: string) => {
  const d = new Date(iso);
  const offset = d.getTimezoneOffset() * 60000;
  return new Date(d.getTime() - offset).toISOString().slice(0, 16);
};

export const fromAuction = (a: Auction): AuctionFormState => ({
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

export const toPayload = (form: AuctionFormState) => ({
  ...form,
  reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
  start_time: new Date(form.start_time).toISOString(),
  end_time: new Date(form.end_time).toISOString(),
});

export default function AuctionForm({
  form,
  onChange,
  onSubmit,
  submitLabel,
}: {
  form: AuctionFormState;
  onChange: (next: AuctionFormState) => void;
  onSubmit: (e: React.FormEvent) => void;
  submitLabel: string;
}) {
  return (
    <form className="form-grid" onSubmit={onSubmit}>
      <label>Title</label>
      <input value={form.title} onChange={(e) => onChange({ ...form, title: e.target.value })} required />

      <label>Description</label>
      <textarea value={form.description} rows={4} onChange={(e) => onChange({ ...form, description: e.target.value })} />

      <label>Starting Price</label>
      <input
        type="number"
        value={form.starting_price}
        onChange={(e) => onChange({ ...form, starting_price: Number(e.target.value) })}
        required
      />

      <label>Reserve Price</label>
      <input value={form.reserve_price} onChange={(e) => onChange({ ...form, reserve_price: e.target.value })} />

      <label>Min Increment</label>
      <input
        type="number"
        value={form.min_increment}
        onChange={(e) => onChange({ ...form, min_increment: Number(e.target.value) })}
        required
      />

      <label>Start Time</label>
      <input
        type="datetime-local"
        value={form.start_time}
        onChange={(e) => onChange({ ...form, start_time: e.target.value })}
        required
      />

      <label>End Time</label>
      <input
        type="datetime-local"
        value={form.end_time}
        onChange={(e) => onChange({ ...form, end_time: e.target.value })}
        required
      />

      <label>Bid Visibility</label>
      <select
        value={form.bid_visibility}
        onChange={(e) => onChange({ ...form, bid_visibility: e.target.value as AuctionFormState["bid_visibility"] })}
      >
        <option value="PUBLIC">PUBLIC</option>
        <option value="ANONYMOUS">ANONYMOUS</option>
        <option value="SEALED">SEALED</option>
      </select>

      <label>Auto Extend</label>
      <input
        type="checkbox"
        checked={form.auto_extend_enabled}
        onChange={(e) => onChange({ ...form, auto_extend_enabled: e.target.checked })}
      />

      <label>Sniping Window (sec)</label>
      <input
        type="number"
        value={form.sniping_window_seconds}
        onChange={(e) => onChange({ ...form, sniping_window_seconds: Number(e.target.value) })}
      />

      <label>Sniping Extension (sec)</label>
      <input
        type="number"
        value={form.sniping_extension_seconds}
        onChange={(e) => onChange({ ...form, sniping_extension_seconds: Number(e.target.value) })}
      />

      <button type="submit" className="btn-brand">
        {submitLabel}
      </button>
    </form>
  );
}
