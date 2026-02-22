import { useEffect, useMemo, useState } from "react";
import { api, type TenantAsset, type TenantSitePayload } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

type TabKey = "branding" | "homepage" | "pages" | "assets";

type BrandingDraft = {
  logo_url: string;
  primary_color: string;
  secondary_color: string;
  favicon_url: string;
  font_family: string;
};

type Highlight = { title: string; body: string; icon: string };
type SocialLink = { label: string; href: string };

type ContentDraft = {
  hero: {
    headline: string;
    subheadline: string;
    banner_image_url: string;
    cta_label: string;
    cta_href: string;
  };
  featured: {
    mode: "AUTO" | "MANUAL";
    auction_id: string;
  };
  highlights: Highlight[];
  footer: {
    text: string;
    contact_email: string;
    social_links: SocialLink[];
  };
  about_markdown: string;
  contact: {
    address: string;
    phone: string;
    email: string;
  };
};

const tabs: Array<{ key: TabKey; label: string }> = [
  { key: "branding", label: "Branding" },
  { key: "homepage", label: "Homepage" },
  { key: "pages", label: "Pages" },
  { key: "assets", label: "Assets" },
];

const emptyContent: ContentDraft = {
  hero: {
    headline: "",
    subheadline: "",
    banner_image_url: "",
    cta_label: "",
    cta_href: "/auctions",
  },
  featured: { mode: "AUTO", auction_id: "" },
  highlights: [{ title: "", body: "", icon: "" }],
  footer: { text: "", contact_email: "", social_links: [] },
  about_markdown: "",
  contact: { address: "", phone: "", email: "" },
};

