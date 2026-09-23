"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowDownLeft, ArrowLeft, CheckCircle2, Clock3, Download, History, Loader2, RefreshCcw, Search, ShieldCheck, UserCog, WalletCards, XCircle } from "lucide-react";
import type { MutasiRow, TrxRow } from "./_components/types";

type Tab = "overview" | "mutasi" | "transaksi";

type MemberInfo = {
  id: number;
  email: string;
  nama: string;
  role: string;
  aktif: boolean;
  saldo: number;
  dibuat_pada: string;
};

type MonthStats = {
  trx_success_count: number;
  trx_failed_count: number;
  trx_success_amount: number;
  trx_failed_amount: number;
  dep_approved_count: number;
  dep_rejected_count: number;
  dep_approved_amount: number;
  dep_rejected_amount: number;
};

type ActivityItem = {
  id: string;
  time: string;
  title: string;
  detail: string;
  amount: number;
  status: string;
  kind: "mutasi" | "transaksi";
};

const PAGE_SIZE = 12;

function authHeader(): Record<string, string> {
  const token = localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function money(value: number) {
  return `Rp ${new Intl.NumberFormat("id-ID").format(Number(value || 0))}`;
}

function dateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Jakarta" }).format(date);
}

function shortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(date);
}

function roleLabel(role: string) {
  const normalized = String(role || "").toLowerCase();
  if (normalized === "user") return "Pelanggan";
  if (normalized === "agent") return "Agent";
  if (normalized === "marketing" || normalized === "master") return "Marketing";
  if (normalized === "analis") return "Operator Kredit";
  if (normalized === "operator_trx") return "Operator Transaksi";
  if (normalized === "operator_wallet") return "Operator Wallet";
  if (normalized === "admin") return "Admin";
  return role || "Akun";
}

function escapeCSV(value: unknown) {
  const text = String(value ?? "");
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

function downloadBlob(filename: string, blob: Blob) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}

