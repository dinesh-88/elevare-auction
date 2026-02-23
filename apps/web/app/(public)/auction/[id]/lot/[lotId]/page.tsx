import { normalizeAuction, normalizeLot, normalizeLotBid, type Auction, type Lot } from "../../../../../../src/api/client";
import LotDetail from "../../../../../../src/screens/public/LotDetail";

type Props = {
  params: Promise<{
    id: string;
    lotId: string;
  }>;
};

export default async function Page({ params }: Props) {
  const { lotId, id: auctionId } = await params;

  let initialLot: Lot | null = null;
  let initialAuction: Auction | null = null;
  let initialRelatedLots: Lot[] = [];
  let initialBids: Array<{ bid_id: string; amount_minor: number; created_at: string; bidder_display: string }> = [];

  try {
    const apiBase = process.env.API_BASE_URL ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
    const [lotRes, auctionRes, lotsRes, bidsRes] = await Promise.all([
      fetch(`${apiBase}/lots/${lotId}`, { cache: "no-store" }),
      fetch(`${apiBase}/auctions/${auctionId}`, { cache: "no-store" }),
      fetch(`${apiBase}/auctions/${auctionId}/lots`, { cache: "no-store" }),
      fetch(`${apiBase}/lots/${lotId}/bids`, { cache: "no-store" }),
    ]);
    if (lotRes.ok) initialLot = normalizeLot(await lotRes.json());
    if (auctionRes.ok) initialAuction = normalizeAuction(await auctionRes.json());
    if (lotsRes.ok) initialRelatedLots = ((await lotsRes.json()) as unknown[]).map(normalizeLot).filter((l) => l.id !== lotId);
    if (bidsRes.ok) {
      initialBids = ((await bidsRes.json()) as unknown[])
        .map(normalizeLotBid)
        .map((bid) => ({
          bid_id: bid.id,
          amount_minor: bid.amount_minor,
          created_at: bid.created_at,
          bidder_display: bid.bidder_display,
        }));
    }
  } catch {
    // let client-side hydration fetch recover
  }

  return (
    <LotDetail
      lotID={lotId}
      initialLot={initialLot}
      initialAuction={initialAuction ? { id: initialAuction.id, status: initialAuction.status, currency: initialAuction.currency } : null}
      initialRelatedLots={initialRelatedLots}
      initialBids={initialBids}
    />
  );
}
