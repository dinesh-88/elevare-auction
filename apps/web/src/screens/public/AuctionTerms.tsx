"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, type AuctionTerms as AuctionTermsRecord } from "../../api/client";
import { useAuth } from "../../hooks/useAuth";
import { getLocaleFromPathname } from "../../i18n/config";
import MarkdownPreview from "../../components/MarkdownPreview";
import { extractAuctionIdFromSlug } from "../../utils/auctionRoutes";

export default function AuctionTerms() {
  const { id } = useParams<{ id: string }>();
  const pathname = usePathname();
  const navigate = useNavigate();
  const { isAuthenticated, isAdmin, loading } = useAuth();
  const [terms, setTerms] = useState<AuctionTermsRecord | null>(null);
  const [hasActiveTerms, setHasActiveTerms] = useState(true);
  const [accepted, setAccepted] = useState(false);
  const [checked, setChecked] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const locale = getLocaleFromPathname(pathname);
  const auctionId = useMemo(() => extractAuctionIdFromSlug(id ?? "") ?? id ?? "", [id]);
  const returnTo = useMemo(() => {
    if (typeof window === "undefined") return id ? `/auction/${id}` : "/auctions";
    const params = new URLSearchParams(window.location.search);
    return params.get("returnTo") || (id ? `/auction/${id}` : "/auctions");
  }, [id]);

  useEffect(() => {
    if (!auctionId) return;
    let active = true;
    setError(null);
    (async () => {
      try {
        const activeTerms = await api.getActiveAuctionTerms(auctionId, locale);
        if (!active) return;
        if (!activeTerms) {
          setHasActiveTerms(false);
          setTerms(null);
          return;
        }
        setHasActiveTerms(true);
        setTerms(activeTerms);
        if (isAuthenticated && !isAdmin) {
          const status = await api.getAuctionTermsStatus(auctionId);
          if (!active) return;
          setAccepted(status.accepted);
        } else {
          setAccepted(false);
        }
      } catch (err) {
        if (!active) return;
        setError(err instanceof Error ? err.message : "Failed to load auction terms");
      }
    })();
    return () => {
      active = false;
    };
  }, [auctionId, isAuthenticated, isAdmin, locale]);

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      navigate("/login");
    }
  }, [isAuthenticated, loading, navigate]);

  const onAccept = async () => {
    if (!auctionId || !checked) return;
    setSubmitting(true);
    setError(null);
    try {
      await api.acceptAuctionTerms(auctionId);
      navigate(returnTo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to accept terms");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <p className="text-sm text-neutral-500">Loading...</p>;

  return (
    <section className="page-shell max-w-4xl space-y-6">
      <header className="space-y-2">
        <Link to={returnTo} className="btn-link-brand">
          ← Back to Auction
        </Link>
        <h1 className="font-heading text-3xl text-neutral-950">Auction Terms</h1>
        <p className="text-sm text-neutral-500">You must accept the active auction terms before bidding.</p>
      </header>

      {error ? <p className="error">{error}</p> : null}

      {!hasActiveTerms ? (
        <div className="panel space-y-3">
          <p className="text-neutral-700">This auction does not have published terms yet. Bidding is unavailable until the organizer publishes a version.</p>
        </div>
      ) : null}

      {terms ? (
        <div className="panel space-y-6">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-neutral-200 pb-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-neutral-500">Active Version</p>
              <p className="mt-1 text-lg font-medium text-neutral-900">Version {terms.version_number}</p>
            </div>
            {accepted ? <span className="rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">Accepted</span> : null}
          </div>

          <MarkdownPreview markdown={terms.content_markdown} />

          {!isAdmin ? (
            accepted ? (
              <div className="rounded-lg border border-green-200 bg-green-50 p-4 text-sm text-green-800">
                Your acceptance matches the active version. You can continue to bid.
              </div>
            ) : (
              <div className="space-y-4 border-t border-neutral-200 pt-5">
                <label className="flex items-start gap-3 text-sm text-neutral-700">
                  <input type="checkbox" className="mt-1" checked={checked} onChange={(e) => setChecked(e.target.checked)} />
                  <span>I have read and accept the active auction terms.</span>
                </label>
                <button type="button" className="btn-brand" disabled={!checked || submitting} onClick={onAccept}>
                  {submitting ? "Accepting..." : "Accept Terms"}
                </button>
              </div>
            )
          ) : null}
        </div>
      ) : null}
    </section>
  );
}
