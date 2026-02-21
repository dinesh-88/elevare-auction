import { Link, useLocation } from "react-router-dom";

const links = [
  { to: "/admin", label: "Dashboard" },
  { to: "/admin/auctions", label: "Auctions" },
  { to: "/admin/financial-rules", label: "Financial Rules" },
  { to: "/admin/create-auction", label: "Create Auction" },
  { to: "/admin/bidders", label: "Bidders" },
];

export default function AdminSidebar() {
  const location = useLocation();

  return (
    <aside className="admin-sidebar">
      <h2 className="mb-4 font-heading text-2xl text-neutral-900">Elevare Admin</h2>
      <p className="mb-4 text-xs text-neutral-500">Manage auctions, lots, and bidders.</p>
      <nav>
        {links.map((link) => (
          <Link key={link.to} to={link.to} className={location.pathname === link.to ? "active" : ""}>
            {link.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
