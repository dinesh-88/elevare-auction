import { Outlet } from "react-router-dom";
import Header from "../components/Header";
import { useTheme } from "../theme/ThemeProvider";
import { useAuth } from "../hooks/useAuth";

export default function PublicLayout() {
  const { config } = useTheme();
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-white text-neutral-800 font-body">
      <Header
        tenantName={config?.name ?? "Elevare"}
        tenantLogoUrl={config?.branding?.logo_url}
        isAuthenticated={isAuthenticated}
        userEmail={user?.email}
        onLogout={logout}
      />

      <main>
        <Outlet />
      </main>
    </div>
  );
}
