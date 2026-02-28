import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Link, useNavigate } from "react-router-dom";
import { api, type InstitutionOnboardingData, type InstitutionOnboardingState } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

const stepLabels = [
  "Identity",
  "Branding",
  "Currency & VAT",
  "Fee Defaults",
  "Legal Terms",
  "Email Setup",
  "Review & Activate",
] as const;

const emptyData: InstitutionOnboardingData = {
  institution_name: "",
  country: "",
  logo_url: "",
  primary_color: "",
  secondary_color: "",
  favicon_url: "",
  font_family: "",
  default_currency: "USD",
  vat_rate_percent: 0,
  buyers_premium_percent: 0,
  legal_terms_markdown: "",
  support_email: "",
};

function stepSlice(step: number, data: InstitutionOnboardingData) {
  switch (step) {
    case 1:
      return { institution_name: data.institution_name, country: data.country };
    case 2:
      return {
        logo_url: data.logo_url,
        primary_color: data.primary_color,
        secondary_color: data.secondary_color,
        favicon_url: data.favicon_url,
        font_family: data.font_family,
      };
    case 3:
      return { default_currency: data.default_currency, vat_rate_percent: data.vat_rate_percent };
    case 4:
      return { buyers_premium_percent: data.buyers_premium_percent };
    case 5:
      return { legal_terms_markdown: data.legal_terms_markdown };
    case 6:
      return { support_email: data.support_email };
    default:
      return data;
  }
}

function nextIncompleteStep(state: InstitutionOnboardingState | null) {
  if (!state) return 1;
  if (state.status === "completed") return 7;
  return Math.min(Math.max((state.step || 0) + 1, 1), 7);
}

