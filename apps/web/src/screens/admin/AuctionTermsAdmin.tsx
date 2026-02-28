"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Link, useParams } from "react-router-dom";
import { api, type AuctionTerms } from "../../api/client";
import MarkdownPreview from "../../components/MarkdownPreview";
import { LOCALES, type AppLocale } from "../../i18n/config";

export default function AuctionTermsAdmin() {
  const t = useTranslations("termsAdmin");
  const tLocales = useTranslations("locales");
  const { auctionId } = useParams<{ auctionId: string }>();
  const [items, setItems] = useState<AuctionTerms[]>([]);
  const [locale, setLocale] = useState<AppLocale>("en");
  const [content, setContent] = useState("");
  const [publishOnCreate, setPublishOnCreate] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [publishingId, setPublishingId] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    if (!auctionId) return;
    setLoading(true);
    setError(null);
    try {
      const next = await api.listAuctionTerms(auctionId, locale);
      setItems(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load terms");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [auctionId, locale]);

  const onCreate = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!auctionId) return;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.createAuctionTerms(auctionId, { locale, content_markdown: content, publish: publishOnCreate });
      setContent("");
      setPublishOnCreate(false);
      setMessage(publishOnCreate ? t("createdAndPublished") : t("created"));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create terms");
    } finally {
      setSaving(false);
    }
  };

  const onPublish = async (termsId: string) => {
    if (!auctionId) return;
    setPublishingId(termsId);
    setError(null);
    setMessage(null);
    try {
      await api.publishAuctionTerms(auctionId, termsId, locale);
      setMessage(t("published"));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to publish terms");
    } finally {
      setPublishingId(null);
    }
  };

  const current = items.find((item) => item.is_active) ?? null;

  return (
    <section className="page-shell max-w-6xl space-y-6">
      <header className="section-head">
        <div>
          <h1 className="section-title">{t("title")}</h1>
          <p className="section-copy">{t("description")}</p>
        </div>
        <Link className="btn-secondary" to={auctionId ? `/admin/auctions/${auctionId}/settings` : "/admin/auctions"}>
          {t("back")}
        </Link>
      </header>

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
      <div className="panel max-w-sm">
        <label className="space-y-2 text-sm">
          <span className="text-neutral-700">{t("locale")}</span>
          <select value={locale} onChange={(event) => setLocale(event.target.value as AppLocale)}>
            {LOCALES.map((item) => (
              <option key={item} value={item}>
                {tLocales(item)}
              </option>
            ))}
          </select>
        </label>
      </div>

      {loading ? <p className="text-sm text-neutral-500">{t("loading")}</p> : null}

      {!loading ? (
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
          <div className="space-y-6">
            <div className="panel space-y-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-neutral-500">{t("currentVersion")}</p>
                {current ? (
                  <p className="mt-1 text-lg font-medium text-neutral-900">Version {current.version_number}</p>
                ) : (
                  <p className="mt-1 text-sm text-neutral-500">{t("noActive")}</p>
                )}
              </div>
              {current ? <MarkdownPreview markdown={current.content_markdown} /> : null}
            </div>

            <div className="panel space-y-4">
              <h2 className="font-heading text-xl text-neutral-900">{t("versionHistory")}</h2>
              {items.length === 0 ? <p className="text-sm text-neutral-500">{t("noVersions")}</p> : null}
              <div className="space-y-3">
                {items.map((item) => (
                  <div key={item.id} className="rounded-lg border border-neutral-200 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p className="font-medium text-neutral-900">Version {item.version_number}</p>
                        <p className="text-xs text-neutral-500">
                          {item.created_at
                            ? new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(item.created_at))
                            : t("unknownDate")}
                        </p>
                      </div>
                      {item.is_active ? (
                        <span className="rounded-full bg-neutral-900 px-3 py-1 text-xs font-medium text-white">{t("active")}</span>
                      ) : (
                        <button
                          type="button"
                          className="btn-secondary"
                          disabled={publishingId === item.id}
                          onClick={() => void onPublish(item.id)}
                        >
                          {publishingId === item.id ? t("publishing") : t("publish")}
                        </button>
                      )}
                    </div>
                    <div className="mt-4 max-h-56 overflow-auto rounded-lg bg-neutral-50 p-4">
                      <MarkdownPreview markdown={item.content_markdown} className="text-sm" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <form className="panel space-y-4" onSubmit={onCreate}>
            <div>
              <h2 className="font-heading text-xl text-neutral-900">{t("createVersion")}</h2>
              <p className="mt-1 text-sm text-neutral-500">{t("immutable")}</p>
            </div>
            <label className="space-y-2 text-sm">
              <span className="text-neutral-700">{t("termsMarkdown")}</span>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={16}
                className="w-full rounded-lg border border-neutral-200 px-3 py-2 font-mono text-sm"
                placeholder={t("placeholder")}
              />
            </label>
            <label className="flex items-start gap-3 text-sm text-neutral-700">
              <input type="checkbox" className="mt-1" checked={publishOnCreate} onChange={(e) => setPublishOnCreate(e.target.checked)} />
              <span>{t("publishNow")}</span>
            </label>
            <div className="space-y-3 rounded-lg border border-neutral-200 bg-neutral-50 p-4">
              <p className="text-sm font-medium text-neutral-900">{t("preview")}</p>
              <MarkdownPreview markdown={content} className="text-sm" />
            </div>
            <button type="submit" className="btn-brand" disabled={saving || content.trim() === ""}>
              {saving ? t("saving") : t("create")}
            </button>
          </form>
        </div>
      ) : null}
    </section>
  );
}
