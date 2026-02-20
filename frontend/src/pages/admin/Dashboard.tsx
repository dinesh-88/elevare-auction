import { useEffect, useState } from "react";
import { api } from "../../api/client";

export default function Dashboard() {
  const [auctionsCount, setAuctionsCount] = useState(0);
  const [pendingBidders, setPendingBidders] = useState(0);

  useEffect(() => {
    Promise.all([api.listAuctions(), api.listBidders()])
      .then(([auctions, bidders]) => {
        setAuctionsCount(auctions.length);
        setPendingBidders(bidders.filter((b) => b.approval_status === "PENDING").length);
      })
      .catch(() => {
        setAuctionsCount(0);
        setPendingBidders(0);
      });
  }, []);

  return (
    <section>
      <h1>Dashboard</h1>
      <div className="stats-grid">
        <article className="stat-card">
          <h3>Visible Auctions</h3>
          <p>{auctionsCount}</p>
        </article>
        <article className="stat-card">
          <h3>Pending Bidders</h3>
          <p>{pendingBidders}</p>
        </article>
      </div>
    </section>
  );
}
