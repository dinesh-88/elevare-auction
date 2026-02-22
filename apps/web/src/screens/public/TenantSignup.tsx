import { useState } from "react";
import { api } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

export default function TenantSignup() {
  const [form, setForm] = useState({
    organization_name: "",
    desired_subdomain: "",
    admin_email: "",
    password: "",
    country: "",
    estimated_auction_volume: "",
  });
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setMessage(null);
    try {
      await api.tenantSignup({
        organization_name: form.organization_name,
        desired_subdomain: form.desired_subdomain,
        admin_email: form.admin_email,
        password: form.password,
        country: form.country,
        estimated_auction_volume: form.estimated_auction_volume ? Number(form.estimated_auction_volume) : undefined,
      });
      setMessage("Application submitted. Your tenant is pending platform approval.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Signup failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="auth-card onboarding-card">
      <h1 className="font-heading text-3xl">Organization Signup</h1>
      <p className="text-sm text-neutral-600">Create your Elevare tenant request.</p>
      <form onSubmit={submit}>
        <label>Organization name</label>
        <input value={form.organization_name} onChange={(e) => setForm({ ...form, organization_name: e.target.value })} required />

        <label>Desired subdomain</label>
        <input
          placeholder="museum"
          value={form.desired_subdomain}
          onChange={(e) => setForm({ ...form, desired_subdomain: e.target.value })}
          required
        />

        <label>Admin email</label>
        <input
          type="email"
          value={form.admin_email}
          onChange={(e) => setForm({ ...form, admin_email: e.target.value })}
          required
        />

        <label>Password</label>
        <input
          type="password"
          minLength={8}
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
          required
        />

        <label>Country</label>
        <input value={form.country} onChange={(e) => setForm({ ...form, country: e.target.value })} required />

        <label>Estimated auction volume (optional)</label>
        <input
          type="number"
          min={0}
          value={form.estimated_auction_volume}
          onChange={(e) => setForm({ ...form, estimated_auction_volume: e.target.value })}
        />

        <ElevareButton type="submit" disabled={loading}>
          {loading ? "Submitting..." : "Submit request"}
        </ElevareButton>
      </form>
      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
