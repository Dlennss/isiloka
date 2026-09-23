import { NextResponse } from "next/server";
import { forwardAuth, requireApiBase } from "@/lib/adminApi";
import { verifyTurnstileToken } from "@/lib/serverTurnstile";

export const runtime = "nodejs";

export async function POST(req: Request) {
  const base = requireApiBase();
  const incoming = new Headers(req.headers);
  const auth = forwardAuth(incoming);
  const turnstileRequired = incoming.get("x-turnstile-required") === "1";
  const turnstileToken = incoming.get("x-turnstile-token") || "";
  const body = await req.text();

  if (!auth && turnstileRequired) {
    if (!turnstileToken) {
      return NextResponse.json({ ok: false, error: "turnstile token required" }, { status: 400 });
    }
    const ip = incoming.get("x-forwarded-for")?.split(",")[0]?.trim();
    const v = await verifyTurnstileToken(turnstileToken, ip);
    if (!v.success) {
      return NextResponse.json(
        { ok: false, error: "turnstile failed", codes: v.errorCodes, detail: v.error ?? "" },
        { status: 403 }
      );
    }
  }

  const r = await fetch(`${base}/v1/app/billing-checks`, {
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
