import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api, setAuthTokenGetter, type Role, type User } from "../api/client";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<User>;
  updateLocale: (locale: User["locale"]) => Promise<User>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function normalizeUser(raw: unknown): User | null {
  if (!raw || typeof raw !== "object") return null;
  const r = raw as Record<string, unknown>;
  const role = (r.role ?? r.Role) as Role | undefined;
  const id = (r.id ?? r.ID) as string | undefined;
  const tenantId = (r.tenant_id ?? r.TenantID) as string | undefined;
  const email = (r.email ?? r.Email) as string | undefined;
  if (!id || !tenantId || !email || !role) return null;

  return {
    id,
    tenant_id: tenantId,
    email,
    locale: String(r.locale ?? r.Locale ?? "en") as User["locale"],
    role,
    email_verified: Boolean(r.email_verified ?? r.EmailVerified),
    approval_status: String(r.approval_status ?? r.ApprovalStatus ?? "PENDING") as User["approval_status"],
    kyc_status: String(r.kyc_status ?? r.KYCStatus ?? "NONE") as User["kyc_status"],
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setAuthTokenGetter(() => null);
  }, []);

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const res = await api.getSession();
        const normalizedUser = normalizeUser(res.user);
        if (active) setUser(normalizedUser);
      } catch {
        if (active) setUser(null);
      } finally {
        if (active) setLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, []);

  const login = async (email: string, password: string) => {
    const res = await api.login({ email, password });
    const normalizedUser = normalizeUser(res.user);
    if (!normalizedUser) throw new Error("Invalid user payload");
    setUser(normalizedUser);
    return normalizedUser;
  };

  const updateLocale = async (locale: User["locale"]) => {
    const res = await api.updateMe({ locale });
    const normalizedUser = normalizeUser(res.user);
    if (!normalizedUser) throw new Error("Invalid user payload");
    setUser(normalizedUser);
    return normalizedUser;
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // Clear local auth state even if logout request fails.
    }
    setUser(null);
  };

  const value = useMemo(
    () => ({
      user,
      loading,
      isAuthenticated: Boolean(user),
      isAdmin: user?.role !== ("BIDDER" as Role),
      login,
      updateLocale,
      logout,
    }),
    [user, loading],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
