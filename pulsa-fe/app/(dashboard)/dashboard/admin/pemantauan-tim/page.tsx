"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Activity, CheckCircle2, Clock3, RefreshCcw,
  ShieldCheck, UserCog, Users, XCircle,
} from "lucide-react";

type TeamMember = {
  id: number;
  nama?: string;
  email?: string;
  phone?: string;
  role: string;
  aktif: boolean;
};

type TeamActivity = {
  id: number;
  actor_id: number;
  actor_name?: string;
  actor_email?: string;
  actor_role: string;
  action: string;
  reason?: string;
  created_at: string;
};

type CreditApplication = {
  id: number;
  marketing_id?: number;
  analyst_id?: number;
  status: string;
  loan_status?: string;
  outstanding_amount?: number;
  approved_amount?: number;
  loan_approved_at?: string;
  created_at: string;
  updated_at: string;
};

type ApiResult = {
  ok: boolean;
  status: number;
  data: Record<string, unknown>;
  error: string;
};

const MARKETING_QUEUE = new Set(["submitted", "marketing_review"]);
const OPERATOR_QUEUE = new Set(["analysis_review", "ready_to_disburse"]);

async function fetchAdminData(url: string): Promise<ApiResult> {
  let lastError = "Koneksi ke server terputus";
  for (let attempt = 0; attempt < 2; attempt += 1) {
    try {
      const response = await fetch(url, { cache: "no-store", credentials: "same-origin" });
      const data = await response.json().catch(() => ({})) as Record<string, unknown>;
      const error = typeof data.error === "string" ? data.error : "Data belum dapat dimuat";
      return response.ok
        ? { ok: true, status: response.status, data, error: "" }
        : { ok: false, status: response.status, data, error };
    } catch (cause) {
      lastError = cause instanceof Error && cause.message !== "Failed to fetch"
        ? cause.message
        : "Koneksi ke server terputus";
      if (attempt === 0) await new Promise((resolve) => setTimeout(resolve, 600));
    }
  }
  return { ok: false, status: 0, data: {}, error: lastError };
}

function roleLabel(role: string) {
  if (role === "marketing" || role === "master") return "Marketing";
  if (role === "operator_credit" || role === "analis") return "Operator Kredit";
  return role || "Sistem";
}

function actionLabel(action: string) {
  switch (action) {
    case "agent_credit_marketing_review": return "Dokumen diteruskan ke operator";
    case "agent_credit_final_decision": return "Keputusan kredit dibuat";
    case "agent_credit_operational_status_changed": return "Status kredit agent diubah";
    default: return action.replaceAll("_", " ");
  }
}

function dateTime(value?: string) {
  if (!value) return "Belum ada aktivitas";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Belum ada aktivitas";
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Jakarta",
  }).format(date);
}

