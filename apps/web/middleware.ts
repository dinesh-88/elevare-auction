import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

function normalizeHost(request: NextRequest): string {
  const forwardedHost = request.headers.get("x-forwarded-host");
  const host = (forwardedHost ?? request.headers.get("host") ?? "").toLowerCase().trim();
  return host.split(":")[0];
}

function isTenantHost(host: string): boolean {
  if (!host) return false;
  if (host === "elevareauctions.com" || host === "www.elevareauctions.com") return false;
  if (host === "localhost" || host === "127.0.0.1") return false;
  if (host.startsWith("api.")) return false;
  if (host.endsWith(".elevareauctions.com")) return true;
  if (host.endsWith(".localhost")) return true;
  return false;
}

export function middleware(request: NextRequest) {
  const host = normalizeHost(request);
  const { pathname } = request.nextUrl;

  if (pathname === "/" && isTenantHost(host)) {
    const url = request.nextUrl.clone();
    url.pathname = "/auctions";
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/"]
};
