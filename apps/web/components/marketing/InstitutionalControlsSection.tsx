import Section from "./Section";

const controls = [
  "Bidder approval workflows and permissioned access to live events",
  "Risk controls for account review, suspension, and audit visibility",
  "Real-time operational oversight with bidder, lot, and activity records",
  "Role-based administration for governance, finance, and catalogue teams"
];

export default function InstitutionalControlsSection() {
  return (
    <Section background="white">
      <div className="grid gap-10 lg:grid-cols-[minmax(0,1fr)_minmax(320px,0.9fr)]">
        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Institutional Control</p>
          <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
            Operational controls designed for accountable auction practice.
          </h2>
          <p className="mt-5 max-w-2xl text-lg leading-8 text-brand-700">
            Governance should not sit outside the auction platform. Elevare keeps approvals, permissions, fee policy,
            and audit readiness within a single institution-managed environment.
          </p>
        </div>
        <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-8">
          <ul className="space-y-4 text-base leading-7 text-brand-700">
            {controls.map((control) => (
              <li key={control} className="border-b border-neutral-200 pb-4 last:border-b-0 last:pb-0">
                {control}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </Section>
  );
}
