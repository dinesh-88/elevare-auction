export type Role = "OWNER" | "ADMIN" | "STAFF" | "ACCOUNTING" | "BIDDER";
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
  amount_minor: number;
  created_at: string;
  bidder_display: string;
}

export interface NotificationItem {
  id: string;
  type: "OUTBIDDED" | "LOT_WON" | "INVOICE_ISSUED";
  title: string;
  body: string;
  link?: string | null;
  metadata?: Record<string, unknown>;
  read_at?: string | null;
  created_at: string;
}

export interface LotAuditActor {
  user_id: string | null;
  display_name: string;
  role: "SYSTEM" | "ADMIN" | "OWNER" | "BIDDER" | "ACCOUNTING" | "STAFF" | string;
}

export interface LotAuditItem {
  id: string;
  created_at: string;
  event_type: string;
  entity_type: string;
  entity_id: string;
  actor: LotAuditActor | null;
  summary: string | null;
  metadata: Record<string, unknown>;
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
  require_bidder_approval?: boolean;
  created_at: string;
}

export interface AuctionBidderAdminRow {
  user_id: string;
  email: string;
  approval_status: "PENDING" | "APPROVED" | "REJECTED";
  tenant_status: "ACTIVE" | "SUSPENDED";
  risk_level: "LOW" | "NORMAL" | "HIGH";
  requested_at: string;
  decided_at?: string | null;
}

