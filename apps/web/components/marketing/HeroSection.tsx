import Link from "next/link";

export default function HeroSection() {
  return (
    <section className="bg-neutral-50 pb-28 pt-40">
      <div className="mx-auto max-w-7xl px-6">
        <div className="max-w-5xl">
          <div className="inline-flex items-center gap-3 rounded-full border border-neutral-200 bg-white px-3 py-2">
            <img src="/brand/elevare-monogram.svg" alt="" className="h-8 w-8 rounded-sm" />
            <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-accent-600">
              White-Label Auction Infrastructure
            </p>
          </div>
          <h1 className="mt-7 max-w-5xl font-serif text-5xl font-semibold leading-[1.04] tracking-[-0.025em] text-brand-900 md:text-7xl">
            Operate fully branded online auctions under your own domain, governance, and commercial model.
          </h1>
          <p className="mt-8 max-w-3xl text-lg leading-8 text-brand-700 md:text-xl">
            Elevare provides the public-facing presentation and institutional controls required to run serious online
            auctions without handing over bidder ownership, commercial rules, or brand trust.
          </p>
          <div className="mt-11 flex flex-wrap gap-4">
            <Link
              href="/contact"
              className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white hover:bg-accent-600"
            >
              Arrange a Demonstration
            </Link>
            <a
              href="https://demo.elevareauctions.com"
              className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 hover:bg-brand-900 hover:text-white"
            >
              View Live Demo
            </a>
          </div>
          <p className="mt-10 text-sm text-brand-700">
            Institutional-grade controls · Transparent fee configuration · Real-time bidding engine
          </p>
        </div>
      </div>
    </section>
  );
}
