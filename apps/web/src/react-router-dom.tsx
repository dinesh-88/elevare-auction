"use client";

import React, { useEffect } from "react";
import NextLink from "next/link";
import { usePathname, useRouter, useParams as useNextParams } from "next/navigation";

type LinkProps = Omit<React.ComponentPropsWithoutRef<"a">, "href"> & { to: string; replace?: boolean };

export function Link({ to, children, ...rest }: LinkProps) {
  return (
    <NextLink href={to} {...rest}>
      {children}
    </NextLink>
  );
}

export function useNavigate() {
  const router = useRouter();
  return (to: string, options?: { replace?: boolean }) => {
    if (options?.replace) {
      router.replace(to);
      return;
    }
    router.push(to);
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
  useEffect(() => {
    if (replace) {
      router.replace(to);
      return;
    }
    router.push(to);
  }, [replace, router, to]);
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
