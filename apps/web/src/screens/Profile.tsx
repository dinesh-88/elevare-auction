"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useAuth } from "../hooks/useAuth";
import { LOCALES, type AppLocale, localizePath, stripLocalePrefix } from "../i18n/config";

export default function Profile() {
  const t = useTranslations("profile");
  const tLocales = useTranslations("locales");
  const pathname = usePathname();
  const { user, loading, updateLocale } = useAuth();
  const [locale, setLocale] = useState<AppLocale>((user?.locale ?? "en") as AppLocale);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (user?.locale) {
      setLocale(user.locale);
    }
  }, [user?.locale]);

  const save = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const updated = await updateLocale(locale);
      setMessage(t("success"));
      const nextPath = localizePath(stripLocalePrefix(pathname || "/profile"), updated.locale);
      window.location.assign(nextPath);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update locale");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <section className="page-shell">
        <p className="text-sm text-neutral-500">{t("loading")}</p>
      </section>
    );
  }

  if (!user) {
    return (
      <section className="page-shell">
        <p className="error">Unauthorized</p>
      </section>
    );
  }

  return (
    <section className="page-shell max-w-3xl space-y-6">
      <header className="section-head">
        <div>
          <h1 className="section-title">{t("title")}</h1>
          <p className="section-copy">{t("description")}</p>
        </div>
      </header>

      <div className="panel space-y-6">
        <div className="grid gap-x-6 gap-y-4 sm:grid-cols-[160px_minmax(0,1fr)] sm:items-center">
          <label className="text-sm font-medium text-neutral-700">{t("email")}</label>
          <input value={user.email} readOnly disabled className="min-w-0" />

          <label className="text-sm font-medium text-neutral-700">{t("language")}</label>
          <select value={locale} onChange={(event) => setLocale(event.target.value as AppLocale)} className="min-w-0">
            {LOCALES.map((item) => (
              <option key={item} value={item}>
                {tLocales(item)}
              </option>
            ))}
          </select>

          <div className="sm:col-start-2 flex justify-start">
            <button type="button" className="btn-brand" disabled={saving || locale === user.locale} onClick={() => void save()}>
              {saving ? t("saving") : t("save")}
            </button>
          </div>
        </div>

        {error ? <p className="error">{error}</p> : null}
        {message ? <p className="success">{message}</p> : null}
      </div>
    </section>
  );
}
