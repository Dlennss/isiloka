import { spawn } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import type { AgentCreditApplication, AgentCreditPayment } from "@/lib/api.auth";

function databaseUrl() {
  if (process.env.DATABASE_URL) return process.env.DATABASE_URL;
  const backendEnv = path.resolve(process.cwd(), "..", "pulsa-be", ".env");
  if (!existsSync(backendEnv)) return "";
  const line = readFileSync(backendEnv, "utf8")
    .split(/\r?\n/)
    .find((item) => item.trim().startsWith("DATABASE_URL="));
  return line ? line.replace(/^DATABASE_URL=/, "").trim().replace(/^"|"$/g, "") : "";
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

function parsePaymentNote(payment: AgentCreditPayment) {
  const rawNote = String(payment.note || "").trim();
  if (!rawNote.startsWith("{")) return payment;

  try {
    const parsed = JSON.parse(rawNote) as {
      note?: string;
      payment_method?: string;
      payment_proof?: AgentCreditPayment["payment_proof"];
    };
    return {
      ...payment,
      note: String(parsed.note || ""),
      payment_method: String(parsed.payment_method || ""),
      payment_proof: parsed.payment_proof,
    };
  } catch {
    return payment;
  }
}

export async function getAgentCreditApplicationsDatabaseFallback() {
  const dsn = databaseUrl();
  if (!dsn) return [] as AgentCreditApplication[];

  const sql = `
WITH applications AS (
  SELECT
    a.id,
    a.member_id,
    COALESCE(m.nama, '') AS member_name,
    COALESCE(m.email, '') AS member_email,
    COALESCE(m.phone, '') AS member_phone,
    a.requested_amount,
    a.approved_amount,
    a.status,
    COALESCE(a.applicant_data, '{}'::jsonb) AS applicant_data,
    COALESCE(a.document_data, '{}'::jsonb) AS document_data,
    COALESCE(a.agent_signature_data, '') AS agent_signature_data,
    (a.agent_signature_data IS NOT NULL AND a.agent_signature_data <> '') AS has_agent_signature,
    a.agent_signature_at::text AS agent_signature_at,
    COALESCE(a.marketing_note, '') AS marketing_note,
    COALESCE(a.analyst_note, '') AS analyst_note,
    COALESCE(a.analyst_recommendation, '') AS analyst_recommendation,
    COALESCE(a.analyst_recommended_amount, 0) AS analyst_recommended_amount,
    COALESCE(l.status, '') AS loan_status,
    COALESCE(l.outstanding_amount, 0) AS outstanding_amount,
    COALESCE(l.available_amount, 0) AS credit_available_amount,
    COALESCE(pay.paid_amount, 0) AS paid_amount,
    COALESCE(pay.payment_count, 0) AS payment_count,
    COALESCE(r.code, 'start') AS credit_level_code,
    COALESCE(r.name, 'Kilat Start') AS credit_level_name,
    COALESCE(r.limit_amount, 500000) AS credit_limit_amount,
    l.approved_at::text AS loan_approved_at,
    l.due_date::text AS loan_due_date,
    a.created_at::text AS created_at,
    a.updated_at::text AS updated_at
  FROM public.agent_credit_application a
  JOIN public.member m ON m.id = a.member_id
  LEFT JOIN public.agent_credit_loan l ON l.application_id = a.id
  LEFT JOIN public.agent_credit_rank r ON r.id = COALESCE(l.rank_id, a.rank_id)
  LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(p.amount), 0) AS paid_amount, COUNT(*) AS payment_count
    FROM public.agent_credit_payment p
    WHERE p.loan_id = l.id
  ) pay ON TRUE
  ORDER BY a.created_at DESC, a.id DESC
  LIMIT 200
)
SELECT COALESCE(json_agg(row_to_json(applications)), '[]'::json)::text FROM applications;
`;

  try {
    const stdout = await runPsql(dsn, sql);
    return JSON.parse(stdout.trim() || "[]") as AgentCreditApplication[];
  } catch {
    return [] as AgentCreditApplication[];
  }
}

export async function attachAgentCreditPaymentsFallback(applications: AgentCreditApplication[]) {
  if (!applications.length) return applications;

  const ids = applications.map((item) => Number(item.id)).filter((id) => Number.isFinite(id) && id > 0);
  if (!ids.length) return applications;

  const dsn = databaseUrl();
  if (!dsn) return applications;

  const sql = `
WITH selected AS (
  SELECT unnest(ARRAY[${ids.join(",")}]::bigint[]) AS application_id
),
rows AS (
  SELECT
    p.id,
    p.loan_id,
    l.application_id,
    p.member_id,
    p.amount,
    p.due_date::text AS due_date,
    p.paid_at::text AS paid_at,
    p.days_late,
    p.status,
    COALESCE(p.note, '') AS note
  FROM public.agent_credit_payment p
  JOIN public.agent_credit_loan l ON l.id = p.loan_id
  JOIN selected s ON s.application_id = l.application_id
  ORDER BY p.paid_at DESC, p.id DESC
)
SELECT COALESCE(json_agg(row_to_json(rows)), '[]'::json)::text FROM rows;
`;

  try {
    const stdout = await runPsql(dsn, sql);
    const rows = JSON.parse(stdout.trim() || "[]") as AgentCreditPayment[];
    const grouped = new Map<number, AgentCreditPayment[]>();

    rows.map(parsePaymentNote).forEach((payment) => {
      const key = Number(payment.application_id || 0);
      if (!key) return;
      grouped.set(key, [...(grouped.get(key) || []), payment]);
    });

    return applications.map((item) => {
      const payments = grouped.get(Number(item.id));
      if (!payments) return item;
      return {
        ...item,
        payments,
        payment_count: payments.length,
      };
    });
  } catch {
    return applications;
  }
}
