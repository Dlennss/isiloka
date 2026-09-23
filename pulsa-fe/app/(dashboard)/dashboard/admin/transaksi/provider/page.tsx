"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { Download, SlidersHorizontal } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

type Row = {
  id: number;
  dibuat_pada: string;
  provider: string;
  status_provider: string;
  transaksi_member_id: number;
  member_id?: number;
  member_nama?: string;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  trx_id_provider?: string;
  kode_respon?: string;
  no_referensi?: string;
  pesan?: string;
  http_status?: number;
  percobaan?: number;
  harga?: number;
  saldo_terakhir?: number;
};

type ProviderItem = {
  id: number;
  nama?: string;
  aktif?: boolean;
};

type MemberItem = {
  id: number;
  nama?: string;
  email?: string;
};

type ExportKind = "" | "csv" | "excel" | "pdf";
type ExportColumnKey = "no" | "waktu" | "provider" | "status_provider" | "member" | "member_trx_id" | "ref_id" | "perintah" | "kode_produk" | "tujuan" | "qty" | "http_status" | "kode_respon" | "harga" | "saldo_terakhir" | "pesan";
type ExportRow = Record<ExportColumnKey, string | number>;

const PAGE_SIZE = 10;
const EXPORT_BATCH_SIZE = 500;
const EXPORT_COLUMNS: Array<{ key: ExportColumnKey; label: string }> = [
  { key: "no", label: "No" },
  { key: "waktu", label: "Waktu" },
  { key: "provider", label: "Provider" },
  { key: "status_provider", label: "Status" },
  { key: "member", label: "Member" },
  { key: "member_trx_id", label: "MemberTrxID" },
  { key: "ref_id", label: "Ref ID" },
  { key: "perintah", label: "Perintah" },
  { key: "kode_produk", label: "Produk" },
  { key: "tujuan", label: "Tujuan" },
  { key: "qty", label: "Qty" },
  { key: "http_status", label: "HTTP" },
  { key: "kode_respon", label: "RC" },
  { key: "harga", label: "Harga" },
  { key: "saldo_terakhir", label: "Saldo" },
  { key: "pesan", label: "Pesan" },
];
const DEFAULT_EXPORT_COLUMNS: ExportColumnKey[] = ["no", "waktu", "provider", "status_provider", "member", "ref_id", "kode_produk", "tujuan", "qty", "kode_respon", "harga"];

