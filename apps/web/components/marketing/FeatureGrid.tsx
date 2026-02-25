const capabilities = [
  {
    title: "Branded Auctions",
    points: ["Custom domain support", "White-label interface", "Institution-controlled content"]
  },
  {
    title: "Real-Time Engine",
    points: ["Proxy bidding", "Anti-sniping soft close", "Live updates"]
  },
  {
    title: "Financial Control",
    points: ["Configurable fee components", "Buyer premium logic", "Settlement dashboard", "Exportable reporting"]
  }
];

export default function FeatureGrid() {
  return (
    <div className="grid gap-6 md:grid-cols-3">
      {capabilities.map((capability) => (
        <article key={capability.title} className="rounded-xl border border-neutral-200 bg-white p-8">
          <h3 className="text-2xl font-semibold text-neutral-900">{capability.title}</h3>
          <ul className="mt-5 space-y-3">
            {capability.points.map((point) => (
              <li key={point} className="text-base text-neutral-600">
                {point}
              </li>
            ))}
          </ul>
        </article>
      ))}
    </div>
  );
}
