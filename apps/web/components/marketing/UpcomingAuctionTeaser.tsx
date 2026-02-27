import Link from "next/link";
import { normalizeFeaturedAuction } from "../../src/api/client";
import { fetchPublicFeaturedAuction } from "../../src/server/api";
import { buildAuctionSlug } from "../../src/utils/auctionRoutes";
import Countdown from "./Countdown";
import Section from "./Section";

export default async function UpcomingAuctionTeaser() {
  let featured = null;
  try {
    featured = normalizeFeaturedAuction(await fetchPublicFeaturedAuction());
  } catch {
    return null;
  }

  const auctionSlug = buildAuctionSlug({ id: featured.auction.id, title: featured.auction.title });

  return (
    <Section background="white">
      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(320px,0.8fr)] lg:items-start">
        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Upcoming Auction</p>
          <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
            {featured.auction.title}
          </h2>
          <p className="mt-5 text-lg leading-8 text-brand-700">
            Presented by{" "}
            <Link href={`/institutions/${featured.institution.slug}`} className="text-brand-900 hover:text-accent-600">
              {featured.institution.name}
            </Link>
          </p>
          <p className="mt-3 text-base text-brand-700">
            Opens{" "}
            {new Intl.DateTimeFormat("en-GB", { dateStyle: "full", timeStyle: "short" }).format(
              new Date(featured.auction.start_time)
            )}
          </p>
        </div>
        <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-8">
          <p className="text-xs uppercase tracking-[0.12em] text-brand-600">Countdown</p>
          <div className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900">
            <Countdown startTime={featured.auction.start_time} status={featured.auction.status} />
          </div>
          <div className="mt-8">
            <Link
              href={`/auction/${auctionSlug}/preview`}
              className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white hover:bg-accent-600"
            >
              View Preview
            </Link>
          </div>
        </div>
      </div>
    </Section>
  );
}
