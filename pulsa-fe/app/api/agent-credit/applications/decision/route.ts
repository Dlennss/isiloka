import { NextResponse } from "next/server";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

export const runtime = "nodejs";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

export async function POST(req: Request) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const incomingAuthorization = String(req.headers.get("authorization") || "").trim();
  const token = incomingAuthorization || (session?.backendToken ? `Bearer ${session.backendToken}` : "");
  if (!token) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const rawBody = await req.text();
  let body = rawBody;
  try {
    const parsed = JSON.parse(rawBody) as { decision?: unknown };
    if (parsed.decision === "forward_to_analysis") {
      parsed.decision = "kirim_analis";
      body = JSON.stringify(parsed);
    }
  } catch {
    body = rawBody;
  }

  const res = await fetch(`${apiBase()}/v1/master/agent-credit/applications/decision`, {
    method: "POST",
    headers: {
      Authorization: token,
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
