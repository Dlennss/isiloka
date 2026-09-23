"use client";

import { type FormEvent, useCallback, useEffect, useState } from "react";
import { Download, FileText, RotateCcw, Search, SlidersHorizontal } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type TrxRow = {
  id: number;
  member_id: number;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  qty_provider?: number;
  charge_receiver_applied?: boolean;
  fee_member_rp?: number;
  status: string;
  keterangan?: string | null;
  biaya_perkiraan: number;
  biaya_aktual: number;
  dibuat_pada: string;
  diperbarui_pada: string;
};

type ApiResp = {
  ok: boolean;
  rows?: TrxRow[];
  total?: number;
  total_pages?: number;
  error?: string;
};

type ExportKind = "" | "csv" | "excel" | "pdf";
type ExportColumnKey =
  | "id"
  | "waktu"
  | "perintah"
  | "produk"
  | "tujuan"
  | "qty"
  | "qty_provider"
  | "beban_transaksi"
  | "fee"
  | "status"
  | "biaya_perkiraan"
  | "biaya_aktual"
  | "keterangan"
  | "ref_id";

const EXPORT_PAGE_SIZE = 200;
const PAGE_SIZE = 10;

const EXPORT_COLUMNS: Array<{ key: ExportColumnKey; label: string }> = [
  { key: "id", label: "ID" },
  { key: "waktu", label: "Waktu" },
  { key: "perintah", label: "Perintah" },
  { key: "produk", label: "Produk" },
  { key: "tujuan", label: "Tujuan" },
  { key: "qty", label: "Qty Input" },
  { key: "qty_provider", label: "Qty Kirim" },
  { key: "beban_transaksi", label: "Beban Transaksi" },
  { key: "fee", label: "Fee" },
  { key: "status", label: "Status" },
  { key: "biaya_perkiraan", label: "Biaya Est" },
  { key: "biaya_aktual", label: "Biaya Aktual" },
  { key: "keterangan", label: "Keterangan" },
  { key: "ref_id", label: "Ref ID" },
];

const DEFAULT_EXPORT_COLUMNS: ExportColumnKey[] = [
  "id",
  "waktu",
  "perintah",
  "produk",
  "tujuan",
  "qty",
  "qty_provider",
  "beban_transaksi",
  "fee",
  "status",
  "ref_id",
];

async function apiFetch(path: string, init?: RequestInit) {
  const t = localStorage.getItem("auth_token") || "";
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (t) headers.Authorization = `Bearer ${t}`;
  const res = await fetch(path, {
    ...init,
    headers: { ...headers, ...(init?.headers || {}) },
    cache: "no-store",
  });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

function buildQS(params: Record<string, string>) {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    const vv = (v || "").trim();
    if (vv) sp.set(k, vv);
  }
  return sp.toString();
}

function formatDateYYYYMMDD(d: Date) {
  return d.toISOString().slice(0, 10);
}

function statusLabel(status: string): string {
  const s = String(status || "").toLowerCase();
  if (s === "success") return "Berhasil";
  if (s === "failed") return "Gagal";
  if (s === "pending") return "Pending";
  return status || "-";
}

function monthToRange(ym: string): { from: string; to: string } | null {
  const m = (ym || "").match(/^(\d{4})-(\d{2})$/);
  if (!m) return null;

  const year = Number(m[1]);
  const month = Number(m[2]);
  if (!Number.isFinite(year) || !Number.isFinite(month) || month < 1 || month > 12) return null;

  const from = `${m[1]}-${m[2]}-01`;
  const nextMonth = new Date(Date.UTC(year, month, 1));
  const lastDay = new Date(nextMonth.getTime() - 24 * 60 * 60 * 1000);
  const to = formatDateYYYYMMDD(lastDay);
  return { from, to };
}

function formatTanggalID(ymd: string): string {
  if (!ymd) return "-";
  const d = new Date(`${ymd}T00:00:00`);
  if (Number.isNaN(d.getTime())) return ymd;
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "long", year: "numeric" });
}