export interface TenantBidderProfile {
  id: string;
  tenant_id: string;
  user_id: string;
  phone?: string | null;
  company_name?: string | null;
  address_line1?: string | null;
  address_line2?: string | null;
  city?: string | null;
  postal_code?: string | null;
  country?: string | null;
  status: "ACTIVE" | "SUSPENDED";
  risk_level: "LOW" | "NORMAL" | "HIGH";
  internal_notes?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Lot {
  id: string;
  lot_id?: string;
  auction_id: string;
  lot_number: string;
  order_index: number;
  title: string;
  description: string;
  starting_price: number;
  current_price: number;
  current_price_minor?: number;
  minimum_next_bid_minor?: number;
  reserve_price?: number | null;
  min_increment: number;
  increment_minor?: number;
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
  hammer_price_minor?: number;
  fees_total_minor?: number;
  tax_total_minor?: number;
  total_due_minor?: number;
  payment_status?: "UNPAID" | "PAID" | "CANCELLED";
  paid_at?: string | null;
  cancelled_at?: string | null;
  payment_reference?: string | null;
  cancel_reason?: string | null;
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

export interface AuctionSettlementSummary {
  auction_id: string;
  currency: string;
  invoice_count: number;
  paid_count: number;
  unpaid_count: number;
  cancelled_count: number;
  total_hammer_minor: number;
  total_fees_minor: number;
  total_tax_minor: number;
  total_due_minor: number;
  total_paid_minor: number;
  outstanding_minor: number;
}

export interface SettlementInvoiceRow {
  id: string;
  invoice_number: string;
  bidder_name: string;
  bidder_email: string;
  lot_summary: string;
  hammer_minor: number;
  fees_minor: number;
  tax_minor: number;
  total_due_minor: number;
  currency: string;
  payment_status: "UNPAID" | "PAID" | "CANCELLED";
  issued_at: string;
  paid_at?: string | null;
  payment_reference?: string | null;
  cancel_reason?: string | null;
}

function normalizeSettlementSummary(raw: unknown): AuctionSettlementSummary {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    auction_id: String(r.auction_id ?? r.AuctionID ?? ""),
    currency: String(r.currency ?? r.Currency ?? "USD"),
    invoice_count: Number(r.invoice_count ?? r.InvoiceCount ?? 0),
    paid_count: Number(r.paid_count ?? r.PaidCount ?? 0),
    unpaid_count: Number(r.unpaid_count ?? r.UnpaidCount ?? 0),
    cancelled_count: Number(r.cancelled_count ?? r.CancelledCount ?? 0),
    total_hammer_minor: Number(r.total_hammer_minor ?? r.TotalHammerMinor ?? 0),
    total_fees_minor: Number(r.total_fees_minor ?? r.TotalFeesMinor ?? 0),
    total_tax_minor: Number(r.total_tax_minor ?? r.TotalTaxMinor ?? 0),
    total_due_minor: Number(r.total_due_minor ?? r.TotalDueMinor ?? 0),
    total_paid_minor: Number(r.total_paid_minor ?? r.TotalPaidMinor ?? 0),
    outstanding_minor: Number(r.outstanding_minor ?? r.OutstandingMinor ?? 0),
  };
}

function normalizeNotificationItem(raw: unknown): NotificationItem {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    type: String(r.type ?? r.Type ?? "OUTBIDDED") as NotificationItem["type"],
    title: String(r.title ?? r.Title ?? ""),
    body: String(r.body ?? r.Body ?? ""),
    link: (r.link ?? r.Link ?? null) as string | null,
    metadata: (r.metadata ?? r.Metadata ?? {}) as Record<string, unknown>,
    read_at: (r.read_at ?? r.ReadAt ?? null) as string | null,
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeSettlementInvoiceRow(raw: unknown): SettlementInvoiceRow {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    invoice_number: String(r.invoice_number ?? r.InvoiceNumber ?? ""),
    bidder_name: String(r.bidder_name ?? r.BidderName ?? ""),
    bidder_email: String(r.bidder_email ?? r.BidderEmail ?? ""),
    lot_summary: String(r.lot_summary ?? r.LotSummary ?? ""),
    hammer_minor: Number(r.hammer_minor ?? r.HammerMinor ?? 0),
    fees_minor: Number(r.fees_minor ?? r.FeesMinor ?? 0),
    tax_minor: Number(r.tax_minor ?? r.TaxMinor ?? 0),
    total_due_minor: Number(r.total_due_minor ?? r.TotalDueMinor ?? 0),
    currency: String(r.currency ?? r.Currency ?? "USD"),
    payment_status: String(r.payment_status ?? r.PaymentStatus ?? "UNPAID") as SettlementInvoiceRow["payment_status"],
    issued_at: String(r.issued_at ?? r.IssuedAt ?? ""),
    paid_at: (r.paid_at ?? r.PaidAt ?? null) as string | null,
    payment_reference: (r.payment_reference ?? r.PaymentReference ?? null) as string | null,
    cancel_reason: (r.cancel_reason ?? r.CancelReason ?? null) as string | null,
  };
}

function normalizeLotAuditItem(raw: unknown): LotAuditItem {
  const r = (raw ?? {}) as Record<string, unknown>;
  const actorRaw = (r.actor ?? null) as Record<string, unknown> | null;
  return {
    id: String(r.id ?? r.ID ?? ""),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
    event_type: String(r.event_type ?? r.eventType ?? r.action ?? r.Action ?? ""),
    entity_type: String(r.entity_type ?? r.entityType ?? r.EntityType ?? "LOT"),
    entity_id: String(r.entity_id ?? r.entityId ?? r.EntityID ?? ""),
    actor: actorRaw
      ? {
          user_id: actorRaw.user_id ? String(actorRaw.user_id) : null,
          display_name: String(actorRaw.display_name ?? "System"),
          role: String(actorRaw.role ?? "SYSTEM"),
        }
      : null,
    summary: r.summary == null ? null : String(r.summary),
    metadata: (r.metadata && typeof r.metadata === "object" ? r.metadata : {}) as Record<string, unknown>,
  };
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
    favicon_url?: string;
    font_family?: string;
  };
  content?: Record<string, unknown>;
  content_version?: number;
  content_schema_version?: number;
  require_admin_approval: boolean;
  require_kyc: boolean;
}

export interface TenantSettings {
  default_currency: string;
}

export interface TenantSitePayload {
  branding: Record<string, unknown>;
  content: Record<string, unknown>;
}

export interface TenantSiteState {
  branding_published: Record<string, unknown>;
  content_published: Record<string, unknown>;
  branding_draft: Record<string, unknown>;
  content_draft: Record<string, unknown>;
  content_version: number;
  content_schema_version: number;
}

export interface TenantAsset {
  id: string;
  tenant_id: string;
  kind: "logo" | "banner" | "image";
  url: string;
  width?: number;
  height?: number;
  size_bytes?: number;
  mime?: string;
  created_at: string;
}

