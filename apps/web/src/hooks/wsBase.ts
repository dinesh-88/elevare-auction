const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const adaptLocalBaseToTenantHost = (base: string) => {
  if (typeof window === "undefined" || !base) return base;
  try {
    const url = new URL(base);
    const currentHost = window.location.hostname;
    const baseHost = url.hostname;
    const isLocalBase = baseHost === "localhost" || baseHost === "127.0.0.1";
    const isTenantHost = currentHost.endsWith(".localhost") && currentHost !== "localhost";
    if (isLocalBase && isTenantHost) {
      url.hostname = currentHost;
      return trimTrailingSlash(url.toString());
    }
    return base;
  } catch {
    return base;
  }
};

const toWsProtocol = (protocol: string) => (protocol === "https:" ? "wss:" : "ws:");

const normalizeConfiguredWsBase = (value: string) => adaptLocalBaseToTenantHost(trimTrailingSlash(value));

const deriveFromApiBase = (apiBase: string) => {
  try {
    const url = new URL(apiBase);
    url.protocol = toWsProtocol(url.protocol);
    return normalizeConfiguredWsBase(url.toString());
  } catch {
    return "";
  }
};

const runtimeFallbackWsBase = () => {
  if (typeof window === "undefined") return "";
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const host = window.location.hostname;
  const isLocalHost =
    host === "localhost" ||
    host === "127.0.0.1" ||
    host.endsWith(".localhost");
  if (isLocalHost) {
    return `${protocol}://${host}:${process.env.NEXT_PUBLIC_WS_PORT ?? "8080"}`;
  }
  return `${protocol}://${host}`;
};

export const resolveWsBase = () => {
  const envWsBase = process.env.NEXT_PUBLIC_WS_BASE_URL?.trim();
  if (envWsBase) return normalizeConfiguredWsBase(envWsBase);

  const envApiBase = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (envApiBase) {
    const derived = deriveFromApiBase(envApiBase);
    if (derived) return derived;
  }

  return runtimeFallbackWsBase();
};
