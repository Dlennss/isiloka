import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

type RouteContext = {
  params: Promise<{ id: string }>;
};

export async function GET(req: Request, ctx: RouteContext) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await ctx.params;

  const r = await fetch(`${base}/v1/history/mutasi/${encodeURIComponent(id)}`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
