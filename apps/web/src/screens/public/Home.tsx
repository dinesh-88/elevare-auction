import { Link } from "react-router-dom";
import { useTheme } from "../../theme/ThemeProvider";

export default function Home() {
  const { config } = useTheme();

  return (
    <section className="min-h-screen bg-white">
      <section className="mx-auto max-w-[1440px] px-4 py-12 md:px-8 md:py-[120px]">
        <div className="grid grid-cols-1 items-center gap-8 md:gap-16 lg:grid-cols-5">
          <div className="space-y-6 md:space-y-8 lg:col-span-3">
            <h1 className="font-heading text-4xl leading-tight sm:text-[42px] md:text-[56px]">{config?.name ?? "Elevare"}</h1>
            <p className="max-w-xl text-base text-neutral-500 sm:text-lg md:text-xl">Auction Infrastructure for Institutions</p>
            <div className="flex flex-col gap-4 sm:flex-row">
              <Link to="/signup" className="inline-flex rounded-lg bg-brand-primary px-8 py-3 font-medium text-white hover:opacity-90">
                Request Access
              </Link>
              <Link to="/auctions" className="inline-flex rounded-lg border border-neutral-200 bg-neutral-50 px-8 py-3 font-medium text-neutral-900 hover:bg-neutral-200">
                View Demo
              </Link>
            </div>
          </div>
          <div className="lg:col-span-2">
            <div className="aspect-square overflow-hidden rounded-2xl shadow-sm">
              {config?.branding?.hero_image ? (
                <img src={config.branding.hero_image} alt="Hero" className="h-full w-full object-cover" />
              ) : (
                <div className="h-full w-full bg-neutral-100" />
              )}
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-[1440px] border-t border-neutral-200 px-4 py-12 md:px-8 md:py-24">
        <div className="grid grid-cols-1 gap-12 md:grid-cols-3 md:gap-16">
          <div className="space-y-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-50 text-brand-primary">LOT</div>
            <h3 className="font-heading text-xl">Multi-Lot Auctions</h3>
            <p className="leading-relaxed text-neutral-500">Run sophisticated multi-lot auctions with advanced lot management and flexible bidding structures.</p>
          </div>
          <div className="space-y-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-50 text-brand-primary">RT</div>
            <h3 className="font-heading text-xl">Real-Time Bidding</h3>
            <p className="leading-relaxed text-neutral-500">Live bidding with instant updates, ensuring seamless experiences for all participants.</p>
          </div>
          <div className="space-y-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-neutral-50 text-brand-primary">SEC</div>
            <h3 className="font-heading text-xl">Institutional-Grade Security</h3>
            <p className="leading-relaxed text-neutral-500">Enterprise-level security and compliance built for institutional requirements.</p>
          </div>
        </div>
      </section>
    </section>
  );
}