export default function AccountActivityClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const memberID = Number(searchParams.get("member_id") || 0);
  const initialTab = searchParams.get("tab") === "transaksi" ? "transaksi" : searchParams.get("tab") === "mutasi" ? "mutasi" : "overview";
  const [tab, setTab] = useState<Tab>(initialTab);
  const [member, setMember] = useState<MemberInfo | null>(null);
  const [stats, setStats] = useState<MonthStats[]>([]);
  const [mutations, setMutations] = useState<MutasiRow[]>([]);
  const [transactions, setTransactions] = useState<TrxRow[]>([]);
  const [mutationTotal, setMutationTotal] = useState(0);
  const [transactionTotal, setTransactionTotal] = useState(0);
  const [mutationOffset, setMutationOffset] = useState(0);
  const [transactionOffset, setTransactionOffset] = useState(0);
  const [query, setQuery] = useState("");
  const [direction, setDirection] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [downloadOpen, setDownloadOpen] = useState(false);
  const [exporting, setExporting] = useState(false);

  const load = useCallback(async () => {
    if (!memberID) return;
    setLoading(true);
    setError("");
    try {
      const mutationQuery = new URLSearchParams({ member_id: String(memberID), limit: String(PAGE_SIZE), offset: String(mutationOffset) });
      const transactionQuery = new URLSearchParams({ member_id: String(memberID), limit: String(PAGE_SIZE), offset: String(transactionOffset) });
      if (query.trim()) {
        mutationQuery.set("ref_id", query.trim());
        transactionQuery.set("q", query.trim());
      }
      if (direction) mutationQuery.set("arah", direction);
      if (from) { mutationQuery.set("from", from); transactionQuery.set("from", from); }
      if (to) { mutationQuery.set("to", to); transactionQuery.set("to", to); }

      const [memberResponse, statsResponse, mutationResponse, transactionResponse] = await Promise.all([
        fetch(`/api/admin/members/get?member_id=${memberID}`, { headers: authHeader(), cache: "no-store" }),
        fetch(`/api/admin/members/stats?member_id=${memberID}`, { headers: authHeader(), cache: "no-store" }),
        fetch(`/api/admin/history/mutasi?${mutationQuery}`, { headers: authHeader(), cache: "no-store" }),
        fetch(`/api/admin/history/transaksi?${transactionQuery}`, { headers: authHeader(), cache: "no-store" }),
      ]);
      const [memberData, statsData, mutationData, transactionData] = await Promise.all([
        memberResponse.json().catch(() => ({})), statsResponse.json().catch(() => ({})), mutationResponse.json().catch(() => ({})), transactionResponse.json().catch(() => ({})),
      ]);
      if (!memberResponse.ok || !memberData.item) throw new Error(memberData.error || "Akun tidak ditemukan");
      setMember(memberData.item);
      setStats(Array.isArray(statsData.items) ? statsData.items : []);
      setMutations(Array.isArray(mutationData.items) ? mutationData.items : []);
      setTransactions(Array.isArray(transactionData.items) ? transactionData.items : []);
      setMutationTotal(Number(mutationData.total || 0));
      setTransactionTotal(Number(transactionData.total || 0));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Aktivitas akun belum dapat dimuat");
    } finally {
      setLoading(false);
    }
  }, [direction, from, memberID, mutationOffset, query, to, transactionOffset]);

  useEffect(() => {
    if (!memberID) {
      router.replace("/dashboard/admin/master/members");
      return;
    }
    void load();
  }, [load, memberID, router]);

  const summary = useMemo(() => stats.reduce((result, item) => ({
    successCount: result.successCount + Number(item.trx_success_count || 0),
    failedCount: result.failedCount + Number(item.trx_failed_count || 0),
    depositAmount: result.depositAmount + Number(item.dep_approved_amount || 0),
  }), { successCount: 0, failedCount: 0, depositAmount: 0 }), [stats]);

  const recentActivity = useMemo<ActivityItem[]>(() => [
    ...mutations.map((item) => ({ id: `m-${item.id}`, time: item.dibuat_pada, title: item.alasan || "Mutasi saldo", detail: item.ref_id || item.catatan || "-", amount: Number(item.jumlah || 0), status: item.arah, kind: "mutasi" as const })),
    ...transactions.map((item) => ({ id: `t-${item.id}`, time: item.dibuat_pada, title: item.kode_produk || item.perintah || "Transaksi", detail: `${item.tujuan || "-"} · ${item.ref_id || "-"}`, amount: Number(item.biaya_aktual || item.biaya_perkiraan || 0), status: item.status, kind: "transaksi" as const })),
  ].sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime()).slice(0, 10), [mutations, transactions]);

  function applyFilter(event: FormEvent) {
    event.preventDefault();
    setMutationOffset(0);
    setTransactionOffset(0);
    void load();
  }

  function resetFilter() {
    setQuery(""); setDirection(""); setFrom(""); setTo(""); setMutationOffset(0); setTransactionOffset(0);
  }

  async function exportCurrent(kind: "csv" | "excel" | "pdf") {
    setExporting(true);
    setDownloadOpen(false);
    try {
      const rows = tab === "transaksi"
        ? transactions.map((item) => ({ waktu: dateTime(item.dibuat_pada), referensi: item.ref_id, aktivitas: item.kode_produk || item.perintah, tujuan: item.tujuan, nominal: Number(item.biaya_aktual || item.biaya_perkiraan || 0), status: item.status }))
        : mutations.map((item) => ({ waktu: dateTime(item.dibuat_pada), referensi: item.ref_id, aktivitas: item.alasan, catatan: item.catatan || "", nominal: Number(item.jumlah || 0), status: item.arah }));
      if (!rows.length) return;
      const filename = `aktivitas-${memberID}-${tab}`;
      if (kind === "csv") {
        const headers = Object.keys(rows[0]);
        const csv = [headers.join(","), ...rows.map((row) => headers.map((key) => escapeCSV(row[key as keyof typeof row])).join(","))].join("\n");
        downloadBlob(`${filename}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8" }));
      } else if (kind === "excel") {
        const XLSX = await import("xlsx");
        const book = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(book, XLSX.utils.json_to_sheet(rows), "Aktivitas");
        downloadBlob(`${filename}.xlsx`, new Blob([XLSX.write(book, { bookType: "xlsx", type: "array" })], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
      } else {
        const [{ jsPDF }, tableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
        const doc = new jsPDF({ orientation: "landscape" });
        const headers = Object.keys(rows[0]);
        doc.text(`Aktivitas Akun - ${member?.nama || memberID}`, 14, 14);
        tableModule.default(doc, { startY: 20, head: [headers], body: rows.map((row) => headers.map((key) => String(row[key as keyof typeof row] ?? ""))), styles: { fontSize: 8 } });
        downloadBlob(`${filename}.pdf`, new Blob([doc.output("arraybuffer")], { type: "application/pdf" }));
      }
    } finally {
      setExporting(false);
    }
  }

  const roleAction = member?.role === "agent"
    ? { href: `/dashboard/admin/kredit/pengajuan?member_id=${member.id}`, label: "Kredit & Tagihan" }
    : member?.role === "marketing" || member?.role === "master"
      ? { href: "/dashboard/admin/pemantauan-tim?role=marketing", label: "Aktivitas Marketing" }
      : member?.role === "analis"
        ? { href: "/dashboard/admin/pemantauan-tim?role=operator_credit", label: "Keputusan Kredit" }
        : null;

  return (
    <main className="-m-2 min-h-screen bg-[#eef7f2] p-3 text-slate-950 sm:p-5 lg:p-7">
      <div className="mx-auto w-full max-w-7xl space-y-5">
        <header className="overflow-hidden rounded-[28px] border border-emerald-100 bg-white shadow-[0_20px_50px_rgba(15,23,42,0.07)]">
          <div className="bg-[linear-gradient(135deg,#053a2f_0%,#087a50_62%,#50cf3e_100%)] px-5 py-6 text-white sm:px-7">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div><p className="text-[11px] font-black uppercase tracking-[0.18em] text-lime-100">Pengawasan Akun</p><h1 className="mt-2 text-3xl font-black">Aktivitas Akun</h1><p className="mt-2 text-sm font-semibold text-emerald-50/90">Riwayat transaksi dan pergerakan saldo pengguna Isiloka.</p></div>
              <div className="flex flex-wrap gap-2"><button type="button" onClick={() => router.push("/dashboard/admin/master/members")} className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-white/30 bg-white/10 px-3 text-sm font-black text-white hover:bg-white/20"><ArrowLeft className="h-4 w-4" /> Kembali</button><button type="button" onClick={() => void load()} disabled={loading} className="inline-flex min-h-10 items-center gap-2 rounded-xl bg-white px-3 text-sm font-black text-emerald-800 hover:bg-emerald-50"><RefreshCcw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} /> Muat Ulang</button></div>
            </div>
          </div>
          <div className="grid gap-4 p-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:p-6">
            <div className="flex min-w-0 items-center gap-3"><span className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-emerald-50 text-emerald-700"><UserCog className="h-6 w-6" /></span><div className="min-w-0"><p className="truncate text-lg font-black">{member?.nama || (loading ? "Memuat akun..." : "Akun tidak ditemukan")}</p><p className="truncate text-sm font-semibold text-slate-500">{member?.email || "-"}</p></div></div>
            {member ? <div className="flex flex-wrap gap-2 sm:justify-end"><span className="rounded-full bg-emerald-50 px-3 py-1.5 text-xs font-black text-emerald-700">{roleLabel(member.role)}</span><span className={`rounded-full px-3 py-1.5 text-xs font-black ${member.aktif ? "bg-lime-50 text-lime-700" : "bg-rose-50 text-rose-700"}`}>{member.aktif ? "Aktif" : "Nonaktif"}</span><span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-600">Bergabung {shortDate(member.dibuat_pada)}</span></div> : null}
          </div>
        </header>

        {error ? <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-bold text-rose-700">{error}</div> : null}

        <section className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          {[{ label: "Saldo Utama", value: money(member?.saldo || 0), icon: WalletCards, tone: "text-emerald-700 bg-emerald-50" },{ label: "Transaksi Berhasil", value: `${summary.successCount} transaksi`, icon: CheckCircle2, tone: "text-sky-700 bg-sky-50" },{ label: "Transaksi Gagal", value: `${summary.failedCount} transaksi`, icon: XCircle, tone: "text-rose-700 bg-rose-50" },{ label: "Top Up Diterima", value: money(summary.depositAmount), icon: ArrowDownLeft, tone: "text-amber-700 bg-amber-50" }].map((item) => { const Icon = item.icon; return <div key={item.label} className="flex min-h-28 items-start justify-between gap-3 rounded-[20px] border border-slate-200 bg-white p-4 shadow-[0_12px_28px_rgba(15,23,42,0.05)]"><div><p className="text-xs font-bold text-slate-500">{item.label}</p><p className="mt-2 text-lg font-black sm:text-xl">{loading ? "..." : item.value}</p></div><span className={`grid h-10 w-10 shrink-0 place-items-center rounded-xl ${item.tone}`}><Icon className="h-5 w-5" /></span></div>; })}
        </section>

        <section className="overflow-hidden rounded-[24px] border border-slate-200 bg-white shadow-[0_14px_34px_rgba(15,23,42,0.06)]">
          <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
            <nav className="flex flex-wrap gap-2" aria-label="Jenis aktivitas">{([{ value: "overview", label: "Ringkasan" }, { value: "mutasi", label: "Saldo & Mutasi" }, { value: "transaksi", label: "Transaksi" }] as Array<{ value: Tab; label: string }>).map((item) => <button key={item.value} type="button" onClick={() => setTab(item.value)} className={`min-h-10 rounded-full border px-4 text-sm font-black transition ${tab === item.value ? "border-emerald-700 bg-emerald-700 text-white" : "border-slate-200 bg-white text-slate-600 hover:border-emerald-300"}`}>{item.label}</button>)}{roleAction ? <Link href={roleAction.href} className="inline-flex min-h-10 items-center gap-1.5 rounded-full border border-lime-300 bg-lime-50 px-4 text-sm font-black text-lime-800"><ShieldCheck className="h-4 w-4" /> {roleAction.label}</Link> : null}</nav>
            {tab !== "overview" ? <div className="relative"><button type="button" onClick={() => setDownloadOpen((value) => !value)} disabled={exporting} className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-3 text-sm font-black text-emerald-800"><Download className="h-4 w-4" /> {exporting ? "Menyiapkan..." : "Unduh Laporan"}</button>{downloadOpen ? <div className="absolute right-0 top-12 z-20 w-40 rounded-xl border border-slate-200 bg-white p-1.5 shadow-xl">{(["csv", "excel", "pdf"] as const).map((kind) => <button key={kind} type="button" onClick={() => void exportCurrent(kind)} className="block w-full rounded-lg px-3 py-2 text-left text-sm font-bold uppercase text-slate-700 hover:bg-emerald-50">{kind}</button>)}</div> : null}</div> : null}
          </div>

          {tab !== "overview" ? <form onSubmit={applyFilter} className="grid gap-3 border-b border-slate-200 bg-[#f8fffb] p-4 sm:grid-cols-2 lg:grid-cols-[minmax(220px,1fr)_160px_170px_170px_auto]">
            <label className="relative"><span className="sr-only">Cari aktivitas</span><Search className="pointer-events-none absolute left-3 top-3.5 h-4 w-4 text-slate-400" /><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Cari referensi, produk, atau tujuan" className="h-11 w-full rounded-xl border border-slate-200 bg-white pl-10 pr-3 text-sm outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100" /></label>
            {tab === "mutasi" ? <select value={direction} onChange={(event) => setDirection(event.target.value)} className="h-11 rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none focus:border-emerald-500"><option value="">Semua arah</option><option value="credit">Dana masuk</option><option value="debit">Dana keluar</option></select> : <div className="hidden lg:block" />}
            <input type="date" value={from} onChange={(event) => setFrom(event.target.value)} className="h-11 rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none focus:border-emerald-500" aria-label="Tanggal mulai" />
            <input type="date" value={to} onChange={(event) => setTo(event.target.value)} className="h-11 rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none focus:border-emerald-500" aria-label="Tanggal akhir" />
            <div className="flex gap-2"><button type="submit" className="min-h-11 flex-1 rounded-xl bg-emerald-700 px-4 text-sm font-black text-white hover:bg-emerald-800">Terapkan</button><button type="button" onClick={resetFilter} className="min-h-11 rounded-xl border border-slate-200 bg-white px-3 text-sm font-black text-slate-600 hover:bg-slate-50">Reset</button></div>
          </form> : null}

          <div className="overflow-x-auto">
            {loading ? <div className="grid min-h-64 place-items-center text-sm font-bold text-slate-500"><Loader2 className="mb-2 h-6 w-6 animate-spin text-emerald-600" /> Memuat aktivitas...</div> : tab === "overview" ? <ActivityTimeline rows={recentActivity} /> : tab === "mutasi" ? <MutationTable rows={mutations} /> : <TransactionTable rows={transactions} />}
          </div>

          {tab !== "overview" ? <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3 text-xs font-bold text-slate-500"><span>{tab === "mutasi" ? mutationTotal : transactionTotal} data</span><div className="flex gap-2"><button type="button" onClick={() => tab === "mutasi" ? setMutationOffset((value) => Math.max(0, value - PAGE_SIZE)) : setTransactionOffset((value) => Math.max(0, value - PAGE_SIZE))} disabled={(tab === "mutasi" ? mutationOffset : transactionOffset) === 0} className="rounded-lg border border-slate-200 px-3 py-2 disabled:opacity-40">Sebelumnya</button><button type="button" onClick={() => tab === "mutasi" ? setMutationOffset((value) => value + PAGE_SIZE) : setTransactionOffset((value) => value + PAGE_SIZE)} disabled={(tab === "mutasi" ? mutationOffset + mutations.length >= mutationTotal : transactionOffset + transactions.length >= transactionTotal)} className="rounded-lg border border-slate-200 px-3 py-2 disabled:opacity-40">Berikutnya</button></div></div> : null}
        </section>
      </div>
    </main>
  );
}

function ActivityTimeline({ rows }: { rows: ActivityItem[] }) {
  if (!rows.length) return <EmptyState />;
  return <div className="divide-y divide-slate-100">{rows.map((item) => <article key={item.id} className="grid gap-3 px-4 py-4 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:items-center sm:px-5"><span className={`grid h-10 w-10 place-items-center rounded-xl ${item.kind === "mutasi" ? "bg-emerald-50 text-emerald-700" : "bg-sky-50 text-sky-700"}`}>{item.kind === "mutasi" ? <WalletCards className="h-5 w-5" /> : <History className="h-5 w-5" />}</span><div className="min-w-0"><p className="font-black text-slate-900">{item.title}</p><p className="mt-1 truncate text-xs font-semibold text-slate-500">{item.detail}</p><p className="mt-1 flex items-center gap-1 text-xs text-slate-400"><Clock3 className="h-3.5 w-3.5" /> {dateTime(item.time)}</p></div><div className="sm:text-right"><p className="font-black text-slate-900">{money(item.amount)}</p><Status value={item.status} /></div></article>)}</div>;
}

function MutationTable({ rows }: { rows: MutasiRow[] }) {
  if (!rows.length) return <EmptyState />;
  return <table className="w-full min-w-[850px] text-left text-sm"><thead className="bg-emerald-50 text-[11px] font-black uppercase text-emerald-800"><tr><th className="px-4 py-3">Waktu</th><th className="px-4 py-3">Aktivitas</th><th className="px-4 py-3">Referensi</th><th className="px-4 py-3">Nominal</th><th className="px-4 py-3">Arah</th><th className="px-4 py-3">Saldo</th></tr></thead><tbody className="divide-y divide-slate-100">{rows.map((item) => <tr key={item.id} className="hover:bg-emerald-50/40"><td className="whitespace-nowrap px-4 py-4 text-slate-600">{dateTime(item.dibuat_pada)}</td><td className="px-4 py-4"><p className="font-bold">{item.alasan || "Mutasi saldo"}</p><p className="mt-1 text-xs text-slate-500">{item.catatan || "-"}</p></td><td className="whitespace-nowrap px-4 py-4 font-mono text-xs text-slate-500">{item.ref_id || "-"}</td><td className="whitespace-nowrap px-4 py-4 font-black">{money(item.jumlah)}</td><td className="px-4 py-4"><Status value={item.arah} /></td><td className="whitespace-nowrap px-4 py-4 text-slate-600">{money(Number(item.saldo_sebelum || 0))} → {money(Number(item.saldo_sesudah || 0))}</td></tr>)}</tbody></table>;
}

function TransactionTable({ rows }: { rows: TrxRow[] }) {
  if (!rows.length) return <EmptyState />;
  return <table className="w-full min-w-[900px] text-left text-sm"><thead className="bg-emerald-50 text-[11px] font-black uppercase text-emerald-800"><tr><th className="px-4 py-3">Waktu</th><th className="px-4 py-3">Produk</th><th className="px-4 py-3">Referensi</th><th className="px-4 py-3">Tujuan</th><th className="px-4 py-3">Nominal</th><th className="px-4 py-3">Status</th></tr></thead><tbody className="divide-y divide-slate-100">{rows.map((item) => <tr key={item.id} className="hover:bg-emerald-50/40"><td className="whitespace-nowrap px-4 py-4 text-slate-600">{dateTime(item.dibuat_pada)}</td><td className="px-4 py-4"><p className="font-bold">{item.kode_produk || item.perintah || "Transaksi"}</p><p className="mt-1 text-xs text-slate-500">{item.keterangan || "-"}</p></td><td className="whitespace-nowrap px-4 py-4 font-mono text-xs text-slate-500">{item.ref_id || "-"}</td><td className="whitespace-nowrap px-4 py-4">{item.tujuan || "-"}</td><td className="whitespace-nowrap px-4 py-4 font-black">{money(Number(item.biaya_aktual || item.biaya_perkiraan || 0))}</td><td className="px-4 py-4"><Status value={item.status} /></td></tr>)}</tbody></table>;
}

function Status({ value }: { value: string }) {
  const normalized = String(value || "").toLowerCase();
  const positive = ["success", "sukses", "credit", "berhasil"].includes(normalized);
  const negative = ["failed", "gagal", "error", "debit"].includes(normalized);
  return <span className={`mt-1 inline-flex rounded-full px-2.5 py-1 text-[11px] font-black ${positive ? "bg-emerald-50 text-emerald-700" : negative ? "bg-rose-50 text-rose-700" : "bg-amber-50 text-amber-700"}`}>{normalized === "credit" ? "Dana masuk" : normalized === "debit" ? "Dana keluar" : value || "Diproses"}</span>;
}

function EmptyState() {
  return <div className="grid min-h-64 place-items-center px-5 py-12 text-center"><div><span className="mx-auto grid h-12 w-12 place-items-center rounded-2xl bg-emerald-50 text-emerald-700"><History className="h-6 w-6" /></span><p className="mt-3 font-black text-slate-800">Belum ada aktivitas</p><p className="mt-1 text-sm font-semibold text-slate-500">Aktivitas akun akan muncul setelah transaksi atau perubahan saldo terjadi.</p></div></div>;
}
