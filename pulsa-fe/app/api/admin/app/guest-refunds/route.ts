import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const qs = url.searchParams.toString();

  const res = await fetch(`${base}/v1/admin/app/guest-refunds${qs ? `?${qs}` : ""}`, {
    method: "GET",
    headers: {
      ...(auth ? { Authorization: auth } : {}),
    },
    cache: "no-store",
  });

  const text = await res.text();
  return new NextResponse(text, { status: res.status, headers: { "Content-Type": "application/json" } });
}
