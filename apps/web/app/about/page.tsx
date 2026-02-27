import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";
import Link from "next/link";

const audiences = [
  {
    title: "Museums",
    body: "Present fundraising and collection-led auctions under your own institutional identity, with governance that matches established internal controls."
  },
  {
    title: "Galleries",
    body: "Operate private and public sales with client-facing polish while retaining direct ownership of bidder relationships, catalogues, and reporting."
  },
  {
    title: "Collections & Estates",
    body: "Run time-sensitive disposals with a structured digital process, clear oversight, and an audit trail suitable for trustees, advisors, and stakeholders."
  }
];

const principles = [
  "Your institution remains the primary brand across every bidder touchpoint.",
  "Access controls support manual approval, suspension, and role-based oversight.",
  "Fee logic, settlement views, and exports remain aligned to your operating model.",
  "Activity is recorded in a defensible timeline for operational review and governance."
];

const governancePoints = [
  {
    title: "Role-Based Permissions",
    body: "Separate responsibilities across administrators, finance teams, and operational staff without exposing critical controls more broadly than necessary."
  },
  {
    title: "Bidder Approval",
    body: "Screen participants before they are permitted to bid, supporting higher-trust events and institution-specific risk policies."
  },
  {
    title: "Audit Readiness",
    body: "Track platform actions and bidder-facing events through a clear operational history suitable for internal review."
  },
  {
    title: "Configurable Commercial Rules",
    body: "Apply buyer premiums, fee components, and settlement workflows that reflect the institution's own financial framework."
  }
];

const outcomes = [
  "Reduced dependence on marketplace-led commercial terms",
  "Direct ownership of bidder data and institutional relationships",
  "Clearer reporting for finance, leadership, and compliance review"
];

export const metadata: Metadata = {
  title: "Institution Profile | Elevare Auctions",
  description:
    "Learn how Elevare supports institutional auction operations with white-label delivery, governance controls, and audit-conscious workflows.",
  alternates: {
    canonical: "/about"
  },
  openGraph: {
    title: "Institution Profile | Elevare Auctions",
    description:
      "Learn how Elevare supports institutional auction operations with white-label delivery, governance controls, and audit-conscious workflows.",
    url: "https://elevareauctions.com/about",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function AboutPage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Institution Profile</p>
          <div className="mt-4 grid gap-10 lg:grid-cols-[minmax(0,1.7fr)_minmax(280px,1fr)] lg:items-end">
            <div>
              <h1 className="text-4xl font-semibold leading-tight text-brand-900 md:text-5xl">
                Auction infrastructure designed for institutions that need control, not marketplace dependence.
              </h1>
              <p className="mt-6 max-w-3xl text-lg leading-8 text-brand-700">
                Elevare is built for organisations that require formal governance, operational clarity, and a bidder
                experience that remains fully under institutional ownership. The platform supports branded online
                auctions without forcing institutions into generic marketplace models or fixed commercial rules.
              </p>
            </div>
            <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-6">
              <p className="text-sm font-medium uppercase tracking-[0.12em] text-brand-600">Core Position</p>
              <p className="mt-4 text-2xl font-semibold text-brand-900">White-label, audit-conscious, institution-led.</p>
              <p className="mt-4 text-sm leading-7 text-brand-700">
                Elevare keeps branding, bidder ownership, and operating rules with the institution rather than the
                software provider.
              </p>
            </div>
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="text-3xl font-semibold text-brand-900 md:text-4xl">Who We Serve</h2>
          <p className="mt-4 max-w-3xl text-lg text-brand-700">
            The platform is intended for organisations that need public confidence, disciplined workflows, and a
            presentation standard aligned with their own reputation.
          </p>
          <div className="mt-10 grid gap-6 md:grid-cols-3">
            {audiences.map((audience) => (
              <article key={audience.title} className="rounded-xl border border-neutral-200 bg-white p-8">
                <h3 className="text-2xl font-semibold text-brand-900">{audience.title}</h3>
                <p className="mt-4 leading-7 text-brand-700">{audience.body}</p>
              </article>
            ))}
          </div>
        </Section>

        <Section>
          <div className="grid gap-12 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
            <div>
              <h2 className="text-3xl font-semibold text-brand-900 md:text-4xl">Operating Principles</h2>
              <p className="mt-4 max-w-2xl text-lg text-brand-700">
                Elevare is structured around four expectations common to serious institutional operations.
              </p>
            </div>
            <ul className="space-y-4 text-lg text-brand-700">
              {principles.map((principle) => (
                <li key={principle} className="rounded-xl border border-neutral-200 bg-white px-6 py-5">
                  {principle}
                </li>
              ))}
            </ul>
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="text-3xl font-semibold text-brand-900 md:text-4xl">Governance Overview</h2>
          <p className="mt-4 max-w-3xl text-lg text-brand-700">
            Institutions using Elevare can maintain formal controls across the bidder journey, catalogue operations,
            and financial administration.
          </p>
          <div className="mt-10 grid gap-6 md:grid-cols-2">
            {governancePoints.map((point) => (
              <article key={point.title} className="rounded-xl border border-neutral-200 bg-white p-8">
                <h3 className="text-2xl font-semibold text-brand-900">{point.title}</h3>
                <p className="mt-4 leading-7 text-brand-700">{point.body}</p>
              </article>
            ))}
          </div>
        </Section>

        <Section>
          <div className="grid gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:items-center">
            <div>
              <h2 className="text-3xl font-semibold text-brand-900 md:text-4xl">Institutional Outcomes</h2>
              <p className="mt-4 max-w-2xl text-lg text-brand-700">
                The intended result is not merely a cleaner interface. It is a more defensible operating model for
                digital auctions.
              </p>
            </div>
            <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-8">
              <ul className="space-y-4 text-lg text-brand-800">
                {outcomes.map((outcome) => (
                  <li key={outcome}>{outcome}</li>
                ))}
              </ul>
            </div>
          </div>
        </Section>

        <Section background="neutral">
          <h2 className="text-3xl font-semibold text-brand-900 md:text-4xl">Discuss Your Institutional Requirements</h2>
          <p className="mt-4 max-w-3xl text-lg text-brand-700">
            If your organisation needs branded auction delivery with stronger control over approvals, operations, and
            financial handling, we can walk through the platform in detail.
          </p>
          <div className="mt-8 flex flex-wrap gap-4">
            <Link
              href="/demo"
              className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-accent-600"
            >
              Arrange a Demonstration
            </Link>
            <Link
              href="/contact"
              className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 transition-colors hover:bg-brand-900 hover:text-white"
            >
              Contact Us
            </Link>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
