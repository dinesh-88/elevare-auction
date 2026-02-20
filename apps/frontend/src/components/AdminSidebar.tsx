import { Link, useLocation } from "react-router-dom";

const links = [
  { to: "/admin", label: "Dashboard" },
  { to: "/admin/auctions", label: "Auctions" },
  { to: "/admin/create-auction", label: "Create Auction" },
  { to: "/admin/bidders", label: "Bidders" },
];

export default function AdminSidebar() {
  const location = useLocation();

  return (
    <aside className="admin-sidebar">
      <h2>Elevare Admin</h2>
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
