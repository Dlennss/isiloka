import { NextResponse } from "next/server";
import { requireApiBase } from "@/lib/adminApi";
import { getBackendAuthorization } from "@/lib/server-auth";

export const runtime = "nodejs";

type RouteContext = {
  params: Promise<{ invoiceId: string }>;
};

export async function POST(req: Request, { params }: RouteContext) {
  const base = requireApiBase();
  const incoming = new Headers(req.headers);
  const auth = await getBackendAuthorization(req);
  const guestEmail = incoming.get("x-guest-email") || "";
  const guestPhone = incoming.get("x-guest-phone") || "";
  const { invoiceId } = await params;
  const body = await req.text();

  const r = await fetch(`${base}/v1/app/orders/${encodeURIComponent(invoiceId)}/pay`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(auth ? { Authorization: auth } : {}),
      ...(guestEmail ? { "X-Guest-Email": guestEmail } : {}),
      ...(guestPhone ? { "X-Guest-Phone": guestPhone } : {}),
    },
    body: body || "{}",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
