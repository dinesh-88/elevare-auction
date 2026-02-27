"use client";

import { FormEvent, useMemo, useState } from "react";

type ResolvedTenant = {
  slug: string;
  name: string;
};

export default function LoginHubPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [matches, setMatches] = useState<ResolvedTenant[]>([]);

  const normalizedEmail = useMemo(() => email.trim().toLowerCase(), [email]);

  const redirectToTenant = (slug: string) => {
    const scheme = (process.env.NEXT_PUBLIC_TENANT_APP_SCHEME ?? "https").trim() || "https";
    const baseDomain = (process.env.NEXT_PUBLIC_TENANT_APP_BASE_DOMAIN ?? "elevareauctions.com").trim() || "elevareauctions.com";
    const url = new URL(`${scheme}://${slug}.${baseDomain}/login`);
    if (normalizedEmail) {
      url.searchParams.set("email", normalizedEmail);
    }
    window.location.assign(url.toString());
  };

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setMatches([]);

    if (!normalizedEmail) {
      setError("Enter your email address.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/resolve", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ email: normalizedEmail }),
      });
      const payload = (await response.json().catch(() => ({}))) as {
        error?: string;
        redirect_url?: string;
        tenants?: ResolvedTenant[];
      };

      if (!response.ok) {
        setError(payload.error ?? "Unable to resolve your institution.");
        return;
      }

      if (payload.redirect_url) {
        window.location.assign(payload.redirect_url);
        return;
      }

      if (Array.isArray(payload.tenants) && payload.tenants.length > 0) {
        setMatches(payload.tenants);
        return;
      }

      setError("No associated institutions found.");
    } catch {
      setError("Unable to resolve your institution.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-[linear-gradient(180deg,#f7f4ee_0%,#f1ece3_100%)] px-6 py-16 text-stone-900">
      <div className="mx-auto max-w-3xl">
        <section className="rounded-[28px] border border-stone-300/80 bg-white/90 px-7 py-8 shadow-[0_24px_80px_-48px_rgba(41,37,36,0.38)] sm:px-10 sm:py-10">
          <p className="text-xs uppercase tracking-[0.18em] text-stone-500">Login Hub</p>
          <h1 className="mt-4 font-serif text-4xl leading-tight sm:text-5xl">Resolve Your Institution Access</h1>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-stone-600 sm:text-base">
            Enter your account email and we will direct you to the appropriate institution sign-in page.
          </p>

          <form className="mt-8 space-y-4" onSubmit={onSubmit}>
            <div>
              <label className="mb-2 block text-sm font-medium text-stone-700" htmlFor="email">
                Email Address
              </label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="name@institution.org"
                className="w-full rounded-2xl border border-stone-300 bg-white px-4 py-3 text-base text-stone-900 outline-none transition focus:border-stone-500 focus:ring-2 focus:ring-stone-300/60"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="inline-flex min-w-40 items-center justify-center rounded-2xl bg-stone-900 px-5 py-3 text-sm font-medium text-white transition hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? "Resolving..." : "Continue"}
            </button>
            {error ? <p className="text-sm text-red-700">{error}</p> : null}
          </form>
        </section>

        {matches.length > 1 ? (
          <section className="mt-6 rounded-[28px] border border-stone-300/80 bg-white/90 px-7 py-8 shadow-[0_24px_80px_-48px_rgba(41,37,36,0.28)] sm:px-10">
            <h2 className="font-serif text-2xl">Select Your Institution</h2>
            <p className="mt-3 text-sm leading-7 text-stone-600">
              We found multiple institutions associated with this email address. Choose one to continue.
            </p>
            <div className="mt-6 grid gap-3">
              {matches.map((item) => (
                <button
                  key={item.slug}
                  type="button"
                  onClick={() => redirectToTenant(item.slug)}
                  className="flex items-center justify-between rounded-2xl border border-stone-200 px-4 py-4 text-left transition hover:border-stone-400 hover:bg-stone-50"
                >
                  <span className="font-medium text-stone-900">{item.name}</span>
                  <span className="text-sm text-stone-500">{item.slug}.elevareauctions.com</span>
                </button>
              ))}
            </div>
          </section>
        ) : null}
      </div>
    </main>
  );
}
