"use client";

import { useEffect, useState } from "react";
import { Link, Navigate, useParams } from "../../react-router-dom";
import { api, type TenantBidderProfile } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import { useAuth } from "../../hooks/useAuth";

type FormState = {
  phone: string;
  company_name: string;
  address_line1: string;
  address_line2: string;
  city: string;
  postal_code: string;
  country: string;
  vat_classification: string;
  accept_terms: boolean;
  accept_privacy: boolean;
};

const EMPTY_FORM: FormState = {
  phone: "",
  company_name: "",
  address_line1: "",
  address_line2: "",
  city: "",
  postal_code: "",
  country: "",
  vat_classification: "",
  accept_terms: false,
  accept_privacy: false,
};

export default function BidderOnboarding() {
  const { tenant } = useParams<{ tenant?: string }>();
  const tenantAlias = tenant || "current";
  const { loading: authLoading, isAuthenticated, isAdmin } = useAuth();
  const [profile, setProfile] = useState<TenantBidderProfile | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (authLoading || !isAuthenticated || isAdmin) return;
    setLoading(true);
    setError(null);
    api
      .getMyBidderOnboarding()
      .then((data) => {
        setProfile(data);
        setForm({
          phone: data.phone ?? "",
          company_name: data.company_name ?? "",
          address_line1: data.address_line1 ?? "",
          address_line2: data.address_line2 ?? "",
          city: data.city ?? "",
          postal_code: data.postal_code ?? "",
          country: data.country ?? "",
          vat_classification: data.vat_classification ?? "",
          accept_terms: Boolean(data.terms_accepted_at),
          accept_privacy: Boolean(data.privacy_accepted_at),
        });
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load onboarding"))
      .finally(() => setLoading(false));
  }, [authLoading, isAuthenticated, isAdmin]);

  if (authLoading) {
    return <section className="mx-auto max-w-[960px] px-4 py-10 md:px-8 md:py-12"><p className="text-sm text-neutral-500">Loading onboarding...</p></section>;
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (isAdmin) return <Navigate to="/admin/dashboard" replace />;

  const saveDraft = async (silent = false) => {
    setSaving(true);
    if (!silent) {
      setError(null);
      setMessage(null);
    }
    try {
      const updated = await api.saveMyBidderOnboarding({
        phone: form.phone,
        company_name: form.company_name,
        address_line1: form.address_line1,
        address_line2: form.address_line2,
        city: form.city,
        postal_code: form.postal_code,
        country: form.country,
        vat_classification: form.vat_classification,
        accept_terms: form.accept_terms,
        accept_privacy: form.accept_privacy,
      });
      setProfile(updated);
      if (!silent) {
        setMessage("Draft saved.");
      }
      return updated;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to save onboarding";
      setError(message);
      throw err;
    } finally {
      setSaving(false);
    }
  };

  const submit = async () => {
    setSubmitting(true);
    setError(null);
    setMessage(null);
    try {
      await saveDraft(true);
      const updated = await api.submitMyBidderOnboarding();
      setProfile(updated);
      setMessage(
        updated.bidder_status === "approved"
          ? "Onboarding approved. You can now bid."
          : updated.bidder_status === "pending_review"
            ? "Onboarding submitted. Your institution will review it."
            : "Onboarding updated.",
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to submit onboarding");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="mx-auto max-w-[960px] px-4 py-10 md:px-8 md:py-12">
      <header className="mb-8 flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-neutral-500">Bidder Onboarding</p>
          <h1 className="mt-2 font-heading text-3xl text-neutral-900">Complete your institution profile before bidding</h1>
          <p className="mt-2 text-sm text-neutral-500">Finish identity, VAT, and policy confirmation once per institution.</p>
        </div>
        <Link className="btn-link-brand" to={`/dashboard/${tenantAlias}`}>
          Back to Dashboard
        </Link>
      </header>

      <div className="mb-6 grid gap-3 md:grid-cols-4">
        {[
          "1 Identity Confirmation",
          "2 VAT Classification",
          "3 Terms Acceptance",
          "4 Review & Submit",
        ].map((label) => (
          <div key={label} className="rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm font-medium text-neutral-700">
            {label}
          </div>
        ))}
      </div>

      {error ? <p className="error mb-4">{error}</p> : null}
      {message ? <p className="success mb-4">{message}</p> : null}
      {loading ? <p className="text-sm text-neutral-500">Loading onboarding data...</p> : null}

      {!loading ? (
        <div className="space-y-6">
          <section className="panel space-y-4">
            <h2 className="font-heading text-xl text-neutral-900">Identity Confirmation</h2>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="text-sm">
                <span>Phone</span>
                <input value={form.phone} onChange={(e) => setForm((prev) => ({ ...prev, phone: e.target.value }))} />
              </label>
              <label className="text-sm">
                <span>Company</span>
                <input value={form.company_name} onChange={(e) => setForm((prev) => ({ ...prev, company_name: e.target.value }))} />
              </label>
              <label className="text-sm md:col-span-2">
                <span>Address Line 1</span>
                <input value={form.address_line1} onChange={(e) => setForm((prev) => ({ ...prev, address_line1: e.target.value }))} />
              </label>
              <label className="text-sm md:col-span-2">
                <span>Address Line 2</span>
                <input value={form.address_line2} onChange={(e) => setForm((prev) => ({ ...prev, address_line2: e.target.value }))} />
              </label>
              <label className="text-sm">
                <span>City</span>
                <input value={form.city} onChange={(e) => setForm((prev) => ({ ...prev, city: e.target.value }))} />
              </label>
              <label className="text-sm">
                <span>Postal Code</span>
                <input value={form.postal_code} onChange={(e) => setForm((prev) => ({ ...prev, postal_code: e.target.value }))} />
              </label>
              <label className="text-sm">
                <span>Country</span>
                <input value={form.country} onChange={(e) => setForm((prev) => ({ ...prev, country: e.target.value }))} />
              </label>
            </div>
          </section>

          <section className="panel space-y-4">
            <h2 className="font-heading text-xl text-neutral-900">VAT Classification</h2>
            <label className="text-sm">
              <span>VAT Classification</span>
              <input
                value={form.vat_classification}
                onChange={(e) => setForm((prev) => ({ ...prev, vat_classification: e.target.value }))}
                placeholder="Business, private collector, VAT exempt, etc."
              />
            </label>
          </section>

          <section className="panel space-y-4">
            <h2 className="font-heading text-xl text-neutral-900">Terms Acceptance</h2>
            <label className="flex items-start gap-3 text-sm text-neutral-700">
              <input
                type="checkbox"
                checked={form.accept_terms}
                onChange={(e) => setForm((prev) => ({ ...prev, accept_terms: e.target.checked }))}
                className="mt-1"
              />
              <span>I confirm that I accept the institution bidding terms.</span>
            </label>
            <label className="flex items-start gap-3 text-sm text-neutral-700">
              <input
                type="checkbox"
                checked={form.accept_privacy}
                onChange={(e) => setForm((prev) => ({ ...prev, accept_privacy: e.target.checked }))}
                className="mt-1"
              />
              <span>I confirm that I accept the institution privacy policy.</span>
            </label>
          </section>

          <section className="panel space-y-4">
            <h2 className="font-heading text-xl text-neutral-900">Review & Submit</h2>
            <div className="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm text-neutral-700">
              Current status: <span className="font-semibold">{profile?.bidder_status ?? "incomplete"}</span>
            </div>
            <p className="text-sm text-neutral-500">
              Submission requires a verified email, identity details, VAT classification, and both confirmations.
            </p>
            <div className="flex flex-wrap gap-3">
              <ElevareButton type="button" variant="secondary" onClick={() => void saveDraft()} disabled={saving || submitting}>
                {saving ? "Saving..." : "Save Draft"}
              </ElevareButton>
              <ElevareButton type="button" onClick={() => void submit()} disabled={saving || submitting}>
                {submitting ? "Submitting..." : "Submit"}
              </ElevareButton>
            </div>
          </section>
        </div>
      ) : null}
    </section>
  );
}
