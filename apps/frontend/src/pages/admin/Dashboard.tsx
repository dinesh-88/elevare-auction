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
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Dashboard</h1>
          <p className="section-copy">Snapshot of current auction activity.</p>
        </div>
      </header>
      <div className="stats-grid">
        <article className="panel">
          <p className="text-sm text-neutral-500">Visible Auctions</p>
          <p className="mt-2 font-heading text-4xl">{auctionsCount}</p>
        </article>
        <article className="panel">
          <p className="text-sm text-neutral-500">Pending Bidders</p>
          <p className="mt-2 font-heading text-4xl">{pendingBidders}</p>
        </article>
      </div>
    </section>
  );
}
