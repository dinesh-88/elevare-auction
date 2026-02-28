export const LOCALES = ["en", "fr", "de"] as const;

export type AppLocale = (typeof LOCALES)[number];

export const DEFAULT_LOCALE: AppLocale = "en";

export function isLocale(value: string | null | undefined): value is AppLocale {
  return value === "en" || value === "fr" || value === "de";
}

export function normalizeLocale(value: string | null | undefined): AppLocale {
  return isLocale(value) ? value : DEFAULT_LOCALE;
}

export function getLocaleFromPathname(pathname: string | null | undefined): AppLocale {
  if (!pathname) return DEFAULT_LOCALE;
  const segment = pathname.split("/").filter(Boolean)[0];
  return normalizeLocale(segment);
}

export function stripLocalePrefix(pathname: string): string {
  const parts = pathname.split("/");
  const first = parts[1];
  if (!isLocale(first)) return pathname || "/";
  const next = `/${parts.slice(2).join("/")}`;
  return next === "/" ? "/" : next.replace(/\/+$/, "") || "/";
}

export function localizePath(path: string, locale: AppLocale): string {
  if (!path) return `/${locale}`;
  if (/^(https?:|mailto:|tel:)/.test(path) || path.startsWith("#")) {
    return path;
  }
  const [pathname, search = ""] = path.split("?");
  const normalizedPath = pathname.startsWith("/") ? pathname : `/${pathname}`;
  const stripped = stripLocalePrefix(normalizedPath);
  const localized = stripped === "/" ? `/${locale}` : `/${locale}${stripped}`;
  return search ? `${localized}?${search}` : localized;
}