export interface TenantLoginDiscoveryItem {
  tenant_id: string;
  tenant_name: string;
  host: string;
  login_url: string;
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
    require_bidder_approval: Boolean(r.require_bidder_approval ?? r.RequireBidderApproval ?? false),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
  };
}

function normalizeAuctionBidderAdminRow(raw: unknown): AuctionBidderAdminRow {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    user_id: String(r.user_id ?? r.UserID ?? ""),
    email: String(r.email ?? r.Email ?? ""),
    approval_status: String(r.approval_status ?? r.ApprovalStatus ?? "PENDING") as AuctionBidderAdminRow["approval_status"],
    tenant_status: String(r.tenant_status ?? r.TenantStatus ?? "ACTIVE") as AuctionBidderAdminRow["tenant_status"],
    risk_level: String(r.risk_level ?? r.RiskLevel ?? "NORMAL") as AuctionBidderAdminRow["risk_level"],
    requested_at: String(r.requested_at ?? r.RequestedAt ?? ""),
    decided_at: (r.decided_at ?? r.DecidedAt ?? null) as string | null,
  };
}

function normalizeTenantBidderProfile(raw: unknown): TenantBidderProfile {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    user_id: String(r.user_id ?? r.UserID ?? ""),
    phone: (r.phone ?? r.Phone ?? null) as string | null,
    company_name: (r.company_name ?? r.CompanyName ?? null) as string | null,
    address_line1: (r.address_line1 ?? r.AddressLine1 ?? null) as string | null,
    address_line2: (r.address_line2 ?? r.AddressLine2 ?? null) as string | null,
    city: (r.city ?? r.City ?? null) as string | null,
    postal_code: (r.postal_code ?? r.PostalCode ?? null) as string | null,
    country: (r.country ?? r.Country ?? null) as string | null,
    status: String(r.status ?? r.Status ?? "ACTIVE") as TenantBidderProfile["status"],
    risk_level: String(r.risk_level ?? r.RiskLevel ?? "NORMAL") as TenantBidderProfile["risk_level"],
    internal_notes: (r.internal_notes ?? r.InternalNotes ?? null) as string | null,
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
    updated_at: String(r.updated_at ?? r.UpdatedAt ?? ""),
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
    lot_id: String(r.lot_id ?? r.lotID ?? r.id ?? r.ID ?? ""),
    auction_id: String(r.auction_id ?? r.AuctionID ?? ""),
    lot_number: String(r.lot_number ?? r.LotNumber ?? ""),
    order_index: Number(r.order_index ?? r.OrderIndex ?? 0),
    title: String(r.title ?? r.Title ?? ""),
    description: String(r.description ?? r.Description ?? ""),
    starting_price: Number(r.starting_price ?? r.StartingPrice ?? 0),
    current_price: Number(r.current_price ?? r.CurrentPrice ?? 0),
    current_price_minor: Number(r.current_price_minor ?? r.CurrentMinor ?? r.current_price ?? r.CurrentPrice ?? 0),
    minimum_next_bid_minor: Number(
      r.minimum_next_bid_minor ?? r.MinimumNext ?? Number(r.current_price ?? r.CurrentPrice ?? 0) + Number(r.min_increment ?? r.MinIncrement ?? 0),
    ),
    reserve_price: (r.reserve_price ?? r.ReservePrice ?? null) as number | null,
    min_increment: Number(r.min_increment ?? r.MinIncrement ?? 0),
    increment_minor: Number(r.increment_minor ?? r.Increment ?? r.min_increment ?? r.MinIncrement ?? 0),
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
    amount_minor: Number(r.amount_minor ?? r.AmountMinor ?? r.amount ?? r.Amount ?? 0),
    created_at: String(r.created_at ?? r.CreatedAt ?? ""),
    bidder_display: String(r.bidder_display ?? r.BidderDisplay ?? "Bidder"),
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

function normalizeSiteState(raw: unknown): TenantSiteState {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    branding_published: ((r.branding_published ?? {}) as Record<string, unknown>) ?? {},
    content_published: ((r.content_published ?? {}) as Record<string, unknown>) ?? {},
    branding_draft: ((r.branding_draft ?? {}) as Record<string, unknown>) ?? {},
    content_draft: ((r.content_draft ?? {}) as Record<string, unknown>) ?? {},
    content_version: Number(r.content_version ?? 1),
    content_schema_version: Number(r.content_schema_version ?? 1),
  };
}

