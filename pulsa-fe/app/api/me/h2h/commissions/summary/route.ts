import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const r = await fetch(`${base}/v1/me/h2h/commissions/summary`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });
  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
