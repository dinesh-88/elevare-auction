import Link from "next/link";
import type { Metadata } from "next";
import Footer from "../components/marketing/Footer";
import HeroSection from "../components/marketing/HeroSection";
import FeatureGrid from "../components/marketing/FeatureGrid";
import Navbar from "../components/marketing/Navbar";
import Section from "../components/marketing/Section";

export const metadata: Metadata = {
  alternates: {
    canonical: "/"
  }
};

export default function HomePage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <HeroSection />

        <Section>
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Most Online Auction Platforms Take Control.</h2>
          <ul className="mt-8 space-y-4 text-lg text-brand-700">
            <li>Your brand becomes secondary.</li>
            <li>Your bidders belong to the platform.</li>
            <li>Your fees are fixed.</li>
            <li>Your data is not fully yours.</li>
          </ul>
          <p className="mt-8 text-xl text-brand-800 font-medium">Elevare gives you full ownership.</p>
        </Section>

        <Section background="neutral">
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Core Capabilities</h2>
          <p className="mt-4 text-lg text-brand-700">Everything required to operate high-trust online auctions under your own institution.</p>
          <div className="mt-10">
            <FeatureGrid />
          </div>
        </Section>

        <Section>
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Built for Institutional Control</h2>
          <ul className="mt-8 space-y-4 text-lg text-brand-700">
            <li>Bidder approval workflows</li>
            <li>Risk controls and suspension</li>
            <li>Full audit timeline</li>
            <li>Role-based permissions</li>
          </ul>
        </Section>

        <Section background="neutral">
          <h2 className="font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">Ready to Run Your Own Branded Auction?</h2>
          <div className="mt-8 flex flex-wrap gap-4">
            <Link
              href="/demo"
              className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white hover:bg-accent-600 transition-colors"
            >
              Book a Demo
            </Link>
            <Link
              href="/contact"
              className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 hover:bg-brand-900 hover:text-white transition-colors"
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
