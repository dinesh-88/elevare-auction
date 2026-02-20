import { Link, Outlet } from "react-router-dom";
import { useTheme } from "../theme/ThemeProvider";
import { useAuth } from "../hooks/useAuth";

export default function PublicLayout() {
  const { config } = useTheme();
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-800 font-body">
      <header className="border-b border-neutral-200 bg-white/95">
        <div className="container-shell flex items-center justify-between py-4">
          <Link to="/" className="flex items-center gap-3 font-heading text-xl text-neutral-800">
            {config?.branding?.logo_url ? (
              <img src={config.branding.logo_url} alt="logo" className="h-9 w-9 rounded-sm object-contain" />
            ) : null}
            <span>{config?.name ?? "Auction"}</span>
          </Link>

          <nav className="flex items-center gap-6 text-sm text-neutral-600">
            <Link className="hover:text-brand-primary" to="/">
              Auctions
            </Link>
            <Link className="hover:text-brand-primary" to="/signup">
              Organization Signup
            </Link>
            {!isAuthenticated ? (
              <Link className="hover:text-brand-primary" to="/login">
                Login
              </Link>
            ) : null}
            {!isAuthenticated ? (
              <Link className="hover:text-brand-primary" to="/register">
                Register
              </Link>
            ) : null}
            {isAuthenticated ? <span className="text-xs text-neutral-500">{user?.email}</span> : null}
            {isAuthenticated ? (
              <button type="button" className="btn-brand" onClick={logout}>
                Logout
              </button>
            ) : null}
          </nav>
        </div>
      </header>

      <main className="container-shell py-12">
        <Outlet />
      </main>

      <footer className="border-t border-neutral-200 bg-white py-6 text-center text-sm text-neutral-500">
        {config?.branding?.footer_text ?? "Powered by Elevare Auction"}
      </footer>
    </div>
  );
}
