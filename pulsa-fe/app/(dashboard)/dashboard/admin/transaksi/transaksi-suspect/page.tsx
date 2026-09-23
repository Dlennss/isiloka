"use client";

import { useEffect, useMemo, useState } from "react";
import { Banknote, SlidersHorizontal } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { alertConfirm, alertError, alertSuccess } from "@/components/ui/alerts";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { decodeJwt } from "@/lib/jwt";

type AuditRow = {
  transaksi_member_id: number;
  member_id?: number;
  member_nama?: string;
  status_member?: string;
  ref_id_member?: string;
  produk_member?: string;
  tujuan_member?: string;
  qty_member?: number;
  transaksi_provider_id: number;
  provider: string;
  status_provider: string;
  ref_id_provider: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  nominal_provider_request?: number;
  harga?: number;
  kode_respon?: string;
  pesan: string;
  no_referensi?: string;
  provider_dibuat_pada: string;
  resolved?: boolean;
  resolved_at?: string;
  resolved_by_user_id?: number;
  resolved_by_name?: string;
  resolve_note?: string;
  can_retry_refund?: boolean;
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
  has_next?: boolean;
  total_exact?: boolean;
  error?: string;
};

type ResolveResponse = {
  ok?: boolean;
  item?: {
    resolved?: boolean;
    trx_id?: number;
    user_id?: number;
    processed?: number;
    resolved_count?: number;
    failed_count?: number;
    processed_ids?: number[];
    provider_row_id?: number;
    refid?: string;
    product_sent?: string;
    http_status?: number;
    accepted?: boolean;
    immediate_reject?: boolean;
    message?: string;
    already_exists?: boolean;
    provider?: string;
    new_provider_trx_id?: number;
    mode?: string;
    status?: string;
    already_settled?: boolean;
    bank_nama?: string;
    saldo_sesudah?: number;
    total_amount?: number;
    ref_id?: string;
  };
  error?: string;
};

const DEFAULT_LIMIT = 10;

