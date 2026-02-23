import { api, normalizeAuction, normalizeLot, normalizeLotBid, type Auction, type Lot, type LotBid } from "../../../../../../../src/api/client";
import LotDetail from "../../../../../../../src/screens/public/LotDetail";

type Props = {
  params: {
    auctionId: string;
    lotId: string;
  };
};

export default async function Page({ params }: Props) {
  const lotId = params.lotId;
  const auctionId = params.auctionId;

  let initialLot: Lot | null = null;
  let initialAuction: Auction | null = null;
  let initialRelatedLots: Lot[] = [];
  let initialBids: LotBid[] = [];

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
    if (bidsRes.ok) initialBids = ((await bidsRes.json()) as unknown[]).map(normalizeLotBid);
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
