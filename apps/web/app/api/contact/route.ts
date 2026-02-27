import { NextResponse } from "next/server";

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

function resolveApiBase() {
  const fromServer = process.env.API_BASE_URL?.trim();
  if (fromServer) return trimTrailingSlash(fromServer);
  const fromPublic = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (fromPublic) return trimTrailingSlash(fromPublic);
  return "http://localhost:8080";
}

export async function POST(request: Request) {
  const payload = (await request.json()) as {
    name?: string;
    institution?: string;
    email?: string;
    message?: string;
  };

  const name = payload.name?.trim() ?? "";
  const institution = payload.institution?.trim() ?? "";
  const email = payload.email?.trim() ?? "";
  const message = payload.message?.trim() ?? "";

  if (!name || !institution || !email || !email.includes("@")) {
    return NextResponse.json({ error: "Please complete all required fields." }, { status: 400 });
  }

  const response = await fetch(`${resolveApiBase()}/contact`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({
      name,
      institution,
      email,
      message,
    }),
    cache: "no-store",
  });

  if (!response.ok) {
    const errorText = await response.text();
    return NextResponse.json(
      { error: errorText || "We could not submit your request at this time." },
      { status: response.status },
    );
  }

  const data = (await response.json()) as Record<string, unknown>;
  return NextResponse.json(data, { status: 201 });
}
