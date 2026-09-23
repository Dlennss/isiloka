"use client";
import { fmtID } from "@/lib/format";

import { type FormEvent, useCallback, useEffect, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type MutasiRow = {
  id: number;
  member_id: number;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string | null;
  saldo_sebelum?: number | null;
  saldo_sesudah?: number | null;
  dibuat_pada: string;
};

type ApiResp = {
  ok: boolean;
  rows?: MutasiRow[];
  total?: number;
  total_pages?: number;
  error?: string;
};

type ExportKind = "" | "csv" | "excel" | "pdf";

const PAGE_SIZE = 10;
const EXPORT_PAGE_SIZE = 200;

async function apiFetch(path: string, init?: RequestInit) {
  const t = localStorage.getItem("auth_token") || "";
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (t) headers.Authorization = `Bearer ${t}`;
  const res = await fetch(path, { ...init, headers: { ...headers, ...(init?.headers || {}) }, cache: "no-store" });
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

export default function HistoryMutasiPage() {
  const [rows, setRows] = useState<MutasiRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [exporting, setExporting] = useState<ExportKind>("");

  // filters
  const [refID, setRefID] = useState("");
  const [arah, setArah] = useState("");
  const [date, setDate] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [exportMonth, setExportMonth] = useState("");
  const [appliedRefID, setAppliedRefID] = useState("");
  const [appliedArah, setAppliedArah] = useState("");
  const [appliedDate, setAppliedDate] = useState("");
  const [appliedFrom, setAppliedFrom] = useState("");
  const [appliedTo, setAppliedTo] = useState("");
  const [applyTick, setApplyTick] = useState(0);

  const load = useCallback(async () => {
    setErr("");
    setLoading(true);
    try {
      const qs = buildQS({
        limit: String(PAGE_SIZE),
        offset: String(offset),
        ref_id: appliedRefID,
        arah: appliedArah,
        date: appliedDate,
        from: appliedFrom,
        to: appliedTo,
      });
      const r = await apiFetch(`/api/me/history/mutasi?${qs}`);
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
  }, [offset, appliedRefID, appliedArah, appliedDate, appliedFrom, appliedTo]);

  useEffect(() => {
    load();
  }, [load, applyTick]);

  async function fetchMonthRows(ym: string): Promise<MutasiRow[] | null> {
    const range = monthToRange(ym);
    if (!range) {
      setErr("Pilih bulan export terlebih dahulu (format YYYY-MM).");
      return null;
    }

    const out: MutasiRow[] = [];
    let offset = 0;

    while (true) {
      const qs = buildQS({
        from: range.from,
        to: range.to,
        limit: String(EXPORT_PAGE_SIZE),
        offset: String(offset),
      });
      const r = await apiFetch(`/api/me/history/mutasi?${qs}`);
      const j: ApiResp = r.data || ({} as ApiResp);
      if (!r.ok || !j.ok) {
        setErr(j.error || "Gagal mengambil data export mutasi");
        return null;
      }

      const chunk = j.rows || [];
      out.push(...chunk);

      if (chunk.length < EXPORT_PAGE_SIZE) break;
      offset += EXPORT_PAGE_SIZE;
    }

    if (!out.length) {
      setErr("Tidak ada data mutasi pada bulan yang dipilih.");
      return null;
    }

    return out;
  }

  function toExportRows(items: MutasiRow[]) {
    return items.map((x, idx) => ({
      no: idx + 1,
      id: x.id,
      waktu: new Date(x.dibuat_pada).toLocaleString("id-ID"),
      arah: x.arah,
      jumlah: x.jumlah,
      alasan: x.alasan,
      catatan: x.catatan || "",
      saldo_sebelum: x.saldo_sebelum ?? "",
      saldo_sesudah: x.saldo_sesudah ?? "",
      ref_id: x.ref_id || "",
    }));
  }

  async function onDownloadCSV() {
    setErr("");
    setExporting("csv");
    try {
      const items = await fetchMonthRows(exportMonth);
      if (!items) return;

      const normalized = toExportRows(items);
      const header = Object.keys(normalized[0]);
      const body = normalized.map((row) => header.map((h) => row[h as keyof typeof row]));
      const csv = [header, ...body].map((row) => row.map(escapeCSV).join(",")).join("\n");

      downloadBlob(`mutasi-${exportMonth}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8;" }));
    } finally {
      setExporting("");
    }
  }

  async function onDownloadExcel() {
    setErr("");
    setExporting("excel");
    try {
      const items = await fetchMonthRows(exportMonth);
      if (!items) return;

      const XLSX = await import("xlsx");
      const normalized = toExportRows(items);
      const ws = XLSX.utils.json_to_sheet(normalized);
      const wb = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(wb, ws, "Mutasi");
      const out = XLSX.write(wb, { bookType: "xlsx", type: "array" });

      downloadBlob(`mutasi-${exportMonth}.xlsx`, new Blob([out], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
    } finally {
      setExporting("");
    }
  }

  async function onDownloadPDF() {
    setErr("");
    setExporting("pdf");
    try {
      const items = await fetchMonthRows(exportMonth);
      if (!items) return;

      const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
      const autoTable = autoTableModule.default;

      const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
      doc.setFontSize(12);
      doc.text(`Laporan Mutasi ${exportMonth}`, 40, 36);

      const rowsPDF = toExportRows(items);
      autoTable(doc, {
        startY: 48,
        styles: { fontSize: 8, cellPadding: 4 },
        headStyles: { fillColor: [45, 45, 45] },
        head: [["No", "ID", "Waktu", "Arah", "Jumlah", "Alasan", "Catatan", "Saldo Sebelum", "Saldo Sesudah", "Ref ID"]],
        body: rowsPDF.map((x) => [x.no, x.id, x.waktu, x.arah, fmtID(x.jumlah), x.alasan, x.catatan, String(x.saldo_sebelum), String(x.saldo_sesudah), x.ref_id]),
      });

      const out = doc.output("arraybuffer");
      downloadBlob(`mutasi-${exportMonth}.pdf`, new Blob([out], { type: "application/pdf" }));
    } finally {
      setExporting("");
    }
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setAppliedRefID(refID.trim());
    setAppliedArah(arah);
    setAppliedDate(date);
    setAppliedFrom(from);
    setAppliedTo(to);
    setOffset(0);
    setApplyTick((v) => v + 1);
  }

  function onReset() {
    setRefID("");
    setArah("");
    setDate("");
    setFrom("");
    setTo("");
    setAppliedRefID("");
    setAppliedArah("");
    setAppliedDate("");
    setAppliedFrom("");
    setAppliedTo("");
    setOffset(0);
    setApplyTick((v) => v + 1);
  }
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const columns: DataTableColumn<MutasiRow>[] = [
    { id: "waktu", header: "Waktu", render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID") },
    { id: "arah", header: "Arah", render: (x) => x.arah },
    { id: "jumlah", header: "Jumlah", render: (x) => `Rp ${fmtID(x.jumlah)}` },
    { id: "alasan", header: "Alasan", render: (x) => x.alasan },
    { id: "catatan", header: "Catatan", render: (x) => x.catatan || "-" },
    { id: "saldo_sebelum", header: "Saldo Sebelum", render: (x) => x.saldo_sebelum == null ? "-" : `Rp ${Number(x.saldo_sebelum).toLocaleString("id-ID")}` },
    { id: "saldo_sesudah", header: "Saldo Sesudah", render: (x) => x.saldo_sesudah == null ? "-" : `Rp ${Number(x.saldo_sesudah).toLocaleString("id-ID")}` },
    { id: "ref_id", header: "Ref ID", render: (x) => x.ref_id || "-" },
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-xs text-white/60">History</div>
          <h1 className="text-lg font-semibold">Mutasi</h1>
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

      <form onSubmit={onSubmit} className="rounded-xl border border-white/10 bg-white/5 p-3">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-12">
          <div className="md:col-span-5">
            <label className="mb-1 block text-xs text-white/60">Ref ID</label>
            <input
              value={refID}
              onChange={(e) => setRefID(e.target.value)}
              placeholder="Cari ref_id"
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none placeholder:text-white/40 focus:border-white/20"
            />
          </div>

          <div className="md:col-span-3">
            <label className="mb-1 block text-xs text-white/60">Arah</label>
            <select
              value={arah}
              onChange={(e) => setArah(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-white/20"
            >
              <option value="">Semua</option>
              <option value="debit">Debit</option>
              <option value="credit">Credit</option>
            </select>
          </div>

          <div className="md:col-span-4">
            <label className="mb-1 block text-xs text-white/60">Tanggal (single day)</label>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-white/20 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
          </div>

          <div className="md:col-span-4">
            <label className="mb-1 block text-xs text-white/60">Dari tanggal</label>
            <input
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-white/20 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
          </div>

          <div className="md:col-span-4">
            <label className="mb-1 block text-xs text-white/60">Sampai tanggal</label>
            <input
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-white/20 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            {(from || to) && (
              <div className="mt-1 text-xs text-white/50">
                Periode aktif: {formatTanggalID(from)} - {formatTanggalID(to)}
              </div>
            )}
          </div>

          <div className="md:col-span-4 flex items-end gap-2">
            <button
              type="submit"
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15"
              disabled={loading}
            >
              Terapkan
            </button>
            <button
              type="button"
              onClick={onReset}
              className="rounded-lg border border-white/10 bg-transparent px-3 py-2 text-sm text-white/70 hover:bg-white/5"
              disabled={loading}
            >
              Reset
            </button>

            <div className="ml-auto text-xs text-white/50">{loading ? "Memuat..." : rows.length ? `${rows.length} data` : "0 data"}</div>
          </div>

          <div className="md:col-span-4">
            <label className="mb-1 block text-xs text-white/60">Download Bulanan</label>
            <input
              type="month"
              value={exportMonth}
              onChange={(e) => setExportMonth(e.target.value)}
              className="w-full rounded-lg border border-white/10 bg-black/20 px-3 py-2 text-sm text-white/90 outline-none focus:border-white/20 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-85 [&::-webkit-calendar-picker-indicator]:invert hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            {exportMonth && <div className="mt-1 text-xs text-white/50">Bulan terpilih: {formatBulanID(exportMonth)}</div>}
          </div>

          <div className="md:col-span-8 flex items-end gap-2">
            <button
              type="button"
              onClick={onDownloadCSV}
              disabled={loading || !!exporting}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {exporting === "csv" ? "Menyiapkan CSV..." : "Download CSV"}
            </button>
            <button
              type="button"
              onClick={onDownloadExcel}
              disabled={loading || !!exporting}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {exporting === "excel" ? "Menyiapkan Excel..." : "Download Excel"}
            </button>
            <button
              type="button"
              onClick={onDownloadPDF}
              disabled={loading || !!exporting}
              className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-sm text-white/90 hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {exporting === "pdf" ? "Menyiapkan PDF..." : "Download PDF"}
            </button>
            <span className="text-xs text-white/50">Export berdasarkan bulan terpilih.</span>
          </div>
        </div>
      </form>

      {err ? <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-200">{err}</div> : null}

      {rows.length > 0 ? (
        <DataTable<MutasiRow>
          columns={columns}
          rows={rows}
          rowKey={(x) => x.id}
          rowNumberStart={offset + 1}
          emptyText="Belum ada data"
          minWidthClassName="min-w-[1120px]"
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
    </div>
  );
}
