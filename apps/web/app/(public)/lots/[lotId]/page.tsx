import LotDetail from "../../../../src/screens/public/LotDetail";
import { normalizeAuction, normalizeLot, normalizeLotBid } from "../../../../src/api/client";
import type { Lot } from "../../../../src/api/client";
import { fetchAuction, fetchLot, fetchLotBids, fetchLotsByAuction } from "../../../../src/server/api";

export default async function Page({ params }: { params: Promise<{ lotId: string }> }) {
  const { lotId } = await params;

  let initialLot: Lot | null = null;
  let initialAuction: { id: string; status: "DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED"; currency: string } | null = null;
  let initialRelatedLots: Lot[] = [];
  let initialBids: Array<{ bid_id: string; amount: number; created_at: string }> = [];

  try {
    const rawLot = await fetchLot(lotId);
    initialLot = normalizeLot(rawLot);
    const [rawAuction, rawLots, rawBids] = await Promise.all([
      fetchAuction(initialLot.auction_id),
      fetchLotsByAuction(initialLot.auction_id),
      fetchLotBids(lotId),
    ]);
    const auction = normalizeAuction(rawAuction);
    initialAuction = { id: auction.id, status: auction.status, currency: auction.currency };
    initialRelatedLots = rawLots.map(normalizeLot).filter((item) => item.id !== initialLot?.id);
    initialBids = rawBids.map(normalizeLotBid).map((bid) => ({
      bid_id: bid.id,
      amount: bid.amount,
      created_at: bid.created_at,
    }));
  } catch {
    initialLot = null;
  }

  return (
    <LotDetail
      lotID={lotId}
      initialLot={initialLot}
      initialAuction={initialAuction}
      initialRelatedLots={initialRelatedLots}
      initialBids={initialBids}
    />
  );
}
