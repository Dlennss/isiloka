import { NextResponse } from "next/server";
import { requireApiBase, forwardAuth } from "@/lib/adminApi";

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = forwardAuth(new Headers(req.headers));
  const url = new URL(req.url);
  const memberID = url.searchParams.get("member_id") || "";
  const id = Number(memberID);
  if (!Number.isFinite(id) || id <= 0) {
    return NextResponse.json({ ok: false, error: "member_id invalid" }, { status: 400 });
  }

  const r = await fetch(`${base}/v1/admin/users/${id}`, {
    headers: auth ? { Authorization: auth } : {},
    cache: "no-store",
  });

  const j = await r.json().catch(() => ({}));
  // normalize shape for existing UI that expects { item: {...} }
  if (j && typeof j === "object" && !Array.isArray(j)) {
    const item = (j as { item?: unknown }).item;
    const rows = (j as { rows?: unknown[] }).rows;
    if (!item && Array.isArray(rows) && rows.length > 0) {
      (j as Record<string, unknown>).item = rows[0];
    }
  }
  return NextResponse.json(j, { status: r.status });
}
