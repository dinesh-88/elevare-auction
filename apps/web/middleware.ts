import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

const DEFAULT_MARKETING_HOSTS = [
  "elevareauctions.com",
  "www.elevareauctions.com",
  "login.elevareauctions.com",
  "localhost",
  "127.0.0.1"
];

function normalizeHost(request: NextRequest): string {
  const forwardedHost = request.headers.get("x-forwarded-host");
  const host = (forwardedHost ?? request.headers.get("host") ?? "").toLowerCase().trim();
  return host.split(":")[0];
}

function getMarketingHosts(): Set<string> {
  const raw = process.env.MARKETING_HOSTS ?? "";
  const configured = raw
    .split(",")
    .map((value) => value.trim().toLowerCase())
    .filter(Boolean);
  return new Set(configured.length > 0 ? configured : DEFAULT_MARKETING_HOSTS);
}

function isTenantHost(host: string): boolean {
  if (!host) return false;
  if (host.startsWith("api.")) return false;
  return !getMarketingHosts().has(host);
}

function getLoginHubHost(): string {
  return (process.env.NEXT_PUBLIC_LOGIN_HUB_HOST ?? "login.elevareauctions.com").toLowerCase().trim();
}

export function middleware(request: NextRequest) {
  const host = normalizeHost(request);
  const { pathname } = request.nextUrl;

  if (host === getLoginHubHost()) {
    if (pathname === "/") {
      const url = request.nextUrl.clone();
      url.pathname = "/login";
      return NextResponse.redirect(url);
    }
    if (pathname === "/login") {
      const url = request.nextUrl.clone();
      url.pathname = "/login-hub";
      return NextResponse.rewrite(url);
    }
  }

  if (pathname === "/" && isTenantHost(host)) {
    const url = request.nextUrl.clone();
    url.pathname = "/auctions";
    return NextResponse.rewrite(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/login"]
};
