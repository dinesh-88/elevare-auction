import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export default function PrivacyPage() {
  return (
    <div className="bg-white text-neutral-900">
      <Navbar />
      <main>
        <Section>
          <h1 className="text-4xl md:text-5xl font-semibold text-neutral-900">Privacy</h1>
          <p className="mt-6 text-lg text-neutral-600 max-w-3xl">
            Elevare is committed to protecting institutional and bidder data. Full policy text can be published here as
            legal copy is finalized.
          </p>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
