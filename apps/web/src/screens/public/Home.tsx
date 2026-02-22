import { Link } from "react-router-dom";
import { useTheme } from "../../theme/ThemeProvider";

type HomepageContent = {
  hero?: {
    headline?: string;
    subheadline?: string;
    banner_image_url?: string;
    cta?: { label?: string; href?: string };
  };
  highlights?: Array<{ title?: string; body?: string; icon?: string }>;
};

type SiteContent = {
  homepage?: HomepageContent;
  about?: { markdown?: string };
  contact?: { address?: string; phone?: string; email?: string };
};

export default function Home() {
  const { config } = useTheme();
  const content = (config?.content ?? {}) as SiteContent;
  const hero = content.homepage?.hero;
  const highlights = content.homepage?.highlights ?? [];
  const aboutMarkdown = content.about?.markdown ?? "";
  const contact = content.contact;

  return (
    <section className="min-h-screen bg-white">
      <section className="mx-auto max-w-[1440px] px-4 py-12 md:px-8 md:py-[120px]">
        <div className="grid grid-cols-1 items-center gap-8 md:gap-16 lg:grid-cols-5">
          <div className="space-y-6 md:space-y-8 lg:col-span-3">
            <h1 className="font-heading text-4xl leading-tight sm:text-[42px] md:text-[56px]">
              {hero?.headline || config?.name || "Elevare"}
            </h1>
            <p className="max-w-xl text-base text-neutral-500 sm:text-lg md:text-xl">
              {hero?.subheadline || "Auction Infrastructure for Institutions"}
            </p>
            <div className="flex flex-col gap-4 sm:flex-row">
              <Link
                to={hero?.cta?.href || "/auctions"}
                className="inline-flex rounded-lg bg-brand-primary px-8 py-3 font-medium text-white hover:opacity-90"
              >
                {hero?.cta?.label || "View Auctions"}
              </Link>
              <Link to="/auctions" className="inline-flex rounded-lg border border-neutral-200 bg-neutral-50 px-8 py-3 font-medium text-neutral-900 hover:bg-neutral-200">
                Browse Lots
              </Link>
            </div>
          </div>
          <div className="lg:col-span-2">
            <div className="aspect-square overflow-hidden rounded-2xl shadow-sm">
              {hero?.banner_image_url || config?.branding?.hero_image ? (
                <img src={hero?.banner_image_url || config?.branding?.hero_image} alt="Hero" className="h-full w-full object-cover" />
              ) : (
                <div className="h-full w-full bg-neutral-100" />
              )}
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-[1440px] border-t border-neutral-200 px-4 py-12 md:px-8 md:py-24">
        <div className="grid grid-cols-1 gap-12 md:grid-cols-3 md:gap-16">
          {(highlights.length > 0
            ? highlights.slice(0, 6)
            : [
                { title: "Multi-Lot Auctions", body: "Run sophisticated multi-lot auctions with flexible bidding structures.", icon: "LOT" },
                { title: "Real-Time Bidding", body: "Live bidding with instant updates for all participants.", icon: "RT" },
                { title: "Institutional Security", body: "Enterprise-level security and compliance for institutional clients.", icon: "SEC" },
              ]
          ).map((item, idx) => (
            <div key={`${item.title || "highlight"}-${idx}`} className="space-y-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-50 text-brand-primary">
                {(item.icon || "•").slice(0, 4).toUpperCase()}
              </div>
              <h3 className="font-heading text-xl">{item.title || "Highlight"}</h3>
              <p className="leading-relaxed text-neutral-500">{item.body || ""}</p>
            </div>
          ))}
        </div>
      </section>

      {aboutMarkdown || contact?.email || contact?.address || contact?.phone ? (
        <section className="mx-auto max-w-[1440px] border-t border-neutral-200 px-4 py-12 md:px-8 md:py-16">
          <div className="grid gap-10 md:grid-cols-2">
            <div className="space-y-3">
              <h2 className="font-heading text-2xl text-neutral-900">About</h2>
              <p className="whitespace-pre-wrap leading-relaxed text-neutral-600">{aboutMarkdown || ""}</p>
            </div>
            <div className="space-y-3">
              <h2 className="font-heading text-2xl text-neutral-900">Contact</h2>
              {contact?.address ? <p className="text-neutral-600">{contact.address}</p> : null}
              {contact?.phone ? <p className="text-neutral-600">{contact.phone}</p> : null}
              {contact?.email ? (
                <p>
                  <a href={`mailto:${contact.email}`} className="text-brand-primary hover:underline">
                    {contact.email}
                  </a>
                </p>
              ) : null}
            </div>
          </div>
        </section>
      ) : null}
    </section>
  );
}
