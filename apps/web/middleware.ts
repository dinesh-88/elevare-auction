import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { DEFAULT_LOCALE, LOCALES, isLocale, normalizeLocale } from "./src/i18n/config";

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

function getAccessTokenCookieName(): string {
  return (process.env.ACCESS_TOKEN_COOKIE ?? "access_token").trim() || "access_token";
}

function decodeJwtLocale(token: string | undefined): string | null {
  if (!token) return null;
  const [, payload] = token.split(".");
  if (!payload) return null;
  try {
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    const json = JSON.parse(atob(base64)) as { locale?: string };
    return typeof json.locale === "string" ? normalizeLocale(json.locale) : null;
  } catch {
    return null;
  }
}

function detectPathLocale(pathname: string): string | null {
  const first = pathname.split("/").filter(Boolean)[0];
  return isLocale(first) ? first : null;
}

function stripLocale(pathname: string): string {
  const locale = detectPathLocale(pathname);
  if (!locale) return pathname;
  const stripped = pathname.slice(locale.length + 1);
  if (!stripped) return "/";
  return stripped.startsWith("/") ? stripped : `/${stripped}`;
}

function localize(pathname: string, locale: string): string {
  const stripped = stripLocale(pathname);
  return stripped === "/" ? `/${locale}` : `/${locale}${stripped}`;
}

function inferLocaleFromHeaders(request: NextRequest): string {
  const value = request.headers.get("accept-language") ?? "";
  const matched = value
    .split(",")
    .map((part) => part.trim().split(";")[0]?.toLowerCase())
    .find((part) => part && LOCALES.some((locale) => part === locale || part.startsWith(`${locale}-`)));
  if (!matched) return DEFAULT_LOCALE;
  return normalizeLocale(matched.split("-")[0]);
}

function shouldBypass(pathname: string): boolean {
  return pathname.startsWith("/_next") || pathname.startsWith("/api") || pathname.startsWith("/uploads/") || /\.[^/]+$/.test(pathname);
}

export function middleware(request: NextRequest) {
  const host = normalizeHost(request);
  const { pathname } = request.nextUrl;

  if (shouldBypass(pathname)) {
    return NextResponse.next();
  }

  const tokenLocale = decodeJwtLocale(request.cookies.get(getAccessTokenCookieName())?.value);
  const pathLocale = detectPathLocale(pathname);
  const preferredLocale = tokenLocale ?? pathLocale ?? inferLocaleFromHeaders(request);
  const barePathname = stripLocale(pathname);

  if (!pathLocale) {
    const url = request.nextUrl.clone();
    url.pathname = localize(pathname, preferredLocale);
    return NextResponse.redirect(url);
  }

  if (tokenLocale && pathLocale !== tokenLocale) {
    const url = request.nextUrl.clone();
    url.pathname = localize(pathname, tokenLocale);
    return NextResponse.redirect(url);
  }

  if (host === getLoginHubHost()) {
    if (barePathname === "/") {
      const url = request.nextUrl.clone();
      url.pathname = localize("/login", preferredLocale);
      return NextResponse.redirect(url);
    }
    if (barePathname === "/login") {
      const url = request.nextUrl.clone();
      url.pathname = "/login-hub";
      return NextResponse.rewrite(url);
    }
  }

  if (barePathname === "/" && isTenantHost(host)) {
    const url = request.nextUrl.clone();
    url.pathname = "/auctions";
    return NextResponse.rewrite(url);
  }

  const url = request.nextUrl.clone();
  url.pathname = barePathname;
  return NextResponse.rewrite(url);
}

export const config = {
  matcher: ["/((?!_next|.*\\..*).*)"]
};
