"use client";

import { useMemo, useState } from "react";
import { usePathname, useSearchParams } from "next/navigation";
import { api } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";
import { getLocaleFromPathname, localizePath } from "../../i18n/config";

export default function AcceptInvite() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const token = useMemo(() => searchParams.get("token") ?? "", [searchParams]);
  const locale = getLocaleFromPathname(pathname);
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setMessage(null);
    try {
      await api.acceptStaffInvitation({ token, password, locale });
      setMessage("Invitation accepted. Redirecting…");
      window.location.assign(localizePath("/admin/staff", locale));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to accept invitation");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="auth-card">
      <h1 className="font-heading text-3xl">Accept Staff Invitation</h1>
      <p className="text-sm text-neutral-600">
        Accept your invitation to join this institution. If you already have an account for this email, the password can be left blank.
      </p>
      <form onSubmit={submit}>
        <label>Invitation Token</label>
        <input value={token} readOnly />
        <label>Password (new accounts only)</label>
        <input type="password" minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} />
        <ElevareButton type="submit" disabled={submitting || !token}>
          {submitting ? "Accepting…" : "Accept Invitation"}
        </ElevareButton>
        {error ? <p className="error">{error}</p> : null}
        {message ? <p className="success">{message}</p> : null}
      </form>
    </section>
  );
}
