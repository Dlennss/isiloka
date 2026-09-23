import { NextResponse } from "next/server";
import { getBackendAuthorization } from "@/lib/server-auth";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

export async function GET(req: Request) {
  const token = await getBackendAuthorization(req);
  if (!token) return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  const query = new URL(req.url).searchParams.toString();
  const res = await fetch(`${apiBase()}/v1/master/agent-credit/manual-applications?${query}`, {
    headers: { Authorization: token },
    cache: "no-store",
  });
  return new NextResponse(await res.text(), { status: res.status, headers: { "Content-Type": "application/json" } });
}

export async function POST(req: Request) {
  const token = await getBackendAuthorization(req);
  if (!token) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(`${apiBase()}/v1/master/agent-credit/manual-applications`, {
    method: "POST",
    headers: {
      Authorization: token,
      "Content-Type": "application/json",
    },
    body: await req.text(),
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

export async function PUT(req: Request) {
  const token = await getBackendAuthorization(req);
  if (!token) return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  const res = await fetch(`${apiBase()}/v1/master/agent-credit/manual-applications`, {
    method: "PUT",
    headers: { Authorization: token, "Content-Type": "application/json" },
    body: await req.text(),
    cache: "no-store",
  });
  return new NextResponse(await res.text(), {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}
