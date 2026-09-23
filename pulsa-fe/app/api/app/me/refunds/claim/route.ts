import { NextResponse } from "next/server";

export async function POST(req: Request) {
  const auth = req.headers.get("authorization") || "";
  if (!auth.trim()) {
    return NextResponse.json({ ok: false, error: "missing bearer token" }, { status: 401 });
  }

  const base = process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";
  const body = await req.text();
  const res = await fetch(`${base}/v1/app/me/refunds/claim`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: auth,
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
