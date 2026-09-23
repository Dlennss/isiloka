import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";

type RouteContext = {
  params: Promise<{ invoiceId: string }>;
};

export async function GET(req: Request, { params }: RouteContext) {
  const base = requireApiBase();
  const incoming = new Headers(req.headers);
  const auth = forwardAuth(incoming);
  const guestEmail = incoming.get("x-guest-email") || "";
  const guestPhone = incoming.get("x-guest-phone") || "";
  const { invoiceId } = await params;

  const r = await fetch(`${base}/v1/app/orders/${encodeURIComponent(invoiceId)}`, {
    method: "GET",
    headers: {
      ...(auth ? { Authorization: auth } : {}),
      ...(guestEmail ? { "X-Guest-Email": guestEmail } : {}),
      ...(guestPhone ? { "X-Guest-Phone": guestPhone } : {}),
    },
    cache: "no-store",
  });

  const text = await r.text();
  return new NextResponse(text, { status: r.status, headers: { "Content-Type": "application/json" } });
}
