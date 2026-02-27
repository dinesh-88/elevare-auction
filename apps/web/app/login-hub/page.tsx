"use client";

import { FormEvent, useMemo, useState } from "react";
import type { TenantLoginDiscoveryItem } from "../../src/api/client";
import { api } from "../../src/api/client";

export default function LoginHubPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [matches, setMatches] = useState<TenantLoginDiscoveryItem[]>([]);

  const normalizedEmail = useMemo(() => email.trim(), [email]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setMatches([]);

    if (!normalizedEmail) {
      setError("Enter your email address.");
      return;
    }

    try {
      setLoading(true);
      const items = await api.discoverTenantLogin(normalizedEmail);
      if (items.length === 0) {
        setError("No active tenant account found for this email.");
        return;
      }
      if (items.length === 1) {
        window.location.assign(items[0].login_url);
        return;
      }
      setMatches(items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not resolve tenant.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-neutral-50 text-brand-900">
      <main className="mx-auto w-full max-w-2xl px-6 py-20">
        <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Login Hub</p>
        <h1 className="mt-4 text-4xl font-semibold">Find Your Tenant Login</h1>
        <p className="mt-4 text-brand-700">
          Enter your account email and we will redirect you to the correct tenant login page.
        </p>

        <form className="mt-8 space-y-4 rounded-xl border border-neutral-200 bg-white p-6" onSubmit={onSubmit}>
          <label className="block text-sm font-medium text-brand-800" htmlFor="email">
            Work Email
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="name@institution.com"
            className="w-full rounded-md border border-neutral-200 px-4 py-3 text-brand-900 focus:outline-none focus:ring-2 focus:ring-accent-400/50"
          />
          <button
            type="submit"
            disabled={loading}
            className="inline-flex items-center rounded-md bg-accent-500 px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-accent-600 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading ? "Finding tenant..." : "Continue"}
          </button>
          {error ? <p className="text-sm text-danger-500">{error}</p> : null}
        </form>

        {matches.length > 1 ? (
          <section className="mt-6 rounded-xl border border-neutral-200 bg-white p-6">
            <h2 className="text-lg font-semibold text-brand-900">Multiple Tenants Found</h2>
            <p className="mt-2 text-sm text-brand-700">Select your organization to continue.</p>
            <ul className="mt-4 space-y-3">
              {matches.map((item) => (
                <li key={item.tenant_id}>
                  <a
                    href={item.login_url}
                    className="flex items-center justify-between rounded-md border border-neutral-200 px-4 py-3 text-brand-900 transition-colors hover:border-accent-400 hover:bg-neutral-50"
                  >
                    <span>{item.tenant_name}</span>
                    <span className="text-sm text-brand-700">{item.host}</span>
                  </a>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
      </main>
    </div>
  );
}
