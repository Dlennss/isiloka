import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

type Params = { params: Promise<{ id: string }> };

export async function GET(req: Request, { params }: Params) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await params;

  const r = await fetch(`${base}/v1/admin/kategori-fee-app/${id}`, {
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

export async function PUT(req: Request, { params }: Params) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await params;
  const body = await req.text();

  const r = await fetch(`${base}/v1/admin/kategori-fee-app/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      ...(auth ? { Authorization: auth } : {}),
    },
    body,
  });
  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}

export async function DELETE(req: Request, { params }: Params) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const { id } = await params;

  const r = await fetch(`${base}/v1/admin/kategori-fee-app/${id}`, {
    method: "DELETE",
    headers: auth ? { Authorization: auth } : {},
  });
  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}
