import FeatureGrid from "./FeatureGrid";
import Section from "./Section";

export default function CapabilitiesSection() {
  return (
    <Section background="neutral">
      <div className="max-w-3xl">
        <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Platform Capabilities</p>
        <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
          Institutional tooling for catalogues, bidding, and settlement.
        </h2>
        <p className="mt-5 text-lg leading-8 text-brand-700">
          Elevare is designed for organisations that require transparent governance, precise fee logic, and a public
          experience that remains fully under their own control.
        </p>
      </div>
      <div className="mt-10">
        <FeatureGrid />
      </div>
    </Section>
  );
}