function formatBulanID(ym: string): string {
  const m = (ym || "").match(/^(\d{4})-(\d{2})$/);
  if (!m) return "-";
  const d = new Date(Number(m[1]), Number(m[2]) - 1, 1);
  if (Number.isNaN(d.getTime())) return ym;
  return d.toLocaleDateString("id-ID", { month: "long", year: "numeric" });
}

function escapeCSV(v: unknown): string {
  const s = String(v ?? "");
  if (s.includes('"') || s.includes(",") || s.includes("\n")) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
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

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export default function HistoryTransaksiPage() {
  const [rows, setRows] = useState<TrxRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");
  const [exporting, setExporting] = useState<ExportKind>("");

  // filters
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [from, setFrom] = useState(""); // YYYY-MM-DD
  const [to, setTo] = useState(""); // YYYY-MM-DD
  const [exportMonth, setExportMonth] = useState("");
  const [appliedQ, setAppliedQ] = useState("");
  const [appliedStatus, setAppliedStatus] = useState("");
  const [appliedFrom, setAppliedFrom] = useState("");
  const [appliedTo, setAppliedTo] = useState("");
  const [applyTick, setApplyTick] = useState(0);
  const [exportOpen, setExportOpen] = useState(false);
  const [selectedExportColumns, setSelectedExportColumns] = useState<ExportColumnKey[]>(DEFAULT_EXPORT_COLUMNS);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const [selectedReceipt, setSelectedReceipt] = useState<TrxRow | null>(null);
  const [selectedDetail, setSelectedDetail] = useState<TrxRow | null>(null);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  const load = useCallback(async () => {
    setErr("");
    setLoading(true);
    try {
      const qs = buildQS({
        limit: String(PAGE_SIZE),
        offset: String(offset),
        q: appliedQ,
        status: appliedStatus,
        from: appliedFrom,
        to: appliedTo,
      });
      const r = await apiFetch(`/api/me/history/transaksi?${qs}`);
      const j: ApiResp = r.data || ({} as ApiResp);
      if (!r.ok || !j.ok) {
        setErr(j.error || "Gagal mengambil data");
        setRows([]);
        setTotal(0);
        setTotalPages(1);
        return;
      }
      setRows(j.rows || []);
      setTotal(Number(j.total || 0));
      setTotalPages(Math.max(1, Number(j.total_pages || 1)));
    } finally {
      setLoading(false);
    }
  }, [appliedQ, appliedStatus, appliedFrom, appliedTo, offset]);

  useEffect(() => {
    load();
  }, [load, applyTick]);

  async function fetchMonthRows(ym: string): Promise<TrxRow[] | null> {
    const range = monthToRange(ym);
    if (!range) {
      setErr("Pilih bulan export terlebih dahulu (format YYYY-MM).");
      return null;
    }

    const out: TrxRow[] = [];
    let offset = 0;

    while (true) {
      const qs = buildQS({
        from: range.from,
        to: range.to,
        status: appliedStatus,
        limit: String(EXPORT_PAGE_SIZE),
        offset: String(offset),
      });
      const r = await apiFetch(`/api/me/history/transaksi?${qs}`);
      const j: ApiResp = r.data || ({} as ApiResp);
      if (!r.ok || !j.ok) {
        setErr(j.error || "Gagal mengambil data export transaksi");
        return null;
      }

      const chunk = j.rows || [];
      out.push(...chunk);

      if (chunk.length < EXPORT_PAGE_SIZE) break;
      offset += EXPORT_PAGE_SIZE;
    }

    if (!out.length) {
      setErr("Tidak ada data transaksi pada bulan yang dipilih.");
      return null;
    }

    return out;
  }

  function toExportRows(items: TrxRow[]) {
    return items.map((x, idx) => ({
      no: idx + 1,
      id: x.id,
      waktu: new Date(x.dibuat_pada).toLocaleString("id-ID"),
      perintah: x.perintah,
      produk: x.kode_produk,
      tujuan: x.tujuan,
      qty: x.qty,
      qty_provider: typeof x.qty_provider === "number" ? x.qty_provider : x.qty,
      beban_transaksi: x.charge_receiver_applied ? "Penerima" : "Member",
      fee: x.fee_member_rp,
      status: x.status,
      biaya_perkiraan: x.biaya_perkiraan,
      biaya_aktual: x.biaya_aktual,
      keterangan: x.keterangan || "",
      ref_id: x.ref_id || "",
    }));
  }

  async function runExport(kind: ExportKind) {
    setErr("");
    if (!selectedExportColumns.length) {
      setErr("Pilih minimal satu kolom untuk export.");
      return;
    }
    setExporting(kind);
    try {
      const items = await fetchMonthRows(exportMonth);
      if (!items) return;

      const normalized = toExportRows(items);
      const pickedLabels = ["No", ...EXPORT_COLUMNS.filter((col) => selectedExportColumns.includes(col.key)).map((col) => col.label)];
      const body = normalized.map((row) =>
        [
          row.no,
          ...selectedExportColumns.map((key) => {
            const value = row[key];
            return value == null ? "" : value;
          }),
        ]
      );

      if (kind === "csv") {
        const csv = [pickedLabels, ...body].map((row) => row.map(escapeCSV).join(",")).join("\n");
        downloadBlob(`transaksi-${exportMonth}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8;" }));
      } else if (kind === "excel") {
        const XLSX = await import("xlsx");
        const sheetRows = normalized.map((row) => ({
          No: row.no,
          ...Object.fromEntries(selectedExportColumns.map((key) => [EXPORT_COLUMNS.find((col) => col.key === key)?.label || String(key), row[key]])),
        }));
        const ws = XLSX.utils.json_to_sheet(sheetRows);
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "Transaksi");
        const out = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        downloadBlob(`transaksi-${exportMonth}.xlsx`, new Blob([out], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
      } else if (kind === "pdf") {
        const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
        const autoTable = autoTableModule.default;

        const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
        doc.setFontSize(12);
        doc.text(`Laporan Transaksi ${exportMonth}`, 40, 36);
        autoTable(doc, {
          startY: 48,
          styles: { fontSize: 8, cellPadding: 4 },
          headStyles: { fillColor: [45, 45, 45] },
          head: [pickedLabels],
          body,
        });

        const out = doc.output("arraybuffer");
        downloadBlob(`transaksi-${exportMonth}.pdf`, new Blob([out], { type: "application/pdf" }));
      }

      setExportOpen(false);
    } finally {
      setExporting("");
    }
  }

  function toggleExportColumn(key: ExportColumnKey) {
    setSelectedExportColumns((prev) => (prev.includes(key) ? prev.filter((x) => x !== key) : [...prev, key]));
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setOffset(0);
    setAppliedQ(q.trim());
    setAppliedStatus(status.trim());
    setAppliedFrom(from);
    setAppliedTo(to);
    setApplyTick((v) => v + 1);
    setMobileFilterOpen(false);
  }

  function onReset() {
    setOffset(0);
    setQ("");
    setStatus("");
    setFrom("");
    setTo("");
    setAppliedQ("");
    setAppliedStatus("");
    setAppliedFrom("");
    setAppliedTo("");
    setApplyTick((v) => v + 1);
    setMobileFilterOpen(false);
  }

  function printReceipt(row: TrxRow) {
    const biayaAktual = Number(row.biaya_aktual || row.biaya_perkiraan || 0);
    const qty = Number(row.qty || 0);
    const printedAt = new Date(row.dibuat_pada).toLocaleString("id-ID");
    const html = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Struk ${escapeHtml(row.ref_id)}</title>
    <style>
      @page { size: 58mm auto; margin: 2mm; }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        background: #fff;
        color: #000;
        font-family: "Courier New", Courier, monospace;
        font-size: 12px;
        line-height: 1.45;
      }
      .receipt {
        width: 54mm;
        max-width: 54mm;
        margin: 0;
        padding: 2mm 0 4mm;
      }
      .title { font-size: 16px; font-weight: 700; text-align: center; }
      .muted { color: #222; font-size: 11px; text-align: center; }
      .line { border-top: 1px dashed #000; margin: 8px 0; }
      .row {
        display: flex;
        gap: 8px;
        justify-content: space-between;
        align-items: flex-start;
        margin: 3px 0;
      }
      .label { width: 20mm; }
      .value { flex: 1; text-align: right; word-break: break-word; }
      .value.left { text-align: left; }
      .footer { margin-top: 10px; text-align: center; font-size: 11px; }
    </style>
  </head>
  <body>
    <div class="receipt">
      <div class="title">Bukti Transaksi</div>
      <div class="muted">Transaksi sukses</div>
      <div class="line"></div>
      <div class="row"><div class="label">Waktu</div><div class="value">${escapeHtml(printedAt)}</div></div>
      <div class="row"><div class="label">Ref ID</div><div class="value">${escapeHtml(row.ref_id)}</div></div>
      <div class="row"><div class="label">Perintah</div><div class="value">${escapeHtml(row.perintah)}</div></div>
      <div class="row"><div class="label">Produk</div><div class="value">${escapeHtml(row.kode_produk)}</div></div>
      <div class="row"><div class="label">Tujuan</div><div class="value">${escapeHtml(row.tujuan)}</div></div>
      <div class="row"><div class="label">Qty</div><div class="value">Rp ${qty.toLocaleString("id-ID")}</div></div>
      <div class="row"><div class="label">Biaya Aktual</div><div class="value">Rp ${biayaAktual.toLocaleString("id-ID")}</div></div>
      <div class="row"><div class="label">Status</div><div class="value">${escapeHtml(String(row.status).toUpperCase())}</div></div>
      <div class="line"></div>
      <div class="row"><div class="label">Keterangan / SN</div></div>
      <div class="row"><div class="value left">${escapeHtml(row.keterangan || "-")}</div></div>
      <div class="line"></div>
      <div class="footer">Terima kasih</div>
    </div>
    <script>
      window.onload = function() {
        window.print();
        setTimeout(function() { window.close(); }, 250);
      };
    </script>
  </body>
</html>`;

    const win = window.open("", "_blank", "width=420,height=720");
    if (!win) return;
    win.document.open();
    win.document.write(html);
    win.document.close();
  }

  const columns: DataTableColumn<TrxRow>[] = [
    {
      id: "waktu",
      header: "Waktu",
      thClassName: "whitespace-nowrap",
      render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID"),
    },
    { id: "status", header: "Status", thClassName: "whitespace-nowrap", render: (x) => statusLabel(x.status) },
    { id: "tujuan", header: "Tujuan", thClassName: "whitespace-nowrap", render: (x) => x.tujuan },
    { id: "produk", header: "Produk", thClassName: "whitespace-nowrap", render: (x) => x.kode_produk },
    { id: "ref_id", header: "Ref ID", thClassName: "whitespace-nowrap", render: (x) => x.ref_id || "-" },
    {
      id: "aksi",
      header: "Aksi",
      thClassName: "whitespace-nowrap",
      render: (x) => (
        <div className="flex gap-2">
          <Button variant="outline" className="h-8 px-3" onClick={() => setSelectedDetail(x)}>
            Detail
          </Button>
          {String(x.status).toLowerCase() === "success" ? (
            <Button variant="outline" className="h-8 px-3" onClick={() => setSelectedReceipt(x)}>
              <FileText className="h-4 w-4" />
              Cetak
            </Button>
          ) : null}
        </div>
      ),
    },
  ];
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;

  return (
    <div className="min-w-0 space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-xs text-white/60">History</div>
          <h1 className="text-lg font-semibold">Transaksi</h1>
          <div className="text-xs text-white/50">Total {total} data</div>
        </div>
        <button
          onClick={load}
          className="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/80 hover:bg-white/10"
          disabled={loading}
        >
          {loading ? "Memuat..." : "Refresh"}
        </button>
      </div>

      <div className="md:hidden">
        <button
          type="button"
          onClick={() => setMobileFilterOpen((v) => !v)}
          className="flex h-10 w-full items-center justify-between rounded-xl border border-white/12 bg-linear-to-r from-slate-900/85 via-slate-900/75 to-cyan-950/25 px-3 text-sm text-white/90 shadow-[0_16px_36px_-28px_rgba(6,182,212,0.8)]"
        >
          <span className="inline-flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
          </span>
        </button>
      </div>

      <form
        onSubmit={onSubmit}
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-cyan-950/20 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="grid grid-cols-1 gap-3 md:grid-cols-12">
          <div className="md:col-span-3">
            <label className="mb-1 block text-xs text-white/60">Pencarian</label>
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Cari ref_id / tujuan (nomor) / kode_produk / perintah / status / keterangan"
              className="w-full rounded-xl border border-white/15 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none placeholder:text-white/40 focus:border-cyan-400/40"
            />
          </div>

          <div className="md:col-span-3">
            <label className="mb-1 block text-xs text-white/60">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full rounded-xl border border-white/15 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-cyan-400/40"
            >
              <option value="">Semua status</option>
              <option value="pending">Pending</option>
              <option value="success">Berhasil</option>
              <option value="failed">Gagal</option>
            </select>
          </div>

          <div className="md:col-span-2">
            <label className="mb-1 block text-xs text-white/60">Dari tanggal</label>
            <input
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className="w-full rounded-xl border border-white/15 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-cyan-400/40 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
          </div>

          <div className="md:col-span-2">
            <label className="mb-1 block text-xs text-white/60">Sampai tanggal</label>
            <input
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className="w-full rounded-xl border border-white/15 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-cyan-400/40 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            {(from || to) && (
              <div className="mt-1 text-xs text-white/50">
                Periode aktif: {formatTanggalID(from)} - {formatTanggalID(to)}
              </div>
            )}
          </div>

          <div className="md:col-span-2">
            <label className="mb-1 block text-xs text-white/60">Download Bulanan</label>
            <input
              type="month"
              value={exportMonth}
              onChange={(e) => setExportMonth(e.target.value)}
              className="w-full rounded-xl border border-white/15 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-cyan-400/40 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            {exportMonth && <div className="mt-1 text-xs text-white/50">Bulan terpilih: {formatBulanID(exportMonth)}</div>}
          </div>

          <div className="md:col-span-12 flex flex-wrap items-end gap-2">
            <button
              type="submit"
              className="inline-flex items-center gap-2 rounded-xl border border-cyan-400/25 bg-cyan-500/15 px-3 py-2 text-sm font-medium text-cyan-100 hover:bg-cyan-500/20"
              disabled={loading}
            >
              <Search className="h-4 w-4" />
              Terapkan
            </button>
            <button
              type="button"
              onClick={onReset}
              className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-transparent px-3 py-2 text-sm text-white/70 hover:bg-white/5"
              disabled={loading}
            >
              <RotateCcw className="h-4 w-4" />
              Reset
            </button>
            <button
              type="button"
              onClick={() => setExportOpen(true)}
              disabled={loading || !!exporting}
              className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <Download className="h-4 w-4" />
              Export
            </button>

          </div>
        </div>
      </form>

      {err ? (
        <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-200">{err}</div>
      ) : null}

      {rows.length > 0 ? (
        <DataTable<TrxRow>
          columns={columns}
          rows={rows}
          rowKey={(x) => x.id}
          emptyText="Belum ada data"
          minWidthClassName="min-w-300"
          rowNumberStart={offset + 1}
          wrapperClassName="w-full overflow-x-auto overflow-y-hidden rounded-xl border border-white/10 bg-slate-950/35"
          pagination={{
            page: currentPage,
            totalPages,
            onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
            onNext: () => setOffset((v) => v + PAGE_SIZE),
            onPageChange: (page) => setOffset((page - 1) * PAGE_SIZE),
            disablePrev: loading || currentPage <= 1,
            disableNext: loading || currentPage >= totalPages,
          }}
        />
      ) : null}

      <AppModal
        open={exportOpen}
        onClose={() => {
          if (exporting) return;
          setExportOpen(false);
        }}
        title="Export Transaksi"
        subtitle="Pilih kolom yang ingin dimasukkan ke file export."
        maxWidthClassName="max-w-3xl"
        footer={
          <div className="flex flex-wrap items-center justify-end gap-2">
            <button
              type="button"
              onClick={() => setSelectedExportColumns(EXPORT_COLUMNS.map((x) => x.key))}
              className="rounded-lg border border-white/10 bg-transparent px-3 py-2 text-sm text-white/70 hover:bg-white/5"
              disabled={!!exporting}
            >
              Pilih Semua
            </button>
            <button
              type="button"
              onClick={() => setSelectedExportColumns(DEFAULT_EXPORT_COLUMNS)}
              className="rounded-lg border border-white/10 bg-transparent px-3 py-2 text-sm text-white/70 hover:bg-white/5"
              disabled={!!exporting}
            >
              Default
            </button>
            <button
              type="button"
              onClick={() => void runExport("csv")}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!!exporting}
            >
              {exporting === "csv" ? "Menyiapkan CSV..." : "CSV"}
            </button>
            <button
              type="button"
              onClick={() => void runExport("excel")}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!!exporting}
            >
              {exporting === "excel" ? "Menyiapkan Excel..." : "Excel"}
            </button>
            <button
              type="button"
              onClick={() => void runExport("pdf")}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!!exporting}
            >
              {exporting === "pdf" ? "Menyiapkan PDF..." : "PDF"}
            </button>
          </div>
        }
      >
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {EXPORT_COLUMNS.map((col) => {
            const checked = selectedExportColumns.includes(col.key);
            return (
              <label
                key={col.key}
                className={`flex cursor-pointer items-center gap-3 rounded-xl border px-3 py-2 text-sm transition ${
                  checked
                    ? "border-cyan-400/40 bg-cyan-500/10 text-cyan-100"
                    : "border-white/10 bg-slate-950/35 text-slate-200 hover:bg-slate-900/55"
                }`}
              >
                <input
                  type="checkbox"
                  className="h-4 w-4 accent-cyan-500"
                  checked={checked}
                  disabled={!!exporting}
                  onChange={() => toggleExportColumn(col.key)}
                />
                <span>{col.label}</span>
              </label>
            );
          })}
        </div>
      </AppModal>

      <AppModal
        open={!!selectedReceipt}
        onClose={() => setSelectedReceipt(null)}
        title="Struk Transaksi"
        subtitle="Format ringkas yang siap dicetak ke kertas thermal."
      >
        {selectedReceipt ? (
          <div className="space-y-4 text-sm text-slate-200">
            <div className="mx-auto w-full max-w-[320px] rounded-2xl border border-dashed border-white/15 bg-white px-4 py-5 font-mono text-black shadow-[0_20px_50px_-25px_rgba(0,0,0,0.8)]">
              <div className="text-center">
                <div className="text-lg font-bold tracking-wide"> RESI TRANSAKSI</div>
                <div className="text-[11px] uppercase tracking-[0.2em] text-slate-500">Struk Transaksi Sukses</div>
              </div>

              <div className="my-3 border-t border-dashed border-slate-400" />

              <div className="space-y-1.5 text-[12px] leading-5">
                <div className="flex items-start justify-between gap-3">
                  <span>Waktu</span>
                  <span className="max-w-[160px] text-right">{new Date(selectedReceipt.dibuat_pada).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Ref ID</span>
                  <span className="max-w-[160px] break-all text-right">{selectedReceipt.ref_id}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Perintah</span>
                  <span className="text-right">{selectedReceipt.perintah}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Produk</span>
                  <span className="text-right">{selectedReceipt.kode_produk}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Tujuan</span>
                  <span className="text-right">{selectedReceipt.tujuan}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Qty</span>
                  <span className="text-right">Rp {Number(selectedReceipt.qty || 0).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Biaya Aktual</span>
                  <span className="text-right">Rp {Number(selectedReceipt.biaya_aktual || selectedReceipt.biaya_perkiraan || 0).toLocaleString("id-ID")}</span>
                </div>
                <div className="flex items-start justify-between gap-3">
                  <span>Status</span>
                  <span className="text-right font-bold uppercase">{selectedReceipt.status}</span>
                </div>
              </div>

              <div className="my-3 border-t border-dashed border-slate-400" />

              <div className="text-[11px] uppercase tracking-wide text-slate-500">Keterangan / SN</div>
              <div className="mt-1 break-words text-[12px] leading-5">{selectedReceipt.keterangan || "-"}</div>

              <div className="my-3 border-t border-dashed border-slate-400" />
              <div className="text-center text-[11px] text-slate-500">Terima kasih</div>
            </div>

            <div className="flex justify-center">
              <Button variant="primary" className="h-10 px-4" onClick={() => printReceipt(selectedReceipt)}>
                <FileText className="h-4 w-4" />
                Cetak
              </Button>
            </div>
          </div>
        ) : null}
      </AppModal>

      <AppModal
        open={!!selectedDetail}
        onClose={() => setSelectedDetail(null)}
        title="Detail Transaksi"
        subtitle="Informasi lengkap transaksi member."
      >
        {selectedDetail ? (
          <div className="grid gap-3 md:grid-cols-2 text-sm text-slate-200">
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Waktu</div>
              <div className="mt-1 font-semibold text-white">{new Date(selectedDetail.dibuat_pada).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Ref ID</div>
              <div className="mt-1 break-all font-mono text-cyan-200">{selectedDetail.ref_id}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Perintah</div>
              <div className="mt-1 font-semibold text-white">{selectedDetail.perintah}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Produk</div>
              <div className="mt-1 font-semibold text-white">{selectedDetail.kode_produk}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Tujuan</div>
              <div className="mt-1 font-semibold text-white">{selectedDetail.tujuan}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Status</div>
              <div className="mt-1 font-semibold uppercase text-white">{selectedDetail.status}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Qty Input</div>
              <div className="mt-1 font-semibold text-white">{Number(selectedDetail.qty || 0).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Qty Kirim</div>
              <div className="mt-1 font-semibold text-white">{Number(typeof selectedDetail.qty_provider === "number" ? selectedDetail.qty_provider : selectedDetail.qty).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Beban Transaksi</div>
              <div className="mt-1 font-semibold text-white">{selectedDetail.charge_receiver_applied ? "Penerima" : "Member"}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Fee</div>
              <div className="mt-1 font-semibold text-white">Rp {Number(selectedDetail.fee_member_rp || 0).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Biaya Est</div>
              <div className="mt-1 font-semibold text-white">Rp {Number(selectedDetail.biaya_perkiraan || 0).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3">
              <div className="text-xs uppercase tracking-wide text-slate-400">Biaya Aktual</div>
              <div className="mt-1 font-semibold text-white">Rp {Number(selectedDetail.biaya_aktual || 0).toLocaleString("id-ID")}</div>
            </div>
            <div className="rounded-xl border border-white/10 bg-slate-950/40 p-3 md:col-span-2">
              <div className="text-xs uppercase tracking-wide text-slate-400">Keterangan</div>
              <div className="mt-1 break-words font-semibold text-white">{selectedDetail.keterangan || "-"}</div>
            </div>
          </div>
        ) : null}
      </AppModal>
    </div>
  );
}
