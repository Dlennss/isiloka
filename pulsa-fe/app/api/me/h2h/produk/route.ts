import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

type ProfileResponse = {
  ok?: boolean;
  profile?: {
    id?: number;
    nama?: string;
    email?: string;
  };
  api_keys?: Array<{
    id?: number;
    api_key?: string;
    aktif?: boolean;
  }>;
  error?: string;
};

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  if (!auth) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const profileRes = await fetch(`${base}/v1/me/profile`, {
    headers: { Authorization: auth },
    cache: "no-store",
  });
  const profileJson = (await profileRes.json().catch(() => ({}))) as ProfileResponse;
  if (!profileRes.ok || !profileJson?.ok) {
    return NextResponse.json(
      { ok: false, error: profileJson?.error || "Gagal memuat profil member." },
      { status: profileRes.status || 500 },
    );
  }

  const apiKey = (profileJson.api_keys || []).find((item) => item?.aktif && item?.api_key)?.api_key || "";
  if (!apiKey) {
    return NextResponse.json({ ok: false, error: "API key aktif member tidak ditemukan." }, { status: 400 });
  }

  const url = new URL(req.url);
  const upstream = await fetch(`${base}/v1/h2h/produk?${url.searchParams.toString()}`, {
    headers: { "X-Api-Key": apiKey },
    cache: "no-store",
  });
  const text = await upstream.text();
  return new NextResponse(text, { status: upstream.status, headers: { "Content-Type": "application/json" } });
}
