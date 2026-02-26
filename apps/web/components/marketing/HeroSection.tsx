import Image from "next/image";
import Link from "next/link";

export default function HeroSection() {
  return (
    <section className="bg-neutral-50 py-24">
      <div className="max-w-7xl mx-auto px-6 grid gap-14 items-center lg:grid-cols-2">
        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-accent-600">White-Label Auction Infrastructure</p>
          <h1 className="mt-4 text-4xl md:text-5xl font-semibold leading-tight text-brand-900">
            Run Fully Branded Online Auctions — Without Marketplace Fees.
          </h1>
          <p className="mt-6 text-lg text-brand-700 max-w-2xl">
            Elevare enables museums, galleries, and collectors to run secure, real-time auctions under their own
            domain and brand.
          </p>
          <div className="mt-8 flex flex-wrap gap-4">
            <Link
              href="/demo"
              className="inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white hover:bg-accent-600 transition-colors"
            >
              Request a Demo
            </Link>
            <Link
              href="/auctions"
              className="inline-flex items-center rounded-md border border-brand-900 px-6 py-3 text-sm font-medium text-brand-900 hover:bg-brand-900 hover:text-white transition-colors"
            >
              View Demo Auction
            </Link>
          </div>
        </div>

        <div className="rounded-2xl border border-neutral-200 bg-white p-4 shadow-sm">
          <Image
            src="/marketing-demo-placeholder.svg"
            alt="Elevare branded auction dashboard preview"
            width={1280}
            height={820}
            className="h-auto w-full rounded-xl"
            priority
          />
        </div>
      </div>
    </section>
  );
}