function asObject(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function parseBranding(raw: Record<string, unknown>): BrandingDraft {
  return {
    logo_url: asString(raw.logo_url),
    primary_color: asString(raw.primary_color),
    secondary_color: asString(raw.secondary_color),
    favicon_url: asString(raw.favicon_url),
    font_family: asString(raw.font_family),
  };
}

function parseContent(raw: Record<string, unknown>): ContentDraft {
  const homepage = asObject(raw.homepage);
  const hero = asObject(homepage.hero);
  const cta = asObject(hero.cta);
  const featuredAuction = asObject(homepage.featuredAuction);
  const footer = asObject(raw.footer);
  const about = asObject(raw.about);
  const contact = asObject(raw.contact);

  const highlightsRaw = Array.isArray(homepage.highlights) ? homepage.highlights : [];
  const highlights: Highlight[] = highlightsRaw
    .map((row) => asObject(row))
    .map((row) => ({ title: asString(row.title), body: asString(row.body), icon: asString(row.icon) }));

  const socialLinksObj = asObject(footer.social_links);
  const social_links: SocialLink[] = Object.entries(socialLinksObj).map(([label, href]) => ({ label, href: asString(href) }));

  return {
    hero: {
      headline: asString(hero.headline),
      subheadline: asString(hero.subheadline),
      banner_image_url: asString(hero.banner_image_url),
      cta_label: asString(cta.label),
      cta_href: asString(cta.href),
    },
    featured: {
      mode: asString(featuredAuction.mode) === "MANUAL" ? "MANUAL" : "AUTO",
      auction_id: asString(featuredAuction.auction_id),
    },
    highlights: highlights.length > 0 ? highlights : [{ title: "", body: "", icon: "" }],
    footer: {
      text: asString(footer.text),
      contact_email: asString(footer.contact_email),
      social_links,
    },
    about_markdown: asString(about.markdown),
    contact: {
      address: asString(contact.address),
      phone: asString(contact.phone),
      email: asString(contact.email),
    },
  };
}

function toSitePayload(branding: BrandingDraft, content: ContentDraft): TenantSitePayload {
  const highlights = content.highlights
    .map((h) => ({ title: h.title.trim(), body: h.body.trim(), icon: h.icon.trim() }))
    .filter((h) => h.title || h.body || h.icon);
  const social_links: Record<string, string> = {};
  content.footer.social_links.forEach((s) => {
    const label = s.label.trim();
    const href = s.href.trim();
    if (label && href) social_links[label] = href;
  });

  return {
    branding: {
      logo_url: branding.logo_url.trim(),
      primary_color: branding.primary_color.trim(),
      secondary_color: branding.secondary_color.trim(),
      favicon_url: branding.favicon_url.trim(),
      font_family: branding.font_family.trim(),
    },
    content: {
      homepage: {
        hero: {
          headline: content.hero.headline.trim(),
          subheadline: content.hero.subheadline.trim(),
          banner_image_url: content.hero.banner_image_url.trim(),
          cta: {
            label: content.hero.cta_label.trim(),
            href: content.hero.cta_href.trim(),
          },
        },
        featuredAuction: {
          mode: content.featured.mode,
          auction_id: content.featured.mode === "MANUAL" ? content.featured.auction_id.trim() || undefined : undefined,
        },
        highlights,
      },
      footer: {
        text: content.footer.text.trim(),
        contact_email: content.footer.contact_email.trim(),
        social_links,
      },
      about: {
        markdown: content.about_markdown,
      },
      contact: {
        address: content.contact.address.trim(),
        phone: content.contact.phone.trim(),
        email: content.contact.email.trim(),
      },
    },
  };
}

export default function Site() {
  const [activeTab, setActiveTab] = useState<TabKey>("branding");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [contentVersion, setContentVersion] = useState<number>(1);

  const [branding, setBranding] = useState<BrandingDraft>({
    logo_url: "",
    primary_color: "",
    secondary_color: "",
    favicon_url: "",
    font_family: "",
  });
  const [content, setContent] = useState<ContentDraft>(emptyContent);
  const [assets, setAssets] = useState<TenantAsset[]>([]);
  const [assetKind, setAssetKind] = useState<"logo" | "banner" | "image">("image");
  const [uploadingAsset, setUploadingAsset] = useState(false);

  const previewUrl = useMemo(() => {
    if (typeof window === "undefined") return "";
    const token = process.env.NEXT_PUBLIC_SITE_PREVIEW_TOKEN;
    if (!token) return "";
    return `${window.location.origin}/?previewToken=${token}`;
  }, []);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const [siteState, assetRows] = await Promise.all([api.getSite(), api.listAssets(undefined, 100)]);
      setBranding(parseBranding(asObject(siteState.branding_draft)));
      setContent(parseContent(asObject(siteState.content_draft)));
      setAssets(assetRows);
      setContentVersion(siteState.content_version || 1);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load site settings");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const saveDraft = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.updateSiteDraft(toSitePayload(branding, content));
      setMessage("Draft saved.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save draft");
    } finally {
      setSaving(false);
    }
  };

  const publish = async () => {
    setPublishing(true);
    setError(null);
    setMessage(null);
    try {
      const result = await api.publishSite();
      setContentVersion(result.after_version);
      setMessage(`Published. Version ${result.after_version}.`);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to publish");
    } finally {
      setPublishing(false);
    }
  };

  const uploadAsset = async (file: File) => {
    setUploadingAsset(true);
    setError(null);
    try {
      const uploaded = await api.uploadAsset(assetKind, file);
      setAssets((prev) => [uploaded, ...prev]);
      if (assetKind === "logo") setBranding((prev) => ({ ...prev, logo_url: uploaded.url }));
      if (assetKind === "banner") {
        setContent((prev) => ({ ...prev, hero: { ...prev.hero, banner_image_url: uploaded.url } }));
      }
      setMessage(`${assetKind} uploaded.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to upload asset");
    } finally {
      setUploadingAsset(false);
    }
  };

  if (loading) {
    return <section className="page-shell"><p className="text-sm text-neutral-500">Loading site settings...</p></section>;
  }

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Site Content</h1>
          <p className="section-copy">Manage branding, homepage content, and pages. Current version: {contentVersion}</p>
        </div>
        <div className="actions">
          {previewUrl ? (
            <a className="btn-link-brand" href={previewUrl} target="_blank" rel="noreferrer">
              Preview Draft
            </a>
          ) : null}
          <ElevareButton type="button" variant="secondary" onClick={saveDraft} disabled={saving || publishing}>
            {saving ? "Saving..." : "Save Draft"}
          </ElevareButton>
          <ElevareButton type="button" onClick={publish} disabled={publishing || saving}>
            {publishing ? "Publishing..." : "Publish"}
          </ElevareButton>
        </div>
      </header>

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}

      <div className="panel">
        <div className="mb-4 flex flex-wrap gap-2">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              type="button"
              onClick={() => setActiveTab(tab.key)}
              className={`rounded-md border px-3 py-1.5 text-sm ${activeTab === tab.key ? "border-neutral-900 bg-neutral-900 text-white" : "border-neutral-300 bg-white text-neutral-700"}`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {activeTab === "branding" ? (
          <div className="form-grid">
            <label>Logo URL</label>
            <input value={branding.logo_url} onChange={(e) => setBranding((p) => ({ ...p, logo_url: e.target.value }))} />
            <label>Primary Color (#RRGGBB)</label>
            <input value={branding.primary_color} onChange={(e) => setBranding((p) => ({ ...p, primary_color: e.target.value }))} />
            <label>Secondary Color (#RRGGBB)</label>
            <input value={branding.secondary_color} onChange={(e) => setBranding((p) => ({ ...p, secondary_color: e.target.value }))} />
            <label>Favicon URL</label>
            <input value={branding.favicon_url} onChange={(e) => setBranding((p) => ({ ...p, favicon_url: e.target.value }))} />
            <label>Font Family (optional)</label>
            <input value={branding.font_family} onChange={(e) => setBranding((p) => ({ ...p, font_family: e.target.value }))} />
          </div>
        ) : null}

        {activeTab === "homepage" ? (
          <div className="space-y-6">
            <div className="form-grid">
              <label>Hero Headline</label>
              <input value={content.hero.headline} onChange={(e) => setContent((p) => ({ ...p, hero: { ...p.hero, headline: e.target.value } }))} />
              <label>Hero Subheadline</label>
              <input value={content.hero.subheadline} onChange={(e) => setContent((p) => ({ ...p, hero: { ...p.hero, subheadline: e.target.value } }))} />
              <label>Hero Banner Image URL</label>
              <input value={content.hero.banner_image_url} onChange={(e) => setContent((p) => ({ ...p, hero: { ...p.hero, banner_image_url: e.target.value } }))} />
              <label>CTA Label</label>
              <input value={content.hero.cta_label} onChange={(e) => setContent((p) => ({ ...p, hero: { ...p.hero, cta_label: e.target.value } }))} />
              <label>CTA Link</label>
              <input value={content.hero.cta_href} onChange={(e) => setContent((p) => ({ ...p, hero: { ...p.hero, cta_href: e.target.value } }))} />
            </div>

            <div className="form-grid">
              <label>Featured Auction Mode</label>
              <select
                value={content.featured.mode}
                onChange={(e) =>
                  setContent((p) => ({
                    ...p,
                    featured: { ...p.featured, mode: e.target.value === "MANUAL" ? "MANUAL" : "AUTO" },
                  }))
                }
              >
                <option value="AUTO">AUTO</option>
                <option value="MANUAL">MANUAL</option>
              </select>
              {content.featured.mode === "MANUAL" ? (
                <>
                  <label>Featured Auction ID</label>
                  <input value={content.featured.auction_id} onChange={(e) => setContent((p) => ({ ...p, featured: { ...p.featured, auction_id: e.target.value } }))} />
                </>
              ) : null}
            </div>

            <div className="space-y-3">
              <div className="section-head">
                <h3 className="font-heading text-lg text-neutral-900">Highlights</h3>
                <ElevareButton
                  type="button"
                  variant="secondary"
                  onClick={() => setContent((p) => ({ ...p, highlights: [...p.highlights, { title: "", body: "", icon: "" }] }))}
                >
                  Add Highlight
                </ElevareButton>
              </div>
              {content.highlights.map((item, index) => (
                <div key={`highlight-${index}`} className="rounded-md border border-neutral-200 p-4">
                  <div className="grid gap-3 md:grid-cols-3">
                    <input
                      placeholder="Title"
                      value={item.title}
                      onChange={(e) =>
                        setContent((p) => ({
                          ...p,
                          highlights: p.highlights.map((row, i) => (i === index ? { ...row, title: e.target.value } : row)),
                        }))
                      }
                    />
                    <input
                      placeholder="Body"
                      value={item.body}
                      onChange={(e) =>
                        setContent((p) => ({
                          ...p,
                          highlights: p.highlights.map((row, i) => (i === index ? { ...row, body: e.target.value } : row)),
                        }))
                      }
                    />
                    <input
                      placeholder="Icon key"
                      value={item.icon}
                      onChange={(e) =>
                        setContent((p) => ({
                          ...p,
                          highlights: p.highlights.map((row, i) => (i === index ? { ...row, icon: e.target.value } : row)),
                        }))
                      }
                    />
                  </div>
                  <div className="mt-3">
                    <ElevareButton
                      type="button"
                      variant="secondary"
                      onClick={() =>
                        setContent((p) => ({
                          ...p,
                          highlights: p.highlights.filter((_, i) => i !== index),
                        }))
                      }
                    >
                      Remove
                    </ElevareButton>
                  </div>
                </div>
              ))}
            </div>

            <div className="form-grid">
              <label>Footer Text</label>
              <input value={content.footer.text} onChange={(e) => setContent((p) => ({ ...p, footer: { ...p.footer, text: e.target.value } }))} />
              <label>Footer Contact Email</label>
              <input value={content.footer.contact_email} onChange={(e) => setContent((p) => ({ ...p, footer: { ...p.footer, contact_email: e.target.value } }))} />
            </div>

            <div className="space-y-3">
              <div className="section-head">
                <h3 className="font-heading text-lg text-neutral-900">Footer Social Links</h3>
                <ElevareButton
                  type="button"
                  variant="secondary"
                  onClick={() =>
                    setContent((p) => ({
                      ...p,
                      footer: { ...p.footer, social_links: [...p.footer.social_links, { label: "", href: "" }] },
                    }))
                  }
                >
                  Add Link
                </ElevareButton>
              </div>
              {content.footer.social_links.map((link, index) => (
                <div key={`social-${index}`} className="grid gap-2 md:grid-cols-[1fr_2fr_auto]">
                  <input
                    placeholder="Label"
                    value={link.label}
                    onChange={(e) =>
                      setContent((p) => ({
                        ...p,
                        footer: {
                          ...p.footer,
                          social_links: p.footer.social_links.map((row, i) =>
                            i === index ? { ...row, label: e.target.value } : row,
                          ),
                        },
                      }))
                    }
                  />
                  <input
                    placeholder="https://..."
                    value={link.href}
                    onChange={(e) =>
                      setContent((p) => ({
                        ...p,
                        footer: {
                          ...p.footer,
                          social_links: p.footer.social_links.map((row, i) =>
                            i === index ? { ...row, href: e.target.value } : row,
                          ),
                        },
                      }))
                    }
                  />
                  <ElevareButton
                    type="button"
                    variant="secondary"
                    onClick={() =>
                      setContent((p) => ({
                        ...p,
                        footer: {
                          ...p.footer,
                          social_links: p.footer.social_links.filter((_, i) => i !== index),
                        },
                      }))
                    }
                  >
                    Remove
                  </ElevareButton>
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {activeTab === "pages" ? (
          <div className="space-y-4">
            <div>
              <label>About (Markdown)</label>
              <textarea
                rows={8}
                value={content.about_markdown}
                onChange={(e) => setContent((p) => ({ ...p, about_markdown: e.target.value }))}
              />
            </div>
            <div className="form-grid">
              <label>Contact Address</label>
              <input value={content.contact.address} onChange={(e) => setContent((p) => ({ ...p, contact: { ...p.contact, address: e.target.value } }))} />
              <label>Contact Phone</label>
              <input value={content.contact.phone} onChange={(e) => setContent((p) => ({ ...p, contact: { ...p.contact, phone: e.target.value } }))} />
              <label>Contact Email</label>
              <input value={content.contact.email} onChange={(e) => setContent((p) => ({ ...p, contact: { ...p.contact, email: e.target.value } }))} />
            </div>
          </div>
        ) : null}

        {activeTab === "assets" ? (
          <div className="space-y-4">
            <div className="grid gap-3 md:grid-cols-[220px_1fr_auto] md:items-end">
              <div>
                <label>Asset Kind</label>
                <select value={assetKind} onChange={(e) => setAssetKind(e.target.value as "logo" | "banner" | "image")}> 
                  <option value="logo">Logo</option>
                  <option value="banner">Banner</option>
                  <option value="image">Image</option>
                </select>
              </div>
              <div>
                <label>Upload File</label>
                <input
                  type="file"
                  accept={assetKind === "logo" ? "image/png,image/jpeg,image/webp,image/svg+xml" : "image/png,image/jpeg,image/webp"}
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) {
                      void uploadAsset(file);
                    }
                  }}
                />
              </div>
              <p className="text-xs text-neutral-500">Max 10MB</p>
            </div>

            {uploadingAsset ? <p className="text-sm text-neutral-500">Uploading asset...</p> : null}

            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Preview</th>
                    <th>Kind</th>
                    <th>URL</th>
                    <th>Created</th>
                    <th>Use</th>
                  </tr>
                </thead>
                <tbody>
                  {assets.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="text-sm text-neutral-500">
                        No assets uploaded yet.
                      </td>
                    </tr>
                  ) : (
                    assets.map((asset) => (
                      <tr key={asset.id}>
                        <td>
                          <img src={asset.url} alt={asset.kind} className="h-10 w-16 rounded object-cover border border-neutral-200" />
                        </td>
                        <td className="capitalize">{asset.kind}</td>
                        <td className="max-w-[360px] truncate text-xs text-neutral-600">{asset.url}</td>
                        <td>{new Date(asset.created_at).toLocaleString()}</td>
                        <td>
                          {asset.kind === "logo" ? (
                            <ElevareButton type="button" variant="secondary" onClick={() => setBranding((p) => ({ ...p, logo_url: asset.url }))}>
                              Use as Logo
                            </ElevareButton>
                          ) : null}
                          {asset.kind === "banner" ? (
                            <ElevareButton
                              type="button"
                              variant="secondary"
                              onClick={() => setContent((p) => ({ ...p, hero: { ...p.hero, banner_image_url: asset.url } }))}
                            >
                              Use as Banner
                            </ElevareButton>
                          ) : null}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}
