import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const url = new URL(req.url);
  const auth = forwardAuth(new Headers(req.headers));
  const response = await fetch(`${base}/v1/admin/agent-credit/inactive-agents?${url.searchParams.toString()}`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });
  const body = await response.text();
  return new NextResponse(body, { status: response.status, headers: { "Content-Type": "application/json" } });
}
