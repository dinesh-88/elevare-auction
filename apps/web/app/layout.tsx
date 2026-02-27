import type { Metadata } from "next";
import { Cormorant_Garamond, Inter } from "next/font/google";
import AppProviders from "../src/app-providers";
import "./globals.css";

const fontSerif = Cormorant_Garamond({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-serif"
});

const fontSans = Inter({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-sans"
});

export const metadata: Metadata = {
  metadataBase: new URL("https://elevareauctions.com"),
  title: "Elevare Auctions — White-Label Online Auction Platform",
  description:
    "Run fully branded online auctions with real-time bidding, proxy logic, and institutional-grade controls.",
  alternates: {
    canonical: "/"
  },
  openGraph: {
    title: "Elevare Auctions — White-Label Online Auction Platform",
    description:
      "Run fully branded online auctions with real-time bidding, proxy logic, and institutional-grade controls.",
    url: "https://elevareauctions.com",
    siteName: "Elevare Auctions",
    type: "website",
    images: [
      {
        url: "/og/elevare-marketing-og.svg",
        width: 1200,
        height: 630,
        alt: "Elevare Auctions"
      }
    ]
  },
  icons: {
    icon: "/favicon.ico",
    apple: "/apple-touch-icon.png"
  },
  twitter: {
    card: "summary_large_image",
    title: "Elevare Auctions — White-Label Online Auction Platform",
    description:
      "Run fully branded online auctions with real-time bidding, proxy logic, and institutional-grade controls.",
    images: ["/og/elevare-marketing-og.svg"]
  }
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${fontSerif.variable} ${fontSans.variable} font-sans`}>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
