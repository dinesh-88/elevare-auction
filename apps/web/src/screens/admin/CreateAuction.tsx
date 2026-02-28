import { Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { api } from "../../api/client";
import AuctionForm, { defaultAuctionFormState, toPayload } from "../../components/admin/AuctionForm";

export default function CreateAuction() {
  const [form, setForm] = useState(defaultAuctionFormState());
  const [blocked, setBlocked] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    Promise.all([api.getTenantSettings(), api.getInstitutionOnboarding()])
      .then(([settings, onboarding]) => {
        if (cancelled) return;
        setForm((prev) => ({ ...prev, currency: settings.default_currency || "USD" }));
        setBlocked(onboarding.status !== "completed");
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (blocked) {
      setError("Complete institution onboarding before creating auctions.");
      return;
    }
    setError(null);
    setMessage(null);

    try {
      await api.createAuction(toPayload(form));
      setMessage("Auction created.");
      setForm(defaultAuctionFormState());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create auction");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Create Auction</h1>
          <p className="section-copy">Set schedule, mode, and bid behavior.</p>
        </div>
      </header>

      {blocked ? (
        <div className="panel">
          <p className="mb-3">Institution onboarding is not complete yet. Auction creation stays locked until activation.</p>
          <Link className="btn-brand" to="/admin/onboarding">Open Onboarding Wizard</Link>
        </div>
      ) : null}

      {!blocked ? (
        <div className="panel">
          <AuctionForm form={form} onChange={setForm} onSubmit={submit} submitLabel="Create Auction" />
        </div>
      ) : null}

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
