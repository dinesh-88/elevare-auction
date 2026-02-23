import { useState } from "react";
import { api, type Auction } from "../../api/client";
import ImageUploader from "./ImageUploader";

export type AuctionFormState = {
  currency: string;
  title: string;
  description: string;
  images: string[];
  start_time: string;
  end_time: string;
  allow_manual_start: boolean;
  allow_manual_pause: boolean;
  allow_manual_end: boolean;
  auto_extend_enabled: boolean;
  auction_mode: "TIMED" | "SEQUENTIAL";
  auto_progress_enabled: boolean;
  lot_duration_seconds: number;
  sniping_window_seconds: number;
  sniping_extension_seconds: number;
  bid_visibility: "PUBLIC" | "ANONYMOUS" | "SEALED";
  require_bidder_approval: boolean;
};

export const now = new Date();

export const defaultAuctionFormState = (): AuctionFormState => ({
  currency: "USD",
  title: "",
  description: "",
  images: [],
  start_time: new Date(now.getTime() + 10 * 60 * 1000).toISOString().slice(0, 16),
  end_time: new Date(now.getTime() + 70 * 60 * 1000).toISOString().slice(0, 16),
  auto_extend_enabled: true,
  sniping_window_seconds: 60,
  sniping_extension_seconds: 120,
  allow_manual_start: true,
  allow_manual_pause: true,
  allow_manual_end: true,
  bid_visibility: "PUBLIC",
  require_bidder_approval: false,
  auction_mode: "TIMED",
  auto_progress_enabled: true,
  lot_duration_seconds: 120,
});

export const toInputDateTime = (iso: string) => {
  const d = new Date(iso);
  const offset = d.getTimezoneOffset() * 60000;
  return new Date(d.getTime() - offset).toISOString().slice(0, 16);
};

export const fromAuction = (a: Auction): AuctionFormState => ({
  currency: a.currency,
  title: a.title,
  description: a.description,
  images:
    a.metadata?.images && a.metadata.images.length > 0
      ? a.metadata.images
      : a.metadata?.image
        ? [a.metadata.image]
        : [],
  start_time: toInputDateTime(a.start_time),
  end_time: toInputDateTime(a.end_time),
  allow_manual_start: a.allow_manual_start,
  allow_manual_pause: a.allow_manual_pause,
  allow_manual_end: a.allow_manual_end,
  auto_extend_enabled: a.auto_extend_enabled,
  auction_mode: a.auction_mode,
  auto_progress_enabled: a.auto_progress_enabled,
  lot_duration_seconds: a.lot_duration_seconds,
  sniping_window_seconds: a.sniping_window_seconds,
  sniping_extension_seconds: a.sniping_extension_seconds,
  bid_visibility: a.bid_visibility,
  require_bidder_approval: Boolean(a.require_bidder_approval),
});

export const toPayload = (form: AuctionFormState) => ({
  currency: form.currency,
  title: form.title,
  description: form.description,
  metadata: {
    image: form.images[0] ?? "",
    images: form.images,
  },
  bid_visibility: form.bid_visibility,
  require_bidder_approval: form.require_bidder_approval,
  auction_mode: form.auction_mode,
  allow_manual_start: form.allow_manual_start,
  allow_manual_pause: form.allow_manual_pause,
  allow_manual_end: form.allow_manual_end,
  auto_extend_enabled: form.auto_extend_enabled,
  auto_progress_enabled: form.auction_mode === "SEQUENTIAL" ? form.auto_progress_enabled : false,
  lot_duration_seconds: form.auction_mode === "SEQUENTIAL" ? form.lot_duration_seconds : 0,
  sniping_window_seconds: form.sniping_window_seconds,
  sniping_extension_seconds: form.sniping_extension_seconds,
  start_time: new Date(form.start_time).toISOString(),
  end_time: new Date(form.end_time).toISOString(),
});

export default function AuctionForm({
  form,
  onChange,
  onSubmit,
  submitLabel,
  currencyDisabled = false,
  currencyHelpText,
}: {
  form: AuctionFormState;
  onChange: (next: AuctionFormState) => void;
  onSubmit: (e: React.FormEvent) => void;
  submitLabel: string;
  currencyDisabled?: boolean;
  currencyHelpText?: string;
}) {
  const [uploadingHint, setUploadingHint] = useState<string | null>(null);

  return (
    <form className="form-grid" onSubmit={onSubmit}>
      <label>Title</label>
      <input value={form.title} onChange={(e) => onChange({ ...form, title: e.target.value })} required />

      <label>Description</label>
      <textarea value={form.description} rows={4} onChange={(e) => onChange({ ...form, description: e.target.value })} />

      <label>Currency</label>
      <select value={form.currency} disabled={currencyDisabled} onChange={(e) => onChange({ ...form, currency: e.target.value })}>
        <option value="USD">USD</option>
        <option value="GBP">GBP</option>
        <option value="EUR">EUR</option>
        <option value="AUD">AUD</option>
        <option value="CAD">CAD</option>
        <option value="CHF">CHF</option>
        <option value="JPY">JPY</option>
      </select>
      {currencyHelpText ? <p className="text-xs text-neutral-500">{currencyHelpText}</p> : null}

      <ImageUploader
        label="Auction Images"
        images={form.images}
        altText={form.title || "Auction image"}
        onChange={(images) => onChange({ ...form, images })}
        onUploadOne={async (file) => {
          setUploadingHint("Uploading...");
          try {
            return await api.uploadAuctionImage(file);
          } finally {
            setUploadingHint(null);
          }
        }}
      />
      {uploadingHint ? <p className="text-sm text-neutral-500">{uploadingHint}</p> : null}

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

      <label>Auction Mode</label>
      <select
        value={form.auction_mode}
        onChange={(e) => onChange({ ...form, auction_mode: e.target.value as AuctionFormState["auction_mode"] })}
      >
        <option value="TIMED">TIMED</option>
        <option value="SEQUENTIAL">SEQUENTIAL</option>
      </select>

      <label>Auto Progress (Sequential)</label>
      <input
        type="checkbox"
        checked={form.auto_progress_enabled}
        disabled={form.auction_mode !== "SEQUENTIAL"}
        onChange={(e) => onChange({ ...form, auto_progress_enabled: e.target.checked })}
      />

      <label>Lot Duration Seconds (Sequential)</label>
      <input
        type="number"
        min={0}
        value={form.lot_duration_seconds}
        disabled={form.auction_mode !== "SEQUENTIAL"}
        onChange={(e) => onChange({ ...form, lot_duration_seconds: Number(e.target.value) })}
      />

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
