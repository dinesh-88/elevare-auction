import Link from "next/link";
import Section from "./Section";

export default function ClosingCTASection() {
  return (
    <Section background="neutral">
      <div className="mx-auto max-w-4xl text-center">
        <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Next Step</p>
        <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
          Arrange a detailed review of your institution&apos;s auction requirements.
        </h2>
        <p className="mx-auto mt-5 max-w-3xl text-lg leading-8 text-brand-700">
          We will walk through brand control, bidder governance, fee configuration, and the public experience your
          team wishes to operate under its own domain.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-4">
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
      </div>
    </Section>
  );
}
