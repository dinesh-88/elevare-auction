import { useEffect, useState } from "react";
import { api } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

const CURRENCIES = ["USD", "GBP", "EUR", "AUD", "CAD", "CHF", "JPY"];

export default function Settings() {
  const [defaultCurrency, setDefaultCurrency] = useState("USD");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    api
      .getTenantSettings()
      .then((settings) => {
        if (cancelled) return;
        setDefaultCurrency(settings.default_currency || "USD");
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Failed to load tenant settings");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const save = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const updated = await api.updateTenantSettings({ default_currency: defaultCurrency });
      setDefaultCurrency(updated.default_currency);
      setMessage("Default currency updated.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update default currency");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Organization Settings</h1>
          <p className="section-copy">Set the default currency used for new auctions.</p>
        </div>
      </header>

      <div className="panel">
        {loading ? (
          <p className="text-sm text-neutral-500">Loading settings...</p>
        ) : (
          <div className="form-grid">
            <label>Default Currency</label>
            <select value={defaultCurrency} onChange={(e) => setDefaultCurrency(e.target.value)}>
              {CURRENCIES.map((currency) => (
                <option key={currency} value={currency}>
                  {currency}
                </option>
              ))}
            </select>
            <div className="actions">
              <ElevareButton type="button" onClick={save} disabled={saving}>
                {saving ? "Saving..." : "Save"}
              </ElevareButton>
            </div>
          </div>
        )}
      </div>

      {error ? <p className="error mt-3">{error}</p> : null}
      {message ? <p className="success mt-3">{message}</p> : null}
    </section>
  );
}
