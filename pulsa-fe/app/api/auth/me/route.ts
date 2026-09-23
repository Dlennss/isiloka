import { NextResponse } from "next/server";

export const runtime = "nodejs";

export async function GET(req: Request) {
  const base = process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";
  const auth = req.headers.get("authorization") || "";

  const r = await fetch(`${base}/v1/auth/me`, {
    method: "GET",
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}

