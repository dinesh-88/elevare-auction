import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export const metadata: Metadata = {
  title: "Request a Demo — Elevare Auctions",
  description:
    "Book a guided demo of Elevare's white-label auction platform for museums, galleries, and collectors.",
  alternates: {
    canonical: "/demo"
  },
  openGraph: {
    title: "Request a Demo — Elevare Auctions",
    description:
      "Book a guided demo of Elevare's white-label auction platform for museums, galleries, and collectors.",
    url: "https://elevareauctions.com/demo",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function DemoPage() {
  return (
    <div className="bg-white text-neutral-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-neutral-500">Demo</p>
          <h1 className="mt-4 text-4xl md:text-5xl font-semibold text-neutral-900">Request a Guided Elevare Demo</h1>
          <p className="mt-6 text-lg text-neutral-600 max-w-3xl">
            We will walk your team through branded auction setup, bidder management, live bidding controls, and
            settlement workflows tailored to institutional operations.
          </p>
          <div className="mt-8 rounded-xl border border-neutral-200 bg-neutral-50 p-6">
            <p className="text-neutral-700">Email us at <a className="text-neutral-900 underline" href="mailto:demo@elevareauctions.com">demo@elevareauctions.com</a> to schedule a session.</p>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
