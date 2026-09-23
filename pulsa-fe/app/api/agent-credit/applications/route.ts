import { NextResponse } from "next/server";
import { getBackendAuthorization } from "@/lib/server-auth";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

async function proxy(path: string, method: "GET" | "POST", req?: Request) {
  const token = await getBackendAuthorization(req);
  if (!token) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  try {
    const body = method === "POST" && req ? await req.text() : undefined;
    const res = await fetch(`${apiBase()}${path}`, {
      method,
      headers: {
        Authorization: token,
        "Content-Type": "application/json",
      },
      body,
      cache: "no-store",
      signal: AbortSignal.timeout(15000),
    });
    const text = await res.text();
    return new NextResponse(text, {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch {
    return NextResponse.json({ ok: false, error: "Layanan pengajuan kredit belum dapat dihubungi" }, { status: 502 });
  }
}

export async function GET(req: Request) {
  const query = new URL(req.url).searchParams.toString();
  return proxy(`/v1/master/agent-credit/applications${query ? `?${query}` : ""}`, "GET", req);
}

export async function POST(req: Request) {
  return proxy("/v1/me/agent-credit/applications", "POST", req);
}
