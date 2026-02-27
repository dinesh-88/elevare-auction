import type { Metadata } from "next";
import Footer from "../../components/marketing/Footer";
import Navbar from "../../components/marketing/Navbar";
import Section from "../../components/marketing/Section";

export const metadata: Metadata = {
  title: "How Online Auctions Operate | Elevare",
  description:
    "Learn how proxy bidding, soft close rules, bidder approvals, and governance controls operate within Elevare's auction model.",
  alternates: {
    canonical: "/how-it-works"
  },
  openGraph: {
    title: "How Online Auctions Operate | Elevare",
    description:
      "Learn how proxy bidding, soft close rules, bidder approvals, and governance controls operate within Elevare's auction model.",
    url: "https://elevareauctions.com/how-it-works",
    images: ["/og/elevare-marketing-og.svg"]
  },
  twitter: {
    card: "summary_large_image",
    title: "How Online Auctions Operate | Elevare",
    description:
      "Learn how proxy bidding, soft close rules, bidder approvals, and governance controls operate within Elevare's auction model.",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

const sections = [
  {
    title: "Proxy Bidding",
    body: "Bidders may set a maximum bid in advance. The platform then increases their standing bid automatically only as required, using the configured increment schedule. This allows participants to remain competitive without manually returning to the screen for every movement in price.",
  },
  {
    title: "Soft Close Logic",
    body: "Where anti-sniping protection is enabled, late bids inside the defined extension window move the closing time forward. This preserves fairness by ensuring a lot is not won simply because another bidder had no reasonable opportunity to respond in the final moments.",
  },
  {
    title: "Bidder Approval",
    body: "Institutions may require bidders to be reviewed before bidding rights are granted. This supports controlled participation, allows higher-risk registrations to be assessed manually, and creates a more suitable operating model for regulated or reputation-sensitive sales.",
  },
  {
    title: "Governance & Audit",
    body: "Operational actions can be governed through role-based permissions and recorded in an audit trail. In practice, this means approvals, catalogue changes, and financial handling can be managed with a clearer line of accountability than consumer-grade marketplace tooling usually provides.",
  },
];

export default function HowItWorksPage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <Section>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Process</p>
          <h1 className="mt-4 font-serif text-4xl font-semibold md:text-5xl">How Online Auctions Operate</h1>
          <p className="mt-6 max-w-3xl text-lg leading-8 text-brand-700">
            Elevare is designed to help institutions run timed online auctions with clearer rules, stronger controls,
            and more predictable bidder handling.
          </p>
        </Section>

        {sections.map((section, index) => (
          <Section key={section.title} background={index % 2 === 0 ? "neutral" : "white"}>
            <div className="max-w-4xl">
              <h2 className="font-serif text-3xl font-semibold md:text-4xl">{section.title}</h2>
              <p className="mt-6 text-lg leading-8 text-brand-700">{section.body}</p>
            </div>
          </Section>
        ))}
      </main>
      <Footer />
    </div>
  );
}
