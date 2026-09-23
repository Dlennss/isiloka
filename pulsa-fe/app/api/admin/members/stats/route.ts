import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const qs = new URLSearchParams(url.searchParams);
  if (qs.get("member_id") && !qs.get("user_id")) {
    qs.set("user_id", qs.get("member_id") || "");
  }

  const r = await fetch(`${base}/v1/admin/users/stats?${qs.toString()}`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, {
    status: r.status,
    headers: { "Content-Type": "application/json" },
  });
}
