import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type Auction } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

type FormState = {
  auto_extend_enabled: boolean;
  sniping_window_seconds: number;
  sniping_extension_seconds: number;
};

export default function AuctionSettings() {
  const params = useParams();
  const auctionId = params.auctionId;
  const [auction, setAuction] = useState<Auction | null>(null);
  const [form, setForm] = useState<FormState | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!auctionId) return;
    setLoading(true);
    setError(null);
    api
      .getAuction(auctionId)
      .then((a) => {
        setAuction(a);
        setForm({
          auto_extend_enabled: a.auto_extend_enabled,
          sniping_window_seconds: a.sniping_window_seconds,
          sniping_extension_seconds: a.sniping_extension_seconds,
        });
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auction settings"))
      .finally(() => setLoading(false));
  }, [auctionId]);

  const validationError = useMemo(() => {
    if (!form) return null;
    if (!Number.isInteger(form.sniping_window_seconds) || !Number.isInteger(form.sniping_extension_seconds)) {
      return "Window and extension must be integers.";
    }
    if (form.sniping_window_seconds < 5) {
      return "Window must be at least 5 seconds.";
    }
    if (form.sniping_extension_seconds < 30) {
      return "Extension must be at least 30 seconds.";
    }
    if (form.sniping_extension_seconds <= form.sniping_window_seconds) {
      return "Extension must be greater than window.";
    }
    return null;
  }, [form]);

  const onSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!auctionId || !auction || !form) return;
    if (validationError) {
      setError(validationError);
      return;
    }
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const updated = await api.updateAuction(auctionId, {
        title: auction.title,
        description: auction.description,
        currency: auction.currency,
        metadata: auction.metadata ?? {},
        start_time: auction.start_time,
        end_time: auction.end_time,
        allow_manual_start: auction.allow_manual_start,
        allow_manual_pause: auction.allow_manual_pause,
        allow_manual_end: auction.allow_manual_end,
        auction_mode: auction.auction_mode,
        auto_progress_enabled: auction.auto_progress_enabled,
        lot_duration_seconds: auction.lot_duration_seconds,
        bid_visibility: auction.bid_visibility,
        auto_extend_enabled: form.auto_extend_enabled,
        sniping_window_seconds: form.sniping_window_seconds,
        sniping_extension_seconds: form.sniping_extension_seconds,
      });
      setAuction(updated);
      setMessage("Soft-close settings saved.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <p className="text-sm text-neutral-500">Loading settings...</p>;
  if (!auction || !form) return <p className="error">Auction not found.</p>;

  return (
    <section className="page-shell max-w-3xl">
      <header className="section-head">
        <div>
          <h1 className="section-title">Auction Settings</h1>
          <p className="section-copy">Configure soft close behaviour for this auction.</p>
        </div>
        <Link className="btn-secondary" to={`/admin/auctions/${auction.id}/lots`}>
          Back to Lots
        </Link>
      </header>

      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <form className="panel space-y-5" onSubmit={onSave}>
        <div className="flex items-center justify-between rounded-lg border border-neutral-200 p-4">
          <div>
            <h2 className="font-medium text-neutral-900">Enable Soft Close</h2>
            <p className="text-sm text-neutral-500">Extend lot end time when bids arrive near closing.</p>
          </div>
          <input
            type="checkbox"
            checked={form.auto_extend_enabled}
            onChange={(e) => setForm((prev) => (prev ? { ...prev, auto_extend_enabled: e.target.checked } : prev))}
          />
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1 text-sm">
            <span className="text-neutral-700">Window (seconds)</span>
            <input
              type="number"
              min={5}
              value={form.sniping_window_seconds}
              onChange={(e) => setForm((prev) => (prev ? { ...prev, sniping_window_seconds: Number(e.target.value) } : prev))}
              className="w-full rounded-lg border border-neutral-200 px-3 py-2"
            />
          </label>
          <label className="space-y-1 text-sm">
            <span className="text-neutral-700">Extension Duration (seconds)</span>
            <input
              type="number"
              min={30}
              value={form.sniping_extension_seconds}
              onChange={(e) => setForm((prev) => (prev ? { ...prev, sniping_extension_seconds: Number(e.target.value) } : prev))}
              className="w-full rounded-lg border border-neutral-200 px-3 py-2"
            />
          </label>
        </div>

        {validationError ? <p className="text-sm text-red-600">{validationError}</p> : null}

        <div className="actions">
          <ElevareButton type="submit" disabled={saving || Boolean(validationError)}>
            {saving ? "Saving..." : "Save"}
          </ElevareButton>
        </div>
      </form>
    </section>
  );
}
