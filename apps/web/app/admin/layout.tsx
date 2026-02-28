"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import AdminSidebar from "../../src/components/AdminSidebar";
import { api } from "../../src/api/client";
import { useAuth } from "../../src/hooks/useAuth";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { loading, isAuthenticated, isAdmin } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [checkingOnboarding, setCheckingOnboarding] = useState(true);

  useEffect(() => {
    if (loading) return;
    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }
    if (!isAdmin) {
      router.replace("/");
    }
  }, [loading, isAdmin, isAuthenticated, router]);

  useEffect(() => {
    let cancelled = false;
    const exempt = pathname === "/admin/onboarding" || pathname === "/admin/settings";

    if (loading || !isAuthenticated || !isAdmin) {
      setCheckingOnboarding(true);
      return;
    }
    if (exempt) {
      setCheckingOnboarding(false);
      return;
    }

    setCheckingOnboarding(true);
    api
      .getInstitutionOnboarding()
      .then((state) => {
        if (cancelled) return;
        if (state.status !== "completed") {
          router.replace("/admin/onboarding");
          return;
        }
        setCheckingOnboarding(false);
      })
      .catch(() => {
        if (!cancelled) setCheckingOnboarding(false);
      });

    return () => {
      cancelled = true;
    };
  }, [loading, isAdmin, isAuthenticated, pathname, router]);

  if (loading || !isAuthenticated || !isAdmin || checkingOnboarding) {
    return null;
  }

  return (
    <div className="admin-shell">
      <AdminSidebar />
      <main className="admin-main page-shell">{children}</main>
    </div>
  );
}
