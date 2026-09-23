import { NextResponse } from "next/server";
import { spawn } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { getAppServerSession } from "@/lib/server-auth";

type SessionShape = {
  backendToken?: string;
};

export const runtime = "nodejs";

const apiBase = () => process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083";

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
  return new Promise<void>((resolve, reject) => {
    const child = spawn("psql", [dsn, "-v", "ON_ERROR_STOP=1", "-qAt"], { stdio: ["pipe", "ignore", "pipe"] });
    let stderr = "";
    const timeout = setTimeout(() => {
      child.kill();
      reject(new Error("psql timeout"));
    }, 15000);

    child.stderr.on("data", (chunk) => {
      stderr += String(chunk);
    });
    child.on("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });
    child.on("close", (code) => {
      clearTimeout(timeout);
      if (code === 0) resolve();
      else reject(new Error(stderr || `psql exited with code ${code}`));
    });
    child.stdin.end(sql);
  });
}

function compactPaymentProof(value: unknown) {
  if (!value || typeof value !== "object") return null;
  const proof = value as Record<string, unknown>;
  const dataURL = String(proof.data_url || "");
  if (!dataURL.startsWith("data:image/")) return null;

  return {
    name: String(proof.name || "bukti-transfer.png").slice(0, 120),
    type: String(proof.type || "image/png").slice(0, 80),
    size: Number(proof.size || 0),
    data_url: dataURL,
  };
}

async function persistPaymentProofFallback(payload: Record<string, unknown>) {
  const dsn = databaseUrl();
  const applicationID = Number(payload.application_id || 0);
  const amount = Number(payload.amount || 0);
  const proof = compactPaymentProof(payload.payment_proof);
  if (!dsn || !applicationID || !amount || !proof) return;

  const notePayload = JSON.stringify({
    payment_method: String(payload.payment_method || "transfer"),
    note: String(payload.note || ""),
    payment_proof: proof,
  });

  const sql = `
WITH target AS (
  SELECT p.id
  FROM public.agent_credit_payment p
  JOIN public.agent_credit_loan l ON l.id = p.loan_id
  WHERE l.application_id = ${applicationID}
    AND p.amount = ${amount}
  ORDER BY p.paid_at DESC, p.id DESC
  LIMIT 1
)
UPDATE public.agent_credit_payment p
SET note = ${sqlLiteral(notePayload)}
FROM target
WHERE p.id = target.id;
`;

  await runPsql(dsn, sql);
}

export async function POST(req: Request) {
  const session = (await getAppServerSession()) as SessionShape | null;
  const token = String(session?.backendToken || "").trim();
  if (!token) {
    return NextResponse.json({ ok: false, error: "unauthorized" }, { status: 401 });
  }

  const requestText = await req.text();
  let payload: Record<string, unknown> = {};
  try {
    payload = JSON.parse(requestText || "{}") as Record<string, unknown>;
  } catch {
    return NextResponse.json({ ok: false, error: "Payload pembayaran tidak valid" }, { status: 400 });
  }
  const res = await fetch(`${apiBase()}/v1/me/agent-credit/payments`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: requestText,
    cache: "no-store",
  });
  const text = await res.text();
  if (res.ok) {
    await persistPaymentProofFallback(payload).catch(() => undefined);
  }
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}