function TeamMonitoringContent() {
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [activities, setActivities] = useState<TeamActivity[]>([]);
  const [applications, setApplications] = useState<CreditApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadData = useCallback(async () => {
    setLoading(true);
    setError("");
    const results = await Promise.all([
      fetchAdminData("/api/admin/members?scope=retail&role=marketing&limit=200"),
      fetchAdminData("/api/admin/agent-credit/team-activity?limit=100"),
      fetchAdminData("/api/agent-credit/applications"),
    ]);

    if (results.every((item) => item.status === 401)) {
      setError("Sesi login Admin tidak valid. Silakan login kembali.");
      setLoading(false);
      return;
    }

    const rows = (result: ApiResult) => result.ok
      ? (Array.isArray(result.data.items) ? result.data.items : Array.isArray(result.data.rows) ? result.data.rows : [])
      : [];
    setMembers(rows(results[0]) as TeamMember[]);
    setActivities(rows(results[1]) as TeamActivity[]);
    setApplications(rows(results[2]) as CreditApplication[]);

    const failures = results.filter((item) => !item.ok);
    if (failures.length === results.length) setError("Data pemantauan belum dapat dimuat. Silakan muat ulang.");
    else if (failures.some((item) => item.status === 403)) setError("Sebagian data operasional belum dapat dimuat. Tekan Muat Ulang untuk mencoba kembali.");
    else if (failures.length > 0) setError("Sebagian data belum dapat dimuat. Tekan Muat Ulang untuk mencoba kembali.");
    setLoading(false);
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => { void loadData(); }, 0);
    return () => window.clearTimeout(timer);
  }, [loadData]);

  const marketing = useMemo(() => members.filter((item) => item.role === "marketing"), [members]);
  const marketingQueue = applications.filter((item) => MARKETING_QUEUE.has(item.status));
  const operatorQueue = applications.filter((item) => OPERATOR_QUEUE.has(item.status));

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <div className="mx-auto w-full max-w-7xl space-y-5">
        <header className="overflow-hidden rounded-[28px] bg-[linear-gradient(135deg,#053a2f_0%,#087a50_58%,#50cf3e_100%)] px-5 py-7 text-white shadow-[0_22px_50px_rgba(6,78,59,0.16)] sm:px-8">
          <p className="text-[11px] font-black uppercase tracking-[0.2em] text-lime-100">Pengawasan Admin</p>
          <div className="mt-2 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div>
              <h1 className="text-3xl font-black tracking-normal sm:text-4xl">Pemantauan Operasional Kredit</h1>
              <p className="mt-2 max-w-3xl text-sm font-semibold leading-6 text-emerald-50/90">
                Pantau kinerja seluruh marketing, antrean operator kredit, dan pengajuan yang perlu perhatian.
              </p>
            </div>
            <button type="button" onClick={() => void loadData()} className="inline-flex min-h-11 items-center justify-center gap-2 rounded-2xl border border-white/30 bg-white/15 px-4 text-sm font-black text-white transition hover:bg-white/25">
              <RefreshCcw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} /> Muat Ulang
            </button>
          </div>
        </header>

        <section className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          {[
            { label: "Total Marketing", value: marketing.length, icon: Users },
            { label: "Marketing Aktif", value: marketing.filter((item) => item.aktif).length, icon: CheckCircle2 },
            { label: "Menunggu Marketing", value: marketingQueue.length, icon: Clock3 },
            { label: "Antrean Operator", value: operatorQueue.length, icon: ShieldCheck },
          ].map((item) => {
            const Icon = item.icon;
            return <div key={item.label} className="flex min-h-28 items-center justify-between rounded-[22px] border border-slate-200 bg-white p-4 shadow-[0_12px_28px_rgba(15,23,42,0.06)]"><div><p className="text-xs font-bold text-slate-600">{item.label}</p><p className="mt-2 text-2xl font-black text-[#043f32]">{loading ? "..." : item.value}</p></div><span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-100 text-emerald-800"><Icon className="h-5 w-5" /></span></div>;
          })}
        </section>

        {error ? <div className="rounded-2xl border border-rose-300 bg-rose-50 px-4 py-3 text-sm font-bold text-rose-800">{error}</div> : null}

        <section>
          <div className="overflow-hidden rounded-[24px] border border-slate-200 bg-white shadow-[0_14px_34px_rgba(15,23,42,0.06)]">
            <div className="flex items-center justify-between gap-3 border-b border-slate-200 px-5 py-4">
              <div><h2 className="text-lg font-black">Kinerja Marketing</h2><p className="mt-1 text-xs font-semibold text-slate-600">Ringkasan penanganan kredit per akun marketing.</p></div>
              <Link href="/dashboard/admin/master/members" className="inline-flex min-h-10 shrink-0 items-center gap-2 rounded-xl border border-emerald-400 bg-emerald-50 px-3 text-xs font-black !text-emerald-900"><UserCog className="h-4 w-4" /> Kelola Akun</Link>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[780px] text-left text-sm">
                <thead className="bg-[#e5f6ee] text-[11px] font-black uppercase text-emerald-900"><tr><th className="px-5 py-3">Marketing</th><th className="px-4 py-3">Status</th><th className="px-4 py-3 text-center">Ditangani</th><th className="px-4 py-3 text-center">Ke Operator</th><th className="px-4 py-3 text-center">Disetujui</th><th className="px-5 py-3">Aktivitas Terakhir</th></tr></thead>
                <tbody className="divide-y divide-slate-100">
                  {marketing.map((member) => {
                    const assigned = applications.filter((item) => item.marketing_id === member.id);
                    const forwarded = assigned.filter((item) => !MARKETING_QUEUE.has(item.status));
                    const approved = assigned.filter((item) => item.status === "approved");
                    const last = activities.find((item) => item.actor_id === member.id);
                    return <tr key={member.id} className="hover:bg-emerald-50/40"><td className="px-5 py-4"><p className="font-black">{member.nama || "Marketing"}</p><p className="mt-1 text-xs text-slate-600">{member.email || member.phone || "-"}</p></td><td className="px-4 py-4"><span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-black ${member.aktif ? "bg-emerald-100 text-emerald-800" : "bg-rose-100 text-rose-800"}`}>{member.aktif ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}{member.aktif ? "Aktif" : "Nonaktif"}</span></td><td className="px-4 py-4 text-center font-black">{assigned.length}</td><td className="px-4 py-4 text-center font-black">{forwarded.length}</td><td className="px-4 py-4 text-center font-black text-emerald-700">{approved.length}</td><td className="px-5 py-4 text-xs font-semibold text-slate-600">{dateTime(last?.created_at)}</td></tr>;
                  })}
                  {!loading && marketing.length === 0 ? <tr><td colSpan={6} className="px-5 py-10 text-center font-semibold text-slate-600">Belum ada akun marketing.</td></tr> : null}
                </tbody>
              </table>
            </div>
          </div>

        </section>

        <section>
          <div className="overflow-hidden rounded-[24px] border border-slate-200 bg-white shadow-[0_14px_34px_rgba(15,23,42,0.06)]">
            <div className="border-b border-slate-200 px-5 py-4"><h2 className="flex items-center gap-2 text-lg font-black"><Activity className="h-5 w-5 text-emerald-700" /> Aktivitas Kredit Terbaru</h2></div>
            <div className="max-h-[360px] divide-y divide-slate-100 overflow-y-auto">
              {activities.slice(0, 12).map((item) => <article key={item.id} className="grid gap-2 px-5 py-4 sm:grid-cols-[minmax(0,1fr)_auto]"><div><div className="flex flex-wrap items-center gap-2"><p className="font-black">{item.actor_name || item.actor_email || roleLabel(item.actor_role)}</p><span className="rounded-full bg-emerald-100 px-2 py-1 text-[11px] font-black text-emerald-800">{roleLabel(item.actor_role)}</span></div><p className="mt-1 text-sm font-bold text-slate-700">{actionLabel(item.action)}</p>{item.reason ? <p className="mt-1 text-xs font-semibold text-slate-600">Catatan: {item.reason}</p> : null}</div><div className="flex items-center gap-1.5 text-xs font-semibold text-slate-600"><Clock3 className="h-3.5 w-3.5" /> {dateTime(item.created_at)}</div></article>)}
              {!loading && activities.length === 0 ? <div className="px-5 py-10 text-center font-semibold text-slate-600">Belum ada aktivitas kredit yang tercatat.</div> : null}
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}

export default function AdminTeamMonitoringPage() {
  return <TeamMonitoringContent />;
}