function toLocalDateInputValue(date: Date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function downloadBlob(filename: string, blob: Blob) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

function statusClassName(status: string) {
  switch (status) {
    case "success":
      return "border-emerald-400/25 bg-emerald-500/15 text-emerald-200";
    case "pending":
      return "border-amber-400/25 bg-amber-500/15 text-amber-200";
    case "failed":
      return "border-rose-400/25 bg-rose-500/15 text-rose-200";
    default:
      return "border-white/15 bg-slate-900/50 text-slate-300";
  }
}

export default function ProviderTrxPage() {
  const pathname = usePathname();
  const now = useMemo(() => new Date(), []);
  const defaultFrom = useMemo(() => toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)), [now]);
  const defaultTo = useMemo(() => toLocalDateInputValue(now), [now]);
  const canExport = !pathname?.startsWith("/dashboard/operator") && !pathname?.startsWith("/dashboard/wallet");

  const [items, setItems] = useState<Row[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);

  const [provider, setProvider] = useState("");
  const [memberID, setMemberID] = useState("");
  const [status, setStatus] = useState("");
  const [kodeProduk, setKodeProduk] = useState("");
  const [refID, setRefID] = useState("");
  const [tujuan, setTujuan] = useState("");
  const [from, setFrom] = useState(defaultFrom);
  const [to, setTo] = useState(defaultTo);
  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [members, setMembers] = useState<MemberItem[]>([]);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [exporting, setExporting] = useState<ExportKind>("");
  const [selectedExportColumns, setSelectedExportColumns] = useState<ExportColumnKey[]>(DEFAULT_EXPORT_COLUMNS);
  const [resendingProviderTrxID, setResendingProviderTrxID] = useState<number | null>(null);
  const activeFilterCount = [provider, memberID, status, kodeProduk, refID, tujuan, from, to].filter((v) => v.trim()).length;

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (provider) qs.set("provider", provider);
      if (memberID) qs.set("member_id", memberID);
      if (status) qs.set("status", status);
      if (kodeProduk) qs.set("kode_produk", kodeProduk);
      if (refID) qs.set("ref_id", refID);
      if (tujuan) qs.set("tujuan", tujuan);
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/provider/trx?${qs.toString()}`, { headers: authHeader() });
      const j = await r.json().catch(() => ({}));
      const all: Row[] = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > PAGE_SIZE);
      setItems(all.slice(0, PAGE_SIZE));
    } finally {
      setLoading(false);
    }
  }

  async function fetchAllForExport() {
    const out: Row[] = [];
    let nextOffset = 0;

    for (;;) {
      const qs = new URLSearchParams();
      if (provider) qs.set("provider", provider);
      if (memberID) qs.set("member_id", memberID);
      if (status) qs.set("status", status);
      if (kodeProduk) qs.set("kode_produk", kodeProduk);
      if (refID) qs.set("ref_id", refID);
      if (tujuan) qs.set("tujuan", tujuan);
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);
      qs.set("limit", String(EXPORT_BATCH_SIZE));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/provider/trx?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal mengambil data export");

      const batch: Row[] = Array.isArray(j.items) ? j.items : [];
      out.push(...batch);
      if (batch.length < EXPORT_BATCH_SIZE) break;
      nextOffset += batch.length;
    }

    return out;
  }

  function memberLabel(x: Row) {
    if (x.member_nama && x.member_id) return `${x.member_nama} (${x.member_id})`;
    if (x.member_nama) return x.member_nama;
    if (x.member_id) return String(x.member_id);
    return "-";
  }

  function toExportRows(rows: Row[]): ExportRow[] {
    return rows.map((x, idx) => ({
      no: idx + 1,
      waktu: new Date(x.dibuat_pada).toLocaleString("id-ID"),
      provider: x.provider,
      status_provider: x.status_provider || "-",
      member: memberLabel(x),
      member_trx_id: x.transaksi_member_id,
      ref_id: x.ref_id,
      perintah: x.perintah,
      kode_produk: x.kode_produk,
      tujuan: x.tujuan,
      qty: x.qty,
      http_status: x.http_status ?? "-",
      kode_respon: x.kode_respon ?? "-",
      harga: x.harga ?? 0,
      saldo_terakhir: x.saldo_terakhir ?? 0,
      pesan: [x.pesan || "", x.no_referensi || ""].filter(Boolean).join(" - ") || "-",
    }));
  }

  async function runExport(kind: ExportKind) {
    if (!selectedExportColumns.length) {
      await alertWarning("Pilih minimal satu kolom untuk export.");
      return;
    }
    setExporting(kind);
    try {
      const rows = await fetchAllForExport();
      const normalized = toExportRows(rows);
      const pickedLabels = EXPORT_COLUMNS.filter((col) => selectedExportColumns.includes(col.key)).map((col) => col.label);
      const body = normalized.map((row) => selectedExportColumns.map((key) => row[key] ?? ""));

      if (kind === "csv") {
        const csv = [pickedLabels, ...body].map((row) => row.map((cell) => `"${String(cell ?? "").replaceAll('"', '""')}"`).join(",")).join("\n");
        downloadBlob(`transaksi-provider-${from || "all"}_${to || "all"}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8;" }));
      } else if (kind === "excel") {
        const XLSX = await import("xlsx");
        const ws = XLSX.utils.json_to_sheet(normalized.map((row) => Object.fromEntries(selectedExportColumns.map((key, idx) => [pickedLabels[idx], row[key]]))));
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "Transaksi Provider");
        const out = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        downloadBlob(`transaksi-provider-${from || "all"}_${to || "all"}.xlsx`, new Blob([out], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
      } else if (kind === "pdf") {
        const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
        const autoTable = autoTableModule.default;
        const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
        autoTable(doc, { head: [pickedLabels], body, styles: { fontSize: 8, cellPadding: 4 }, headStyles: { fillColor: [15, 23, 42] }, margin: { top: 36, left: 24, right: 24, bottom: 24 } });
        const out = doc.output("arraybuffer");
        downloadBlob(`transaksi-provider-${from || "all"}_${to || "all"}.pdf`, new Blob([out], { type: "application/pdf" }));
      }
      setExportOpen(false);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal export transaksi provider");
    } finally {
      setExporting("");
    }
  }

  useEffect(() => {
    load(offset);
     
  }, [offset]);

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/master/provider", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const items: ProviderItem[] = Array.isArray(j?.items) ? j.items : [];
      setProviders(items.filter((x) => x?.aktif !== false));
    } catch {
      setProviders([]);
    }
  }

  async function loadMembers() {
    try {
      const qs = new URLSearchParams();
      qs.set("limit", "500");
      qs.set("offset", "0");
      const r = await fetch(`/api/admin/members?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const items: MemberItem[] = Array.isArray(j?.items) ? j.items : [];
      setMembers(items);
    } catch {
      setMembers([]);
    }
  }

  useEffect(() => {
    void Promise.all([loadProviders(), loadMembers()]);
  }, []);

  async function resendPending(row: Row) {
    if ((row.status_provider || "").toLowerCase() !== "pending") return;
    if ((row.perintah || "").toUpperCase() !== "PAY") return;
    setResendingProviderTrxID(row.id);
    try {
      const r = await fetch("/api/admin/history/transaksi/resend", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({ provider_trx_id: row.id }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal mengirim ulang transaksi.");
        return;
      }
      await alertSuccess(`Kirim ulang dimulai. Provider: ${j?.item?.provider || "-"}.`);
      await load(offset);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal mengirim ulang transaksi.");
    } finally {
      setResendingProviderTrxID(null);
    }
  }

  const columns: DataTableColumn<Row>[] = useMemo(() => [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID") },
    { id: "provider", header: "Provider", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.provider },
    { id: "member", header: "Member", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => memberLabel(x) },
    { id: "memberTrxId", header: "MemberTrxID", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.transaksi_member_id },
    { id: "refId", header: "RefID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (x) => x.ref_id },
    { id: "perintah", header: "Perintah", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.perintah },
    { id: "produk", header: "Produk", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_produk },
    { id: "tujuan", header: "Tujuan", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.tujuan },
    { id: "qty", header: "Qty", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => fmtID(x.qty) },
    { id: "http", header: "HTTP", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.http_status ?? "-" },
    { id: "rc", header: "RC", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_respon ?? "-" },
    { id: "harga", header: "Harga", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => `Rp ${fmtID(x.harga ?? 0)}` },
    { id: "saldo", header: "Saldo", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => `Rp ${fmtID(x.saldo_terakhir ?? 0)}` },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => (
        <span className={`inline-flex min-w-20 justify-center rounded-full border px-2 py-1 text-xs font-semibold uppercase tracking-[0.08em] ${statusClassName(x.status_provider || "")}`}>
          {x.status_provider || "unknown"}
        </span>
      ),
    },
    { id: "pesan", header: "Pesan", tdClassName: "max-w-90 wrap-break-word text-slate-300", render: (x) => ([x.pesan || "", x.no_referensi || ""].filter(Boolean).join(" - ") || "-") },
    {
      id: "aksi",
      header: "Aksi",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) =>
        (x.status_provider || "").toLowerCase() === "pending" && (x.perintah || "").toUpperCase() === "PAY" ? (
          <Button
            type="button"
            variant="outline"
            className="h-8 border-cyan-400/30 bg-cyan-500/10 px-3 text-xs font-semibold text-cyan-100 hover:bg-cyan-500/20"
            disabled={resendingProviderTrxID === x.id}
            onClick={() => void resendPending(x)}
          >
            {resendingProviderTrxID === x.id ? "Memproses..." : "Kirim Ulang"}
          </Button>
        ) : (
          "-"
        ),
    },
  ], [resendingProviderTrxID]);

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Provider Transactions</div>
          <div className="text-sm text-muted-foreground">Log transaksi ke provider (multi provider).</div>
        </div>

        <div className="w-full md:hidden">
          <Button type="button" variant="outline" className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" onClick={() => setMobileFilterOpen((v) => !v)}>
            <span className="inline-flex items-center gap-2"><SlidersHorizontal className="h-4 w-4" />{mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}</span>
            {activeFilterCount > 0 ? <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span> : null}
          </Button>
        </div>

        <div className={`${mobileFilterOpen ? "block" : "hidden"} w-full rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-slate-800/45 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}>
          <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-slate-300"><SlidersHorizontal className="h-3.5 w-3.5 text-cyan-300" />Filter Transaksi</div>
          <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7">
            <div className="col-span-1 grid grid-cols-2 gap-2 sm:col-span-2 lg:col-span-2 xl:col-span-2">
              <select className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50" value={provider} onChange={(e) => setProvider(e.target.value)}>
                <option value="">provider</option>
                {providers.map((p) => <option key={p.id} value={(p.nama || "").trim().toLowerCase()}>{p.nama || `provider-${p.id}`}</option>)}
              </select>
              <select className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50" value={status} onChange={(e) => setStatus(e.target.value)}>
                <option value="">status</option><option value="success">success</option><option value="pending">pending</option><option value="failed">failed</option>
              </select>
            </div>
            <div className="col-span-1 grid grid-cols-2 gap-2 sm:col-span-2 lg:col-span-2 xl:col-span-2">
              <select className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50" value={memberID} onChange={(e) => setMemberID(e.target.value)}>
                <option value="">member</option>
                {members.map((m) => <option key={m.id} value={String(m.id)}>{m.nama || m.email || `member-${m.id}`}</option>)}
              </select>
              <Input className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40" value={kodeProduk} onChange={(e) => setKodeProduk(e.target.value)} placeholder="kode_produk" />
            </div>
            <div className="col-span-1 grid grid-cols-2 gap-2 sm:col-span-2 lg:col-span-2 xl:col-span-2">
              <Input className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40" value={refID} onChange={(e) => setRefID(e.target.value)} placeholder="ref_id" />
              <Input className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40" value={tujuan} onChange={(e) => setTujuan(e.target.value)} placeholder="tujuan" />
            </div>
            <div className="col-span-1 grid grid-cols-2 gap-2 sm:col-span-2 lg:col-span-2 xl:col-span-2">
              <div className="min-w-0"><label className="mb-1 block text-[11px] font-medium text-slate-300 md:hidden">Dari</label><Input className="h-10 w-full min-w-0 rounded-xl border-white/15 bg-slate-950/55 text-base text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40 md:text-sm scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" type="date" value={from} onChange={(e) => setFrom(e.target.value)} placeholder="from" /></div>
              <div className="min-w-0"><label className="mb-1 block text-[11px] font-medium text-slate-300 md:hidden">Sampai</label><Input className="h-10 w-full min-w-0 rounded-xl border-white/15 bg-slate-950/55 text-base text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40 md:text-sm scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" type="date" value={to} onChange={(e) => setTo(e.target.value)} placeholder="to" /></div>
            </div>
            <div className="col-span-1 grid grid-cols-3 gap-2 sm:col-span-2 lg:col-span-4 xl:col-span-7">
              <Button variant="primary" className="h-10 w-full min-w-0" onClick={() => { setOffset(0); void load(0); setMobileFilterOpen(false); }} disabled={loading}>{loading ? "Loading..." : "Filter"}</Button>
              <Button variant="outline" className="h-10 w-full min-w-0 border-white/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" onClick={() => { setProvider(""); setMemberID(""); setStatus(""); setKodeProduk(""); setRefID(""); setTujuan(""); setFrom(defaultFrom); setTo(defaultTo); setOffset(0); setMobileFilterOpen(false); void load(0); }} disabled={loading}>Reset</Button>
              {canExport ? (
                <Button variant="outline" className="h-10 w-full min-w-0 border-white/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" onClick={() => setExportOpen(true)} disabled={loading}><span className="inline-flex items-center gap-2"><Download className="h-4 w-4" />Export</span></Button>
              ) : null}
            </div>
          </div>
        </div>
      </div>

      <DataTable columns={columns} rows={items} rowKey={(x) => x.id} rowNumberStart={offset + 1} minWidthClassName="min-w-[2000px]" emptyText="Tidak ada data." pagination={{ page: Math.floor(offset / PAGE_SIZE) + 1, totalPages: hasNext ? Math.floor(offset / PAGE_SIZE) + 2 : Math.floor(offset / PAGE_SIZE) + 1, onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)), onNext: () => setOffset((v) => v + PAGE_SIZE), onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE), disablePrev: loading || offset === 0, disableNext: loading || !hasNext }} />

      {canExport ? (
        <AppModal open={exportOpen} onClose={() => !exporting && setExportOpen(false)} title="Export Transaksi Provider">
          <div className="space-y-4 text-sm text-slate-200">
            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => setSelectedExportColumns(EXPORT_COLUMNS.map((x) => x.key))}>Pilih Semua</Button>
              <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => setSelectedExportColumns(DEFAULT_EXPORT_COLUMNS)}>Default</Button>
              <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("csv")}>CSV</Button>
              <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("excel")}>Excel</Button>
              <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("pdf")}>PDF</Button>
            </div>
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              {EXPORT_COLUMNS.map((col) => {
                const checked = selectedExportColumns.includes(col.key);
                return (
                  <label key={col.key} className="flex items-center gap-3 rounded-xl border border-white/10 bg-slate-950/45 px-3 py-2">
                    <input type="checkbox" className="h-4 w-4 rounded border-white/20 bg-slate-900 text-cyan-400 focus:ring-cyan-400/40" checked={checked} onChange={() => setSelectedExportColumns((prev) => (prev.includes(col.key) ? prev.filter((x) => x !== col.key) : [...prev, col.key]))} />
                    <span>{col.label}</span>
                  </label>
                );
              })}
            </div>
          </div>
        </AppModal>
      ) : null}
    </div>
  );
}
