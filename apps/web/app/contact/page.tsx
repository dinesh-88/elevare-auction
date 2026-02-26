import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export const metadata: Metadata = {
  title: "Contact — Elevare Auctions",
  description: "Contact Elevare to discuss demos, implementation planning, and institutional onboarding.",
  alternates: {
    canonical: "/contact"
  },
  openGraph: {
    title: "Contact — Elevare Auctions",
    description: "Contact Elevare to discuss demos, implementation planning, and institutional onboarding.",
    url: "https://elevareauctions.com/contact",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function ContactPage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Contact</p>
          <h1 className="mt-4 text-4xl md:text-5xl font-semibold text-brand-900">Speak With the Elevare Team</h1>
          <p className="mt-6 text-lg text-brand-700 max-w-3xl">
            For product discussions, implementation planning, or institutional onboarding, contact us directly.
          </p>
          <div className="mt-8 rounded-xl border border-neutral-200 bg-neutral-50 p-6 space-y-3 text-brand-700">
            <p>Email: <a className="text-brand-900 underline decoration-accent-500 underline-offset-2" href="mailto:contact@elevareauctions.com">contact@elevareauctions.com</a></p>
            <p>Website: <a className="text-brand-900 underline decoration-accent-500 underline-offset-2" href="https://elevareauctions.com">elevareauctions.com</a></p>
          </div>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