function toLocalDateInputValue(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function lastFiveDaysRange() {
  const toDate = new Date();
  const fromDate = new Date(toDate);
  fromDate.setDate(toDate.getDate() - 4);
  return {
    from: toLocalDateInputValue(fromDate),
    to: toLocalDateInputValue(toDate),
  };
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

function money(n?: number) {
  return `Rp ${Number(n || 0).toLocaleString("id-ID")}`;
}

function getRefID(row: AuditRow) {
  return row.ref_id_provider || row.ref_id_member || "-";
}

function getNominalProviderRequest(row: AuditRow) {
  return Number(row.nominal_provider_request || row.qty || row.qty_member || 0);
}

export default function AdminProviderSuccessSuspiciousMessagePage() {
  const canResolve = useMemo(() => {
    if (typeof window === "undefined") return false;
    const token = localStorage.getItem("auth_token") || "";
    const role = String(decodeJwt(token)?.role || "").trim().toLowerCase();
    return role === "admin" || role === "staff";
  }, []);

  const [rows, setRows] = useState<AuditRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [resendingId, setResendingId] = useState<number | null>(null);
  const [settlingId, setSettlingId] = useState<number | null>(null);
  const [settleRow, setSettleRow] = useState<AuditRow | null>(null);
  const [settleNominal, setSettleNominal] = useState("");
  const [settleFee, setSettleFee] = useState("0");
  const [settleNote, setSettleNote] = useState("");
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [totalPages, setTotalPages] = useState(1);
  const [hasNext, setHasNext] = useState(false);
  const [totalExact, setTotalExact] = useState(false);

  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [provider, setProvider] = useState("");
  const [refID, setRefID] = useState("");
  const [resolveStatus, setResolveStatus] = useState("unresolved");
  const [from, setFrom] = useState(() => lastFiveDaysRange().from);
  const [to, setTo] = useState(() => lastFiveDaysRange().to);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const hasRefIDFilter = refID.trim() !== "";
  const activeFilterCount = [provider, refID, ...(hasRefIDFilter ? [] : [from, to])].filter((v) => v.trim()).length + (resolveStatus !== "all" ? 1 : 0);
  const page = Math.floor(offset / limit) + 1;
  const settleTotal = useMemo(() => Number(settleNominal || 0) + Number(settleFee || 0), [settleNominal, settleFee]);

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
    overrides?: Partial<{ provider: string; refID: string; resolveStatus: string; from: string; to: string }>,
  ) {
    setLoading(true);
    setLoadError("");
    try {
      const providerValue = overrides?.provider ?? provider;
      const refValue = overrides?.refID ?? refID;
      const resolveStatusValue = overrides?.resolveStatus ?? resolveStatus;
      const fromValue = overrides?.from ?? from;
      const toValue = overrides?.to ?? to;
      const refValueTrimmed = refValue.trim();

      const qs = new URLSearchParams();
      qs.set("limit", String(limit));
      qs.set("offset", String(nextOffset));
      qs.set("fast_page", "1");
      if (providerValue.trim()) qs.set("provider", providerValue.trim());
      if (refValueTrimmed) qs.set("ref_id", refValueTrimmed);
      qs.set("resolve_status", resolveStatusValue);
      if (!refValueTrimmed && fromValue.trim()) qs.set("from", fromValue.trim());
      if (!refValueTrimmed && toValue.trim()) qs.set("to", toValue.trim());

      const r = await fetch(`/api/admin/audit/transaksi/transaksi-suspect?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        setRows([]);
        setTotal(0);
        setTotalPages(1);
        setHasNext(false);
        setTotalExact(false);
        setLoadError(j.error || `Gagal memuat data audit (HTTP ${r.status}).`);
        return;
      }

      const nextRows = Array.isArray(j.items) ? j.items : [];
      setRows(nextRows);
      setTotal(Number(j.total || 0));
      setTotalPages(Math.max(1, Number(j.total_pages || 1)));
      setHasNext(Boolean(j.has_next));
      setTotalExact(Boolean(j.total_exact));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadProviders();
    void load(0);
     
  }, []);

  async function retryRefundNoSuccess(row: AuditRow) {
    if (!canResolve || !row.can_retry_refund) return;
    const confirmed = await alertConfirm({
      title: "Proses Ulang Transaksi",
      text: `Transaksi provider ${row.transaksi_provider_id} akan diproses ulang karena refund dan status member masih success dan belum ada provider lain yang sukses. Lanjutkan?`,
      confirmButtonText: "Proses Ulang",
    });
    if (!confirmed) return;

    setResendingId(row.transaksi_provider_id);
    try {
      const r = await fetch("/api/admin/history/transaksi/resend", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({
          mode: "refund_no_success",
          provider_trx_id: row.transaksi_provider_id,
        }),
      });
      const j: ResolveResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memproses ulang transaksi.");
        return;
      }
      await alertSuccess(`Proses ulang dimulai. Provider: ${j.item?.provider || "-"}.`);
      await load(offset);
    } finally {
      setResendingId(null);
    }
  }

  function openSettleModal(row: AuditRow) {
    if (!canResolve || row.resolved) return;
    setSettleRow(row);
    setSettleNominal(String(getNominalProviderRequest(row)));
    setSettleFee("0");
    setSettleNote("");
  }

  function closeSettleModal() {
    if (settlingId) return;
    setSettleRow(null);
    setSettleNominal("");
    setSettleFee("0");
    setSettleNote("");
  }

  async function submitSettleBankDebit() {
    if (!canResolve || !settleRow) return;
    const nominal = Number(settleNominal || 0);
    const fee = Number(settleFee || 0);
    if (!Number.isFinite(nominal) || nominal <= 0) {
      await alertError("Nominal wajib lebih dari 0.");
      return;
    }
    if (!Number.isFinite(fee) || fee < 0) {
      await alertError("Fee tidak boleh minus.");
      return;
    }
    const confirmed = await alertConfirm({
      title: "Selesaikan Suspect",
      text: `Debit BCA SUSPECTTT Marwan 8761518283 sebesar ${money(nominal + fee)} untuk transaksi provider ${settleRow.transaksi_provider_id}. Lanjutkan?`,
      confirmButtonText: "Proses",
    });
    if (!confirmed) return;

    setSettlingId(settleRow.transaksi_provider_id);
    try {
      const r = await fetch("/api/admin/audit/transaksi/transaksi-suspect", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeader(),
        },
        body: JSON.stringify({
          action: "settle_bank_debit",
          transaksi_provider_id: settleRow.transaksi_provider_id,
          nominal,
          fee,
          note: settleNote.trim(),
        }),
      });
      const j: ResolveResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menyelesaikan transaksi suspect.");
        return;
      }
      const suffix = j.item?.already_settled ? " Mutasi bank sebelumnya sudah ada, saldo tidak dipotong ulang." : "";
      await alertSuccess(`Transaksi suspect selesai. Saldo ${j.item?.bank_nama || "BCA SUSPECTTT"} sekarang ${money(j.item?.saldo_sesudah)}.${suffix}`);
      closeSettleModal();
      await load(offset);
    } finally {
      setSettlingId(null);
    }
  }

  const columns: DataTableColumn<AuditRow>[] = [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => fmtDate(x.provider_dibuat_pada) },
    { id: "provider", header: "Provider", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.provider },
    {
      id: "markingStatus",
      header: "Keterangan",
      tdClassName: "whitespace-nowrap",
      render: (x) => (
        <span
          className={`inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold ${
            x.resolved
              ? "border-emerald-400/30 bg-emerald-500/15 text-emerald-200"
              : "border-amber-400/30 bg-amber-500/15 text-amber-200"
          }`}
        >
          {x.resolved ? "Diselesaikan" : "Belum selesai"}
        </span>
      ),
    },
    {
      id: "resolvedAt",
      header: "Selesai Pada",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => (x.resolved_at ? fmtDate(x.resolved_at) : "-"),
    },
    {
      id: "resolvedBy",
      header: "Diselesaikan Oleh",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => x.resolved_by_name || "-",
    },
    { id: "refID", header: "Ref ID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (x) => getRefID(x) },
    { id: "produk", header: "Produk", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_produk || x.produk_member || "-" },
    { id: "tujuan", header: "Tujuan", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (x) => x.tujuan },
    { id: "qty", header: "Qty", tdClassName: "whitespace-nowrap text-right text-slate-200", render: (x) => Number(x.qty || 0).toLocaleString("id-ID") },
    { id: "nominalProviderRequest", header: "Nominal Request", tdClassName: "whitespace-nowrap text-right text-cyan-100", render: (x) => money(getNominalProviderRequest(x)) },
    { id: "harga", header: "Harga Provider", tdClassName: "whitespace-nowrap text-right text-emerald-200", render: (x) => money(x.harga) },
    { id: "rc", header: "RC", tdClassName: "whitespace-nowrap text-slate-200", render: (x) => x.kode_respon || "-" },
    { id: "pesan", header: "Pesan Provider", tdClassName: "min-w-90 text-slate-200", render: (x) => <div className="max-w-[440px] whitespace-normal break-words">{x.pesan || "-"}</div> },
    { id: "statusMember", header: "Status Member", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.status_member || "-" },

    ...(canResolve
      ? [
          {
            id: "actions",
            header: "Aksi",
            tdClassName: "whitespace-nowrap",
            render: (x: AuditRow) => {
              const canResend = Boolean(x.can_retry_refund);
              const canSettleBank = !x.resolved;
              if (!canResend && !canSettleBank) return <span className="text-slate-500">-</span>;
              return (
                <div className="flex items-center gap-2">
                  {canSettleBank ? (
                    <Button
                      type="button"
                      variant="success"
                      className="h-8 px-2 text-xs"
                      disabled={settlingId === x.transaksi_provider_id}
                      onClick={() => openSettleModal(x)}
                    >
                      <Banknote className="h-3.5 w-3.5" />
                      {settlingId === x.transaksi_provider_id ? "Memproses..." : "Selesaikan"}
                    </Button>
                  ) : null}
                  {canResend ? (
                    <Button
                      type="button"
                      variant="outline"
                      className="h-8 border-cyan-400/30 bg-cyan-500/10 px-2 text-xs text-cyan-100 hover:bg-cyan-500/20"
                      disabled={resendingId === x.transaksi_provider_id}
                      onClick={() => void retryRefundNoSuccess(x)}
                    >
                      {resendingId === x.transaksi_provider_id ? "Memproses..." : "Proses Ulang"}
                    </Button>
                  ) : null}
                </div>
              );
            },
          } satisfies DataTableColumn<AuditRow>,
        ]
      : []),
  ];

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Transaksi Provider Suspect</div>
        <div className="text-sm text-muted-foreground">Transaksi provider suspect dengan status member success, terutama provider refund/saldo dikembalikan yang perlu diproses ulang jika belum ada provider lain yang sukses.</div>
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
          {activeFilterCount > 0 ? <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span> : null}
        </Button>
      </div>

      <div className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-cyan-950/25 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-6">
          <select value={provider} onChange={(e) => setProvider(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="">semua provider</option>
            {providers.map((p) => (
              <option key={p.id} value={(p.nama || "").trim().toLowerCase()}>
                {p.nama || `provider-${p.id}`}
              </option>
            ))}
          </select>

          <input value={refID} onChange={(e) => setRefID(e.target.value)} placeholder="Ref ID" className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none placeholder:text-slate-500 focus:border-cyan-400/50" />

          <select value={resolveStatus} onChange={(e) => setResolveStatus(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
            <option value="all">semua status</option>
            <option value="unresolved">belum selesai</option>
            <option value="resolved">diselesaikan</option>
          </select>

          <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" />

          <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert" />

          <div className="grid grid-cols-2 gap-2">
            <Button
              type="button"
              variant="primary"
              className="h-10"
              disabled={loading}
              onClick={() => {
                const nextOffset = 0;
                const refSearch = refID.trim() !== "";
                setOffset(nextOffset);
                if (refSearch) {
                  setFrom("");
                  setTo("");
                }
                void load(nextOffset, refSearch ? { from: "", to: "" } : undefined);
                setMobileFilterOpen(false);
              }}
            >
              {loading ? "Memuat..." : "Terapkan"}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={loading}
              onClick={() => {
                const range = lastFiveDaysRange();
                setProvider("");
                setRefID("");
                setResolveStatus("unresolved");
                setFrom(range.from);
                setTo(range.to);
                setOffset(0);
                void load(0, { provider: "", refID: "", resolveStatus: "unresolved", from: range.from, to: range.to });
                setMobileFilterOpen(false);
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-1 text-sm text-slate-300 sm:flex-row sm:items-center sm:justify-between">
        <div>Total transaksi: {total.toLocaleString("id-ID")}{!totalExact && hasNext ? "+" : ""}</div>
        {loadError ? <div className="text-rose-300">{loadError}</div> : null}
      </div>

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(x) => `${x.transaksi_provider_id}`}
        emptyText={loading ? "Memuat..." : "Tidak ada transaksi provider suspect yang belum/masih perlu ditangani manual."}
        minWidthClassName="min-w-300"
        showRowNumber
        rowNumberStart={offset + 1}
        pagination={{
          page,
          totalPages,
          onPrev: () => {
            const nextOffset = Math.max(0, offset - limit);
            setOffset(nextOffset);
            void load(nextOffset);
          },
          onNext: () => {
            const nextOffset = offset + limit;
            setOffset(nextOffset);
            void load(nextOffset);
          },
          disablePrev: loading || offset <= 0,
          disableNext: loading || (!hasNext && page >= totalPages),
        }}
      />

      <AppModal
        open={!!settleRow}
        onClose={closeSettleModal}
        title="Selesaikan Suspect"
        subtitle="Saldo BCA SUSPECTTT Marwan 8761518283 akan didebit dan dicatat ke mutasi rekening."
        footer={
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button type="button" variant="outline" onClick={closeSettleModal} disabled={!!settlingId}>
              Batal
            </Button>
            <Button type="button" variant="success" onClick={() => void submitSettleBankDebit()} disabled={!settleRow || !!settlingId}>
              {settlingId ? "Memproses..." : "Proses"}
            </Button>
          </div>
        }
      >
        {settleRow ? (
          <div className="space-y-4 text-sm text-slate-200">
            <div className="rounded-md border border-white/10 bg-slate-950/55 p-3">
              <div className="font-medium text-white">Provider trx #{settleRow.transaksi_provider_id}</div>
              <div className="mt-1 text-slate-400">Ref ID: {getRefID(settleRow)}</div>
              <div className="mt-1 text-slate-400">Provider: {settleRow.provider || "-"}</div>
              <div className="mt-1 text-slate-400">Nominal request: {money(getNominalProviderRequest(settleRow))}</div>
              <div className="mt-1 text-slate-400">Harga provider: {money(settleRow.harga)}</div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              <label className="space-y-1">
                <div className="text-xs uppercase tracking-wide text-slate-400">Nominal Request</div>
                <Input className="h-10 bg-slate-950/60" inputMode="numeric" value={settleNominal} readOnly />
              </label>
              <label className="space-y-1">
                <div className="text-xs uppercase tracking-wide text-slate-400">Fee</div>
                <Input className="h-10 bg-slate-950/60" inputMode="numeric" value={settleFee} onChange={(e) => setSettleFee(e.target.value.replace(/\D+/g, ""))} />
              </label>
            </div>

            <label className="block space-y-1">
              <div className="text-xs uppercase tracking-wide text-slate-400">Catatan</div>
              <textarea
                className="min-h-24 w-full rounded-md border border-input bg-slate-950/60 px-3 py-2 text-sm outline-none focus-visible:ring-1 focus-visible:ring-ring"
                value={settleNote}
                onChange={(e) => setSettleNote(e.target.value)}
                placeholder="Catatan opsional"
              />
            </label>

            <div className="rounded-md border border-emerald-400/20 bg-emerald-500/10 p-3">
              <div>Total debit bank: <span className="font-semibold text-emerald-100">{money(settleTotal)}</span></div>
            </div>
          </div>
        ) : null}
      </AppModal>
    </div>
  );
}
