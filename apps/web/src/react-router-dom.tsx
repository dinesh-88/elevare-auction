"use client";

import React, { useEffect } from "react";
import NextLink from "next/link";
import { usePathname, useRouter, useParams as useNextParams } from "next/navigation";
import { getLocaleFromPathname, localizePath } from "./i18n/config";

type LinkProps = Omit<React.ComponentPropsWithoutRef<"a">, "href"> & { to: string; replace?: boolean };

export function Link({ to, children, ...rest }: LinkProps) {
  const pathname = usePathname();
  const locale = getLocaleFromPathname(pathname);
  return (
    <NextLink href={localizePath(to, locale)} {...rest}>
      {children}
    </NextLink>
  );
}

export function useNavigate() {
  const router = useRouter();
  const pathname = usePathname();
  const locale = getLocaleFromPathname(pathname);
  return (to: string, options?: { replace?: boolean }) => {
    const target = localizePath(to, locale);
    if (options?.replace) {
      router.replace(target);
      return;
    }
    router.push(target);
  };
}

export function useLocation() {
  const pathname = usePathname();
  return { pathname };
}

export function useParams<T extends Record<string, string>>() {
  const params = useNextParams() ?? {};
  const normalized: Record<string, string> = {};
  Object.entries(params).forEach(([key, value]) => {
    normalized[key] = Array.isArray(value) ? value[0] ?? "" : value ?? "";
  });
  return normalized as T;
}

export function Navigate({ to, replace = false }: { to: string; replace?: boolean }) {
  const router = useRouter();
  const pathname = usePathname();
  const locale = getLocaleFromPathname(pathname);
  useEffect(() => {
    const target = localizePath(to, locale);
    if (replace) {
      router.replace(target);
      return;
    }
    router.push(target);
  }, [locale, replace, router, to]);
  return null;
}

export function Outlet() {
  return null;
}

export function BrowserRouter({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}

export function Routes({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}

export function Route(props: any) {
  return <>{props?.element ?? props?.children ?? null}</>;
}
