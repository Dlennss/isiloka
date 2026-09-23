import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

function idPath(params: { id: string }) {
  return `/v1/admin/kategori/${encodeURIComponent(params.id)}`;
}

export async function GET(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const p = await params;

  const r = await fetch(`${base}${idPath(p)}`, {
    method: "GET",
    headers: { ...(auth ? { Authorization: auth } : {}) },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const p = await params;
  const body = await req.text();

  const r = await fetch(`${base}${idPath(p)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      ...(auth ? { Authorization: auth } : {}),
    },
    body,
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}

export async function DELETE(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const p = await params;

  const r = await fetch(`${base}${idPath(p)}`, {
    method: "DELETE",
    headers: { ...(auth ? { Authorization: auth } : {}) },
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
