import { headers } from "next/headers";

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const resolveApiBase = () => {
  const fromServer = process.env.API_BASE_URL?.trim();
  if (fromServer) return trimTrailingSlash(fromServer);
  const fromPublic = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (fromPublic) return trimTrailingSlash(fromPublic);
  return "http://localhost:8080";
};

async function fetchJSON<T>(path: string): Promise<T> {
  const h = await headers();
  const host = h.get("host") ?? "";
  const protocol = h.get("x-forwarded-proto") ?? "http";
  const cookie = h.get("cookie") ?? "";
  const url = `${resolveApiBase()}${path}`;
  const requestHeaders: HeadersInit = {
    Host: host,
    "X-Forwarded-Proto": protocol,
    Accept: "application/json",
  };
  if (cookie) {
    requestHeaders.Cookie = cookie;
  }
  const response = await fetch(url, {
    cache: "no-store",
    headers: requestHeaders,
  });
  if (!response.ok) {
    throw new Error(`${response.status}: failed to fetch ${path}`);
  }
  return (await response.json()) as T;
}

export async function fetchPublicAuctions() {
  return fetchJSON<any[]>("/auctions");
}

export async function fetchAuction(auctionID: string) {
  return fetchJSON<any>(`/auctions/${auctionID}`);
}

export async function fetchLotsByAuction(auctionID: string) {
  return fetchJSON<any[]>(`/auctions/${auctionID}/lots`);
}

export async function fetchLot(lotID: string) {
  return fetchJSON<any>(`/lots/${lotID}`);
}

export async function fetchLotBids(lotID: string) {
  return fetchJSON<any[]>(`/lots/${lotID}/bids`);
}
