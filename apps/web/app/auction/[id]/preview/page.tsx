import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import Footer from "../../../../components/marketing/Footer";
import Countdown from "../../../../components/marketing/Countdown";
import Navbar from "../../../../components/marketing/Navbar";
import Section from "../../../../components/marketing/Section";
import { normalizeAuction, normalizeLot, type Auction, type Lot } from "../../../../src/api/client";
import { fetchAuction, fetchLotsByAuction, fetchPublicAuctions } from "../../../../src/server/api";
import { extractAuctionIdFromSlug, slugifyAuctionTitle } from "../../../../src/utils/auctionRoutes";

type Props = {
  params: Promise<{ id: string }>;
};

async function resolveAuction(slug: string): Promise<Auction | null> {
  const id = extractAuctionIdFromSlug(slug);
  if (id) {
    try {
      return normalizeAuction(await fetchAuction(id));
    } catch {
      return null;
    }
  }

  try {
    const auctions = (await fetchPublicAuctions()).map(normalizeAuction);
    return auctions.find((item) => slugifyAuctionTitle(item.title) === slug) ?? null;
  } catch {
    return null;
  }
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const auction = await resolveAuction(id);
  if (!auction) {
    return {
      title: "Auction Preview | Elevare Auctions"
    };
  }
  return {
    title: `${auction.title} — Preview`,
    description: auction.description,
    alternates: {
      canonical: `/auction/${id}/preview`
    },
    openGraph: {
      title: `${auction.title} — Preview`,
      description: auction.description,
      url: `https://elevareauctions.com/auction/${id}/preview`,
      images: ["/og/elevare-marketing-og.svg"]
    }
  };
}

export default async function AuctionPreviewPage({ params }: Props) {
  const { id } = await params;
  const auction = await resolveAuction(id);
  if (!auction) notFound();

  if (auction.status !== "SCHEDULED" && auction.status !== "DRAFT") {
    redirect(`/auction/${auction.id}`);
  }

  let lots: Lot[] = [];
  try {
    lots = (await fetchLotsByAuction(auction.id)).map(normalizeLot).slice(0, 6);
  } catch {
    lots = [];
  }

  const image = auction.metadata?.image ?? auction.metadata?.images?.[0];
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <div className="grid gap-10 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)] lg:items-center">
            <div>
              <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Upcoming Auction</p>
              <h1 className="mt-4 text-4xl font-semibold leading-tight md:text-5xl">{auction.title}</h1>
              <p className="mt-6 max-w-3xl text-lg leading-8 text-brand-700">{auction.description}</p>
              <div className="mt-8 grid gap-4 rounded-2xl border border-neutral-200 bg-neutral-50 p-6 sm:grid-cols-2">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.12em] text-brand-600">Opening Date</p>
                  <p className="mt-2 text-lg font-medium text-brand-900">
                    {new Intl.DateTimeFormat("en-GB", {
                      dateStyle: "full",
                      timeStyle: "short",
                    }).format(new Date(auction.start_time))}
                  </p>
                </div>
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.12em] text-brand-600">Preview Access</p>
                  <p className="mt-2 text-lg text-brand-800">Catalogue available before bidding opens</p>
                </div>
              </div>
              <div className="mt-6">
                <div className="font-serif text-3xl font-semibold tracking-tight text-brand-900">
                  <Countdown startTime={auction.start_time} status={auction.status} />
                </div>
              </div>
            </div>

            <div className="overflow-hidden rounded-2xl border border-neutral-200 bg-neutral-50">
              {image ? (
                <img src={image} alt={auction.title} className="h-full w-full object-cover" />
              ) : (
                <div className="grid aspect-[4/3] place-content-center text-sm text-neutral-500">Preview image coming soon</div>
              )}
            </div>
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="text-3xl font-semibold md:text-4xl">Auction Description</h2>
          <p className="mt-6 max-w-4xl text-lg leading-8 text-brand-700">{auction.description}</p>
        </Section>

        <Section>
          <div className="flex items-end justify-between gap-6">
            <div>
              <h2 className="text-3xl font-semibold md:text-4xl">Catalogue Preview</h2>
              <p className="mt-4 text-lg text-brand-700">A first look at selected lots scheduled for this sale.</p>
            </div>
          </div>
          <div className="mt-10 grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            {lots.map((lot) => {
              const lotImage = lot.metadata?.image ?? lot.metadata?.images?.[0];
              return (
                <article key={lot.id} className="overflow-hidden rounded-xl border border-neutral-200 bg-white">
                  <div className="aspect-[4/3] bg-neutral-100">
                    {lotImage ? (
                      <img src={lotImage} alt={lot.title} className="h-full w-full object-cover" />
                    ) : (
                      <div className="grid h-full place-content-center text-sm text-neutral-500">No image</div>
                    )}
                  </div>
                  <div className="p-6">
                    <p className="text-xs uppercase tracking-[0.12em] text-brand-600">Lot {lot.lot_number}</p>
                    <h3 className="mt-3 text-xl font-semibold text-brand-900">{lot.title}</h3>
                  </div>
                </article>
              );
            })}
            {lots.length === 0 ? <p className="text-sm text-neutral-500">Catalogue preview will be published shortly.</p> : null}
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="text-3xl font-semibold md:text-4xl">Register to Bid</h2>
          <p className="mt-4 max-w-3xl text-lg text-brand-700">
            Registration may be completed in advance so approved bidders are ready when the auction opens.
          </p>
          <div className="mt-8 flex flex-wrap gap-4">
            <Link href="/register" className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white hover:bg-accent-600">
              Register to Bid
            </Link>
            <Link href={`/auction/${auction.id}`} className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 hover:bg-brand-900 hover:text-white">
              View Main Auction Page
            </Link>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
