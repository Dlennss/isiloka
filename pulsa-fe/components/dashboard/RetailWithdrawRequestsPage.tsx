"use client";

import { useEffect, useMemo, useState } from "react";
import { BadgeCheck, Clock3, RefreshCw, Search, ShieldX, WalletCards } from "lucide-react";
import { alertError, alertSuccess } from "@/components/ui/alerts";
import WithdrawApproveModal from "@/components/dashboard/WithdrawApproveModal";

type WithdrawRow = {
  id: number;
  ref_id: string;
  member_nama?: string | null;
  member_email?: string | null;
  amount: number;
  source_type?: "main_balance" | "credit";
  bank_name: string;
  account_name: string;
  account_number: string;
  status: string;
  note?: string;
  reject_reason?: string;
  created_at?: string;
};

function authHeader(): Record<string, string> {
  const token = localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function fmtIDR(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function sourceLabel(source?: string) {
  return source === "credit" ? "Saldo Kredit" : "Saldo Utama";
}

function statusBadge(status: string) {
  if (status === "approved") return "bg-emerald-100 text-emerald-700";
  if (status === "rejected") return "bg-rose-100 text-rose-700";
  if (status === "processing_provider") return "bg-sky-100 text-sky-700";
  return "bg-amber-100 text-amber-700";
}

function statusLabel(status: string) {
  if (status === "approved") return "Berhasil";
  if (status === "rejected") return "Ditolak";
  if (status === "processing_provider") return "Diproses Pulsa24Jam";
  return "Menunggu";
}

export default function RetailWithdrawRequestsPage() {
  const [items, setItems] = useState<WithdrawRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState("pending");
  const [query, setQuery] = useState("");
  const [approveTarget, setApproveTarget] = useState<WithdrawRow | null>(null);
  const [approving, setApproving] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const params = new URLSearchParams({ status, q: query, limit: "100", offset: "0" });
      const response = await fetch(`/api/admin/retail/withdraw-requests?${params.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const body = await response.json().catch(() => ({}));
      if (!response.ok || !body?.ok) throw new Error(body?.error || "Gagal memuat penarikan agent.");
      setItems(Array.isArray(body.items) ? body.items : []);
    } catch (loadError) {
      await alertError(loadError instanceof Error ? loadError.message : "Gagal memuat penarikan agent.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [status]);

  async function approve(payload: { bank_id: number; fee: number; note: string }) {
    if (!approveTarget) return;
    setApproving(true);
    try {
      const response = await fetch(`/api/admin/retail/withdraw-requests/approve?id=${approveTarget.id}`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
      });
      const body = await response.json().catch(() => ({}));
      if (!response.ok || !body?.ok) throw new Error(body?.error || "Penarikan gagal disetujui.");
      await alertSuccess("Penarikan agent berhasil disetujui.");
      setApproveTarget(null);
      await load();
    } catch (approveError) {
      await alertError(approveError instanceof Error ? approveError.message : "Penarikan gagal disetujui.");
    } finally {
      setApproving(false);
    }
  }

  async function reject(item: WithdrawRow) {
    const reason = window.prompt(`Alasan menolak penarikan ${item.ref_id}:`);
    if (!reason?.trim()) return;
    const response = await fetch(`/api/admin/retail/withdraw-requests/reject?id=${item.id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({ reason: reason.trim() }),
    });
    const body = await response.json().catch(() => ({}));
    if (!response.ok || !body?.ok) {
      await alertError(body?.error || "Penarikan gagal ditolak.");
      return;
    }
    await alertSuccess(`Penarikan ditolak dan dana kembali ke ${sourceLabel(item.source_type).toLowerCase()}.`);
    await load();
  }

  const stats = useMemo(() => ({
    pending: items.filter((item) => item.status === "pending" || item.status === "processing_provider").length,
    approved: items.filter((item) => item.status === "approved").length,
    rejected: items.filter((item) => item.status === "rejected").length,
    amount: items.filter((item) => item.status === "pending" || item.status === "processing_provider").reduce((total, item) => total + Number(item.amount || 0), 0),
  }), [items]);

  return (
    <main className="min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <section className="mx-auto w-full max-w-7xl space-y-4">
        <header className="border-b border-emerald-200 bg-white px-4 py-5 sm:px-6">
          <p className="text-[10px] font-black uppercase tracking-[0.2em] text-emerald-700">Operator Kredit</p>
          <h1 className="mt-1 text-2xl font-black sm:text-3xl">Penarikan Agent</h1>
          <p className="mt-2 max-w-2xl text-sm font-semibold leading-6 text-slate-500">Periksa sumber saldo, rekening tujuan, dan nominal sebelum dana dikirim.</p>
        </header>

        <section className="grid grid-cols-2 gap-2 lg:grid-cols-4">
          {[
            { label: "Menunggu", value: stats.pending, hint: fmtIDR(stats.amount), icon: Clock3, color: "text-amber-700 bg-amber-50" },
            { label: "Disetujui", value: stats.approved, hint: "Sudah diproses", icon: BadgeCheck, color: "text-emerald-700 bg-emerald-50" },
            { label: "Ditolak", value: stats.rejected, hint: "Dana dikembalikan", icon: ShieldX, color: "text-rose-700 bg-rose-50" },
            { label: "Total Data", value: items.length, hint: "Sesuai filter", icon: WalletCards, color: "text-sky-700 bg-sky-50" },
          ].map((item) => {
            const Icon = item.icon;
            return (
              <div key={item.label} className="rounded-lg border border-slate-200 bg-white p-3 sm:p-4">
                <span className={`grid h-9 w-9 place-items-center rounded-lg ${item.color}`}><Icon className="h-4 w-4" /></span>
                <p className="mt-3 text-[10px] font-black uppercase text-slate-400">{item.label}</p>
                <p className="mt-1 text-xl font-black">{item.value}</p>
                <p className="mt-1 truncate text-[10px] font-semibold text-slate-500">{item.hint}</p>
              </div>
            );
          })}
        </section>

        <section className="flex flex-col gap-2 border border-slate-200 bg-white p-3 sm:flex-row">
          <div className="flex h-11 flex-1 items-center gap-2 rounded-lg border border-slate-200 px-3">
            <Search className="h-4 w-4 text-slate-400" />
            <input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter") void load(); }} placeholder="Cari agent, referensi, atau bank" className="min-w-0 flex-1 text-sm font-semibold outline-none" />
          </div>
          <select value={status} onChange={(event) => setStatus(event.target.value)} className="h-11 rounded-lg border border-slate-200 bg-white px-3 text-sm font-bold outline-none">
            <option value="pending">Menunggu</option>
            <option value="processing_provider">Diproses Pulsa24Jam</option>
            <option value="approved">Disetujui</option>
            <option value="rejected">Ditolak</option>
            <option value="all">Semua</option>
          </select>
          <button type="button" onClick={() => void load()} className="inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-emerald-700 px-4 text-xs font-black text-white"><RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />Muat Ulang</button>
        </section>

        <section className="overflow-hidden border border-slate-200 bg-white">
          <div className="hidden overflow-x-auto md:block">
            <table className="w-full min-w-[900px] border-collapse text-left">
              <thead className="bg-emerald-50 text-[10px] font-black uppercase text-emerald-800">
                <tr><th className="px-4 py-3">Agent</th><th className="px-4 py-3">Sumber</th><th className="px-4 py-3">Rekening Tujuan</th><th className="px-4 py-3">Nominal</th><th className="px-4 py-3">Status</th><th className="px-4 py-3 text-right">Aksi</th></tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-sm">
                {items.map((item) => (
                  <tr key={item.id} className="align-top">
                    <td className="px-4 py-4"><p className="font-black">{item.member_nama || "Agent Isiloka"}</p><p className="mt-1 text-[10px] font-semibold text-slate-500">{item.member_email || item.ref_id}</p></td>
                    <td className="px-4 py-4"><span className="rounded-full bg-sky-50 px-2.5 py-1 text-[10px] font-black text-sky-700">{sourceLabel(item.source_type)}</span></td>
                    <td className="px-4 py-4"><p className="font-bold">{item.bank_name} · {item.account_number}</p><p className="mt-1 text-[10px] font-semibold text-slate-500">{item.account_name}</p></td>
                    <td className="px-4 py-4 font-black text-emerald-700">{fmtIDR(item.amount)}</td>
                    <td className="px-4 py-4"><span className={`rounded-full px-2.5 py-1 text-[10px] font-black ${statusBadge(item.status)}`}>{statusLabel(item.status)}</span></td>
                    <td className="px-4 py-4"><div className="flex justify-end gap-2">{item.status === "pending" ? <><button onClick={() => setApproveTarget(item)} className="h-9 rounded-lg bg-emerald-700 px-3 text-[10px] font-black text-white">Setujui</button><button onClick={() => void reject(item)} className="h-9 rounded-lg border border-rose-200 px-3 text-[10px] font-black text-rose-600">Tolak</button></> : <span className="text-[10px] font-semibold text-slate-400">Selesai</span>}</div></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="divide-y divide-slate-100 md:hidden">
            {items.map((item) => (
              <article key={item.id} className="p-4">
                <div className="flex items-start justify-between gap-3"><div><p className="text-sm font-black">{item.member_nama || "Agent Isiloka"}</p><p className="mt-1 text-[10px] font-bold text-sky-700">{sourceLabel(item.source_type)}</p></div><p className="text-sm font-black text-emerald-700">{fmtIDR(item.amount)}</p></div>
                <p className="mt-3 text-xs font-semibold text-slate-600">{item.bank_name} · {item.account_number} · {item.account_name}</p>
                <div className="mt-3 flex items-center justify-between gap-2"><span className={`rounded-full px-2.5 py-1 text-[9px] font-black ${statusBadge(item.status)}`}>{statusLabel(item.status)}</span>{item.status === "pending" ? <div className="flex gap-2"><button onClick={() => setApproveTarget(item)} className="h-9 rounded-lg bg-emerald-700 px-3 text-[10px] font-black text-white">Setujui</button><button onClick={() => void reject(item)} className="h-9 rounded-lg border border-rose-200 px-3 text-[10px] font-black text-rose-600">Tolak</button></div> : null}</div>
              </article>
            ))}
          </div>

          {!loading && items.length === 0 ? <div className="px-4 py-12 text-center text-sm font-semibold text-slate-500">Belum ada penarikan untuk filter ini.</div> : null}
          {loading ? <div className="px-4 py-12 text-center text-sm font-semibold text-slate-500">Memuat antrean penarikan...</div> : null}
        </section>
      </section>

      <WithdrawApproveModal open={Boolean(approveTarget)} title="Setujui Penarikan Agent" amount={approveTarget?.amount || 0} banksEndpoint="/api/admin/retail/withdraw-requests/banks" submitting={approving} onClose={() => { if (!approving) setApproveTarget(null); }} onSubmit={approve} />
    </main>
  );
}
