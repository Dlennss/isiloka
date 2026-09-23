import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export async function POST(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const body = await req.text();

  try {
    const r = await fetch(`${base}/v1/admin/loketbayar-transfer/transfers/create`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(auth ? { Authorization: auth } : {}),
      },
      body,
      cache: "no-store",
    });

    const text = await r.text();
    return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
  } catch (err) {
    const message = err instanceof Error ? err.message : "backend tidak bisa dihubungi";
    return NextResponse.json({ ok: false, error: `LoketBayar transfer backend error: ${message}` }, { status: 502 });
  }
}
