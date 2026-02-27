import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import Footer from "../../../../components/marketing/Footer";
import Navbar from "../../../../components/marketing/Navbar";
import Section from "../../../../components/marketing/Section";
import { normalizeAuction, normalizeLot, type Auction, type Lot } from "../../../../src/api/client";
import { fetchAuction, fetchLotsByAuction } from "../../../../src/server/api";
import { extractAuctionIdFromSlug } from "../../../../src/utils/auctionRoutes";
import { currencySymbol } from "../../../../src/utils/currency";

export const metadata: Metadata = {
  title: "Auction Results | Elevare Auctions",
  description: "Review published auction outcomes, sold lots, and realised totals.",
};

type Props = {
  params: Promise<{ slug: string }>;
};

export default async function AuctionResultsPage({ params }: Props) {
  const { slug } = await params;
  const auctionID = extractAuctionIdFromSlug(slug);
  if (!auctionID) notFound();

  let auction: Auction | null = null;
  let lots: Lot[] = [];
  try {
    auction = normalizeAuction(await fetchAuction(auctionID));
    lots = (await fetchLotsByAuction(auctionID)).map(normalizeLot);
  } catch {
    notFound();
  }

  if (!auction || auction.status !== "ENDED") {
    notFound();
  }

  const soldLots = lots.filter((lot) => lot.sale_result === "SOLD");
  const totalRealized = soldLots.reduce((sum, lot) => sum + (lot.hammer_price ?? 0), 0);
  const symbol = currencySymbol(auction.currency);

  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Published Results</p>
          <h1 className="mt-4 text-4xl font-semibold md:text-5xl">{auction.title}</h1>
          <p className="mt-6 text-lg text-brand-700">
            Completed on{" "}
            {new Intl.DateTimeFormat("en-GB", {
              dateStyle: "full",
              timeStyle: "short",
            }).format(new Date(auction.end_time))}
          </p>
        </Section>

        <Section background="neutral">
          <div className="overflow-x-auto rounded-2xl border border-neutral-200 bg-white">
            <div className="min-w-[720px]">
              <div className="grid grid-cols-[1fr_1.2fr_1fr_1fr] gap-4 border-b border-neutral-200 px-6 py-4 text-xs font-semibold uppercase tracking-[0.12em] text-brand-600">
                <div>Lot</div>
                <div>Title</div>
                <div>Final Price</div>
                <div>Status</div>
              </div>
              {lots.map((lot) => (
                <div key={lot.id} className="grid grid-cols-[1fr_1.2fr_1fr_1fr] gap-4 border-b border-neutral-100 px-6 py-5 text-sm last:border-b-0">
                  <div className="font-medium text-brand-900">{lot.lot_number}</div>
                  <div className="text-brand-800">{lot.title}</div>
                  <div className="text-brand-800">
                    {lot.hammer_price != null ? `${symbol}${lot.hammer_price.toLocaleString()}` : `${symbol}0`}
                  </div>
                  <div className={lot.sale_result === "SOLD" ? "text-green-700" : "text-neutral-500"}>
                    {lot.sale_result === "SOLD" ? "Sold" : "Unsold"}
                  </div>
                </div>
              ))}
              {lots.length === 0 ? <div className="px-6 py-5 text-sm text-neutral-500">No lots were published for this auction.</div> : null}
            </div>
          </div>
        </Section>

        <Section>
          <div className="grid gap-6 md:grid-cols-3">
            <div className="rounded-xl border border-neutral-200 bg-neutral-50 p-6">
              <p className="text-xs uppercase tracking-[0.12em] text-brand-600">Total Lots</p>
              <p className="mt-3 text-3xl font-semibold text-brand-900">{lots.length}</p>
            </div>
            <div className="rounded-xl border border-neutral-200 bg-neutral-50 p-6">
              <p className="text-xs uppercase tracking-[0.12em] text-brand-600">Total Sold</p>
              <p className="mt-3 text-3xl font-semibold text-brand-900">{soldLots.length}</p>
            </div>
            <div className="rounded-xl border border-neutral-200 bg-neutral-50 p-6">
              <p className="text-xs uppercase tracking-[0.12em] text-brand-600">Total Realised</p>
              <p className="mt-3 text-3xl font-semibold text-brand-900">
                {symbol}
                {totalRealized.toLocaleString()}
              </p>
            </div>
          </div>
          <div className="mt-8">
            <Link href={`/auction/${auction.id}`} className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 hover:bg-brand-900 hover:text-white">
              View Auction Page
            </Link>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
