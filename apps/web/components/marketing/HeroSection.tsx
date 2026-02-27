import Link from "next/link";

export default function HeroSection() {
  return (
    <section className="bg-neutral-50 py-28">
      <div className="mx-auto max-w-7xl px-6">
        <div className="max-w-5xl">
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">White-Label Auction Infrastructure</p>
          <h1 className="mt-5 font-serif text-5xl font-semibold leading-[1.1] tracking-[-0.02em] text-brand-900 md:text-7xl">
            Operate fully branded online auctions under your own domain, governance, and commercial model.
          </h1>
          <p className="mt-7 max-w-3xl text-lg leading-8 text-brand-700 md:text-xl">
            Elevare provides the public-facing experience and operational controls required to run institutional online
            auctions with confidence, without handing over the bidder relationship or brand presentation.
          </p>
          <div className="mt-10 flex flex-wrap gap-4">
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
          <div className="mt-10 h-px w-full max-w-4xl bg-neutral-200" />
          <p className="mt-5 text-sm text-brand-700">
            Institutional-grade controls · Transparent fee configuration · Real-time bidding engine
          </p>
        </div>
      </div>
    </section>
  );
}
