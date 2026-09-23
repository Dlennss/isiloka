import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function POST(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const body = await req.text();

  const r = await fetch(`${base}/v1/admin/deposit/requests/va/approve?${url.searchParams.toString()}`, {
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
