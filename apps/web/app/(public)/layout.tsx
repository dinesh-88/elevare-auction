"use client";

import Header from "../../src/components/Header";
import { useAuth } from "../../src/hooks/useAuth";
import { useTheme } from "../../src/theme/ThemeProvider";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  const { config } = useTheme();
  const { isAuthenticated, isAdmin, user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-white text-neutral-800 font-body">
      <Header
        tenantName={config?.name ?? "Elevare"}
        tenantLogoUrl={config?.branding?.logo_url}
        isAuthenticated={isAuthenticated}
        isAdmin={isAdmin}
        userEmail={user?.email}
        onLogout={logout}
      />
      <main>{children}</main>
    </div>
  );
}
