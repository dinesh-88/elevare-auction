import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../hooks/useAuth";
import ElevareButton from "../../components/ui/ElevareButton";

export default function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const currentUser = await login(email, password);
      navigate(currentUser.role === "ADMIN" ? "/admin" : "/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="auth-card">
      <h1 className="font-heading text-3xl">Login</h1>
      <p className="text-sm text-neutral-600">Access your tenant dashboard or bidder account.</p>
      <form onSubmit={submit}>
        <label>Email</label>
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <label>Password</label>
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        <ElevareButton type="submit" disabled={submitting}>
          {submitting ? "Signing in..." : "Login"}
        </ElevareButton>
        {error ? <p className="error">{error}</p> : null}
      </form>
      <p className="text-sm text-neutral-600">
        No account? <Link to="/register">Register</Link>
      </p>
    </section>
  );
}