function normalizeTenantAsset(raw: unknown): TenantAsset {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    id: String(r.id ?? r.ID ?? ""),
    tenant_id: String(r.tenant_id ?? r.TenantID ?? ""),
    kind: String(r.kind ?? r.Kind ?? "image") as TenantAsset["kind"],
    url: resolveMediaUrl(String(r.url ?? r.URL ?? "")),
    width: r.width != null ? Number(r.width) : undefined,
    height: r.height != null ? Number(r.height) : undefined,
    size_bytes: r.size_bytes != null ? Number(r.size_bytes) : undefined,
    mime: r.mime != null ? String(r.mime) : undefined,
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
  const method = (init?.method ?? "GET").toUpperCase();
  if ((method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE") && !headers.has("X-CSRF-Token")) {
    const csrfToken = readCookie("csrf_token");
    if (csrfToken) headers.set("X-CSRF-Token", csrfToken);
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers, credentials: "include" });
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

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers, credentials: "include" });
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
  const method = (init?.method ?? "GET").toUpperCase();
  if ((method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE") && !headers.has("X-CSRF-Token")) {
    const csrfToken = readCookie("csrf_token");
    if (csrfToken) headers.set("X-CSRF-Token", csrfToken);
  }
  const res = await fetch(`${API_BASE}${path}`, { ...init, headers, credentials: "include" });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || "Request failed");
  }
  return res.blob();
}

