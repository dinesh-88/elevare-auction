import type { Metadata } from "next";
import ContactRequestForm from "../../components/marketing/ContactRequestForm";
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
        <Section background="neutral" className="text-center">
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Contact</p>
          <h1 className="mt-4 text-4xl md:text-5xl font-semibold text-brand-900">Arrange a Demonstration</h1>
          <p className="mx-auto mt-6 max-w-3xl text-lg text-brand-700">
            If your institution would value a formal review of the platform, please send your details below and we
            shall arrange a suitable time to discuss implementation, governance, and operational requirements.
          </p>
          <ContactRequestForm />
        </Section>
      </main>
      <Footer />
    </div>
  );
}
