import { Link } from "react-router-dom";
import type { Auction } from "../api/client";
import Countdown from "./Countdown";

export default function AuctionCard({ auction }: { auction: Auction }) {
  const image = auction.metadata?.image ?? auction.metadata?.images?.[0];

  return (
    <article className="auction-card">
      <div className="card-image-wrap">
        {image ? <img src={image} alt={auction.title} className="card-image" /> : <div className="card-placeholder">No image</div>}
      </div>
      <div className="card-body">
        <h3>{auction.title}</h3>
        <p className="price">${auction.current_price.toLocaleString()}</p>
        <Countdown endTime={auction.end_time} />
        <Link to={`/auction/${auction.id}`} className="btn">
          View
        </Link>
      </div>
    </article>
  );
}
