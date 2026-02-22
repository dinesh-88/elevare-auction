import { Link } from "react-router-dom";
import type { Auction } from "../api/client";
import CountdownTimer from "./CountdownTimer";
import { currencySymbol } from "../utils/currency";

export default function AuctionCard({ auction }: { auction: Auction }) {
  const image = auction.metadata?.image ?? auction.metadata?.images?.[0];
  const symbol = currencySymbol(auction.currency);

  return (
    <Link to={`/auction/${auction.id}`} className="group block no-underline">
      <article className="overflow-hidden rounded-lg border border-neutral-200 bg-white transition-shadow hover:shadow-md">
        <div className="aspect-[4/3] overflow-hidden bg-neutral-100">
          {image ? (
            <img src={image} alt={auction.title} className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />
          ) : (
            <div className="grid h-full place-content-center text-sm text-neutral-500">No image</div>
          )}
        </div>

        <div className="space-y-3 p-5 md:p-6">
          <div className="text-sm text-neutral-500">Auction</div>
          <h3 className="font-heading text-lg font-medium text-neutral-900 md:text-xl">{auction.title}</h3>
          <div className="flex items-baseline gap-2">
            <span className="text-xl font-heading font-semibold text-neutral-900 md:text-2xl">
              {symbol}
              {auction.current_price.toLocaleString()}
            </span>
          </div>
          <CountdownTimer endTime={new Date(auction.end_time)} />
          <div className="flex items-center gap-2 pt-2 text-brand-primary">
            <span className="font-medium">View</span>
            <span className="transition-transform group-hover:translate-x-1">→</span>
          </div>
        </div>
      </article>
    </Link>
  );
}
