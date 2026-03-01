import { Link, useNavigate } from "react-router-dom";
import { useEffect, useMemo, useRef, useState } from "react";
import { api } from "../../api/client";
import ImageUploader from "../../components/admin/ImageUploader";
import ElevareButton from "../../components/ui/ElevareButton";
import { defaultAuctionFormState, fromAuction, toPayload, type AuctionFormState } from "../../components/admin/AuctionForm";

const stepLabels = ["Basic Info", "Schedule", "Financial Defaults", "Bidding Rules", "Review & Publish"] as const;

export default function CreateAuction() {
  const navigate = useNavigate();
  const [auctionId, setAuctionId] = useState("");
  const [form, setForm] = useState<AuctionFormState>(defaultAuctionFormState());
  const [blocked, setBlocked] = useState(false);
  const [activeStep, setActiveStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const lastSignatureRef = useRef("");
  const readyRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([api.getTenantSettings(), api.getInstitutionOnboarding()])
      .then(async ([settings, onboarding]) => {
        if (cancelled) return;
        const defaultCurrency = settings.default_currency || "USD";
        setBlocked(onboarding.status !== "completed");
        if (onboarding.status !== "completed") {
          setForm((prev) => ({ ...prev, currency: defaultCurrency }));
          return;
        }

        try {
          const draft = await api.getAuctionWizardDraft();
          if (cancelled) return;
          setAuctionId(draft.id);
          setForm(fromAuction(draft));
          setActiveStep(Math.min(Math.max(draft.wizard_step || 1, 1), 5));
        } catch {
          if (cancelled) return;
          const draft = await api.createAuctionWizardDraft();
          if (cancelled) return;
          setAuctionId(draft.id);
          setForm(fromAuction(draft));
          setActiveStep(1);
        }
        readyRef.current = true;
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to initialize auction wizard");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const signature = useMemo(() => JSON.stringify({ step: activeStep, payload: toPayload(form) }), [activeStep, form]);

  useEffect(() => {
    if (!readyRef.current || !auctionId || blocked || loading || activeStep > 4) return;
    if (signature === lastSignatureRef.current) return;
    const handle = window.setTimeout(() => {
      void persistStep(activeStep, true);
    }, 700);
    return () => window.clearTimeout(handle);
  }, [activeStep, auctionId, blocked, loading, signature]);

  const persistStep = async (step: number, silent = false) => {
    if (!auctionId || blocked || step > 4) return true;
    setSaving(true);
    if (!silent) {
      setError(null);
      setMessage(null);
    }
    try {
      const updated = await api.saveAuctionWizard(auctionId, { ...toPayload(form), wizard_step: step });
      setForm(fromAuction(updated));
      setActiveStep(Math.min(Math.max(updated.wizard_step || step, 1), 5));
      lastSignatureRef.current = JSON.stringify({ step, payload: toPayload(fromAuction(updated)) });
      if (!silent) setMessage(`Step ${step} saved.`);
      return true;
    } catch (err) {
      if (!silent) setError(err instanceof Error ? err.message : "Failed to save step");
      return false;
    } finally {
      setSaving(false);
    }
  };

  const publish = async () => {
    if (!auctionId) return;
    setPublishing(true);
    setError(null);
    setMessage(null);
    try {
      await persistStep(4, true);
      await api.publishAuctionWizard(auctionId);
      setMessage("Auction published.");
      navigate("/admin/auctions");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to publish auction");
    } finally {
      setPublishing(false);
    }
  };

  if (loading) {
    return (
      <section className="page-shell">
        <div className="panel">
          <p>Loading auction wizard...</p>
        </div>
      </section>
    );
  }

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Auction Creation Wizard</h1>
          <p className="section-copy">Build the auction step-by-step, auto-save progress, then publish once the configuration is complete.</p>
        </div>
      </header>

      {blocked ? (
        <div className="panel">
          <p className="mb-3">Institution onboarding is not complete yet. Auction creation stays locked until activation.</p>
          <Link className="btn-brand" to="/admin/onboarding">Open Onboarding Wizard</Link>
        </div>
      ) : (
        <div className="panel">
          <div className="mb-6 flex flex-wrap gap-2">
            {stepLabels.map((label, index) => {
              const step = index + 1;
              const active = activeStep === step;
              const completed = activeStep > step;
              return (
                <button
                  key={label}
                  type="button"
                  className={active ? "btn-brand" : completed ? "btn-secondary" : "btn-ghost"}
                  onClick={() => setActiveStep(step)}
                >
                  {step}. {label}
                </button>
              );
            })}
          </div>

          <div className="mb-6 rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm text-neutral-600">
            Draft step: <strong>{activeStep}</strong>/5
            {saving ? " | Auto-saving..." : message ? ` | ${message}` : ""}
          </div>

          {activeStep === 1 ? (
            <div className="form-grid">
              <label>
                Title
                <input value={form.title} onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))} />
              </label>
              <label>
                Description
                <textarea rows={4} value={form.description} onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))} />
              </label>
              <ImageUploader
                label="Auction Images"
                images={form.images}
                altText={form.title || "Auction image"}
                onChange={(images) => setForm((prev) => ({ ...prev, images }))}
                onUploadOne={api.uploadAuctionImage}
              />
            </div>
          ) : null}

          {activeStep === 2 ? (
            <div className="form-grid">
              <label>
                Start Time
                <input type="datetime-local" value={form.start_time} onChange={(e) => setForm((prev) => ({ ...prev, start_time: e.target.value }))} />
              </label>
              <label>
                End Time
                <input type="datetime-local" value={form.end_time} onChange={(e) => setForm((prev) => ({ ...prev, end_time: e.target.value }))} />
              </label>
              <label>
                <span>Allow Manual Start</span>
                <input type="checkbox" checked={form.allow_manual_start} onChange={(e) => setForm((prev) => ({ ...prev, allow_manual_start: e.target.checked }))} />
              </label>
              <label>
                <span>Allow Manual Pause</span>
                <input type="checkbox" checked={form.allow_manual_pause} onChange={(e) => setForm((prev) => ({ ...prev, allow_manual_pause: e.target.checked }))} />
              </label>
              <label>
                <span>Allow Manual End</span>
                <input type="checkbox" checked={form.allow_manual_end} onChange={(e) => setForm((prev) => ({ ...prev, allow_manual_end: e.target.checked }))} />
              </label>
            </div>
          ) : null}

          {activeStep === 3 ? (
            <div className="form-grid">
              <label>
                Currency
                <select value={form.currency} onChange={(e) => setForm((prev) => ({ ...prev, currency: e.target.value }))}>
                  <option value="USD">USD</option>
                  <option value="GBP">GBP</option>
                  <option value="EUR">EUR</option>
                  <option value="AUD">AUD</option>
                  <option value="CAD">CAD</option>
                  <option value="CHF">CHF</option>
                  <option value="JPY">JPY</option>
                </select>
              </label>
              <p className="text-xs text-neutral-500">Financial defaults lock once the auction is published, and currency cannot change after the first bid.</p>
            </div>
          ) : null}

          {activeStep === 4 ? (
            <div className="form-grid">
              <label>
                Bid Visibility
                <select value={form.bid_visibility} onChange={(e) => setForm((prev) => ({ ...prev, bid_visibility: e.target.value as AuctionFormState["bid_visibility"] }))}>
                  <option value="PUBLIC">PUBLIC</option>
                  <option value="ANONYMOUS">ANONYMOUS</option>
                  <option value="SEALED">SEALED</option>
                </select>
              </label>
              <label>
                Auction Mode
                <select value={form.auction_mode} onChange={(e) => setForm((prev) => ({ ...prev, auction_mode: e.target.value as AuctionFormState["auction_mode"] }))}>
                  <option value="TIMED">TIMED</option>
                  <option value="SEQUENTIAL">SEQUENTIAL</option>
                </select>
              </label>
              <label>
                <span>Require Bidder Approval</span>
                <input
                  type="checkbox"
                  checked={form.require_bidder_approval}
                  onChange={(e) => setForm((prev) => ({ ...prev, require_bidder_approval: e.target.checked }))}
                />
              </label>
              <label>
                <span>Auto Extend</span>
                <input type="checkbox" checked={form.auto_extend_enabled} onChange={(e) => setForm((prev) => ({ ...prev, auto_extend_enabled: e.target.checked }))} />
              </label>
              <label>
                Sniping Window (sec)
                <input
                  type="number"
                  min={0}
                  value={form.sniping_window_seconds}
                  onChange={(e) => setForm((prev) => ({ ...prev, sniping_window_seconds: Number(e.target.value || 0) }))}
                />
              </label>
              <label>
                Sniping Extension (sec)
                <input
                  type="number"
                  min={0}
                  value={form.sniping_extension_seconds}
                  onChange={(e) => setForm((prev) => ({ ...prev, sniping_extension_seconds: Number(e.target.value || 0) }))}
                />
              </label>
              <label>
                <span>Auto Progress (Sequential)</span>
                <input
                  type="checkbox"
                  checked={form.auto_progress_enabled}
                  disabled={form.auction_mode !== "SEQUENTIAL"}
                  onChange={(e) => setForm((prev) => ({ ...prev, auto_progress_enabled: e.target.checked }))}
                />
              </label>
              <label>
                Lot Duration Seconds
                <input
                  type="number"
                  min={0}
                  disabled={form.auction_mode !== "SEQUENTIAL"}
                  value={form.lot_duration_seconds}
                  onChange={(e) => setForm((prev) => ({ ...prev, lot_duration_seconds: Number(e.target.value || 0) }))}
                />
              </label>
            </div>
          ) : null}

          {activeStep === 5 ? (
            <div className="grid gap-4 md:grid-cols-2">
              <div className="rounded-2xl border border-neutral-200 p-4">
                <h2 className="mb-3 text-lg font-semibold">Review</h2>
                <p><strong>Title:</strong> {form.title || "Not set"}</p>
                <p><strong>Currency:</strong> {form.currency}</p>
                <p><strong>Start:</strong> {form.start_time || "Not set"}</p>
                <p><strong>End:</strong> {form.end_time || "Not set"}</p>
                <p><strong>Mode:</strong> {form.auction_mode}</p>
                <p><strong>Visibility:</strong> {form.bid_visibility}</p>
              </div>
              <div className="rounded-2xl border border-neutral-200 p-4">
                <h2 className="mb-3 text-lg font-semibold">Publish Gate</h2>
                <p className="mb-4 text-sm text-neutral-600">Publishing is blocked until the draft has at least one lot.</p>
                <ElevareButton type="button" onClick={publish} disabled={publishing}>
                  {publishing ? "Publishing..." : "Publish Auction"}
                </ElevareButton>
              </div>
            </div>
          ) : null}

          <div className="mt-6 flex flex-wrap gap-3">
            <ElevareButton type="button" variant="secondary" disabled={activeStep === 1} onClick={() => setActiveStep((step) => Math.max(1, step - 1))}>
              Back
            </ElevareButton>
            {activeStep < 5 ? (
              <ElevareButton
                type="button"
                onClick={async () => {
                  const ok = await persistStep(Math.min(activeStep, 4));
                  if (ok) setActiveStep((step) => Math.min(5, step + 1));
                }}
              >
                Next
              </ElevareButton>
            ) : null}
          </div>
        </div>
      )}

      {error ? <p className="error">{error}</p> : null}
    </section>
  );
}
