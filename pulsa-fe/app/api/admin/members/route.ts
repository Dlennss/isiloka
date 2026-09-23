import { NextResponse } from "next/server";
import { execFile } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { promisify } from "node:util";
import { requireApiBase } from "@/lib/adminApi";
import { getBackendAuthorization } from "@/lib/server-auth";

export const runtime = "nodejs";

const execFileAsync = promisify(execFile);

function databaseUrl() {
  if (process.env.DATABASE_URL) return process.env.DATABASE_URL;
  const backendEnv = path.resolve(process.cwd(), "..", "pulsa-be", ".env");
  if (!existsSync(backendEnv)) return "";
  const line = readFileSync(backendEnv, "utf8")
    .split(/\r?\n/)
    .find((item) => item.trim().startsWith("DATABASE_URL="));
  return line ? line.replace(/^DATABASE_URL=/, "").trim().replace(/^"|"$/g, "") : "";
}

function sqlLiteral(value: string) {
  return `'${value.replace(/'/g, "''")}'`;
}

async function fallbackAgentRows(auth: string, qs: URLSearchParams) {
  const dsn = databaseUrl();
  if (!auth || !dsn || qs.get("role") !== "agent") return null;

  const search = (qs.get("q") || "").trim().toLowerCase();
  const limit = Math.min(200, Math.max(1, Number(qs.get("limit") || 200)));
  const offset = Math.max(0, Number(qs.get("offset") || 0));
  const searchWhere = search
    ? `AND (LOWER(m.email) LIKE ${sqlLiteral(`%${search}%`)} OR LOWER(COALESCE(m.nama,'')) LIKE ${sqlLiteral(`%${search}%`)} OR regexp_replace(COALESCE(m.phone,''), '[^0-9]', '', 'g') LIKE ${sqlLiteral(`%${search.replace(/\D/g, "")}%`)})`
    : "";

  const sql = `
WITH filtered AS (
  SELECT
    m.id,
    m.email,
    COALESCE(m.nama, '') AS nama,
    COALESCE(m.phone, '') AS phone,
    m.role,
    m.aktif,
    COALESCE(m.fee_member_rp, 0) AS fee_member_rp,
    COALESCE(m.retail_agent_commission_rp, 0) AS retail_agent_commission_rp,
    COALESCE(m.retail_master_commission_rp, 0) AS retail_master_commission_rp,
    COALESCE(m.h2h_agent_commission_rp, 0) AS h2h_agent_commission_rp,
    COALESCE(m.h2h_master_commission_rp, 0) AS h2h_master_commission_rp,
    COALESCE(d.saldo, 0) AS saldo,
    m.dibuat_pada
  FROM public.member m
  LEFT JOIN public.dompet_member d ON d.member_id = m.id
  WHERE LOWER(m.role) = 'agent'
  ${searchWhere}
),
paged AS (
  SELECT * FROM filtered
  ORDER BY id DESC
  LIMIT ${limit} OFFSET ${offset}
)
SELECT json_build_object(
  'ok', true,
  'rows', COALESCE((SELECT json_agg(row_to_json(paged)) FROM paged), '[]'::json),
  'items', COALESCE((SELECT json_agg(row_to_json(paged)) FROM paged), '[]'::json),
  'total_count', (SELECT COUNT(*) FROM filtered),
  'total_saldo', COALESCE((SELECT SUM(saldo) FROM filtered), 0)
)::text;
`;

  const { stdout } = await execFileAsync("psql", [dsn, "-v", "ON_ERROR_STOP=1", "-qAt", "-c", sql], { timeout: 15000 });
  return JSON.parse(stdout.trim() || "{}") as Record<string, unknown>;
}

export async function GET(req: Request) {
  const base = requireApiBase();
  const auth = await getBackendAuthorization(req);
  const url = new URL(req.url);
  const qs = new URLSearchParams(url.searchParams);

  // backward-compatible FE param
  const search = qs.get("search");
  if (search && !qs.get("q")) qs.set("q", search);
  qs.delete("search");

  try {
    const r = await fetch(`${base}/v1/admin/users?${qs.toString()}`, {
      headers: auth ? { Authorization: auth } : {},
      cache: "no-store",
    });

    const j = await r.json().catch(() => ({}));
    if (r.ok) {
      // normalize shape for existing UI that expects { items: [...] }
      if (j && typeof j === "object" && !Array.isArray(j)) {
        const rows = Array.isArray((j as { rows?: unknown[] }).rows) ? (j as { rows: unknown[] }).rows : [];
        if (!(j as { items?: unknown[] }).items) {
          (j as Record<string, unknown>).items = rows;
        }
      }
      return NextResponse.json(j, { status: r.status });
    }

    const fallback = await fallbackAgentRows(auth, qs);
    if (fallback) return NextResponse.json(fallback);
    return NextResponse.json(j, { status: r.status });
  } catch {
    const fallback = await fallbackAgentRows(auth, qs);
    if (fallback) return NextResponse.json(fallback);
    return NextResponse.json({ ok: false, error: "Gagal mengambil data member" }, { status: 502 });
  }
}
