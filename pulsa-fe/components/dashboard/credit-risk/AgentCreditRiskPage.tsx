"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { BadgeCheck, ChevronDown, Loader2, Search, ShieldCheck, TrendingUp, WalletCards } from "lucide-react";
import type { AgentCreditApplication } from "@/lib/api.auth";

type Mode = "marketing" | "operator";

type Props = {
  applications: AgentCreditApplication[];
  mode: Mode;
};

type AgentSummary = {
  memberId: number;
  name: string;
  store: string;
  phone: string;
  email: string;
  currentLevel: string;
  currentLimit: number;
  approvedAmount: number;
  available: number;
  outstanding: number;
  paidCount: number;
  paymentCount: number;
  lateCount: number;
  monthlyCredit: number;
  latestStatus: string;
  latestUpdatedAt: string;
  latestMarketingNote: string;
};

function formatIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function formatDate(value?: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

function applicantText(item: AgentCreditApplication, key: string, fallback = "") {
  const value = item.applicant_data?.[key];
  return typeof value === "string" && value.trim() ? value.trim() : fallback;
}

function statusLabel(status: string, loanStatus?: string) {
  if (status === "approved") {
    if (loanStatus === "paid") return "Siklus selesai";
    if (loanStatus === "overdue") return "Perlu perhatian";
    return "Kredit aktif";
  }
  if (status === "analysis_review") return "Menunggu operator";
  if (status === "marketing_review" || status === "submitted") return "Di marketing";
  if (status.includes("rejected")) return "Ditolak";
  return status || "-";
}

function buildAgentSummaries(applications: AgentCreditApplication[]): AgentSummary[] {
  const grouped = new Map<number, AgentCreditApplication[]>();
  applications.forEach((item) => {
    const key = Number(item.member_id || 0);
    if (!key) return;
    grouped.set(key, [...(grouped.get(key) || []), item]);
  });

  return Array.from(grouped.entries()).map(([memberId, items]) => {
    const sorted = [...items].sort((a, b) => new Date(b.updated_at || b.created_at).getTime() - new Date(a.updated_at || a.created_at).getTime());
    const latest = sorted[0];
    const payments = sorted.flatMap((item) => item.payments || []);
    const lateCount = payments.filter((payment) => Number(payment.days_late || 0) > 0 || String(payment.status || "").toLowerCase() === "late").length;
    const paidCount = sorted.filter((item) => String(item.loan_status || "").toLowerCase() === "paid").length;
    const approved = sorted.filter((item) => item.status === "approved");
    const active = approved.find((item) => String(item.loan_status || "").toLowerCase() !== "paid") || approved[0] || latest;
    const now = new Date();
    const monthlyCredit = sorted
      .filter((item) => {
        const date = new Date(item.updated_at || item.created_at);
        return item.status === "approved" && date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear();
      })
      .reduce((sum, item) => sum + Number(item.approved_amount || item.requested_amount || 0), 0);

    return {
      memberId,
      name: applicantText(latest, "agent_name", latest.member_name || latest.agent_name || "Agent Isiloka"),
      store: applicantText(latest, "store_name", latest.store_name || applicantText(latest, "store_address", "Toko belum diisi")),
      phone: applicantText(latest, "whatsapp", latest.member_phone || "-"),
      email: applicantText(latest, "email", latest.member_email || latest.agent_email || "-"),
      currentLevel: latest.credit_level_name || "Kilat Start",
      currentLimit: Number(latest.credit_limit_amount || active.approved_amount || active.requested_amount || 500000),
      approvedAmount: Number(active.approved_amount || latest.approved_amount || 0),
      available: Number(active.credit_available_amount || 0),
      outstanding: Number(active.outstanding_amount || 0),
      paidCount,
      paymentCount: payments.length || sorted.reduce((sum, item) => sum + Number(item.payment_count || 0), 0),
      lateCount,
      monthlyCredit,
      latestStatus: statusLabel(latest.status, latest.loan_status),
      latestUpdatedAt: latest.updated_at || latest.created_at,
      latestMarketingNote: latest.marketing_note || "-",
    };
  });
}

export function AgentCreditRiskPage({ applications, mode }: Props) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [openId, setOpenId] = useState<number | null>(null);
  const [limitByMember, setLimitByMember] = useState<Record<number, string>>({});
  const [noteByMember, setNoteByMember] = useState<Record<number, string>>({});
  const [busyId, setBusyId] = useState<number | null>(null);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const summaries = useMemo(() => buildAgentSummaries(applications), [applications]);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return summaries;
    return summaries.filter((item) => [item.name, item.store, item.phone, item.email, String(item.memberId)].join(" ").toLowerCase().includes(q));
  }, [query, summaries]);

  const activeDebt = summaries.reduce((sum, item) => sum + item.outstanding, 0);
  const availableCredit = summaries.reduce((sum, item) => sum + item.available, 0);
  const upgradeCandidates = summaries.filter((item) => item.lateCount === 0 && item.paidCount >= 3).length;

  async function changeLimit(agent: AgentSummary) {
    const limitAmount = Number(String(limitByMember[agent.memberId] || "").replace(/\D/g, ""));
    const reason = String(noteByMember[agent.memberId] || "").trim();
    if (!limitAmount || !reason) {
      setMessage({ type: "error", text: "Isi nominal limit dan catatan keputusan dulu." });
      return;
    }
    setBusyId(agent.memberId);
    setMessage(null);
    try {
      const token = window.localStorage.getItem("auth_token") || "";
      const response = await fetch("/api/agent-credit/ranks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ member_id: agent.memberId, limit_amount: limitAmount, reason }),
      });
      const body = (await response.json().catch(() => ({}))) as { ok?: boolean; error?: string };
      if (!response.ok || !body.ok) throw new Error(body.error || "Limit gagal diubah");
      setMessage({ type: "success", text: `Limit ${agent.name} berhasil diperbarui oleh Operator Kredit.` });
      router.refresh();
    } catch (err) {
      setMessage({ type: "error", text: err instanceof Error ? err.message : "Limit gagal diubah" });
    } finally {
      setBusyId(null);
    }
  }

  function recommend(agent: AgentSummary) {
    const reason = String(noteByMember[agent.memberId] || "").trim();
    if (!reason) {
      setMessage({ type: "error", text: "Isi catatan rekomendasi dulu." });
      return;
    }
    setMessage({ type: "success", text: `Rekomendasi ${agent.name} siap dibahas Operator Kredit. Keputusan limit tetap hanya bisa dilakukan Operator Kredit.` });
  }

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto flex w-full max-w-7xl flex-col gap-5">
        <div className="overflow-hidden rounded-[28px] border border-emerald-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.08)]">
          <div className="relative isolate overflow-hidden bg-[radial-gradient(circle_at_90%_10%,rgba(190,242,100,0.55),transparent_28%),linear-gradient(135deg,#052e26_0%,#047857_55%,#3bd64a_110%)] px-5 py-7 text-white sm:px-7 lg:px-9">
            <p className="mb-3 inline-flex rounded-full border border-white/15 bg-white/12 px-3 py-1 text-[11px] font-black uppercase tracking-[0.24em] text-lime-100">
              {mode === "operator" ? "Kontrol Risiko" : "Manajemen Agent"}
            </p>
            <h1 className="text-3xl font-black tracking-normal sm:text-4xl">{mode === "operator" ? "Kenaikan Limit Agent" : "Profil Agent Binaan"}</h1>
            <p className="mt-3 max-w-2xl text-sm font-medium leading-6 text-emerald-50/90">
              {mode === "operator"
                ? "Operator Kredit memutuskan kenaikan tier berdasarkan aktivitas modal, saldo utama, transaksi, dan catatan marketing."
                : "Marketing memantau saldo utama dan transaksi agent, lalu memberi rekomendasi tanpa bisa menaikkan limit langsung."}
            </p>
          </div>

          <div className="space-y-5 p-4 sm:p-6 lg:p-7">
            <div className="grid min-w-0 grid-cols-[repeat(auto-fit,minmax(min(100%,220px),1fr))] gap-3">
              {[
                { label: "Total Agent", value: summaries.length, hint: "Punya histori kredit", icon: ShieldCheck },
                { label: "Kandidat Naik", value: upgradeCandidates, hint: "Siklus selesai tepat waktu", icon: TrendingUp },
                { label: "Modal Berjalan", value: formatIDR(activeDebt), hint: "Modal yang masih digunakan", icon: WalletCards },
                { label: "Kredit Tersedia", value: formatIDR(availableCredit), hint: "Bisa dimutasi agent", icon: BadgeCheck },
              ].map((item) => {
                const Icon = item.icon;
                return (
                  <div key={item.label} className="min-w-0 rounded-3xl border border-slate-200 bg-white p-4 shadow-[0_14px_32px_rgba(15,23,42,0.06)]">
                    <div className="flex items-center justify-between gap-4">
                      <div className="min-w-0">
                        <p className="truncate text-xs font-bold text-slate-500">{item.label}</p>
                        <p className="mt-1 break-words text-2xl font-black text-slate-950">{item.value}</p>
                        <p className="mt-1 line-clamp-2 text-xs font-semibold leading-4 text-slate-400">{item.hint}</p>
                      </div>
                      <span className="grid h-12 w-12 place-items-center rounded-2xl bg-emerald-50 text-emerald-700">
                        <Icon className="h-6 w-6" />
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>

            <section className="rounded-[26px] border border-slate-200 bg-white p-3 shadow-[0_18px_42px_rgba(15,23,42,0.06)] sm:p-5">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <label className="flex h-12 min-w-0 flex-1 items-center gap-2 rounded-2xl border border-slate-200 bg-slate-50 px-4 text-sm font-bold text-slate-500 focus-within:border-emerald-300 focus-within:bg-white">
                  <Search className="h-4 w-4" />
                  <input value={query} onChange={(event) => setQuery(event.target.value)} className="w-full bg-transparent outline-none" placeholder="Cari agent, toko, WA, email" />
                </label>
                <span className="w-fit rounded-full bg-emerald-100 px-4 py-2 text-xs font-black text-emerald-700">
                  {filtered.length} agent tampil
                </span>
              </div>

              {message ? (
                <div className={message.type === "success" ? "mt-4 rounded-2xl bg-emerald-50 px-4 py-3 text-xs font-black text-emerald-700" : "mt-4 rounded-2xl bg-rose-50 px-4 py-3 text-xs font-black text-rose-600"}>
                  {message.text}
                </div>
              ) : null}

              <div className="mt-4 overflow-hidden rounded-2xl border border-slate-200">
                <div className="hidden grid-cols-[minmax(220px,1.4fr)_140px_145px_145px_120px] gap-3 border-b border-slate-200 bg-slate-50 px-4 py-3 text-[10px] font-black uppercase tracking-[0.12em] text-slate-500 lg:grid">
                  <span>Agent</span>
                  <span>Tier</span>
                  <span>Saldo Kredit</span>
                  <span>Performa</span>
                  <span className="text-center">Aksi</span>
                </div>
                {filtered.length ? filtered.map((agent) => {
                  const open = openId === agent.memberId;
                  return (
                    <article key={agent.memberId} className="border-b border-slate-100 bg-white last:border-b-0">
                      <div className="grid gap-3 p-4 lg:grid-cols-[minmax(220px,1.4fr)_140px_145px_145px_120px] lg:items-center">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-black text-slate-950">{agent.name}</p>
                          <p className="mt-1 truncate text-xs font-semibold text-slate-500">{agent.store}</p>
                          <p className="mt-1 truncate text-[11px] font-bold text-slate-400">{agent.phone} - {agent.email}</p>
                        </div>
                        <div>
                          <p className="text-sm font-black text-slate-950">{agent.currentLevel}</p>
                          <p className="mt-1 text-xs font-bold text-emerald-700">{formatIDR(agent.currentLimit)}</p>
                        </div>
                        <div>
                          <p className="text-sm font-black text-slate-950">{formatIDR(agent.available)}</p>
                          <p className="mt-1 text-xs font-bold text-amber-700">Modal berjalan {formatIDR(agent.outstanding)}</p>
                        </div>
                        <div>
                          <p className="text-sm font-black text-slate-950">{agent.paidCount} selesai</p>
                          <p className={agent.lateCount ? "mt-1 text-xs font-bold text-rose-600" : "mt-1 text-xs font-bold text-emerald-700"}>{agent.lateCount} telat</p>
                        </div>
                        <button
                          type="button"
                          onClick={() => setOpenId(open ? null : agent.memberId)}
                          className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 text-xs font-black text-slate-600 transition hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-700"
                        >
                          Detail
                          <ChevronDown className={`h-4 w-4 transition ${open ? "rotate-180" : ""}`} />
                        </button>
                      </div>

                      {open ? (
                        <div className="border-t border-emerald-100 bg-emerald-50/40 p-4">
                          <div className="grid gap-3 lg:grid-cols-3">
                            <div className="rounded-2xl bg-white p-4 ring-1 ring-emerald-100">
                              <p className="text-xs font-black uppercase text-emerald-700">Status Terakhir</p>
                              <p className="mt-2 text-sm font-black text-slate-950">{agent.latestStatus}</p>
                              <p className="mt-1 text-xs font-semibold text-slate-500">{formatDate(agent.latestUpdatedAt)}</p>
                            </div>
                            <div className="rounded-2xl bg-white p-4 ring-1 ring-emerald-100">
                              <p className="text-xs font-black uppercase text-emerald-700">Transaksi Bulan Ini</p>
                              <p className="mt-2 text-sm font-black text-slate-950">{formatIDR(agent.monthlyCredit)}</p>
                              <p className="mt-1 text-xs font-semibold text-slate-500">Dari nominal kredit disetujui</p>
                            </div>
                            <div className="rounded-2xl bg-white p-4 ring-1 ring-emerald-100">
                              <p className="text-xs font-black uppercase text-emerald-700">Catatan Marketing</p>
                              <p className="mt-2 text-xs font-semibold leading-5 text-slate-600">{agent.latestMarketingNote}</p>
                            </div>
                          </div>

                          <div className="mt-3 grid gap-3 lg:grid-cols-[1fr_220px]">
                            <label className="block rounded-2xl border border-slate-200 bg-white p-3">
                              <span className="text-[10px] font-black uppercase tracking-[0.12em] text-emerald-700">
                                {mode === "operator" ? "Catatan Keputusan Operator" : "Catatan Rekomendasi Marketing"}
                              </span>
                              <textarea
                                value={noteByMember[agent.memberId] || ""}
                                onChange={(event) => setNoteByMember((current) => ({ ...current, [agent.memberId]: event.target.value }))}
                                rows={3}
                                className="mt-2 w-full resize-none bg-transparent text-sm font-semibold text-slate-700 outline-none placeholder:text-slate-400"
                                placeholder={mode === "operator" ? "Contoh: pembayaran lancar, naik ke Kilat Plus." : "Contoh: agent ramai dan layak diajukan naik limit."}
                              />
                            </label>
                            <div className="rounded-2xl border border-slate-200 bg-white p-3">
                              {mode === "operator" ? (
                                <>
                                  <label className="block">
                                    <span className="text-[10px] font-black uppercase tracking-[0.12em] text-emerald-700">Naikkan/Turunkan Limit</span>
                                    <input
                                      type="number"
                                      min="1"
                                      step="1"
                                      value={limitByMember[agent.memberId] || ""}
                                      onChange={(event) => setLimitByMember((current) => ({ ...current, [agent.memberId]: event.target.value }))}
                                      placeholder={String(agent.currentLimit || 500000)}
                                      className="mt-2 h-11 w-full rounded-2xl border border-slate-200 bg-slate-50 px-3 text-xs font-black text-slate-950 outline-none"
                                    />
                                  </label>
                                  <button
                                    type="button"
                                    onClick={() => void changeLimit(agent)}
                                    disabled={busyId === agent.memberId}
                                    className="mt-3 inline-flex h-11 w-full items-center justify-center gap-2 rounded-2xl bg-[linear-gradient(135deg,#087e8b,#84cc16)] text-xs font-black text-white shadow-[0_12px_24px_rgba(4,120,87,0.18)] disabled:opacity-60"
                                  >
                                    {busyId === agent.memberId ? <Loader2 className="h-4 w-4 animate-spin" /> : <TrendingUp className="h-4 w-4" />}
                                    Simpan Limit
                                  </button>
                                </>
                              ) : (
                                <button
                                  type="button"
                                  onClick={() => recommend(agent)}
                                  className="inline-flex h-full min-h-24 w-full items-center justify-center rounded-2xl bg-emerald-800 px-4 text-xs font-black text-white shadow-[0_12px_24px_rgba(4,120,87,0.18)]"
                                >
                                  Ajukan Rekomendasi
                                </button>
                              )}
                            </div>
                          </div>
                        </div>
                      ) : null}
                    </article>
                  );
                }) : (
                  <div className="p-10 text-center text-sm font-bold text-slate-400">Belum ada data agent untuk dipantau.</div>
                )}
              </div>
            </section>
          </div>
        </div>
      </section>
    </main>
  );
}