export default function Onboarding() {
  const navigate = useNavigate();
  const searchParams = useSearchParams();
  const [state, setState] = useState<InstitutionOnboardingState | null>(null);
  const [form, setForm] = useState<InstitutionOnboardingData>(emptyData);
  const [activeStep, setActiveStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activating, setActivating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const loadedRef = useRef(false);
  const lastSignatureRef = useRef("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    api
      .getInstitutionOnboarding()
      .then((nextState) => {
        if (cancelled) return;
        setState(nextState);
        setForm(nextState.data);
        const requestedStep = Number(searchParams.get("step") ?? "");
        if (requestedStep >= 1 && requestedStep <= 7) {
          setActiveStep(requestedStep);
        } else {
          setActiveStep(nextIncompleteStep(nextState));
        }
        lastSignatureRef.current = "";
        loadedRef.current = true;
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Failed to load onboarding");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [searchParams]);

  const currentSignature = useMemo(() => JSON.stringify({ step: activeStep, data: stepSlice(activeStep, form) }), [activeStep, form]);

  useEffect(() => {
    if (!loadedRef.current || loading || activeStep > 6) return;
    if (currentSignature === lastSignatureRef.current) return;
    const handle = window.setTimeout(() => {
      void persistStep(activeStep, true);
    }, 700);
    return () => window.clearTimeout(handle);
  }, [activeStep, currentSignature, loading]);

  const persistStep = async (step: number, silent = false) => {
    if (step > 6) return true;
    setSaving(true);
    if (!silent) {
      setError(null);
      setMessage(null);
    }
    try {
      const nextState = await api.saveInstitutionOnboarding({ step, data: form });
      setState(nextState);
      setForm(nextState.data);
      lastSignatureRef.current = JSON.stringify({ step, data: stepSlice(step, nextState.data) });
      if (!silent) setMessage(`Step ${step} saved.`);
      return true;
    } catch (err) {
      if (!silent) setError(err instanceof Error ? err.message : "Failed to save onboarding");
      return false;
    } finally {
      setSaving(false);
    }
  };

  const complete = async () => {
    setActivating(true);
    setError(null);
    setMessage(null);
    try {
      const nextState = await api.completeInstitutionOnboarding();
      setState(nextState);
      setForm(nextState.data);
      setActiveStep(7);
      setMessage("Onboarding completed. Auction creation is now enabled.");
      navigate("/admin/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to complete onboarding");
    } finally {
      setActivating(false);
    }
  };

  const currentStepDone = state?.status === "completed" || (state?.step ?? 0) >= activeStep;

  if (loading) {
    return (
      <section className="page-shell">
        <div className="panel">
          <p>Loading onboarding…</p>
        </div>
      </section>
    );
  }

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Institution Onboarding</h1>
          <p className="section-copy">Configure the institution once, then unlock auction creation and publishing.</p>
        </div>
        {state?.status === "completed" ? <Link className="btn-brand" to="/admin/create-auction">Create Auction</Link> : null}
      </header>

      <div className="panel">
        <div className="mb-6 flex flex-wrap gap-2">
          {stepLabels.map((label, index) => {
            const step = index + 1;
            const done = state?.status === "completed" || (state?.step ?? 0) >= step;
            const active = activeStep === step;
            return (
              <button
                key={label}
                type="button"
                className={active ? "btn-brand" : done ? "btn-secondary" : "btn-ghost"}
                onClick={() => setActiveStep(step)}
              >
                {step}. {label}
              </button>
            );
          })}
        </div>

        <div className="mb-6 rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm text-neutral-600">
          Status: <strong>{state?.status ?? "not_started"}</strong> | Progress: <strong>{Math.max(state?.step ?? 0, state?.status === "completed" ? 7 : 0)}/7</strong>
          {saving ? " | Auto-saving…" : currentStepDone ? " | Saved" : ""}
        </div>

        {activeStep === 1 ? (
          <div className="form-grid">
            <label>
              Institution Name
              <input value={form.institution_name} onChange={(e) => setForm((prev) => ({ ...prev, institution_name: e.target.value }))} />
            </label>
            <label>
              Country
              <input value={form.country} onChange={(e) => setForm((prev) => ({ ...prev, country: e.target.value }))} />
            </label>
          </div>
        ) : null}

        {activeStep === 2 ? (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(280px,1fr)]">
            <div className="form-grid">
              <label>
                Logo URL
                <input value={form.logo_url} onChange={(e) => setForm((prev) => ({ ...prev, logo_url: e.target.value }))} />
              </label>
              <label>
                Primary Color
                <input value={form.primary_color} onChange={(e) => setForm((prev) => ({ ...prev, primary_color: e.target.value }))} />
              </label>
              <label>
                Secondary Color
                <input value={form.secondary_color} onChange={(e) => setForm((prev) => ({ ...prev, secondary_color: e.target.value }))} />
              </label>
              <label>
                Favicon URL
                <input value={form.favicon_url} onChange={(e) => setForm((prev) => ({ ...prev, favicon_url: e.target.value }))} />
              </label>
              <label>
                Font Family
                <input value={form.font_family} onChange={(e) => setForm((prev) => ({ ...prev, font_family: e.target.value }))} />
              </label>
            </div>
            <div
              className="rounded-3xl border p-5"
              style={{
                borderColor: form.secondary_color || "#d4d4d4",
                background: form.primary_color || "#fafafa",
                color: form.secondary_color || "#111827",
                fontFamily: form.font_family || "Georgia, serif",
              }}
            >
              <p className="mb-3 text-xs uppercase tracking-[0.2em] opacity-70">Brand Preview</p>
              {form.logo_url ? <img alt="Logo preview" className="mb-4 max-h-12" src={form.logo_url} /> : null}
              <h2 className="mb-2 text-2xl">{form.institution_name || "Your Institution"}</h2>
              <p className="text-sm opacity-80">This preview uses the selected brand tokens and updates as you type.</p>
            </div>
          </div>
        ) : null}

        {activeStep === 3 ? (
          <div className="form-grid">
            <label>
              Default Currency
              <input
                maxLength={3}
                value={form.default_currency}
                onChange={(e) => setForm((prev) => ({ ...prev, default_currency: e.target.value.toUpperCase() }))}
              />
            </label>
            <label>
              VAT Rate (%)
              <input
                type="number"
                min="0"
                step="0.01"
                value={form.vat_rate_percent}
                onChange={(e) => setForm((prev) => ({ ...prev, vat_rate_percent: Number(e.target.value || 0) }))}
              />
            </label>
          </div>
        ) : null}

        {activeStep === 4 ? (
          <div className="form-grid">
            <label>
              Buyer's Premium (%)
              <input
                type="number"
                min="0"
                step="0.01"
                value={form.buyers_premium_percent}
                onChange={(e) => setForm((prev) => ({ ...prev, buyers_premium_percent: Number(e.target.value || 0) }))}
              />
            </label>
          </div>
        ) : null}

        {activeStep === 5 ? (
          <label className="block">
            Legal Terms
            <textarea
              rows={10}
              value={form.legal_terms_markdown}
              onChange={(e) => setForm((prev) => ({ ...prev, legal_terms_markdown: e.target.value }))}
            />
          </label>
        ) : null}

        {activeStep === 6 ? (
          <div className="form-grid">
            <label>
              Support Email
              <input value={form.support_email} onChange={(e) => setForm((prev) => ({ ...prev, support_email: e.target.value }))} />
            </label>
          </div>
        ) : null}

        {activeStep === 7 ? (
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-2xl border border-neutral-200 p-4">
              <h2 className="mb-3 text-lg font-semibold">Review</h2>
              <p><strong>Institution:</strong> {form.institution_name || "Not set"}</p>
              <p><strong>Country:</strong> {form.country || "Not set"}</p>
              <p><strong>Currency:</strong> {form.default_currency || "Not set"}</p>
              <p><strong>VAT:</strong> {form.vat_rate_percent}%</p>
              <p><strong>Buyer's Premium:</strong> {form.buyers_premium_percent}%</p>
              <p><strong>Support Email:</strong> {form.support_email || "Not set"}</p>
            </div>
            <div className="rounded-2xl border border-neutral-200 p-4">
              <h2 className="mb-3 text-lg font-semibold">Activation</h2>
              <p className="mb-4 text-sm text-neutral-600">
                Activation marks onboarding complete and unlocks auction creation and manual auction start.
              </p>
              {state?.status === "completed" ? (
                <Link className="btn-brand" to="/admin/create-auction">Proceed to Create Auction</Link>
              ) : (
                <ElevareButton type="button" onClick={complete} disabled={activating}>
                  {activating ? "Activating…" : "Activate Institution"}
                </ElevareButton>
              )}
            </div>
          </div>
        ) : null}

        <div className="mt-6 flex flex-wrap gap-3">
          <ElevareButton type="button" variant="secondary" disabled={activeStep === 1} onClick={() => setActiveStep((step) => Math.max(1, step - 1))}>
            Back
          </ElevareButton>
          {activeStep < 7 ? (
            <ElevareButton
              type="button"
              onClick={async () => {
                const ok = await persistStep(activeStep);
                if (ok) setActiveStep((step) => Math.min(7, step + 1));
              }}
            >
              Next
            </ElevareButton>
          ) : null}
        </div>
      </div>

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
