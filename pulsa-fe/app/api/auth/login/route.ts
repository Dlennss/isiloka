import { NextResponse } from "next/server";
import { verifyTurnstileToken } from "@/lib/serverTurnstile";
import { PK_AUTH_COOKIE } from "@/lib/server-auth";

export const runtime = "nodejs";

const TURNSTILE_ENABLED = /^(1|true|yes|on)$/i.test(process.env.TURNSTILE_ENABLED || "");

type ReqBody = {
  email: string;
  password: string;
  turnstileToken?: string;
};

function decodeRole(jwt: string): string | null {
  const parts = jwt.split(".");
  if (parts.length !== 3) return null;
  try {
    const payload = JSON.parse(Buffer.from(parts[1], "base64url").toString("utf8")) as { role?: unknown };
    return typeof payload.role === "string" ? payload.role : null;
  } catch {
    return null;
  }
}

type BackendLoginResp = {
  ok?: boolean;
  token?: string;
  error?: string;
};

export async function POST(req: Request) {
  let body: ReqBody;
  try {
    body = (await req.json()) as ReqBody;
  } catch {
    return NextResponse.json({ ok: false, error: "bad json" }, { status: 400 });
  }

  const email = (body.email || "").trim().toLowerCase();
  const password = body.password || "";
  const turnstileToken = body.turnstileToken || "";

  if (!email || !password || (TURNSTILE_ENABLED && !turnstileToken)) {
    return NextResponse.json({ ok: false, error: "email/password/turnstileToken required" }, { status: 400 });
  }

  if (TURNSTILE_ENABLED) {
    const ip = req.headers.get("x-forwarded-for")?.split(",")[0]?.trim();
    const v = await verifyTurnstileToken(turnstileToken, ip);
    if (!v.success) {
      return NextResponse.json(
        { ok: false, error: "turnstile failed", codes: v.errorCodes, detail: v.error ?? "" },
        { status: 403 }
      );
    }
  }

  const base = process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

  const r = await fetch(`${base}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
    cache: "no-store",
  });

  const data = (await r.json().catch(() => ({}))) as BackendLoginResp;

  if (!r.ok || !data.ok || !data.token) {
    return NextResponse.json({ ok: false, error: data.error ?? "login failed" }, { status: 401 });
  }

  const role = decodeRole(data.token) || "member";
  const response = NextResponse.json({ ok: true, token: data.token, role });
  response.cookies.set(PK_AUTH_COOKIE, data.token, {
    httpOnly: true,
    sameSite: "lax",
    secure: req.url.startsWith("https://") || process.env.NODE_ENV === "production",
    path: "/",
    maxAge: 60 * 60 * 24 * 30,
  });
  return response;
}
