import { Outlet } from "react-router-dom";
import Header from "../components/Header";
import { useTheme } from "../theme/ThemeProvider";
import { useAuth } from "../hooks/useAuth";

export default function PublicLayout() {
  const { config } = useTheme();
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-800 font-body">
      <Header tenantName={config?.name ?? "Auction"} isAuthenticated={isAuthenticated} userEmail={user?.email} onLogout={logout} />

      <main className="container-shell py-12">
        <Outlet />
      </main>

      <footer className="border-t border-neutral-200 bg-white py-6 text-center text-sm text-neutral-500">
        {config?.branding?.footer_text ?? "Powered by Elevare Auction"}
      </footer>
    </div>
  );
}
