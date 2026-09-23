"use client";

import { useEffect, useMemo, useState } from "react";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Download, SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { alertError } from "@/components/ui/alerts";
import { AppModal } from "@/components/ui/app-modal";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type AuditRow = {
  transaksi_provider_id: number;
  transaksi_member_id: number;
  member_id?: number;
  member_nama?: string;
  provider: string;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  kode_respon?: string;
  pesan?: string;
  harga?: number;
  http_status?: number;
  percobaan: number;
  dibuat_pada: string;
  no_referensi?: string;
};

type ProviderItem = {
  id: number;
  nama?: string;
  aktif?: boolean;
};

type ApiResponse = {
  ok?: boolean;
  items?: AuditRow[];
  total?: number;
  limit?: number;
  offset?: number;
  page?: number;
  total_pages?: number;
  error?: string;
};

type ExportKind = "" | "csv" | "excel" | "pdf";
type ExportColumnKey =
  | "no"
  | "waktu"
  | "provider"
  | "ref_id"
  | "member"
  | "perintah"
  | "kode_produk"
  | "tujuan"
  | "qty"
  | "rc"
  | "pesan"
  | "http_status"
  | "percobaan"
  | "provider_trx_id"
  | "member_trx_id";

const DEFAULT_LIMIT = 10;

