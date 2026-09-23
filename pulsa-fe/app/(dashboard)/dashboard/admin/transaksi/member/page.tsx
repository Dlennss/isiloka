"use client";

import { useEffect, useMemo, useRef, useState, type FormEventHandler } from "react";
import { usePathname } from "next/navigation";
import { Download, SlidersHorizontal } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { fmtID } from "@/lib/format";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";

type TrxRow = {
  id: number;
  member_id: number;
  member_email?: string;
  member_nama?: string;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  status: string;
  biaya_perkiraan: number;
  biaya_aktual: number;
  fee_member_rp: number;
  keterangan: string;
  has_callback_url?: boolean;
  dibuat_pada: string;
};

type MemberItem = {
  id: number;
  nama?: string;
  email?: string;
};

type ProdukItem = {
  id: number;
  sku?: string;
  aktif?: boolean;
};

type ApiResponse = {
  ok?: boolean;
  items?: TrxRow[];
  total?: number;
  error?: string;
};

type MemberApiResponse = {
  ok?: boolean;
  items?: MemberItem[];
};

type ProdukApiResponse = {
  ok?: boolean;
  items?: ProdukItem[];
};

type ActionModalState = {
  open: boolean;
  type: "complete" | "cancel";
  trx: TrxRow | null;
  trxIDs: number[];
  reason: string;
  submitting: boolean;
};

type ExportKind = "" | "csv" | "excel" | "pdf";
type ExportColumnKey = "no" | "waktu" | "member" | "ref_id" | "perintah" | "kode_produk" | "tujuan" | "qty" | "status" | "biaya" | "keterangan";

type ExportRow = Record<ExportColumnKey, string | number>;

const DEFAULT_PAGE_SIZE = 10;
const PAGE_SIZE_OPTIONS = [10, 25, 50, 100] as const;
const EXPORT_BATCH_SIZE = 500;
const EXPORT_COLUMNS: Array<{ key: ExportColumnKey; label: string }> = [
  { key: "no", label: "No" },
  { key: "waktu", label: "Waktu" },
  { key: "member", label: "Member" },
  { key: "ref_id", label: "Ref ID" },
  { key: "perintah", label: "Perintah" },
  { key: "kode_produk", label: "Produk" },
  { key: "tujuan", label: "Tujuan" },
  { key: "qty", label: "Qty" },
  { key: "status", label: "Status" },
  { key: "biaya", label: "Biaya" },
  { key: "keterangan", label: "Keterangan" },
];
const DEFAULT_EXPORT_COLUMNS: ExportColumnKey[] = ["no", "waktu", "member", "ref_id", "kode_produk", "tujuan", "qty", "status", "biaya"];

