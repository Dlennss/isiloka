import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);

  const r = await fetch(`${base}/v1/admin/loketbayar-transfer/transfers?${url.searchParams.toString()}`, {
    method: "GET",
    headers: { ...(auth ? { Authorization: auth } : {}) },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
