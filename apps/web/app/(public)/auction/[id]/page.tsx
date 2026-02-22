import AuctionDetail from "../../../../src/screens/public/AuctionDetail";
import { normalizeAuction, normalizeLot } from "../../../../src/api/client";
import type { Auction, Lot } from "../../../../src/api/client";
import { fetchAuction, fetchLotsByAuction } from "../../../../src/server/api";

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  let initialAuction: Auction | null = null;
  let initialLots: Lot[] = [];
  try {
    const [rawAuction, rawLots] = await Promise.all([fetchAuction(id), fetchLotsByAuction(id)]);
    initialAuction = normalizeAuction(rawAuction);
    initialLots = rawLots.map(normalizeLot);
  } catch {
    initialAuction = null;
    initialLots = [];
  }

  return <AuctionDetail auctionID={id} initialAuction={initialAuction} initialLots={initialLots} />;
}
