import AuctionListing from "../../../src/screens/public/AuctionListing";
import { normalizeAuction } from "../../../src/api/client";
import type { Auction } from "../../../src/api/client";
import { fetchPublicAuctions } from "../../../src/server/api";

export default async function Page() {
  let auctions: Auction[] = [];
  try {
    const raw = await fetchPublicAuctions();
    auctions = raw.map(normalizeAuction);
  } catch {
    auctions = [];
  }
  return <AuctionListing initialAuctions={auctions} />;
}
