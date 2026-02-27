import Section from "./Section";

const painPoints = [
  "Your institution is made secondary to the platform brand.",
  "Bidder relationships become difficult to retain and govern directly.",
  "Fee structures are constrained by a third party rather than your own policy.",
  "Operational data is fragmented across systems you do not fully control."
];

export default function ProblemSection() {
  return (
    <Section background="white">
      <div className="mx-auto max-w-3xl text-center">
        <p className="text-xs uppercase tracking-[0.16em] text-accent-600">The Common Constraint</p>
        <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
          Most online auction platforms assume control that should remain with the institution.
        </h2>
        <ul className="mt-8 space-y-4 text-left text-lg leading-8 text-brand-700">
          {painPoints.map((point) => (
            <li key={point} className="border-b border-neutral-200 pb-4 last:border-b-0 last:pb-0">
              {point}
            </li>
          ))}
        </ul>
        <p className="mt-8 text-xl font-medium text-brand-800">Elevare restores full ownership.</p>
      </div>
    </Section>
  );
}
