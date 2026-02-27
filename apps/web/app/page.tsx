import type { Metadata } from "next";
import CapabilitiesSection from "../components/marketing/CapabilitiesSection";
import ClosingCTASection from "../components/marketing/ClosingCTASection";
import Footer from "../components/marketing/Footer";
import HeroSection from "../components/marketing/HeroSection";
import InstitutionPreviewSection from "../components/marketing/InstitutionPreviewSection";
import InstitutionalControlsSection from "../components/marketing/InstitutionalControlsSection";
import Navbar from "../components/marketing/Navbar";
import ProblemSection from "../components/marketing/ProblemSection";
import UpcomingAuctionTeaser from "../components/marketing/UpcomingAuctionTeaser";

export const metadata: Metadata = {
  title: "Elevare Auctions — White-Label Online Auction Platform",
  description:
    "Operate fully branded online auctions under your own domain, with institutional-grade controls, real-time bidding, and configurable fees.",
  alternates: {
    canonical: "/"
  },
  openGraph: {
    title: "Elevare Auctions — White-Label Online Auction Platform",
    description:
      "Operate fully branded online auctions under your own domain, with institutional-grade controls, real-time bidding, and configurable fees.",
    type: "website",
    url: "https://elevareauctions.com",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function HomePage() {
  return (
    <div className="bg-white text-brand-900">
      <Navbar />
      <main>
        <HeroSection />
        <ProblemSection />
        <CapabilitiesSection />
        <InstitutionalControlsSection />
        <InstitutionPreviewSection />
        <UpcomingAuctionTeaser />
        <ClosingCTASection />
      </main>
      <Footer />
    </div>
  );
}
