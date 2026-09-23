"use client";

import { useCallback, useEffect, useState } from "react";
import { RefreshCcw, Search, ReceiptText } from "lucide-react";
import { fmtID } from "@/lib/format";

type AgentTransaction = {
  id: number;
  member_email?: string;
  member_nama?: string;
  ref_id: string;
  kode_produk: string;
  produk_nama?: string;
  tujuan: string;
  qty: number;
  status: string;
  biaya_aktual: number;
  biaya_perkiraan: number;
  dibuat_pada: string;
};

function statusClass(status: string) {
  const value = status.toLowerCase();
  if (value === "success" || value === "sukses") return "bg-emerald-100 text-emerald-700";
  if (value === "failed" || value === "gagal" || value === "refunded") return "bg-rose-100 text-rose-700";
  return "bg-amber-100 text-amber-700";
}

function statusLabel(status: string) {
  const value = status.toLowerCase();
  if (value === "success") return "Berhasil";
  if (value === "failed" || value === "refunded") return "Dana Dikembalikan";
  return status || "-";
}

function dateLabel(value: string) {
  if (!value) return "-";
  return new Date(value).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
}

export default function OperatorAgentTransactionsPage() {
  const [items, setItems] = useState<AgentTransaction[]>([]);
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const [draftSearch, setDraftSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = localStorage.getItem("auth_token") || "";
      const query = new URLSearchParams({ limit: "100" });
      if (status) query.set("status", status);
      if (search) query.set("q", search);
      const response = await fetch(`/api/admin/agent-credit/transactions?${query.toString()}`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        cache: "no-store",
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok || !data.ok) throw new Error(data.error || "Data transaksi agent belum dapat dimuat.");
      setItems(Array.isArray(data.items) ? data.items : []);
    } catch (loadError) {
      setItems([]);
      setError(loadError instanceof Error ? loadError.message : "Data transaksi agent belum dapat dimuat.");
    } finally {
      setLoading(false);
    }
  }, [search, status]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="space-y-5 p-3 sm:p-5 lg:p-7">
      <section className="rounded-[28px] bg-[linear-gradient(135deg,#13464b,#07865f_58%,#3dbb42)] p-5 text-white shadow-[0_18px_40px_rgba(6,78,59,0.16)] sm:p-7">
        <p className="text-[10px] font-black uppercase tracking-[0.2em] text-lime-200">Monitoring operator</p>
        <div className="mt-2 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
          <div>
            <h1 className="text-2xl font-black sm:text-3xl">Transaksi Terakhir Agent</h1>
            <p className="mt-2 max-w-2xl text-sm font-semibold leading-6 text-emerald-50/85">Pantau satu aktivitas paling baru dari setiap agent, termasuk pembelian produk dan penarikan saldo.</p>
          </div>
          <button type="button" onClick={() => void load()} className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-white/30 bg-white/10 px-4 text-xs font-black hover:bg-white/20" disabled={loading}>
            <RefreshCcw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} /> Muat Ulang
          </button>
        </div>
      </section>

      <section className="rounded-3xl border border-emerald-100 bg-white p-4 shadow-[0_12px_32px_rgba(6,78,59,0.06)] sm:p-5">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="text-lg font-black text-slate-950">Aktivitas Terakhir per Agent</h2>
            <p className="mt-1 text-xs font-semibold text-slate-500">Setiap agent hanya ditampilkan satu kali berdasarkan transaksi terbarunya.</p>
          </div>
          <form onSubmit={(event) => { event.preventDefault(); setSearch(draftSearch.trim()); }} className="flex w-full gap-2 lg:max-w-xl">
            <label className="flex min-w-0 flex-1 items-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 focus-within:border-emerald-400">
              <Search className="h-4 w-4 shrink-0 text-slate-400" />
              <input value={draftSearch} onChange={(event) => setDraftSearch(event.target.value)} placeholder="Cari agent, produk, tujuan, atau ref ID" className="min-w-0 flex-1 bg-transparent py-2 text-xs font-semibold text-slate-700 outline-none placeholder:text-slate-400" />
            </label>
            <button type="submit" className="rounded-xl bg-emerald-700 px-4 text-xs font-black text-white hover:bg-emerald-800">Cari</button>
          </form>
        </div>
        <div className="mt-4 flex flex-wrap gap-2">
          {[{ value: "", label: "Semua" }, { value: "success", label: "Berhasil" }, { value: "refunded", label: "Dana Dikembalikan" }].map((filter) => (
            <button key={filter.value} type="button" onClick={() => setStatus(filter.value)} className={`rounded-full border px-4 py-2 text-[11px] font-black transition ${status === filter.value ? "border-emerald-700 bg-emerald-700 text-white" : "border-slate-200 bg-white text-slate-600 hover:border-emerald-300 hover:text-emerald-700"}`}>
              {filter.label}
            </button>
          ))}
        </div>

        {error ? <div className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-bold text-rose-700">{error}</div> : null}
        <div className="mt-4 overflow-hidden rounded-2xl border border-emerald-100">
          <div className="hidden overflow-x-auto lg:block">
            <table className="min-w-[900px] w-full text-left text-xs">
              <thead className="bg-emerald-50 text-[10px] font-black uppercase tracking-[0.08em] text-emerald-800">
                <tr><th className="px-4 py-3">Waktu</th><th className="px-4 py-3">Agent</th><th className="px-4 py-3">Aktivitas</th><th className="px-4 py-3">Tujuan</th><th className="px-4 py-3">Nominal</th><th className="px-4 py-3">Status</th></tr>
              </thead>
              <tbody className="divide-y divide-emerald-50">
                {items.map((item) => <tr key={item.id} className="text-slate-600 hover:bg-emerald-50/40"><td className="whitespace-nowrap px-4 py-4 font-semibold">{dateLabel(item.dibuat_pada)}</td><td className="px-4 py-4"><p className="font-black text-slate-900">{item.member_nama || "Agent"}</p><p className="mt-1 text-[11px] text-slate-400">{item.member_email || "-"}</p></td><td className="px-4 py-4"><p className="font-black text-slate-900">{item.produk_nama || item.kode_produk || "-"}</p>{item.produk_nama && item.kode_produk ? <p className="mt-1 text-[10px] font-bold text-slate-400">{item.kode_produk}</p> : null}</td><td className="px-4 py-4 font-semibold">{item.tujuan || "-"}</td><td className="whitespace-nowrap px-4 py-4 font-black text-slate-900">{fmtID(item.biaya_aktual || item.biaya_perkiraan || 0)}</td><td className="px-4 py-4"><span className={`rounded-full px-3 py-1 text-[10px] font-black ${statusClass(item.status)}`}>{statusLabel(item.status)}</span></td></tr>)}
              </tbody>
            </table>
          </div>
          <div className="divide-y divide-emerald-50 lg:hidden">
            {items.map((item) => <article key={item.id} className="space-y-3 p-4 hover:bg-emerald-50/40"><div className="flex items-start justify-between gap-3"><div className="flex min-w-0 items-center gap-3"><span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-emerald-100 text-emerald-700"><ReceiptText className="h-5 w-5" /></span><div className="min-w-0"><p className="truncate text-sm font-black text-slate-950">{item.member_nama || "Agent"}</p><p className="truncate text-[11px] font-semibold text-slate-400">{item.member_email || "-"}</p></div></div><span className={`shrink-0 rounded-full px-2.5 py-1 text-[10px] font-black ${statusClass(item.status)}`}>{statusLabel(item.status)}</span></div><div className="grid grid-cols-2 gap-2 text-xs"><div className="rounded-xl bg-slate-50 p-3"><p className="text-[10px] font-bold uppercase text-slate-400">Aktivitas</p><p className="mt-1 font-black text-slate-800">{item.produk_nama || item.kode_produk || "-"}</p>{item.produk_nama && item.kode_produk ? <p className="mt-1 text-[10px] font-bold text-slate-400">{item.kode_produk}</p> : null}</div><div className="rounded-xl bg-slate-50 p-3"><p className="text-[10px] font-bold uppercase text-slate-400">Nominal</p><p className="mt-1 font-black text-slate-800">{fmtID(item.biaya_aktual || item.biaya_perkiraan || 0)}</p></div><div className="col-span-2 rounded-xl bg-slate-50 p-3"><p className="text-[10px] font-bold uppercase text-slate-400">Tujuan · Waktu</p><p className="mt-1 break-words font-semibold text-slate-700">{item.tujuan || "-"} · {dateLabel(item.dibuat_pada)}</p></div></div></article>)}
          </div>
          {!loading && items.length === 0 ? <div className="px-4 py-12 text-center text-sm font-semibold text-slate-500">Belum ada transaksi agent yang sesuai filter.</div> : null}
          {loading ? <div className="px-4 py-12 text-center text-sm font-semibold text-slate-500">Memuat transaksi agent...</div> : null}
        </div>
      </section>
    </div>
  );
}
