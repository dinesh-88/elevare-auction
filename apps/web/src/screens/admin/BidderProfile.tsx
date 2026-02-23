import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, type TenantBidderProfile } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

export default function BidderProfile() {
  const { userId } = useParams<{ userId: string }>();
  const [profile, setProfile] = useState<TenantBidderProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) return;
    setLoading(true);
    api
      .getBidderProfile(userId)
      .then(setProfile)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load profile"))
      .finally(() => setLoading(false));
  }, [userId]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId || !profile) return;
    setSaving(true);
    setError(null);
    try {
      const updated = await api.updateBidderProfile(userId, {
        status: profile.status,
        risk_level: profile.risk_level,
        internal_notes: profile.internal_notes ?? null,
        phone: profile.phone ?? null,
        company_name: profile.company_name ?? null,
        address_line1: profile.address_line1 ?? null,
        address_line2: profile.address_line2 ?? null,
        city: profile.city ?? null,
        postal_code: profile.postal_code ?? null,
        country: profile.country ?? null,
      });
      setProfile(updated);
      setMessage("Profile saved.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save profile");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <p className="text-sm text-neutral-500">Loading bidder profile...</p>;
  if (!profile) return <p className="error">Bidder profile not found.</p>;

  return (
    <section className="page-shell max-w-3xl">
      <header className="section-head">
        <div>
          <h1 className="section-title">Bidder Profile</h1>
          <p className="section-copy">{profile.user_id}</p>
        </div>
        <Link className="btn-link-brand" to="/admin/bidders">← Back to Bidders</Link>
      </header>
      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}
      <form className="panel space-y-4" onSubmit={save}>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          <label className="text-sm">
            <span>Status</span>
            <select value={profile.status} onChange={(e) => setProfile((p) => (p ? { ...p, status: e.target.value as TenantBidderProfile["status"] } : p))}>
              <option value="ACTIVE">ACTIVE</option>
              <option value="SUSPENDED">SUSPENDED</option>
            </select>
          </label>
          <label className="text-sm">
            <span>Risk Level</span>
            <select value={profile.risk_level} onChange={(e) => setProfile((p) => (p ? { ...p, risk_level: e.target.value as TenantBidderProfile["risk_level"] } : p))}>
              <option value="LOW">LOW</option>
              <option value="NORMAL">NORMAL</option>
              <option value="HIGH">HIGH</option>
            </select>
          </label>
          <label className="text-sm">
            <span>Phone</span>
            <input value={profile.phone ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, phone: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>Company</span>
            <input value={profile.company_name ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, company_name: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>Address 1</span>
            <input value={profile.address_line1 ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, address_line1: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>Address 2</span>
            <input value={profile.address_line2 ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, address_line2: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>City</span>
            <input value={profile.city ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, city: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>Postal Code</span>
            <input value={profile.postal_code ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, postal_code: e.target.value } : p))} />
          </label>
          <label className="text-sm">
            <span>Country</span>
            <input value={profile.country ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, country: e.target.value } : p))} />
          </label>
        </div>
        <label className="text-sm">
          <span>Internal Notes</span>
          <textarea rows={5} value={profile.internal_notes ?? ""} onChange={(e) => setProfile((p) => (p ? { ...p, internal_notes: e.target.value } : p))} />
        </label>
        <div className="actions">
          <ElevareButton type="submit" disabled={saving}>{saving ? "Saving..." : "Save"}</ElevareButton>
        </div>
      </form>
    </section>
  );
}
