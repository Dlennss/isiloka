import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export async function PUT(req: Request, ctx: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await ctx.params;
  const body = await req.text();

  const r = await fetch(`${base}/v1/admin/users/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      ...(auth ? { Authorization: auth } : {}),
    },
    body,
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}

export async function DELETE(req: Request, ctx: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await ctx.params;

  const r = await fetch(`${base}/v1/admin/users/${encodeURIComponent(id)}`, {
    method: "DELETE",
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}