function toLocalDateInputValue(date: Date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
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

function memberLabel(x: Pick<TrxRow, "member_email" | "member_nama" | "member_id">) {
  return x.member_email?.trim() || x.member_nama?.trim() || String(x.member_id);
}

export default function AdminAllTransaksiMemberPage() {
  const pathname = usePathname();
  const now = useMemo(() => new Date(), []);
  const defaultFrom = useMemo(() => toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)), [now]);
  const defaultTo = useMemo(() => toLocalDateInputValue(now), [now]);
  const canExport = !pathname?.startsWith("/dashboard/operator") && !pathname?.startsWith("/dashboard/wallet");
  const isOperatorMemberPage = pathname?.startsWith("/dashboard/operator/transaksi/member");

  const [rows, setRows] = useState<TrxRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [pageSize, setPageSize] = useState<number>(DEFAULT_PAGE_SIZE);

  const [members, setMembers] = useState<MemberItem[]>([]);
  const [produkItems, setProdukItems] = useState<ProdukItem[]>([]);

  const [refIDInput, setRefIDInput] = useState("");
  const [destInput, setDestInput] = useState("");
  const [statusInput, setStatusInput] = useState("");
  const [memberIDInput, setMemberIDInput] = useState("");
  const [produkInput, setProdukInput] = useState("");
  const [fromInput, setFromInput] = useState(defaultFrom);
  const [toInput, setToInput] = useState(defaultTo);

  const [refID, setRefID] = useState("");
  const [dest, setDest] = useState("");
  const [status, setStatus] = useState("");
  const [memberID, setMemberID] = useState("");
  const [produk, setProduk] = useState("");
  const [from, setFrom] = useState(defaultFrom);
  const [to, setTo] = useState(defaultTo);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);

  const [cancellingTrxID, setCancellingTrxID] = useState<number | null>(null);
  const [completingTrxID, setCompletingTrxID] = useState<number | null>(null);
  const [sendingTrxID, setSendingTrxID] = useState<number | null>(null);
  const [sendingBulk, setSendingBulk] = useState(false);
  const [selectedIDs, setSelectedIDs] = useState<number[]>([]);
  const [actionModal, setActionModal] = useState<ActionModalState>({
    open: false,
    type: "complete",
    trx: null,
    trxIDs: [],
    reason: "",
    submitting: false,
  });

  const [exportOpen, setExportOpen] = useState(false);
  const [exporting, setExporting] = useState<ExportKind>("");
  const [selectedExportColumns, setSelectedExportColumns] = useState<ExportColumnKey[]>(DEFAULT_EXPORT_COLUMNS);
  const didInitLoadRef = useRef(false);

  const page = Math.floor(offset / pageSize) + 1;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const activeFilterCount = [memberIDInput, produkInput, refIDInput, destInput, statusInput, fromInput, toInput].filter((v) => v.trim()).length;

  async function loadMembers() {
    try {
      const qs = new URLSearchParams();
      qs.set("limit", "500");
      qs.set("offset", "0");
      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: MemberApiResponse = await r.json().catch(() => ({}));
      setMembers(Array.isArray(j.items) ? j.items : []);
    } catch {
      setMembers([]);
    }
  }

  async function loadProduk() {
    try {
      const r = await fetch("/api/admin/master/produk", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ProdukApiResponse = await r.json().catch(() => ({}));
      const items = Array.isArray(j.items) ? j.items : [];
      setProdukItems(items.filter((x) => x?.aktif !== false));
    } catch {
      setProdukItems([]);
    }
  }

  async function load(nextOffset = offset, draft?: { refID: string; dest: string; status: string; memberID: string; produk: string; from: string; to: string }) {
    setLoading(true);
    try {
      const v = draft || { refID, dest, status, memberID, produk, from, to };
      const qs = new URLSearchParams();
      qs.set("limit", String(pageSize + 1));
      qs.set("offset", String(nextOffset));
      if (v.refID.trim()) qs.set("ref_id", v.refID.trim());
      if (v.dest.trim()) qs.set("dest", v.dest.trim());
      if (v.status.trim()) qs.set("status", v.status.trim());
      if (v.memberID.trim()) qs.set("member_id", v.memberID.trim());
      if (v.produk.trim()) qs.set("kode_produk", v.produk.trim());
      if (v.from.trim()) qs.set("from", v.from.trim());
      if (v.to.trim()) qs.set("to", v.to.trim());

      const r = await fetch(`/api/admin/history/transaksi?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });

      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        setRows([]);
        setTotal(0);
        setHasNext(false);
        return;
      }

      const all = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > pageSize);
      setRows(all.slice(0, pageSize));
      setTotal(Number(j.total || 0));
      setSelectedIDs((prev) => prev.filter((id) => all.slice(0, pageSize).some((row) => row.id === id)));
    } finally {
      setLoading(false);
    }
  }

  async function fetchAllForExport() {
    const out: TrxRow[] = [];
    let nextOffset = 0;

    for (;;) {
      const qs = new URLSearchParams();
      qs.set("limit", String(EXPORT_BATCH_SIZE));
      qs.set("offset", String(nextOffset));
      if (refID.trim()) qs.set("ref_id", refID.trim());
      if (dest.trim()) qs.set("dest", dest.trim());
      if (status.trim()) qs.set("status", status.trim());
      if (memberID.trim()) qs.set("member_id", memberID.trim());
      if (produk.trim()) qs.set("kode_produk", produk.trim());
      if (from.trim()) qs.set("from", from.trim());
      if (to.trim()) qs.set("to", to.trim());

      const r = await fetch(`/api/admin/history/transaksi?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        throw new Error(j.error || "Gagal mengambil data export");
      }

      const batch = Array.isArray(j.items) ? j.items : [];
      out.push(...batch);
      if (batch.length < EXPORT_BATCH_SIZE) break;
      nextOffset += batch.length;
    }

    return out;
  }

  function toExportRows(items: TrxRow[]): ExportRow[] {
    return items.map((x, idx) => ({
      no: idx + 1,
      waktu: fmtDate(x.dibuat_pada),
      member: memberLabel(x),
      ref_id: x.ref_id,
      perintah: x.perintah,
      kode_produk: x.kode_produk,
      tujuan: x.tujuan,
      qty: x.qty,
      status: x.status,
      biaya: x.biaya_aktual || x.biaya_perkiraan || 0,
      keterangan: x.keterangan || "-",
    }));
  }

  async function runExport(kind: ExportKind) {
    if (!selectedExportColumns.length) {
      await alertWarning("Pilih minimal satu kolom untuk export.");
      return;
    }
    setExporting(kind);
    try {
      const items = await fetchAllForExport();
      const normalized = toExportRows(items);
      const pickedLabels = EXPORT_COLUMNS.filter((col) => selectedExportColumns.includes(col.key)).map((col) => col.label);
      const body = normalized.map((row) => selectedExportColumns.map((key) => row[key] ?? ""));

      if (kind === "csv") {
        const csv = [pickedLabels, ...body].map((row) => row.map((cell) => `"${String(cell ?? "").replaceAll('"', '""')}"`).join(",")).join("\n");
        downloadBlob(`transaksi-member-${from || "all"}_${to || "all"}.csv`, new Blob([csv], { type: "text/csv;charset=utf-8;" }));
      } else if (kind === "excel") {
        const XLSX = await import("xlsx");
        const ws = XLSX.utils.json_to_sheet(normalized.map((row) => Object.fromEntries(selectedExportColumns.map((key, idx) => [pickedLabels[idx], row[key]]))));
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "Transaksi Member");
        const out = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        downloadBlob(`transaksi-member-${from || "all"}_${to || "all"}.xlsx`, new Blob([out], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" }));
      } else if (kind === "pdf") {
        const [{ jsPDF }, autoTableModule] = await Promise.all([import("jspdf"), import("jspdf-autotable")]);
        const autoTable = autoTableModule.default;
        const doc = new jsPDF({ orientation: "landscape", unit: "pt", format: "a4" });
        autoTable(doc, {
          head: [pickedLabels],
          body,
          styles: { fontSize: 8, cellPadding: 4 },
          headStyles: { fillColor: [15, 23, 42] },
          margin: { top: 36, left: 24, right: 24, bottom: 24 },
        });
        const out = doc.output("arraybuffer");
        downloadBlob(`transaksi-member-${from || "all"}_${to || "all"}.pdf`, new Blob([out], { type: "application/pdf" }));
      }
      setExportOpen(false);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal export transaksi member");
    } finally {
      setExporting("");
    }
  }

  const onApplyFilter: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const draft = {
      refID: refIDInput,
      dest: destInput,
      status: statusInput,
      memberID: memberIDInput,
      produk: produkInput,
      from: fromInput,
      to: toInput,
    };
    if (draft.from && draft.to && new Date(draft.from).getTime() > new Date(draft.to).getTime()) {
      await alertWarning("Range tanggal tidak valid. Pastikan nilai 'dari' tidak lebih besar dari 'sampai'.");
      return;
    }

    setRefID(draft.refID);
    setDest(draft.dest);
    setStatus(draft.status);
    setMemberID(draft.memberID);
    setProduk(draft.produk);
    setFrom(draft.from);
    setTo(draft.to);

    const nextOffset = 0;
    setOffset(nextOffset);
    await load(nextOffset, draft);
    setMobileFilterOpen(false);
  };

  async function onResetFilter() {
    setRefIDInput("");
    setDestInput("");
    setStatusInput("");
    setMemberIDInput("");
    setProdukInput("");
    setFromInput(defaultFrom);
    setToInput(defaultTo);

    setRefID("");
    setDest("");
    setStatus("");
    setMemberID("");
    setProduk("");
    setFrom(defaultFrom);
    setTo(defaultTo);

    const nextOffset = 0;
    setOffset(nextOffset);
    await load(nextOffset, { refID: "", dest: "", status: "", memberID: "", produk: "", from: defaultFrom, to: defaultTo });
    setMobileFilterOpen(false);
  }

  function openActionModal(type: "complete" | "cancel", trx: TrxRow) {
    setActionModal({ open: true, type, trx, trxIDs: [trx.id], reason: "", submitting: false });
  }

  async function openBulkActionModal(type: "complete" | "cancel") {
    if (!selectedIDs.length) {
      await alertWarning("Pilih minimal satu transaksi.");
      return;
    }
    setActionModal({ open: true, type, trx: null, trxIDs: selectedIDs, reason: "", submitting: false });
  }

  function closeActionModal() {
    if (actionModal.submitting) return;
    setActionModal({ open: false, type: "complete", trx: null, trxIDs: [], reason: "", submitting: false });
  }

  async function submitActionModal() {
    const trx = actionModal.trx;
    const trxIDs = actionModal.trxIDs;
    if (!trxIDs.length) return;
    const reason = actionModal.reason.trim();
    const isBulk = trxIDs.length > 1;

    if (actionModal.type === "cancel" && !reason) {
      await alertWarning("Alasan batal wajib diisi.");
      return;
    }

    if (!isBulk && trx) {
      if (actionModal.type === "complete") {
        setCompletingTrxID(trx.id);
      } else {
        setCancellingTrxID(trx.id);
      }
    }
    setActionModal((prev) => ({ ...prev, submitting: true }));
    try {
      const endpoint = actionModal.type === "complete" ? "/api/admin/history/transaksi/complete" : "/api/admin/history/transaksi/cancel";
      const payload =
        trxIDs.length === 1
          ? { trx_id: trxIDs[0], reason }
          : { trx_ids: trxIDs, reason };
      const r = await fetch(endpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify(payload),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || (actionModal.type === "complete" ? "Gagal menyelesaikan transaksi" : "Gagal membatalkan transaksi"));
        return;
      }
      if (isBulk) {
        const processed = Number(j?.processed_count || 0);
        const failed = Number(j?.failed_count || 0);
        await alertSuccess(
          actionModal.type === "complete"
            ? `Bulk sukseskan selesai. ${processed} berhasil, ${failed} gagal.`
            : `Bulk batalkan selesai. ${processed} berhasil, ${failed} gagal.`,
        );
      } else {
        await alertSuccess(actionModal.type === "complete" ? "Transaksi berhasil diselesaikan" : "Transaksi berhasil dibatalkan");
      }
      setSelectedIDs([]);
      await load(offset);
      closeActionModal();
    } finally {
      if (!isBulk && trx) {
        if (actionModal.type === "complete") setCompletingTrxID(null);
        else setCancellingTrxID(null);
      }
      setActionModal((prev) => ({ ...prev, submitting: false }));
    }
  }

  async function sendCallback(trxIDs: number[]) {
    if (!trxIDs.length) {
      await alertWarning("Pilih minimal satu transaksi.");
      return;
    }
    const invalidRows = trxIDs
      .map((trxID) => rows.find((row) => row.id === trxID))
      .filter((row): row is TrxRow => Boolean(row && !row.has_callback_url));
    if (invalidRows.length > 0) {
      const refs = invalidRows
        .slice(0, 3)
        .map((row) => row.ref_id)
        .filter(Boolean);
      const suffix = invalidRows.length > refs.length ? " dan lainnya" : "";
      const detail = refs.length ? ` Ref: ${refs.join(", ")}${suffix}.` : "";
      await alertWarning(`Callback tidak bisa dikirim karena member belum punya webhook aktif.${detail}`);
      return;
    }
    const isBulk = trxIDs.length > 1;
    if (isBulk) setSendingBulk(true);
    else setSendingTrxID(trxIDs[0]);

    try {
      const payload = isBulk ? { trx_ids: trxIDs } : { trx_id: trxIDs[0] };
      const r = await fetch("/api/admin/history/transaksi/send-callback", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify(payload),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        await alertError(j?.error || "Gagal mengirim callback transaksi");
        return;
      }
      if (isBulk) {
        const processed = Number(j?.processed_count || 0);
        const failed = Number(j?.failed_count || 0);
        await alertSuccess(`Kirim callback selesai. ${processed} berhasil, ${failed} gagal.`);
      } else {
        await alertSuccess("Callback transaksi berhasil dikirim.");
      }
      setSelectedIDs([]);
      await load(offset);
    } finally {
      if (isBulk) setSendingBulk(false);
      else setSendingTrxID(null);
    }
  }

  useEffect(() => {
    loadMembers();
    loadProduk();
    load(0);
    didInitLoadRef.current = true;
     
  }, []);

  useEffect(() => {
    if (!didInitLoadRef.current) return;
    void load(offset);
     
  }, [offset, pageSize]);

  const visibleSelectableRows = useMemo(() => rows.filter((row) => {
    const st = (row.status || "").toLowerCase();
    return st === "pending" || st === "failed" || st === "success";
  }), [rows]);
  const allVisibleSelected = visibleSelectableRows.length > 0 && visibleSelectableRows.every((row) => selectedIDs.includes(row.id));

  function toggleRow(id: number, checked: boolean) {
    setSelectedIDs((prev) => {
      if (checked) return prev.includes(id) ? prev : [...prev, id];
      return prev.filter((value) => value !== id);
    });
  }

  function toggleSelectAll(checked: boolean) {
    setSelectedIDs((prev) => {
      if (checked) {
        const merged = new Set(prev);
        visibleSelectableRows.forEach((row) => merged.add(row.id));
        return Array.from(merged);
      }
      const visibleIDs = new Set(visibleSelectableRows.map((row) => row.id));
      return prev.filter((id) => !visibleIDs.has(id));
    });
  }

  const columns: DataTableColumn<TrxRow>[] = [
    {
      id: "select",
      header: (
        <input
          type="checkbox"
          className="h-4 w-4 rounded border-white/20 bg-slate-900 text-cyan-400 focus:ring-cyan-400/40"
          checked={allVisibleSelected}
          onChange={(e) => toggleSelectAll(e.target.checked)}
          aria-label="Pilih semua"
        />
      ),
      thClassName: "w-10",
      tdClassName: "whitespace-nowrap",
      render: (x) => (
        <input
          type="checkbox"
          className="h-4 w-4 rounded border-white/20 bg-slate-900 text-cyan-400 focus:ring-cyan-400/40"
          checked={selectedIDs.includes(x.id)}
          onChange={(e) => toggleRow(x.id, e.target.checked)}
          aria-label={`Pilih transaksi ${x.ref_id}`}
        />
      ),
    },
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => fmtDate(x.dibuat_pada) },
    { id: "member", header: "Member", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => memberLabel(x) },
    { id: "ref", header: "Ref ID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (x) => x.ref_id },
    { id: "produk", header: "Produk", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_produk },
    { id: "tujuan", header: "Tujuan", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.tujuan },
    { id: "qty", header: "Qty", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => fmtID(x.qty) },
    { id: "status", header: "Status", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.status },
    { id: "biaya", header: "Biaya", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => fmtID(x.biaya_aktual || x.biaya_perkiraan || 0) },
    { id: "ket", header: "Keterangan", tdClassName: "max-w-80 wrap-break-word text-slate-300", render: (x) => x.keterangan || "-" },
  ];

  const actions: DataTableActions<TrxRow> = {
    header: "Aksi",
    align: "right",
    render: (t) => {
      if (isOperatorMemberPage) {
        const canSendCallback = Boolean(t.has_callback_url);
        return (
          <div className="inline-flex gap-2">
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-8 rounded-xl px-3 text-xs"
              title={canSendCallback ? "Kirim callback" : "Member belum punya webhook aktif"}
              disabled={sendingTrxID === t.id || loading || sendingBulk || !canSendCallback}
              onClick={() => void sendCallback([t.id])}
            >
              {sendingTrxID === t.id ? "Mengirim..." : "Kirim"}
            </Button>
          </div>
        );
      }
      const st = (t.status || "").toLowerCase();
      const canComplete = st === "pending" || st === "failed";
      const canCancel = st === "pending" || st === "success";
      if (!canComplete && !canCancel) return <span className="text-xs text-slate-500">-</span>;
      return (
        <div className="inline-flex gap-2">
          {canComplete ? (
            <Button type="button" size="sm" variant="success" className="h-8 rounded-xl px-3 text-xs" disabled={completingTrxID === t.id || loading} onClick={() => openActionModal("complete", t)}>
              {completingTrxID === t.id ? "Memproses..." : "Sukseskan"}
            </Button>
          ) : null}
          {canCancel ? (
            <Button type="button" size="sm" variant="danger" className="h-8 rounded-xl px-3 text-xs" disabled={cancellingTrxID === t.id || loading} onClick={() => openActionModal("cancel", t)}>
              {cancellingTrxID === t.id ? "Memproses..." : "Batalkan"}
            </Button>
          ) : null}
        </div>
      );
    },
  };

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">All Transaksi Member</div>
        <div className="text-sm text-muted-foreground">Riwayat transaksi semua member.</div>
      </div>

      <div className="md:hidden">
        <Button type="button" variant="outline" className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" onClick={() => setMobileFilterOpen((v) => !v)}>
          <span className="inline-flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
          </span>
          {activeFilterCount > 0 ? <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span> : null}
        </Button>
      </div>

      <form onSubmit={onApplyFilter} className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-cyan-950/25 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}>
        <div className="grid grid-cols-2 gap-2 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-7">
          <select value={memberIDInput} onChange={(e) => setMemberIDInput(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="">semua member</option>
            {members.map((m) => <option key={m.id} value={String(m.id)}>{m.nama || m.email || `member-${m.id}`}</option>)}
          </select>
          <input value={refIDInput} onChange={(e) => setRefIDInput(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50" placeholder="ref_id" />
          <input value={destInput} onChange={(e) => setDestInput(e.target.value)} className="col-span-2 h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 md:col-span-1" placeholder="no tujuan" />
          <select value={produkInput} onChange={(e) => setProdukInput(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="">semua produk</option>
            {produkItems.map((p) => <option key={p.id} value={p.sku || ""}>{p.sku || `produk-${p.id}`}</option>)}
          </select>
          <select value={statusInput} onChange={(e) => setStatusInput(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="">semua status</option>
            <option value="pending">pending</option>
            <option value="success">success</option>
            <option value="failed">failed</option>
          </select>
          <div className="col-span-2 grid grid-cols-2 gap-2 md:col-span-2">
            <div className="min-w-0">
              <label className="mb-1 block text-[11px] font-medium text-slate-300 md:hidden">Dari</label>
              <input value={fromInput} onChange={(e) => setFromInput(e.target.value)} type="date" className="h-10 w-full min-w-0 rounded-xl border border-white/15 bg-slate-950/55 px-3 pr-9 text-base text-slate-100 outline-none focus:border-cyan-400/50 md:text-sm scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" />
            </div>
            <div className="min-w-0">
              <label className="mb-1 block text-[11px] font-medium text-slate-300 md:hidden">Sampai</label>
              <input value={toInput} onChange={(e) => setToInput(e.target.value)} type="date" className="h-10 w-full min-w-0 rounded-xl border border-white/15 bg-slate-950/55 px-3 pr-9 text-base text-slate-100 outline-none focus:border-cyan-400/50 md:text-sm scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" />
            </div>
          </div>
          <div className="col-span-2 grid grid-cols-1 gap-2 sm:grid-cols-3 md:col-span-4 lg:col-span-6 xl:col-span-7 xl:flex xl:justify-end">
            <Button type="submit" variant="primary" className="h-10 w-full min-w-0 xl:w-36" disabled={loading}>{loading ? "Memuat..." : "Terapkan"}</Button>
            <Button type="button" variant="outline" className="h-10 w-full min-w-0 border-white/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50 xl:w-32" onClick={() => void onResetFilter()} disabled={loading}>Reset</Button>
            {canExport ? (
              <Button type="button" variant="outline" className="h-10 w-full min-w-0 border-white/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50 xl:w-32" onClick={() => setExportOpen(true)} disabled={loading}>
                <span className="inline-flex items-center justify-center gap-2"><Download className="h-4 w-4" />Export</span>
              </Button>
            ) : null}
          </div>
        </div>
      </form>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <div className="text-sm text-slate-400">{selectedIDs.length} terpilih</div>
          {isOperatorMemberPage ? (
            <Button type="button" variant="outline" className="h-9 rounded-xl px-3 text-xs" disabled={!selectedIDs.length || loading || sendingBulk} onClick={() => void sendCallback(selectedIDs)}>
              {sendingBulk ? "Mengirim..." : "Kirim Terpilih"}
            </Button>
          ) : (
            <>
              <Button type="button" variant="success" className="h-9 rounded-xl px-3 text-xs" disabled={!selectedIDs.length || loading} onClick={() => void openBulkActionModal("complete")}>
                Sukseskan Terpilih
              </Button>
              <Button type="button" variant="danger" className="h-9 rounded-xl px-3 text-xs" disabled={!selectedIDs.length || loading} onClick={() => void openBulkActionModal("cancel")}>
                Batalkan Terpilih
              </Button>
            </>
          )}
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-300">
          <span>Per page</span>
          <select
            value={pageSize}
            onChange={(e) => {
              const next = Number(e.target.value) || DEFAULT_PAGE_SIZE;
              setPageSize(next);
              setOffset(0);
            }}
            className="h-9 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          >
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </label>
      </div>

      <DataTable columns={columns} rows={rows} rowKey={(x) => x.id} rowNumberStart={offset + 1} minWidthClassName="min-w-[1400px]" emptyText="Tidak ada data." actions={actions} pagination={{ page, totalPages, onPrev: () => setOffset((v) => Math.max(0, v - pageSize)), onNext: () => setOffset((v) => v + pageSize), onPageChange: (nextPage) => setOffset((nextPage - 1) * pageSize), disablePrev: loading || offset === 0, disableNext: loading || !hasNext }} />

      {canExport ? (
        <AppModal open={exportOpen} onClose={() => !exporting && setExportOpen(false)} title="Export Transaksi Member">
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

      <AppModal open={!isOperatorMemberPage && actionModal.open} onClose={closeActionModal} title={actionModal.type === "complete" ? (actionModal.trxIDs.length > 1 ? "Sukseskan Transaksi Terpilih" : "Sukseskan Transaksi") : (actionModal.trxIDs.length > 1 ? "Batalkan Transaksi Terpilih" : "Batalkan Transaksi")}>
        <div className="space-y-4 text-sm text-slate-200">
          <div className="rounded-2xl border border-white/10 bg-slate-950/50 p-4 text-sm text-slate-200">
            {actionModal.trx ? (
              <>
                <div className="font-medium text-white">Ref ID: {actionModal.trx.ref_id}</div>
                <div className="mt-1 text-slate-400">Status saat ini: {actionModal.trx.status || "-"}</div>
              </>
            ) : (
              <>
                <div className="font-medium text-white">{actionModal.trxIDs.length} transaksi dipilih</div>
                <div className="mt-1 text-slate-400">Aksi akan diproses dalam satu request bulk.</div>
              </>
            )}
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-200">Alasan {actionModal.type === "complete" ? "(opsional)" : "(wajib)"}</label>
            <textarea value={actionModal.reason} onChange={(e) => setActionModal((prev) => ({ ...prev, reason: e.target.value }))} rows={4} className="w-full rounded-2xl border border-white/12 bg-slate-950/55 px-4 py-3 text-sm text-slate-100 outline-none focus:border-cyan-400/40" placeholder={actionModal.type === "complete" ? "Tambahkan catatan jika diperlukan" : "Jelaskan alasan pembatalan"} disabled={actionModal.submitting} />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" className="border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50" onClick={closeActionModal} disabled={actionModal.submitting}>Tutup</Button>
            <Button type="button" variant={actionModal.type === "complete" ? "success" : "danger"} onClick={() => void submitActionModal()} disabled={actionModal.submitting}>{actionModal.submitting ? "Memproses..." : actionModal.type === "complete" ? "Sukseskan" : "Batalkan"}</Button>
          </div>
        </div>
      </AppModal>
    </div>
  );
}
