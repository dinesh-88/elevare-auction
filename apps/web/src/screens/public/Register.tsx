import { useState } from "react";
import { api } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

export default function Register() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setMessage(null);
    try {
      await api.register({ email, password });
      setMessage("Registered successfully. Check email to verify.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="auth-card">
      <h1 className="font-heading text-3xl">Register</h1>
      <p className="text-sm text-neutral-600">Create a bidder account for this tenant.</p>
      <form onSubmit={submit}>
        <label>Email</label>
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <label>Password</label>
        <input type="password" minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} required />
        <ElevareButton type="submit" disabled={submitting}>
          {submitting ? "Submitting..." : "Register"}
        </ElevareButton>
        {error ? <p className="error">{error}</p> : null}
        {message ? <p className="success">{message}</p> : null}
      </form>
    </section>
  );
}
