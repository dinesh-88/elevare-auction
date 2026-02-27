import Link from "next/link";
import { notFound } from "next/navigation";
import Footer from "../../../components/marketing/Footer";
import Navbar from "../../../components/marketing/Navbar";
import Section from "../../../components/marketing/Section";
import { normalizeInstitutionProfile } from "../../../src/api/client";
import { fetchPublicInstitution } from "../../../src/server/api";
import { buildAuctionSlug } from "../../../src/utils/auctionRoutes";

type Props = {
  params: Promise<{ slug: string }>;
};

export default async function InstitutionProfilePage({ params }: Props) {
  const { slug } = await params;

  let institution = null;
  try {
    institution = normalizeInstitutionProfile(await fetchPublicInstitution(slug));
  } catch {
    notFound();
  }

  if (!institution) {
    notFound();
  }

  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section background="neutral">
          <div className="grid gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(280px,0.8fr)] lg:items-start">
            <div>
              <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Institution Profile</p>
              <h1 className="mt-4 font-serif text-5xl font-semibold leading-[1.1] tracking-[-0.02em] text-brand-900 md:text-6xl">
                {institution.name}
              </h1>
              <p className="mt-6 max-w-3xl text-lg leading-8 text-brand-700">{institution.description}</p>
            </div>
            <div className="rounded-2xl border border-neutral-200 bg-white p-8">
              {institution.logo_url ? (
                <img src={institution.logo_url} alt={institution.name} className="h-16 w-auto object-contain" />
              ) : null}
              {institution.website_url ? (
                <p className="mt-6 text-sm text-brand-700">
                  <a href={institution.website_url} target="_blank" rel="noreferrer" className="hover:text-accent-600">
                    Visit institution website
                  </a>
                </p>
              ) : null}
              {institution.contact_email ? (
                <p className="mt-3 text-sm text-brand-700">
                  <a href={`mailto:${institution.contact_email}`} className="hover:text-accent-600">
                    {institution.contact_email}
                  </a>
                </p>
              ) : null}
            </div>
          </div>
        </Section>

        <Section>
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Active Auctions</h2>
          <div className="mt-8 grid gap-6 lg:grid-cols-2">
            {institution.active_auctions.map((auction) => {
              const auctionSlug = buildAuctionSlug({ id: auction.id, title: auction.title });
              return (
                <article key={auction.id} className="rounded-2xl border border-neutral-200 bg-neutral-50 p-8">
                  <p className="text-xs uppercase tracking-[0.12em] text-accent-600">{auction.status === "LIVE" ? "Live" : "Upcoming"}</p>
                  <h3 className="mt-3 font-serif text-2xl font-semibold tracking-tight text-brand-900">{auction.title}</h3>
                  <p className="mt-4 text-base leading-7 text-brand-700">
                    {new Intl.DateTimeFormat("en-GB", { dateStyle: "full", timeStyle: "short" }).format(new Date(auction.start_time))}
                  </p>
                  <div className="mt-6">
                    <Link href={`/auction/${auctionSlug}/preview`} className="text-sm font-medium text-brand-900 hover:text-accent-600">
                      View preview
                    </Link>
                  </div>
                </article>
              );
            })}
            {institution.active_auctions.length === 0 ? (
              <p className="text-base text-brand-700">No active auctions are currently listed for this institution.</p>
            ) : null}
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Published Results</h2>
          <div className="mt-8 grid gap-6 lg:grid-cols-2">
            {institution.completed_auctions.map((auction) => {
              const auctionSlug = buildAuctionSlug({ id: auction.id, title: auction.title });
              return (
                <article key={auction.id} className="rounded-2xl border border-neutral-200 bg-white p-8">
                  <p className="text-xs uppercase tracking-[0.12em] text-accent-600">Completed Auction</p>
                  <h3 className="mt-3 font-serif text-2xl font-semibold tracking-tight text-brand-900">{auction.title}</h3>
                  <p className="mt-4 text-base leading-7 text-brand-700">
                    Closed{" "}
                    {new Intl.DateTimeFormat("en-GB", { dateStyle: "full", timeStyle: "short" }).format(new Date(auction.end_time))}
                  </p>
                  <div className="mt-6">
                    <Link href={`/auction/${auctionSlug}/results`} className="text-sm font-medium text-brand-900 hover:text-accent-600">
                      View results
                    </Link>
                  </div>
                </article>
              );
            })}
            {institution.completed_auctions.length === 0 ? (
              <p className="text-base text-brand-700">Published results will appear here once auctions have concluded.</p>
            ) : null}
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
