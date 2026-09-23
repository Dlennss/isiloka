import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

type Ctx = { params: Promise<{ invoiceId: string }> };

export async function GET(req: Request, ctx: Ctx) {
  const { invoiceId } = await ctx.params;
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));

  const r = await fetch(`${base}/v1/admin/app/orders/${encodeURIComponent(invoiceId)}`, {
    method: "GET",
    headers: {
      ...(auth ? { Authorization: auth } : {}),
    },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