function toLocalDateInputValue(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtDate(s: string) {
  try {
    return new Date(s).toLocaleString("id-ID");
  } catch {
    return s;
  }
}

function money(v?: number) {
  if (typeof v !== "number") return "-";
  return new Intl.NumberFormat("id-ID").format(v);
}

function visiblePages(current: number, total: number) {
  const pages = new Set<number>([1, total]);
  for (let i = current - 2; i <= current + 2; i += 1) {
    if (i >= 1 && i <= total) pages.add(i);
  }
  return Array.from(pages).sort((a, b) => a - b);
}

function escapeCSV(value: unknown) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
  return text;
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

const EXPORT_COLUMNS: Array<{ key: ExportColumnKey; label: string }> = [
  { key: "no", label: "No" },
  { key: "waktu", label: "Waktu" },
  { key: "provider", label: "Provider" },
  { key: "ref_id", label: "Ref ID" },
  { key: "member", label: "Member" },
  { key: "perintah", label: "Perintah" },
  { key: "kode_produk", label: "Kode Produk" },
  { key: "tujuan", label: "Tujuan" },
  { key: "qty", label: "Qty" },
  { key: "rc", label: "RC" },
  { key: "pesan", label: "Pesan" },
  { key: "http_status", label: "HTTP Status" },
  { key: "percobaan", label: "Percobaan" },
  { key: "provider_trx_id", label: "Provider Trx ID" },
  { key: "member_trx_id", label: "Member Trx ID" },
];

const DEFAULT_EXPORT_COLUMNS: ExportColumnKey[] = [
  "no",
  "waktu",
  "provider",
  "ref_id",
  "member",
  "kode_produk",
  "tujuan",
  "qty",
  "http_status",
  "percobaan",
  "provider_trx_id",
];

export default function AdminProviderEmptyResponsePage() {
  const [rows, setRows] = useState<AuditRow[]>([]);
  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [provider, setProvider] = useState("");
  const [refID, setRefID] = useState("");
  const [from, setFrom] = useState(() => {
    const now = new Date();
    return toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1));
  });
  const [to, setTo] = useState(() => toLocalDateInputValue(new Date()));
  const [offset, setOffset] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const [exporting, setExporting] = useState<ExportKind>("");
  const [exportOpen, setExportOpen] = useState(false);
  const [selectedExportColumns, setSelectedExportColumns] = useState<ExportColumnKey[]>(DEFAULT_EXPORT_COLUMNS);

  const page = Math.floor(offset / limit) + 1;
  const pages = useMemo(() => visiblePages(page, totalPages), [page, totalPages]);
  const activeFilterCount = [provider, refID, from, to].filter((v) => v.trim()).length;

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/master/provider", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const items: ProviderItem[] = Array.isArray(j?.items) ? j.items : [];
      setProviders(items.filter((x) => x?.aktif !== false));
    } catch {
      setProviders([]);
    }
  }

  async function load(
    nextOffset = offset,
    overrides?: Partial<{
      provider: string;
      refID: string;
      from: string;
      to: string;
    }>
  ) {
    setLoading(true);
    try {
      const providerValue = overrides?.provider ?? provider;
      const refIDValue = overrides?.refID ?? refID;
      const fromValue = overrides?.from ?? from;
      const toValue = overrides?.to ?? to;

      const qs = new URLSearchParams();
      qs.set("limit", String(limit));
      qs.set("offset", String(nextOffset));
      if (providerValue.trim()) qs.set("provider", providerValue.trim());
      if (refIDValue.trim()) qs.set("ref_id", refIDValue.trim());
      if (fromValue.trim()) qs.set("from", fromValue.trim());
      if (toValue.trim()) qs.set("to", toValue.trim());

      const r = await fetch(`/api/admin/audit/transaksi/provider-empty-response?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        setRows([]);
        setTotal(0);
        setTotalPages(1);
        return;
      }
      setRows(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
      setTotalPages(Math.max(1, Number(j.total_pages || 1)));
    } finally {
      setLoading(false);
    }
  }

  async function fetchAllFilteredRows() {
    const out: AuditRow[] = [];
    let nextOffset = 0;
    const fetchLimit = 200;

    for (;;) {
      const qs = new URLSearchParams();
      qs.set("limit", String(fetchLimit));
      qs.set("offset", String(nextOffset));
      if (provider.trim()) qs.set("provider", provider.trim());
      if (refID.trim()) qs.set("ref_id", refID.trim());
      if (from.trim()) qs.set("from", from.trim());
      if (to.trim()) qs.set("to", to.trim());

      const r = await fetch(`/api/admin/audit/transaksi/provider-empty-response?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        throw new Error(j.error || "Gagal mengambil data export.");
      }

      const batch = Array.isArray(j.items) ? j.items : [];
      out.push(...batch);

      if (batch.length < fetchLimit) break;
      nextOffset += fetchLimit;
    }

    if (!out.length) {
      throw new Error("Tidak ada data indikasi timeout sesuai filter.");
    }

    return out;
  }

  function toExportRows(items: AuditRow[]) {
    return items.map((x, idx) => ({
      no: idx + 1,
      waktu: fmtDate(x.dibuat_pada),
      provider: x.provider,
      ref_id: x.ref_id,
      member: x.member_nama ? `${x.member_nama}${x.member_id ? ` (#${x.member_id})` : ""}` : x.member_id ? `#${x.member_id}` : "-",
      perintah: x.perintah,
      kode_produk: x.kode_produk,
      tujuan: x.tujuan,
      qty: x.qty,
      rc: x.kode_respon || "",
      pesan: x.pesan || "",
      http_status: typeof x.http_status === "number" ? x.http_status : "",
      percobaan: x.percobaan,
      provider_trx_id: x.transaksi_provider_id,
      member_trx_id: x.transaksi_member_id,
    }));
  }

  async function runExport(kind: ExportKind) {
    if (!selectedExportColumns.length) {
      await alertError("Pilih minimal satu kolom untuk export.");
      return;
    }

    setExporting(kind);
    try {
      const items = await fetchAllFilteredRows();
      const normalized = toExportRows(items);
      const pickedLabels = EXPORT_COLUMNS.filter((col) => selectedExportColumns.includes(col.key)).map((col) => col.label);
      const body = normalized.map((row) => selectedExportColumns.map((key) => row[key]));

      if (kind === "csv") {
        const csv = [pickedLabels, ...body].map((row) => row.map(escapeCSV).join(",")).join("\n");
        downloadBlob(`audit-timeout-provider-${from || "all"}_${to || "all"}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8;" }));
      } else if (kind === "excel") {
        const XLSX = await import("xlsx");
        const sheetRows = normalized.map((row) =>
          Object.fromEntries(selectedExportColumns.map((key, idx) => [pickedLabels[idx], row[key]]))
        );
        const ws = XLSX.utils.json_to_sheet(sheetRows);
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "Audit Timeout");
        const out = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        downloadBlob(
          `audit-timeout-provider-${from || "all"}_${to || "all"}.xlsx`,
          new Blob([out], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" })
        );
      } else if (kind === "pdf") {
        const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
        const autoTable = autoTableModule.default;
        const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
        doc.setFontSize(12);
        doc.text(`Audit Indikasi Timeout Provider (${from || "all"} s/d ${to || "all"})`, 40, 36);
        autoTable(doc, {
          startY: 48,
          styles: { fontSize: 7, cellPadding: 4 },
          headStyles: { fillColor: [45, 45, 45] },
          head: [pickedLabels],
          body,
        });
        const out = doc.output("arraybuffer");
        downloadBlob(`audit-timeout-provider-${from || "all"}_${to || "all"}.pdf`, new Blob([out], { type: "application/pdf" }));
      }

      setExportOpen(false);
    } catch (err) {
      const message =
        err instanceof Error
          ? err.message
          : kind === "csv"
            ? "Gagal export CSV."
            : kind === "excel"
              ? "Gagal export Excel."
              : "Gagal export PDF.";
      await alertError(message);
    } finally {
      setExporting("");
    }
  }

  function toggleExportColumn(key: ExportColumnKey) {
    setSelectedExportColumns((prev) => (prev.includes(key) ? prev.filter((x) => x !== key) : [...prev, key]));
  }

  useEffect(() => {
    void loadProviders();
    void load(0);
     
  }, []);

  const columns: DataTableColumn<AuditRow>[] = [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => fmtDate(x.dibuat_pada) },
    { id: "provider", header: "Provider", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.provider },
    { id: "ref", header: "Ref ID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (x) => x.ref_id },
    { id: "member", header: "Member", tdClassName: "text-slate-200", render: (x) => x.member_nama ? `${x.member_nama}${x.member_id ? ` (#${x.member_id})` : ""}` : x.member_id ? `#${x.member_id}` : "-" },
    { id: "produk", header: "Produk", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_produk },
    { id: "tujuan", header: "Tujuan", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.tujuan },
    { id: "qty", header: "Qty", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => money(x.qty) },
    { id: "harga", header: "Harga", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => money(x.harga) },
    { id: "rc", header: "RC", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.kode_respon || "-" },
    { id: "pesan", header: "Pesan", tdClassName: "max-w-[240px] text-slate-300", render: (x) => x.pesan?.trim() || "-" },
    { id: "http", header: "HTTP", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => (typeof x.http_status === "number" ? String(x.http_status) : "-") },
    { id: "attempt", header: "Percobaan", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => String(x.percobaan) },
    { id: "trx", header: "Provider Trx ID", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => String(x.transaksi_provider_id) },
  ];

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-2xl border border-amber-400/20 bg-linear-to-br from-amber-500/12 via-slate-900/80 to-rose-500/10 p-4 shadow-[0_24px_60px_-34px_rgba(251,191,36,0.45)]">
        <div className="space-y-1">
          <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-amber-200">Audit Provider</div>
          <div className="text-xl font-semibold tracking-tight text-slate-50">Indikasi Timeout</div>
          <div className="text-sm text-slate-300">Transaksi pending dengan RC dan pesan provider yang masih kosong.</div>
        </div>
      </div>

      <div className="md:hidden">
        <Button
          type="button"
          variant="outline"
          className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
          onClick={() => setMobileFilterOpen((v) => !v)}
        >
          <span className="inline-flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
          </span>
          {activeFilterCount > 0 ? (
            <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span>
          ) : null}
        </Button>
      </div>

      <AppModal
        open={exportOpen}
        onClose={() => {
          if (exporting) return;
          setExportOpen(false);
        }}
        title="Export Audit Timeout"
        subtitle="Pilih kolom yang ingin dimasukkan ke file export."
        maxWidthClassName="max-w-3xl"
        footer={
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => setSelectedExportColumns(EXPORT_COLUMNS.map((x) => x.key))}>
              Pilih Semua
            </Button>
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => setSelectedExportColumns(DEFAULT_EXPORT_COLUMNS)}>
              Default
            </Button>
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("csv")}>
              {exporting === "csv" ? "Menyiapkan CSV..." : "CSV"}
            </Button>
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("excel")}>
              {exporting === "excel" ? "Menyiapkan Excel..." : "Excel"}
            </Button>
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={!!exporting} onClick={() => void runExport("pdf")}>
              {exporting === "pdf" ? "Menyiapkan PDF..." : "PDF"}
            </Button>
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

      <div
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-cyan-950/25 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="grid grid-cols-1 gap-2 md:grid-cols-12">
          <select
            value={provider}
            onChange={(e) => setProvider(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 md:col-span-2"
          >
            <option value="">semua provider</option>
            {providers.map((p) => (
              <option key={p.id} value={(p.nama || "").trim().toLowerCase()}>
                {p.nama || `provider-${p.id}`}
              </option>
            ))}
          </select>

          <input
            value={refID}
            onChange={(e) => setRefID(e.target.value)}
            placeholder="Ref ID"
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none placeholder:text-slate-500 focus:border-cyan-400/50 md:col-span-2"
          />

          <div className="grid grid-cols-2 gap-2 md:col-span-5">
            <label className="space-y-1">
              <span className="text-[11px] font-medium uppercase tracking-wide text-slate-400 md:hidden">Dari</span>
              <input
                type="date"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
                className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
              />
            </label>

            <label className="space-y-1">
              <span className="text-[11px] font-medium uppercase tracking-wide text-slate-400 md:hidden">Sampai</span>
              <input
                type="date"
                value={to}
                onChange={(e) => setTo(e.target.value)}
                className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
              />
            </label>
          </div>

          <div className="grid grid-cols-3 gap-2 md:col-span-3">
            <Button
              type="button"
              variant="primary"
              className="h-10"
              disabled={loading}
              onClick={() => {
                const nextOffset = 0;
                setOffset(nextOffset);
                void load(nextOffset);
                setMobileFilterOpen(false);
              }}
            >
              Cari
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-10 border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
              disabled={loading}
              onClick={() => {
                setProvider("");
                setRefID("");
                const now = new Date();
                const defaultFrom = toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1));
                const defaultTo = toLocalDateInputValue(now);
                setFrom(defaultFrom);
                setTo(defaultTo);
                const nextOffset = 0;
                setOffset(nextOffset);
                void load(nextOffset, { provider: "", refID: "", from: defaultFrom, to: defaultTo });
                setMobileFilterOpen(false);
              }}
            >
              Reset
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-10 gap-2 border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
              disabled={!!exporting}
              onClick={() => setExportOpen(true)}
            >
              <Download className="h-4 w-4" />
              Export
            </Button>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-white/12 bg-slate-950/45 p-3">
        <div className="mb-2 flex items-center justify-between text-sm text-slate-400">
          <span>Total {total.toLocaleString("id-ID")} data</span>
          <span>Page {page} / {totalPages}</span>
        </div>
        <DataTable<AuditRow>
          columns={columns}
          data={rows}
          loading={loading}
          rowKey={(x) => `${x.transaksi_provider_id}`}
          emptyMessage="Tidak ada transaksi provider dengan indikasi timeout."
        />

        <div className="mt-3 flex flex-wrap items-center justify-end gap-2">
          <Button type="button" size="icon" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={loading || page <= 1} onClick={() => { const nextOffset = 0; setOffset(nextOffset); void load(nextOffset); }}>
            <ChevronsLeft className="h-4 w-4" />
          </Button>
          <Button type="button" size="icon" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={loading || page <= 1} onClick={() => { const nextOffset = Math.max(0, offset - limit); setOffset(nextOffset); void load(nextOffset); }}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          {pages.map((p) => (
            <Button
              key={p}
              type="button"
              variant="ghost"
              className={p === page ? "border border-cyan-400/70 bg-cyan-500/15 text-cyan-100" : "text-slate-300 hover:bg-slate-800/50"}
              onClick={() => { const nextOffset = (p - 1) * limit; setOffset(nextOffset); void load(nextOffset); }}
              disabled={loading}
            >
              {p}
            </Button>
          ))}
          <Button type="button" size="icon" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={loading || page >= totalPages} onClick={() => { const nextOffset = offset + limit; setOffset(nextOffset); void load(nextOffset); }}>
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button type="button" size="icon" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" disabled={loading || page >= totalPages} onClick={() => { const nextOffset = Math.max(0, (totalPages - 1) * limit); setOffset(nextOffset); void load(nextOffset); }}>
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
