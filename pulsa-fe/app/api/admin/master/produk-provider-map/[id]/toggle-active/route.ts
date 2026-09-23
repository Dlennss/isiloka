import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

function togglePath(params: { id: string }) {
  return `/v1/admin/produk-provider-map/${encodeURIComponent(params.id)}/toggle-active`;
}

export async function POST(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const p = await params;
  const body = await req.text();

  const r = await fetch(`${base}${togglePath(p)}`, {
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
