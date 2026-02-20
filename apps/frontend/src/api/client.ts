export type Role = "ADMIN" | "BIDDER";
export type TenantStatus = "PENDING" | "ACTIVE" | "REJECTED";

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  role: Role;
  email_verified: boolean;
  approval_status: "PENDING" | "APPROVED" | "REJECTED";
  kyc_status: "NONE" | "PENDING" | "VERIFIED" | "REJECTED";
  created_at: string;
}

export interface Auction {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  metadata?: { image?: string; images?: string[] };
  starting_price: number;
  current_price: number;
  reserve_price?: number | null;
  min_increment: number;
  start_time: string;
  end_time: string;
  status: "DRAFT" | "SCHEDULED" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
  allow_manual_start: boolean;
  allow_manual_pause: boolean;
  allow_manual_end: boolean;
  auto_extend_enabled: boolean;
  sniping_window_seconds: number;
  sniping_extension_seconds: number;
  bid_visibility: "PUBLIC" | "ANONYMOUS" | "SEALED";
  created_at: string;
}

export interface TenantPublicConfig {
  name: string;
  branding: {
    logo_url?: string;
    primary_color?: string;
    secondary_color?: string;
    hero_image?: string;
    footer_text?: string;
  };
  require_admin_approval: boolean;
  require_kyc: boolean;
}

type JsonRecord = Record<string, unknown>;

type TokenGetter = () => string | null;
let getToken: TokenGetter = () => null;

export const setAuthTokenGetter = (fn: TokenGetter) => {
  getToken = fn;
};

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Content-Type", "application/json");
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  const text = await res.text();
  const body = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const message = typeof body === "string" ? body : (body?.error as string) || text || "Request failed";
    throw new Error(message);
  }

  return body as T;
}

async function requestWithToken<T>(path: string, bearerToken: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Content-Type", "application/json");
  headers.set("Authorization", `Bearer ${bearerToken}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  const text = await res.text();
  const body = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const message = typeof body === "string" ? body : (body?.error as string) || text || "Request failed";
    throw new Error(message);
  }

  return body as T;
}

export const api = {
  tenantSignup: (payload: {
    organization_name: string;
    desired_subdomain: string;
    admin_email: string;
    password: string;
    country: string;
    estimated_auction_volume?: number;
  }) => request<JsonRecord>("/signup", { method: "POST", body: JSON.stringify(payload) }),
  getPublicConfig: () => request<TenantPublicConfig>("/public/config"),
  listAuctions: () => request<Auction[]>("/auctions"),
  getAuction: (id: string) => request<Auction>(`/auctions/${id}`),
  register: (payload: { email: string; password: string }) =>
    request<JsonRecord>("/auth/register", { method: "POST", body: JSON.stringify(payload) }),
  login: (payload: { email: string; password: string }) =>
    request<{ token: string; user: User }>("/auth/login", { method: "POST", body: JSON.stringify(payload) }),
  verifyEmail: (userId: string) =>
    request<JsonRecord>("/auth/verify-email", { method: "POST", body: JSON.stringify({ user_id: userId }) }),
  placeBid: (auctionId: string, amount: number) =>
    request<JsonRecord>(`/auctions/${auctionId}/bid`, { method: "POST", body: JSON.stringify({ amount }) }),
  createAuction: (payload: JsonRecord) =>
    request<Auction>("/admin/auctions", { method: "POST", body: JSON.stringify(payload) }),
  listAdminAuctions: () => request<Auction[]>("/admin/auctions"),
  updateAuction: (id: string, payload: JsonRecord) =>
    request<Auction>(`/admin/auctions/${id}`, { method: "PATCH", body: JSON.stringify(payload) }),
  deleteAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}`, { method: "DELETE" }),
  startAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/start`, { method: "PATCH" }),
  pauseAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/pause`, { method: "PATCH" }),
  endAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/end`, { method: "PATCH" }),
  listBidders: () => request<User[]>("/admin/bidders"),
  approveBidder: (id: string) => request<JsonRecord>(`/admin/bidders/${id}/approve`, { method: "PATCH" }),
  platformListTenants: (platformToken: string, status: TenantStatus = "PENDING") =>
    requestWithToken<Array<Record<string, unknown>>>(`/platform/admin/tenants?status=${status}`, platformToken),
  platformApproveTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/approve`, platformToken, { method: "PATCH" }),
  platformRejectTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/reject`, platformToken, { method: "PATCH" }),
};
