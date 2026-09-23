import { NextResponse } from "next/server";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

export async function POST(req: Request) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const token = String(session?.backendToken || "").trim();
  if (!token) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const res = await fetch(`${apiBase()}/v1/me/agent-credit/transfers`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
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
