import { NextResponse } from "next/server";

export const runtime = "nodejs";

type ReqBody = {
  email?: string;
  nama?: string;
  google_id?: string;
  id_token?: string;
};

type BackendGoogleLoginResp = {
  ok?: boolean;
  token?: string;
  role?: string;
  created?: boolean;
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
  const nama = (body.nama || "").trim();
  const googleID = (body.google_id || "").trim();
  const idToken = (body.id_token || "").trim();
  if (!email || !googleID || !idToken) {
    return NextResponse.json({ ok: false, error: "email/google_id/id_token required" }, { status: 400 });
  }

  const base = process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";
  const r = await fetch(`${base}/v1/auth/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, nama, google_id: googleID, id_token: idToken }),
    cache: "no-store",
  });

  const data = (await r.json().catch(() => ({}))) as BackendGoogleLoginResp;
  if (!r.ok || !data.ok || !data.token) {
    return NextResponse.json({ ok: false, error: data.error ?? "google login failed" }, { status: r.status || 401 });
  }

  return NextResponse.json({
    ok: true,
    token: data.token,
    role: data.role || "user",
    created: !!data.created,
  });
}
