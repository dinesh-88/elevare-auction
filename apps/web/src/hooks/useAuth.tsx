import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api, setAuthTokenGetter, type Role, type User } from "../api/client";

interface AuthContextValue {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<User>;
  logout: () => void;
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
    role,
    email_verified: Boolean(r.email_verified ?? r.EmailVerified),
    approval_status: String(r.approval_status ?? r.ApprovalStatus ?? "PENDING") as User["approval_status"],
    kyc_status: String(r.kyc_status ?? r.KYCStatus ?? "NONE") as User["kyc_status"],
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    setAuthTokenGetter(() => token);
  }, [token]);

  const login = async (email: string, password: string) => {
    const res = await api.login({ email, password });
    const normalizedUser = normalizeUser(res.user);
    if (!normalizedUser) throw new Error("Invalid user payload");
    setToken(res.token);
    setUser(normalizedUser);
    return normalizedUser;
  };

  const logout = () => {
    setToken(null);
    setUser(null);
  };

  const value = useMemo(
    () => ({
      token,
      user,
      isAuthenticated: Boolean(token),
      isAdmin: user?.role === ("ADMIN" as Role),
      login,
      logout,
    }),
    [token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
