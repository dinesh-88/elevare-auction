import { Link } from "react-router-dom";
import type { Auction } from "../api/client";
import Countdown from "./Countdown";

export default function AuctionCard({ auction }: { auction: Auction }) {
  const image = auction.metadata?.image ?? auction.metadata?.images?.[0];

  return (
    <article className="overflow-hidden rounded-md border border-neutral-200 bg-white transition-shadow hover:shadow-md">
      <div className="aspect-[4/3] bg-neutral-100">
        {image ? (
          <img src={image} alt={auction.title} className="h-full w-full object-cover" />
        ) : (
          <div className="grid h-full place-content-center text-sm text-neutral-500">No image</div>
        )}
      </div>

      <div className="p-6">
        <h3 className="mb-2 font-heading text-lg text-neutral-900">{auction.title}</h3>

        <div className="bid-amount mb-2 font-heading text-2xl text-brand-primary">${auction.current_price.toLocaleString()}</div>

        <div className="mb-4 text-sm text-neutral-500">
          Ends in <Countdown endTime={auction.end_time} />
        </div>

        <Link to={`/auction/${auction.id}`} className="btn-link-brand">
          View Auction →
        </Link>
      </div>
    </article>
  );
}
