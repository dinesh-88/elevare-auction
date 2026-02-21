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
  auction_mode: "TIMED" | "SEQUENTIAL";
  auto_progress_enabled: boolean;
  lot_duration_seconds: number;
  sniping_window_seconds: number;
  sniping_extension_seconds: number;
  bid_visibility: "PUBLIC" | "ANONYMOUS" | "SEALED";
  created_at: string;
}

export interface Lot {
  id: string;
  auction_id: string;
  lot_number: string;
  order_index: number;
  title: string;
  description: string;
  starting_price: number;
  current_price: number;
  reserve_price?: number | null;
  min_increment: number;
  metadata?: { image?: string; images?: string[] };
  status: "PENDING" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
  started_at?: string | null;
  ended_at?: string | null;
  created_at: string;
  updated_at: string;
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

function normalizeAuction(raw: unknown): Auction {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    title: String(r.title ?? r.Title ?? ""),
    description: String(r.description ?? r.Description ?? ""),
    metadata: (r.metadata ?? r.Metadata ?? {}) as Auction["metadata"],
    starting_price: Number(r.starting_price ?? r.StartingPrice ?? 0),
    current_price: Number(r.current_price ?? r.CurrentPrice ?? 0),
    reserve_price: (r.reserve_price ?? r.ReservePrice ?? null) as number | null,
    min_increment: Number(r.min_increment ?? r.MinIncrement ?? 0),
    start_time: String(r.start_time ?? r.StartTime ?? ""),
    end_time: String(r.end_time ?? r.EndTime ?? ""),
    status: String(r.status ?? r.Status ?? "DRAFT") as Auction["status"],
    allow_manual_start: Boolean(r.allow_manual_start ?? r.AllowManualStart),
    allow_manual_pause: Boolean(r.allow_manual_pause ?? r.AllowManualPause),
    allow_manual_end: Boolean(r.allow_manual_end ?? r.AllowManualEnd),
    auto_extend_enabled: Boolean(r.auto_extend_enabled ?? r.AutoExtendEnabled),
    auction_mode: String(r.auction_mode ?? r.AuctionMode ?? "TIMED") as Auction["auction_mode"],
    auto_progress_enabled: Boolean(r.auto_progress_enabled ?? r.AutoProgressEnabled),
    lot_duration_seconds: Number(r.lot_duration_seconds ?? r.LotDurationSeconds ?? 0),
    sniping_window_seconds: Number(r.sniping_window_seconds ?? r.SnipingWindowSeconds ?? 0),
    sniping_extension_seconds: Number(r.sniping_extension_seconds ?? r.SnipingExtensionSeconds ?? 0),
    bid_visibility: String(r.bid_visibility ?? r.BidVisibility ?? "PUBLIC") as Auction["bid_visibility"],
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeLot(raw: unknown): Lot {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    auction_id: String(r.auction_id ?? r.AuctionID ?? ""),
    lot_number: String(r.lot_number ?? r.LotNumber ?? ""),
    order_index: Number(r.order_index ?? r.OrderIndex ?? 0),
    title: String(r.title ?? r.Title ?? ""),
    description: String(r.description ?? r.Description ?? ""),
    starting_price: Number(r.starting_price ?? r.StartingPrice ?? 0),
    current_price: Number(r.current_price ?? r.CurrentPrice ?? 0),
    reserve_price: (r.reserve_price ?? r.ReservePrice ?? null) as number | null,
    min_increment: Number(r.min_increment ?? r.MinIncrement ?? 0),
    metadata: (r.metadata ?? r.Metadata ?? {}) as Lot["metadata"],
    status: String(r.status ?? r.Status ?? "PENDING") as Lot["status"],
    started_at: (r.started_at ?? r.StartedAt ?? null) as string | null,
    ended_at: (r.ended_at ?? r.EndedAt ?? null) as string | null,
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
    updated_at: String(r.updated_at ?? r.UpdatedAt ?? ""),
  };
}

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
  listAuctions: async () => (await request<unknown[]>("/auctions")).map(normalizeAuction),
  getAuction: async (id: string) => normalizeAuction(await request<unknown>(`/auctions/${id}`)),
  listLotsByAuction: async (auctionId: string) => (await request<unknown[]>(`/auctions/${auctionId}/lots`)).map(normalizeLot),
  getLot: async (lotId: string) => normalizeLot(await request<unknown>(`/lots/${lotId}`)),
  register: (payload: { email: string; password: string }) =>
    request<JsonRecord>("/auth/register", { method: "POST", body: JSON.stringify(payload) }),
  login: (payload: { email: string; password: string }) =>
    request<{ token: string; user: User }>("/auth/login", { method: "POST", body: JSON.stringify(payload) }),
  verifyEmail: (userId: string) =>
    request<JsonRecord>("/auth/verify-email", { method: "POST", body: JSON.stringify({ user_id: userId }) }),
  placeBid: async (..._args: unknown[]) => {
    throw new Error("Auction-level bidding has been removed. Use lot bidding.");
  },
  placeLotBid: (lotId: string, amount: number) =>
    request<JsonRecord>(`/lots/${lotId}/bid`, { method: "POST", body: JSON.stringify({ amount }) }),
  createAuction: async (payload: JsonRecord) =>
    normalizeAuction(await request<unknown>("/admin/auctions", { method: "POST", body: JSON.stringify(payload) })),
  listAdminAuctions: async () => (await request<unknown[]>("/admin/auctions")).map(normalizeAuction),
  updateAuction: async (id: string, payload: JsonRecord) =>
    normalizeAuction(await request<unknown>(`/admin/auctions/${id}`, { method: "PATCH", body: JSON.stringify(payload) })),
  deleteAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}`, { method: "DELETE" }),
  startAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/start`, { method: "PATCH" }),
  pauseAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/pause`, { method: "PATCH" }),
  endAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/end`, { method: "PATCH" }),
  listBidders: () => request<User[]>("/admin/bidders"),
  approveBidder: (id: string) => request<JsonRecord>(`/admin/bidders/${id}/approve`, { method: "PATCH" }),
  createLot: async (auctionId: string, payload: JsonRecord) =>
    normalizeLot(await request<unknown>(`/admin/auctions/${auctionId}/lots`, { method: "POST", body: JSON.stringify(payload) })),
  updateLot: async (lotId: string, payload: JsonRecord) =>
    normalizeLot(await request<unknown>(`/admin/lots/${lotId}`, { method: "PATCH", body: JSON.stringify(payload) })),
  deleteLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}`, { method: "DELETE" }),
  reorderLots: (auctionId: string, lotIds: string[]) =>
    request<JsonRecord>(`/admin/auctions/${auctionId}/reorder-lots`, { method: "POST", body: JSON.stringify({ lot_ids: lotIds }) }),
  startLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}/start`, { method: "POST" }),
  endLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}/end`, { method: "POST" }),
  platformListTenants: (platformToken: string, status: TenantStatus = "PENDING") =>
    requestWithToken<Array<Record<string, unknown>>>(`/platform/admin/tenants?status=${status}`, platformToken),
  platformApproveTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/approve`, platformToken, { method: "PATCH" }),
  platformRejectTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/reject`, platformToken, { method: "PATCH" }),
};
