import { Link, Outlet } from "react-router-dom";
import { useTheme } from "../theme/ThemeProvider";
import { useAuth } from "../hooks/useAuth";

export default function PublicLayout() {
  const { config } = useTheme();
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <div className="public-shell">
      <header className="public-header">
        <Link to="/" className="brand">
          {config?.branding?.logo_url ? <img src={config.branding.logo_url} alt="logo" className="logo" /> : null}
          <span>{config?.name ?? "Auction"}</span>
        </Link>
        <nav className="nav-links">
          <Link to="/">Auctions</Link>
          {!isAuthenticated ? <Link to="/login">Login</Link> : null}
          {!isAuthenticated ? <Link to="/register">Register</Link> : null}
          {isAuthenticated ? <span>{user?.email}</span> : null}
          {isAuthenticated ? (
            <button type="button" onClick={logout}>
              Logout
            </button>
          ) : null}
        </nav>
      </header>

      <main>
        <Outlet />
      </main>

      <footer className="public-footer">{config?.branding?.footer_text ?? "Powered by Elevare Auction"}</footer>
    </div>
  );
}
