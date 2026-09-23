import { NextResponse } from "next/server";
import { getBackendAuthorization } from "@/lib/server-auth";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

async function proxyProfile(method: "GET" | "PATCH", req?: Request) {
  const auth = await getBackendAuthorization(req);
  if (!auth) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const body = method === "PATCH" && req ? await req.text() : undefined;
  const res = await fetch(`${apiBase()}/v1/me/profile`, {
    method,
    headers: {
      Authorization: auth,
      "Content-Type": "application/json",
    },
    body,
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

export async function GET() {
  return proxyProfile("GET");
}

export async function PATCH(req: Request) {
  return proxyProfile("PATCH", req);
}
