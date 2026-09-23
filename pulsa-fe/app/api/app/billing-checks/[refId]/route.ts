import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export const runtime = "nodejs";

export async function GET(req: Request, ctx: { params: Promise<{ refId: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { refId } = await ctx.params;
  const incoming = new Headers(req.headers);

  const r = await fetch(`${base}/v1/app/billing-checks/${encodeURIComponent(refId)}`, {
    method: "GET",
    headers: {
      ...(auth ? { Authorization: auth } : {}),
      ...(incoming.get("x-guest-email") ? { "X-Guest-Email": incoming.get("x-guest-email") as string } : {}),
      ...(incoming.get("x-guest-phone") ? { "X-Guest-Phone": incoming.get("x-guest-phone") as string } : {}),
    },
    cache: "no-store",
  });
  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
