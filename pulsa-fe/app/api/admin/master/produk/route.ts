import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const qs = url.searchParams.toString();

  const r = await fetch(`${base}/v1/admin/produk${qs ? `?${qs}` : ""}`, {
    method: "GET",
    headers: { ...(auth ? { Authorization: auth } : {}) },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}

export async function POST(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const body = await req.text();

  const r = await fetch(`${base}/v1/admin/produk`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(auth ? { Authorization: auth } : {}),
    },
    body,
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
