import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export default function AboutPage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <h1 className="text-4xl md:text-5xl font-semibold text-brand-900">About Elevare</h1>
          <p className="mt-6 text-lg text-brand-700 max-w-3xl">
            Elevare provides white-label auction technology for institutions that require brand control, operational
            rigor, and trusted financial workflows.
          </p>
        </Section>
      </main>
      <Footer />
    </div>
  );
}
