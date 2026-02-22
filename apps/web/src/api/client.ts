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

export interface LotBid {
  id: string;
  lot_id: string;
  user_id: string;
  amount: number;
  created_at: string;
}

export interface Auction {
  id: string;
  tenant_id: string;
  currency: string;
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
  metadata?: {
    image?: string;
    images?: string[];
    details?: Array<{ label: string; value: string }>;
    artist?: string;
    date?: string;
    dimensions?: string;
    medium?: string;
    provenance?: string;
    condition_report?: string;
    [key: string]: unknown;
  };
  status: "PENDING" | "LIVE" | "PAUSED" | "ENDED" | "CANCELLED";
  hammer_price?: number | null;
  sale_result?: "SOLD" | "RESERVE_NOT_MET" | "";
  started_at?: string | null;
  ended_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface LotFinancials {
  lot_id: string;
  winner_user?: string | null;
  hammer_price?: number | null;
  invoice_id?: string | null;
  fees: Array<{
    name: string;
    type: "PERCENTAGE" | "FIXED";
    value: number;
    amount: number;
  }>;
  total_due?: number | null;
}

export interface Invoice {
  id: string;
  lot_id: string;
  tenant_id: string;
  winner_user_id: string;
  invoice_number: string;
  issued_at: string;
  due_date?: string | null;
  currency: string;
  total_amount: number;
  snapshot: {
    currency?: string;
    hammer_price: number;
    fees: Array<{
      name: string;
      type: "PERCENTAGE" | "FIXED";
      value: number;
      amount: number;
    }>;
    total_due: number;
  };
  status: "ISSUED" | "PAID" | "CANCELLED";
  created_at: string;
  lot_number?: string;
  lot_title?: string;
  auction_title?: string;
}

export type FeeCalculationType = "PERCENTAGE" | "FIXED";
export type FeeAppliesTo = "HAMMER" | "PREMIUM" | "SUBTOTAL";
export type FeeScopeType = "TENANT" | "AUCTION" | "LOT";

export interface FeeComponent {
  id: string;
  tenant_id: string;
  name: string;
  calculation_type: FeeCalculationType;
  value: number;
  applies_to: FeeAppliesTo;
  taxable: boolean;
  active: boolean;
  created_at: string;
}

export interface FeeAssignment {
  id: string;
  scope_type: FeeScopeType;
  scope_id: string;
  fee_component_id: string;
  created_at: string;
}

export interface TenantPublicConfig {
  name: string;
  default_currency: string;
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

export interface TenantSettings {
  default_currency: string;
}

type JsonRecord = Record<string, unknown>;

type TokenGetter = () => string | null;
let getToken: TokenGetter = () => null;

export const setAuthTokenGetter = (fn: TokenGetter) => {
  getToken = fn;
};

const normalizeBase = (raw: string) => raw.trim().replace(/\/+$/, "");

const adaptBaseToCurrentHostname = (base: string) => {
  if (typeof window === "undefined" || !base) return base;
  try {
    const url = new URL(base);
    const currentHost = window.location.hostname;
    const baseHost = url.hostname;
    const isLocalBase = baseHost === "localhost" || baseHost === "127.0.0.1";
    const isTenantHost = currentHost.endsWith(".localhost") && currentHost !== "localhost";
    if (isLocalBase && isTenantHost) {
      url.hostname = currentHost;
      return url.toString().replace(/\/+$/, "");
    }
    return base;
  } catch {
    return base;
  }
};

const resolveRuntimeApiBase = () => {
  if (typeof window === "undefined") return "";
  const protocol = window.location.protocol === "https:" ? "https" : "http";
  const host = window.location.hostname;
  const port = process.env.NEXT_PUBLIC_API_PORT ?? "8080";
  return `${protocol}://${host}:${port}`;
};

const API_BASE = adaptBaseToCurrentHostname(
  normalizeBase(process.env.NEXT_PUBLIC_API_BASE_URL ?? resolveRuntimeApiBase()),
);

const resolveMediaUrl = (value: string): string => {
  if (!value) return value;
  if (value.startsWith("http://") || value.startsWith("https://")) return value;
  if (!value.startsWith("/")) return value;
  return `${API_BASE}${value}`;
};

export function normalizeAuction(raw: unknown): Auction {
  const r = (raw ?? {}) as Record<string, unknown>;
  const rawMetadata = (r.metadata ?? r.Metadata ?? {}) as Record<string, unknown>;
  const normalizedMetadata: Auction["metadata"] = {
    ...(rawMetadata as Auction["metadata"]),
  };
  if (typeof rawMetadata.image === "string") {
    normalizedMetadata.image = resolveMediaUrl(rawMetadata.image);
  }
  if (Array.isArray(rawMetadata.images)) {
    normalizedMetadata.images = rawMetadata.images.filter((x): x is string => typeof x === "string").map(resolveMediaUrl);
  }

  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    currency: String(r.currency ?? r.Currency ?? "USD"),
    title: String(r.title ?? r.Title ?? ""),
    description: String(r.description ?? r.Description ?? ""),
    metadata: normalizedMetadata,
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

export function normalizeLot(raw: unknown): Lot {
  const r = (raw ?? {}) as Record<string, unknown>;
  const rawMetadata = (r.metadata ?? r.Metadata ?? {}) as Record<string, unknown>;
  const normalizedMetadata: Lot["metadata"] = {
    ...(rawMetadata as Lot["metadata"]),
  };
  if (typeof rawMetadata.image === "string") {
    normalizedMetadata.image = resolveMediaUrl(rawMetadata.image);
  }
  if (Array.isArray(rawMetadata.images)) {
    normalizedMetadata.images = rawMetadata.images.filter((x): x is string => typeof x === "string").map(resolveMediaUrl);
  }
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
    metadata: normalizedMetadata,
    status: String(r.status ?? r.Status ?? "PENDING") as Lot["status"],
    hammer_price: (r.hammer_price ?? r.HammerPrice ?? null) as number | null,
    sale_result: String(r.sale_result ?? r.SaleResult ?? "") as Lot["sale_result"],
    started_at: (r.started_at ?? r.StartedAt ?? null) as string | null,
    ended_at: (r.ended_at ?? r.EndedAt ?? null) as string | null,
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
    updated_at: String(r.updated_at ?? r.UpdatedAt ?? ""),
  };
}

export function normalizeLotBid(raw: unknown): LotBid {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    lot_id: String(r.lot_id ?? r.LotID ?? (r.lotId as string | undefined) ?? ""),
    user_id: String(r.user_id ?? r.UserID ?? ""),
    amount: Number(r.amount ?? r.Amount ?? 0),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeUser(raw: unknown): User {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    email: String(r.email ?? r.Email ?? ""),
    role: String(r.role ?? r.Role ?? "BIDDER") as Role,
    email_verified: Boolean(r.email_verified ?? r.EmailVerified),
    approval_status: String(r.approval_status ?? r.ApprovalStatus ?? "PENDING") as User["approval_status"],
    kyc_status: String(r.kyc_status ?? r.KYCStatus ?? "NONE") as User["kyc_status"],
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeFeeComponent(raw: unknown): FeeComponent {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    name: String(r.name ?? r.Name ?? ""),
    calculation_type: String(r.calculation_type ?? r.CalculationType ?? "PERCENTAGE") as FeeCalculationType,
    value: Number(r.value ?? r.Value ?? 0),
    applies_to: String(r.applies_to ?? r.AppliesTo ?? "HAMMER") as FeeAppliesTo,
    taxable: Boolean(r.taxable ?? r.Taxable),
    active: Boolean(r.active ?? r.Active),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeFeeAssignment(raw: unknown): FeeAssignment {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    scope_type: String(r.scope_type ?? r.ScopeType ?? "TENANT") as FeeScopeType,
    scope_id: String(r.scope_id ?? r.ScopeID ?? ""),
    fee_component_id: String(r.fee_component_id ?? r.FeeComponentID ?? ""),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!(init?.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  const text = await res.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
  }

  if (!res.ok) {
    const bodyRecord = body && typeof body === "object" ? (body as Record<string, unknown>) : null;
    const message = typeof body === "string" ? body : (bodyRecord?.error as string) || text || "Request failed";
    throw new Error(message);
  }

  return body as T;
}

async function requestWithToken<T>(path: string, bearerToken: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!(init?.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Authorization", `Bearer ${bearerToken}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  const text = await res.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
  }

  if (!res.ok) {
    const bodyRecord = body && typeof body === "object" ? (body as Record<string, unknown>) : null;
    const message = typeof body === "string" ? body : (bodyRecord?.error as string) || text || "Request failed";
    throw new Error(message);
  }

  return body as T;
}

async function requestBinary(path: string, init?: RequestInit): Promise<Blob> {
  const headers = new Headers(init?.headers);
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || "Request failed");
  }
  return res.blob();
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
  getTenantSettings: async () => {
    try {
      return await request<TenantSettings>("/admin/settings");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      if (!msg.includes("404")) throw err;
      const publicCfg = await request<TenantPublicConfig>("/public/config");
      return { default_currency: publicCfg.default_currency || "USD" };
    }
  },
  updateTenantSettings: async (payload: TenantSettings) => {
    try {
      return await request<TenantSettings>("/admin/settings", { method: "PATCH", body: JSON.stringify(payload) });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      if (!msg.includes("404")) throw err;
      try {
        return await request<TenantSettings>("/admin/settings", { method: "POST", body: JSON.stringify(payload) });
      } catch (fallbackErr) {
        const fallbackMsg = fallbackErr instanceof Error ? fallbackErr.message : "";
        if (fallbackMsg.includes("404")) {
          throw new Error("Settings API not available (404). Rebuild/restart the backend API container.");
        }
        throw fallbackErr;
      }
    }
  },
  listAuctions: async () => (await request<unknown[]>("/auctions")).map(normalizeAuction),
  getAuction: async (id: string) => normalizeAuction(await request<unknown>(`/auctions/${id}`)),
  listLotsByAuction: async (auctionId: string) => (await request<unknown[]>(`/auctions/${auctionId}/lots`)).map(normalizeLot),
  getLot: async (lotId: string) => normalizeLot(await request<unknown>(`/lots/${lotId}`)),
  getLotFinancials: (lotId: string) => request<LotFinancials>(`/lots/${lotId}/financials`),
  getInvoice: (invoiceId: string) => request<Invoice>(`/invoices/${invoiceId}`),
  updateInvoiceStatus: (invoiceId: string, status: "ISSUED" | "PAID" | "CANCELLED") =>
    request<JsonRecord>(`/admin/invoices/${invoiceId}/status`, { method: "PATCH", body: JSON.stringify({ status }) }),
  generateLotInvoice: (lotId: string) => request<Invoice>(`/admin/lots/${lotId}/invoice/generate`, { method: "POST" }),
  downloadInvoicePDF: (invoiceId: string) => requestBinary(`/invoices/${invoiceId}/pdf`),
  listLotBids: async (lotId: string) => (await request<unknown[]>(`/lots/${lotId}/bids`)).map(normalizeLotBid),
  listMyWinningLots: async () => (await request<unknown[]>("/me/winnings")).map(normalizeLot),
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
  listBidders: async () => (await request<unknown[]>("/admin/bidders")).map(normalizeUser),
  approveBidder: (id: string) => request<JsonRecord>(`/admin/bidders/${id}/approve`, { method: "PATCH" }),
  rejectBidder: (id: string) => request<JsonRecord>(`/admin/bidders/${id}/reject`, { method: "PATCH" }),
  createLot: async (auctionId: string, payload: JsonRecord) =>
    normalizeLot(await request<unknown>(`/admin/auctions/${auctionId}/lots`, { method: "POST", body: JSON.stringify(payload) })),
  updateLot: async (lotId: string, payload: JsonRecord) =>
    normalizeLot(await request<unknown>(`/admin/lots/${lotId}`, { method: "PATCH", body: JSON.stringify(payload) })),
  deleteLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}`, { method: "DELETE" }),
  reorderLots: (auctionId: string, lotIds: string[]) =>
    request<JsonRecord>(`/admin/auctions/${auctionId}/reorder-lots`, { method: "POST", body: JSON.stringify({ lot_ids: lotIds }) }),
  startLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}/start`, { method: "POST" }),
  endLot: (lotId: string) => request<JsonRecord>(`/admin/lots/${lotId}/end`, { method: "POST" }),
  uploadAuctionImage: async (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<{ url: string }>("/admin/uploads/auction-image", { method: "POST", body: form });
  },
  uploadLotImage: async (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return request<{ url: string }>("/admin/uploads/lot-image", { method: "POST", body: form });
  },
  listFeeComponents: async (activeOnly = true) =>
    (await request<unknown[]>(`/admin/fees/components?active_only=${activeOnly}`)).map(normalizeFeeComponent),
  createFeeComponent: async (payload: {
    name: string;
    calculation_type: FeeCalculationType;
    value: number;
    applies_to: FeeAppliesTo;
    taxable: boolean;
  }) => normalizeFeeComponent(await request<unknown>("/admin/fees/components", { method: "POST", body: JSON.stringify(payload) })),
  updateFeeComponent: async (
    id: string,
    payload: {
      name: string;
      calculation_type: FeeCalculationType;
      value: number;
      applies_to: FeeAppliesTo;
      taxable: boolean;
    },
  ) => normalizeFeeComponent(await request<unknown>(`/admin/fees/components/${id}`, { method: "PATCH", body: JSON.stringify(payload) })),
  deleteFeeComponent: (id: string) => request<JsonRecord>(`/admin/fees/components/${id}`, { method: "DELETE" }),
  listFeeAssignments: async (scopeType: FeeScopeType, scopeID: string) =>
    (await request<unknown[]>(`/admin/fees/assignments?scope_type=${scopeType}&scope_id=${scopeID}`)).map(normalizeFeeAssignment),
  createFeeAssignment: async (payload: {
    scope_type: FeeScopeType;
    scope_id: string;
    fee_component_id: string;
  }) => normalizeFeeAssignment(await request<unknown>("/admin/fees/assignments", { method: "POST", body: JSON.stringify(payload) })),
  deleteFeeAssignment: (id: string) => request<JsonRecord>(`/admin/fees/assignments/${id}`, { method: "DELETE" }),
  platformListTenants: (platformToken: string, status: TenantStatus = "PENDING") =>
    requestWithToken<Array<Record<string, unknown>>>(`/platform/admin/tenants?status=${status}`, platformToken),
  platformApproveTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/approve`, platformToken, { method: "PATCH" }),
  platformRejectTenant: (platformToken: string, tenantId: string) =>
    requestWithToken<JsonRecord>(`/platform/admin/tenants/${tenantId}/reject`, platformToken, { method: "PATCH" }),
};
