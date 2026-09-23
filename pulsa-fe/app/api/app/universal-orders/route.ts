import { getServerSession } from "next-auth";
import { NextResponse } from "next/server";
import { spawn } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { authOptions } from "@/lib/nextauth";

export const runtime = "nodejs";

type SessionShape = {
  backendToken?: string;
};

type ProfileResponse = {
  ok?: boolean;
  profile?: {
    id?: number;
    role?: string;
  };
};

type UniversalOrderPayload = {
  service?: string;
  destination?: string;
  provider?: string;
  product?: string;
  total?: number;
};

function apiBase() {
  return (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");
}

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

function runPsql(dsn: string, sql: string) {
  return new Promise<string>((resolve, reject) => {
    const child = spawn("psql", [dsn, "-v", "ON_ERROR_STOP=1", "-qAt"], { stdio: ["pipe", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    const timeout = setTimeout(() => {
      child.kill();
      reject(new Error("psql timeout"));
    }, 15000);

    child.stdout.on("data", (chunk) => {
      stdout += String(chunk);
    });
    child.stderr.on("data", (chunk) => {
      stderr += String(chunk);
    });
    child.on("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });
    child.on("close", (code) => {
      clearTimeout(timeout);
      if (code === 0) resolve(stdout);
      else reject(new Error(stderr || `psql exited with code ${code}`));
    });
    child.stdin.end(sql);
  });
}

async function getProfile(token: string) {
  const res = await fetch(`${apiBase()}/v1/me/profile`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  const body = (await res.json().catch(() => ({}))) as ProfileResponse;
  if (!res.ok || !body.ok || !body.profile?.id) return null;
  return body.profile;
}

function cleanText(value: unknown, fallback = "") {
  return String(value || fallback).trim().slice(0, 160);
}

function normalizeRole(value: unknown) {
  const role = String(value || "user").trim().toLowerCase();
  if (role === "agent" || role === "master" || role === "marketing") return role;
  return "user";
}

export async function POST(req: Request) {
  const session = (await getServerSession(authOptions)) as SessionShape | null;
  const token = String(session?.backendToken || "").trim();
  if (!token) return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });

  const profile = await getProfile(token);
  if (!profile?.id) return NextResponse.json({ ok: false, error: "profile tidak ditemukan" }, { status: 401 });

  const payload = (await req.json().catch(() => ({}))) as UniversalOrderPayload;
  const service = cleanText(payload.service, "Layanan PulsaKilat");
  const destination = cleanText(payload.destination);
  const provider = cleanText(payload.provider, "PulsaKilat");
  const product = cleanText(payload.product, "Transaksi");
  const total = Math.max(0, Math.round(Number(payload.total || 0)));
  if (!destination || total <= 0) {
    return NextResponse.json({ ok: false, error: "tujuan dan nominal wajib diisi" }, { status: 400 });
  }

  const dsn = databaseUrl();
  if (!dsn) return NextResponse.json({ ok: false, error: "database belum dikonfigurasi" }, { status: 500 });

  const invoicePrefix = `PK${Date.now().toString().slice(-9)}`;
  const memberID = Number(profile.id);
  const role = normalizeRole(profile.role);
  const fee = total >= 100000 ? 1500 : 1000;
  const nominal = Math.max(0, total - fee);
  const sku = `UNIVERSAL-${service.toUpperCase().replace(/[^A-Z0-9]+/g, "-").slice(0, 48) || "SERVICE"}`;
  const productName = `${service} - ${product}`;
  const note = JSON.stringify({
    source: "universal_service_fallback",
    service,
    provider,
    product,
  });

  const sql = `
WITH fallback_product AS (
  SELECT id
  FROM public.produk
  WHERE aktif = true
  ORDER BY id
  LIMIT 1
),
inserted AS (
  INSERT INTO public.app_order (
    invoice_id,
    member_id,
    produk_id,
    produk_sku_snapshot,
    produk_nama_snapshot,
    dest,
    qty,
    nominal,
    buyer_type,
    buyer_role,
    harga_dasar,
    fee,
    harga_final,
    fee_user_snapshot,
    fee_agent_snapshot,
    fee_master_snapshot,
    status,
    catatan,
    dibuat_pada,
    diubah_pada
  )
  SELECT
    ${sqlLiteral(invoicePrefix)},
    ${memberID},
    id,
    ${sqlLiteral(sku)},
    ${sqlLiteral(productName)},
    ${sqlLiteral(destination)},
    1,
    ${nominal},
    'user',
    ${sqlLiteral(role)},
    ${nominal},
    ${fee},
    ${total},
    0,
    0,
    0,
    'pending_payment',
    ${sqlLiteral(note)},
    now(),
    now()
  FROM fallback_product
  RETURNING *
)
SELECT COALESCE(row_to_json(inserted)::text, '{}') FROM inserted;
`;

  const stdout = await runPsql(dsn, sql).catch((error) => {
    throw error;
  });
  const item = JSON.parse(stdout.trim() || "{}") as Record<string, unknown>;
  if (!item.id) {
    return NextResponse.json({ ok: false, error: "produk fallback belum tersedia di database" }, { status: 400 });
  }

  return NextResponse.json({ ok: true, item });
}
