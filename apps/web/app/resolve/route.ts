import { NextRequest, NextResponse } from "next/server";

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const WINDOW_MS = 5 * 60 * 1000;
const LIMIT = 8;
const MIN_RESPONSE_MS = 350;
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type RateLimitBucket = {
  count: number;
  resetAt: number;
};

type TenantResolveCandidate = {
  tenant_id?: string;
  tenant_name?: string;
  slug?: string;
  host?: string;
  login_url?: string;
};

type ResolveErrorBody = {
  error: string;
  message: string;
};

const rateLimitBuckets = new Map<string, RateLimitBucket>();

function resolveApiBase() {
  const fromServer = process.env.API_BASE_URL?.trim();
  if (fromServer) return trimTrailingSlash(fromServer);
  const fromPublic = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (fromPublic) return trimTrailingSlash(fromPublic);
  return "";
}

function resolveApiBaseFromRequest(request: NextRequest) {
  const configured = resolveApiBase();
  if (configured) return configured;

  const forwardedHost = request.headers.get("x-forwarded-host")?.trim();
  const rawHost = forwardedHost || request.headers.get("host")?.trim() || "";
  const host = rawHost.split(",")[0]?.trim() || "";
  const protocol = request.headers.get("x-forwarded-proto")?.trim() || "https";

  if (!host) {
    return "http://localhost:8080";
  }

  const hostname = host.split(":")[0] ?? host;
  if (hostname === "localhost" || hostname === "127.0.0.1") {
    return "http://localhost:8080";
  }
  if (hostname.endsWith(".localhost")) {
    return `${protocol}://api.localhost:8080`;
  }

  const parts = hostname.split(".");
  if (parts.length >= 2) {
    parts[0] = "api";
    return `${protocol}://${parts.join(".")}`;
  }

  return `${protocol}://api.${hostname}`;
}

function resolveTenantBase(slug: string) {
  const scheme = (process.env.NEXT_PUBLIC_TENANT_APP_SCHEME ?? "https").trim() || "https";
  const baseDomain = (process.env.NEXT_PUBLIC_TENANT_APP_BASE_DOMAIN ?? "elevareauctions.com").trim() || "elevareauctions.com";
  return `${scheme}://${slug}.${baseDomain}`;
}

function buildTenantLoginUrl(slug: string, email: string) {
  const url = new URL(`${resolveTenantBase(slug)}/login`);
  url.searchParams.set("email", email);
  return url.toString();
}

function sleep(ms: number) {
  if (ms <= 0) return Promise.resolve();
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function clientKey(request: NextRequest) {
  const forwardedFor = request.headers.get("x-forwarded-for");
  if (forwardedFor) {
    const first = forwardedFor.split(",")[0]?.trim();
    if (first) return first;
  }
  return request.headers.get("x-real-ip")?.trim() || "unknown";
}

function isRateLimited(key: string, nowMs: number) {
  const current = rateLimitBuckets.get(key);
  if (!current || current.resetAt <= nowMs) {
    rateLimitBuckets.set(key, { count: 1, resetAt: nowMs + WINDOW_MS });
    return false;
  }
  current.count += 1;
  if (current.count > LIMIT) {
    rateLimitBuckets.set(key, current);
    return true;
  }
  rateLimitBuckets.set(key, current);
  return false;
}

async function fetchCandidates(request: NextRequest, email: string) {
  const response = await fetch(`${resolveApiBaseFromRequest(request)}/auth/discover-tenant`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ email }),
    cache: "no-store",
  });

  const text = await response.text();
  let payload: { items?: TenantResolveCandidate[]; error?: string } | null = null;
  if (text) {
    try {
      payload = JSON.parse(text) as { items?: TenantResolveCandidate[]; error?: string };
    } catch {
      payload = { error: text };
    }
  }

  if (!response.ok) {
    const safeMessage =
      response.status === 400
        ? "Enter a valid email address."
        : response.status === 404
          ? "No associated institutions found."
          : response.status === 429
            ? "Too many requests. Please try again shortly."
            : "Unable to resolve your institution right now. Please try again shortly.";
    throw new Error(safeMessage);
  }

  return Array.isArray(payload?.items) ? payload!.items : [];
}

export async function POST(request: NextRequest) {
  const startedAt = Date.now();
  const finish = async (response: NextResponse) => {
    await sleep(MIN_RESPONSE_MS - (Date.now() - startedAt));
    return response;
  };

  if (isRateLimited(clientKey(request), startedAt)) {
    return finish(
      NextResponse.json<ResolveErrorBody>(
        { error: "rate_limited", message: "Too many requests. Please try again shortly." },
        { status: 429 },
      ),
    );
  }

  const payload = (await request.json().catch(() => ({}))) as { email?: string };
  const email = payload.email?.trim().toLowerCase() ?? "";
  if (!emailPattern.test(email)) {
    return finish(
      NextResponse.json<ResolveErrorBody>(
        { error: "validation_error", message: "Enter a valid email address." },
        { status: 400 },
      ),
    );
  }

  let items: TenantResolveCandidate[] = [];
  try {
    items = await fetchCandidates(request, email);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unable to resolve your institution right now. Please try again shortly.";
    const code =
      message === "Enter a valid email address."
        ? "validation_error"
        : message === "No associated institutions found."
          ? "not_found"
          : message === "Too many requests. Please try again shortly."
            ? "rate_limited"
            : "internal_error";
    const status =
      message === "Enter a valid email address."
        ? 400
        : message === "No associated institutions found."
          ? 404
          : message === "Too many requests. Please try again shortly."
            ? 429
            : 502;
    return finish(NextResponse.json<ResolveErrorBody>({ error: code, message }, { status }));
  }
  const normalized = items
    .map((item) => ({
      slug: String(item.slug ?? "").trim().toLowerCase(),
      name: String(item.tenant_name ?? "").trim(),
    }))
    .filter((item) => item.slug && item.name);

  if (normalized.length === 0) {
    return finish(
      NextResponse.json<ResolveErrorBody>(
        { error: "not_found", message: "No associated institutions found." },
        { status: 404 },
      ),
    );
  }

  if (normalized.length === 1) {
    return finish(
      NextResponse.json({
        redirect_url: buildTenantLoginUrl(normalized[0].slug, email),
      }),
    );
  }

  return finish(
    NextResponse.json({
      tenants: normalized.map((item) => ({ slug: item.slug, name: item.name })),
    }),
  );
}
