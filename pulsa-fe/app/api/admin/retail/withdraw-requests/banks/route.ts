import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const response = await fetch(`${base}/v1/admin/retail/withdraw-requests/banks`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });
  const text = await response.text();
  return new NextResponse(text, { status: response.status, headers: { "Content-Type": "application/json" } });
}