export const api = {
  discoverTenantLogin: async (email: string) => {
    const response = await request<{ items: TenantLoginDiscoveryItem[] }>("/auth/discover-tenant", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
    return response.items ?? [];
  },
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
  getSite: async () => normalizeSiteState(await request<unknown>("/admin/site")),
  updateSiteDraft: (payload: TenantSitePayload) =>
    request<{ status: string }>("/admin/site/draft", { method: "PUT", body: JSON.stringify(payload) }),
  publishSite: () =>
    request<{ status: string; before_version: number; after_version: number }>("/admin/site/publish", { method: "POST" }),
  listAuctions: async () => (await request<unknown[]>("/auctions")).map(normalizeAuction),
  getAuction: async (id: string) => normalizeAuction(await request<unknown>(`/auctions/${id}`)),
  listLotsByAuction: async (auctionId: string) => (await request<unknown[]>(`/auctions/${auctionId}/lots`)).map(normalizeLot),
  getLot: async (lotId: string) => normalizeLot(await request<unknown>(`/lots/${lotId}`)),
  getLotFinancials: (lotId: string) => request<LotFinancials>(`/lots/${lotId}/financials`),
  getInvoice: (invoiceId: string) => request<Invoice>(`/invoices/${invoiceId}`),
  listMyNotifications: async (params?: { limit?: number; cursor?: string }) => {
    const search = new URLSearchParams();
    search.set("limit", String(params?.limit ?? 20));
    if (params?.cursor) search.set("cursor", params.cursor);
    const res = await request<{ items: unknown[]; next_cursor: string }>(`/me/notifications?${search.toString()}`);
    return {
      items: (res.items ?? []).map(normalizeNotificationItem),
      next_cursor: String(res.next_cursor ?? ""),
    };
  },
  listLotAuditLogs: async (
    lotId: string,
    params?: {
      limit?: number;
      cursor?: string;
      event_type?: string;
      actor_user_id?: string;
      from?: string;
      to?: string;
    },
  ) => {
    const search = new URLSearchParams();
    search.set("limit", String(params?.limit ?? 20));
    if (params?.cursor) search.set("cursor", params.cursor);
    if (params?.event_type) search.set("event_type", params.event_type);
    if (params?.actor_user_id) search.set("actor_user_id", params.actor_user_id);
    if (params?.from) search.set("from", params.from);
    if (params?.to) search.set("to", params.to);

    const response = await fetch(`${API_BASE}/admin/lots/${lotId}/audit?${search.toString()}`, {
      credentials: "include",
      headers: (() => {
        const headers = new Headers();
        const token = getToken();
        if (token) headers.set("Authorization", `Bearer ${token}`);
        return headers;
      })(),
    });
    const text = await response.text();
    let body: unknown = null;
    if (text) {
      try {
        body = JSON.parse(text);
      } catch {
        body = text;
      }
    }
    if (!response.ok) {
      const bodyRecord = body && typeof body === "object" ? (body as Record<string, unknown>) : null;
      const message = typeof body === "string" ? body : (bodyRecord?.error as string) || text || "Request failed";
      const err = new Error(message) as Error & { status?: number };
      err.status = response.status;
      throw err;
    }
    const res = (body ?? {}) as { items?: unknown[]; next_cursor?: string | null };
    return {
      items: (res.items ?? []).map(normalizeLotAuditItem),
      next_cursor: res.next_cursor ? String(res.next_cursor) : null,
    };
  },
  markMyNotificationRead: (notificationId: string) => request<{ status: string }>(`/me/notifications/${notificationId}/read`, { method: "POST" }),
  markAllMyNotificationsRead: () => request<{ status: string }>("/me/notifications/read-all", { method: "POST" }),
  getAuctionSettlementSummary: async (auctionId: string) =>
    normalizeSettlementSummary(await request<unknown>(`/admin/auctions/${auctionId}/settlement-summary`)),
  listAuctionInvoices: (auctionId: string, params?: { status?: "UNPAID" | "PAID" | "CANCELLED"; search?: string; limit?: number; offset?: number }) => {
    const search = new URLSearchParams();
    if (params?.status) search.set("status", params.status);
    if (params?.search) search.set("search", params.search);
    search.set("limit", String(params?.limit ?? 50));
    search.set("offset", String(params?.offset ?? 0));
    return request<{ items: unknown[]; total: number; limit: number; offset: number }>(`/admin/auctions/${auctionId}/invoices?${search.toString()}`).then(
      (res) => ({
        ...res,
        items: (res.items ?? []).map(normalizeSettlementInvoiceRow),
      }),
    );
  },
  markInvoicePaid: (invoiceId: string, paymentReference?: string) =>
    request<Invoice>(`/admin/invoices/${invoiceId}/mark-paid`, {
      method: "POST",
      body: JSON.stringify({ payment_reference: paymentReference || undefined }),
    }),
  cancelInvoice: (invoiceId: string, reason: string) =>
    request<Invoice>(`/admin/invoices/${invoiceId}/cancel`, {
      method: "POST",
      body: JSON.stringify({ reason }),
    }),
  downloadSettlementCSV: (auctionId: string, params?: { status?: "UNPAID" | "PAID" | "CANCELLED"; search?: string }) => {
    const search = new URLSearchParams();
    if (params?.status) search.set("status", params.status);
    if (params?.search) search.set("search", params.search);
    const qs = search.toString();
    return requestBinary(`/admin/auctions/${auctionId}/export-settlement-csv${qs ? `?${qs}` : ""}`);
  },
  updateInvoiceStatus: (invoiceId: string, status: "ISSUED" | "PAID" | "CANCELLED") =>
    request<JsonRecord>(`/admin/invoices/${invoiceId}/status`, { method: "PATCH", body: JSON.stringify({ status }) }),
  generateLotInvoice: (lotId: string) => request<Invoice>(`/admin/lots/${lotId}/invoice/generate`, { method: "POST" }),
  downloadInvoicePDF: (invoiceId: string) => requestBinary(`/invoices/${invoiceId}/pdf`),
  listLotBids: async (lotId: string) => (await request<unknown[]>(`/lots/${lotId}/bids`)).map(normalizeLotBid),
  listMyWinningLots: async () => (await request<unknown[]>("/me/winnings")).map(normalizeLot),
  register: (payload: { email: string; password: string }) =>
    request<JsonRecord>("/auth/register", { method: "POST", body: JSON.stringify(payload) }),
  login: (payload: { email: string; password: string }) =>
    request<{ token?: string; user: User }>("/auth/login", { method: "POST", body: JSON.stringify(payload) }),
  getSession: () => request<{ user: User }>("/auth/session"),
  logout: () => request<{ status: string }>("/auth/logout", { method: "POST" }),
  verifyEmail: (userId: string) =>
    request<JsonRecord>("/auth/verify-email", { method: "POST", body: JSON.stringify({ user_id: userId }) }),
  placeBid: async (..._args: unknown[]) => {
    throw new Error("Auction-level bidding has been removed. Use lot bidding.");
  },
  placeLotBid: (
    lotId: string,
    payload: { bid_amount_minor?: number | null; max_bid_minor?: number | null; amount?: number | null },
  ) => request<JsonRecord>(`/lots/${lotId}/bid`, { method: "POST", body: JSON.stringify(payload) }),
  requestToBid: (auctionId: string) => request<{ auction_id: string; user_id: string; status: "PENDING" | "APPROVED" | "REJECTED" }>(`/auctions/${auctionId}/request-to-bid`, { method: "POST" }),
  createAuction: async (payload: JsonRecord) =>
    normalizeAuction(await request<unknown>("/admin/auctions", { method: "POST", body: JSON.stringify(payload) })),
  listAdminAuctions: async () => (await request<unknown[]>("/admin/auctions")).map(normalizeAuction),
  updateAuction: async (id: string, payload: JsonRecord) =>
    normalizeAuction(await request<unknown>(`/admin/auctions/${id}`, { method: "PATCH", body: JSON.stringify(payload) })),
  deleteAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}`, { method: "DELETE" }),
  startAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/start`, { method: "PATCH" }),
  pauseAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/pause`, { method: "PATCH" }),
  endAuction: (id: string) => request<JsonRecord>(`/admin/auctions/${id}/end`, { method: "PATCH" }),
  updateAuctionBidderSettings: (auctionId: string, requireBidderApproval: boolean) =>
    request<{ auction_id: string; require_bidder_approval: boolean }>(`/admin/auctions/${auctionId}/settings`, {
      method: "PUT",
      body: JSON.stringify({ require_bidder_approval: requireBidderApproval }),
    }),
  listAuctionBidders: async (auctionId: string, status?: "PENDING" | "APPROVED" | "REJECTED") => {
    const params = new URLSearchParams();
    if (status) params.set("status", status);
    const qs = params.toString();
    return (await request<unknown[]>(`/admin/auctions/${auctionId}/bidders${qs ? `?${qs}` : ""}`)).map(normalizeAuctionBidderAdminRow);
  },
  approveAuctionBidder: async (auctionId: string, userId: string) =>
    request<JsonRecord>(`/admin/auctions/${auctionId}/bidders/${userId}/approve`, { method: "POST" }),
  rejectAuctionBidder: async (auctionId: string, userId: string) =>
    request<JsonRecord>(`/admin/auctions/${auctionId}/bidders/${userId}/reject`, { method: "POST" }),
  getBidderProfile: async (userId: string) => normalizeTenantBidderProfile(await request<unknown>(`/admin/bidders/${userId}`)),
  updateBidderProfile: async (
    userId: string,
    payload: {
      status?: "ACTIVE" | "SUSPENDED";
      risk_level?: "LOW" | "NORMAL" | "HIGH";
      internal_notes?: string | null;
      phone?: string | null;
      company_name?: string | null;
      address_line1?: string | null;
      address_line2?: string | null;
      city?: string | null;
      postal_code?: string | null;
      country?: string | null;
    },
  ) => normalizeTenantBidderProfile(await request<unknown>(`/admin/bidders/${userId}`, { method: "PUT", body: JSON.stringify(payload) })),
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
  uploadAsset: async (kind: "logo" | "banner" | "image", file: File) => {
    const form = new FormData();
    form.append("file", file);
    return normalizeTenantAsset(await request<unknown>(`/admin/assets/upload?kind=${kind}`, { method: "POST", body: form }));
  },
  listAssets: async (kind?: "logo" | "banner" | "image", limit = 100) => {
    const params = new URLSearchParams();
    if (kind) params.set("kind", kind);
    params.set("limit", String(limit));
    const qs = params.toString();
    return (await request<unknown[]>(`/admin/assets${qs ? `?${qs}` : ""}`)).map(normalizeTenantAsset);
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

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = document.cookie.match(new RegExp(`(?:^|; )${escaped}=([^;]*)`));
  if (!match) return null;
  try {
    return decodeURIComponent(match[1]);
  } catch {
    return match[1];
  }
}
