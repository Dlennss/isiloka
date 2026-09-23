import { NextResponse } from "next/server";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

export const runtime = "nodejs";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

function resolveToken(req: Request, session: SessionShape | null) {
  const incomingAuthorization = String(req.headers.get("authorization") || "").trim();
  return incomingAuthorization || (session?.backendToken ? `Bearer ${session.backendToken}` : "");
}

export async function GET(req: Request) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const token = resolveToken(req, session);
  if (!token) return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });

  const res = await fetch(`${apiBase()}/v1/master/agent-credit/ranks`, {
    headers: { Authorization: token },
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, { status: res.status, headers: { "Content-Type": "application/json" } });
}

export async function POST(req: Request) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const token = resolveToken(req, session);
  if (!token) return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });

  const body = await req.text();
  const res = await fetch(`${apiBase()}/v1/master/agent-credit/ranks`, {
    method: "POST",
    headers: { Authorization: token, "Content-Type": "application/json" },
    body,
    cache: "no-store",
  });
  const text = await res.text();
  return new NextResponse(text, { status: res.status, headers: { "Content-Type": "application/json" } });
}
