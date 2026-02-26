import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export const metadata: Metadata = {
  title: "Pricing Philosophy — Elevare Auctions",
  description:
    "Explore Elevare's institutional pricing philosophy for white-label online auction infrastructure.",
  alternates: {
    canonical: "/pricing"
  },
  openGraph: {
    title: "Pricing Philosophy — Elevare Auctions",
    description:
      "Explore Elevare's institutional pricing philosophy for white-label online auction infrastructure.",
    url: "https://elevareauctions.com/pricing",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function PricingPage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Pricing Philosophy</p>
          <h1 className="mt-4 text-4xl md:text-5xl font-semibold text-brand-900">Transparent, Institution-Aligned Pricing</h1>
          <p className="mt-6 text-lg text-brand-700 max-w-3xl">
            Elevare is built as white-label infrastructure, not a marketplace tax. Pricing is aligned with your
            operational model, auction cadence, and governance requirements.
          </p>
          <ul className="mt-8 space-y-4 text-lg text-brand-700">
            <li>No forced marketplace branding</li>
            <li>Configurable commercial structure</li>
            <li>Support for institutional accounting and controls</li>
          </ul>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
