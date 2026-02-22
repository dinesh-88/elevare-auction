import type { Metadata } from "next";
import AppProviders from "../src/app-providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "Elevare Auction",
  description: "Multi-tenant auction platform"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
