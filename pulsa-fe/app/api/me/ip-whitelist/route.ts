import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

async function proxy(req: Request, method: string) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);

  const body = method === "POST" ? await req.text() : undefined;

  const r = await fetch(`${base}/v1/me/ip-whitelist?${url.searchParams.toString()}`, {
    method,
    headers: {
      ...(method === "POST" ? { "Content-Type": "application/json" } : {}),
      ...(auth ? { Authorization: auth } : {}),
    },
    body,
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}

export async function GET(req: Request) {
  return proxy(req, "GET");
}

export async function POST(req: Request) {
  return proxy(req, "POST");
}

export async function DELETE(req: Request) {
  return proxy(req, "DELETE");
}
