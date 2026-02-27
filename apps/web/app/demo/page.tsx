import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export const metadata: Metadata = {
  title: "Arrange a Demonstration — Elevare Auctions",
  description:
    "Arrange a guided review of Elevare's white-label auction platform for museums, galleries, and specialist institutions.",
  alternates: {
    canonical: "/demo"
  },
  openGraph: {
    title: "Arrange a Demonstration — Elevare Auctions",
    description:
      "Arrange a guided review of Elevare's white-label auction platform for museums, galleries, and specialist institutions.",
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
          <h1 className="mt-4 font-serif text-4xl font-semibold leading-[1.1] tracking-[-0.02em] text-neutral-900 md:text-5xl">
            Arrange a Demonstration
          </h1>
          <p className="mt-6 max-w-3xl text-lg leading-8 text-neutral-600">
            We will review branded auction setup, bidder governance, live bidding controls, and settlement workflows
            tailored to institutional operations.
          </p>
          <div className="mt-8 rounded-xl border border-neutral-200 bg-neutral-50 p-6">
            <p className="text-neutral-700">
              Email us at{" "}
              <a className="text-neutral-900 underline" href="mailto:demo@elevareauctions.com">
                demo@elevareauctions.com
              </a>{" "}
              to arrange a suitable session.
            </p>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
