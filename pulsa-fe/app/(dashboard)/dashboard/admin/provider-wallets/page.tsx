"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { fmtID } from "@/lib/format";

type Row = {
  provider: string;
  saldo_internal: number;
  saldo_provider: number;
  selisih: number;
  snapshot_at?: string;
};

type HistoryRow = {
  id: number;
  provider: string;
  bank_id?: number;
  bank_nama?: string;
  ref_id: string;
  arah: "credit" | "debit";
  jumlah: number;
  alasan: string;
  catatan: string;
  saldo_sebelum: number;
  saldo_sesudah: number;
  transaksi_member_id?: number;
  transaksi_provider_id?: number;
  diubah_oleh?: number;
  diubah_oleh_nama?: string;
  dibuat_pada: string;
};

type BankOption = {
  id: number;
  nama: string;
  nomor_rekening: string;
  atas_nama: string;
  saldo: number;
  aktif: boolean;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

function onlyDigits(s: string): string {
  return (s || "").replace(/[^\d]/g, "");
}

function formatRupiahDigits(digits: string): string {
  const n = Number(digits || "0");
  if (!digits) return "";
  if (!Number.isFinite(n)) return "";
  return `Rp ${fmtID(n)}`;
}

function visiblePages(current: number, total: number): number[] {
  const start = Math.max(1, current - 2);
  const end = Math.min(total, current + 2);
  const pages: number[] = [];
  for (let p = start; p <= end; p += 1) pages.push(p);
  return pages;
}

function providerLabel(provider: string): string {
  return (provider || "-").toUpperCase();
}

export default function ProviderWalletsPage() {
  const pathname = usePathname();
  const isWalletMode = pathname.startsWith("/dashboard/wallet");
  const [items, setItems] = useState<Row[]>([]);
  const [banks, setBanks] = useState<BankOption[]>([]);
  const [loading, setLoading] = useState(false);

  const [depositProvider, setDepositProvider] = useState<string>("yuscom");
  const [depositBankID, setDepositBankID] = useState("");
  const [depositAmountDigits, setDepositAmountDigits] = useState<string>("");
  const [depositAdminDigits, setDepositAdminDigits] = useState<string>("");
  const [depositNote, setDepositNote] = useState<string>("");
  const [depositOpen, setDepositOpen] = useState(false);
  const [depositSubmitting, setDepositSubmitting] = useState(false);

  const [provider, setProvider] = useState<string>("yuscom");
  const [mode, setMode] = useState<"credit" | "debit">("credit");
  const [amountDigits, setAmountDigits] = useState<string>("");
  const [note, setNote] = useState<string>("");
  const [adjustOpen, setAdjustOpen] = useState(false);
  const [adjustSubmitting, setAdjustSubmitting] = useState(false);

  const [historyProvider, setHistoryProvider] = useState<string>("");
  const [historyArah, setHistoryArah] = useState<"" | "credit" | "debit">("credit");
  const [historyItems, setHistoryItems] = useState<HistoryRow[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyLimit] = useState(10);
  const [historyOffset, setHistoryOffset] = useState(0);
  const [historyTotal, setHistoryTotal] = useState(0);

  async function load() {
    setLoading(true);
    try {
      const [walletRes, bankRes] = await Promise.all([
        fetch(`/api/admin/provider/wallets`, {
          headers: authHeader(),
          cache: "no-store",
        }),
        fetch(`/api/admin/master/bank`, {
          headers: authHeader(),
          cache: "no-store",
        }),
      ]);
      const walletJSON = await walletRes.json().catch(() => ({}));
      const bankJSON = await bankRes.json().catch(() => ({}));
      setItems(walletJSON.data || []);
      setBanks(Array.isArray(bankJSON.items) ? bankJSON.items : []);
    } finally {
      setLoading(false);
    }
  }

  async function loadHistory(
    p: string,
    offset = 0,
    arah: "credit" | "" | "debit" = historyArah
  ) {
    setHistoryLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("provider", p);
      qs.set("limit", String(historyLimit));
      qs.set("offset", String(offset));
      if (arah) qs.set("arah", arah);

      const r = await fetch(`/api/admin/provider/wallets/history?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal load riwayat");
        return;
      }

      setHistoryProvider(p);
      setHistoryArah(arah);
      setHistoryItems(Array.isArray(j.items) ? j.items : []);
      setHistoryTotal(Number(j.total || 0));
      setHistoryOffset(offset);
    } finally {
      setHistoryLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function submitAdjust() {
    if (adjustSubmitting) return;
    const amt = Number(amountDigits || "0");
    if (!provider || !Number.isFinite(amt) || amt <= 0) {
      await alertWarning("Nominal koreksi wajib lebih dari 0.");
      return;
    }

    setAdjustSubmitting(true);
    try {
      const r = await fetch(`/api/admin/provider/wallets/adjust`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({ provider, amount: amt, mode, note: note.trim() }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal update saldo");
        return;
      }

      await alertSuccess(`RefID: ${j.refid || "-"} • saldo internal: ${j.saldo_internal}`);
      setAmountDigits("");
      setNote("");
      setMode("credit");
      setAdjustOpen(false);
      await load();
      if (historyProvider) {
        await loadHistory(historyProvider, 0, historyArah);
      }
    } finally {
      setAdjustSubmitting(false);
    }
  }

  async function submitDeposit() {
    if (depositSubmitting) return;
    const amt = Number(depositAmountDigits || "0");
    const adminFee = Number(depositAdminDigits || "0");
    if (!depositBankID) {
      await alertWarning("Bank sumber wajib dipilih.");
      return;
    }
    if (!depositProvider || !Number.isFinite(amt) || amt <= 0) {
      await alertWarning("Nominal tambah saldo wajib lebih dari 0.");
      return;
    }
    if (!Number.isFinite(adminFee) || adminFee < 0) {
      await alertWarning("Admin tidak valid.");
      return;
    }

    setDepositSubmitting(true);
    try {
      const r = await fetch(`/api/admin/provider/wallets/deposit`, {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: Number(depositBankID),
          provider: depositProvider,
          amount: amt,
          admin_fee: adminFee,
          note: depositNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal tambah saldo provider");
        return;
      }

      const savedAdminFee = Number(j.admin_fee || 0);
      await alertSuccess(
        `RefID: ${j.refid || "-"} • saldo bank: Rp ${fmtID(Number(j.bank_saldo || 0))}${savedAdminFee > 0 ? ` • Admin: Rp ${fmtID(savedAdminFee)}` : ""} • saldo provider: Rp ${fmtID(Number(j.saldo_internal || 0))}`
      );
      setDepositAmountDigits("");
      setDepositAdminDigits("");
      setDepositNote("");
      setDepositBankID("");
      setDepositOpen(false);
      await load();
      if (historyProvider) {
        await loadHistory(historyProvider, 0, historyArah);
      }
    } finally {
      setDepositSubmitting(false);
    }
  }

  const selectedLabel = providerLabel(provider);
  const amountPreview = formatRupiahDigits(amountDigits);
  const depositSelectedLabel = providerLabel(depositProvider);
  const depositAmountPreview = formatRupiahDigits(depositAmountDigits);
  const depositAdminPreview = formatRupiahDigits(depositAdminDigits);
  const depositBankDebit = Number(depositAmountDigits || "0") + Number(depositAdminDigits || "0");
  const selectedSourceBank = banks.find((bank) => String(bank.id) === depositBankID) || null;
  const historyTotalPages = Math.max(1, Math.ceil(historyTotal / historyLimit));
  const historyCurrentPage = Math.floor(historyOffset / historyLimit) + 1;
  const historyPageButtons = visiblePages(historyCurrentPage, historyTotalPages);
  const totalSaldoInternal = items.reduce((sum, x) => sum + Number(x.saldo_internal || 0), 0);
  const totalSaldoProvider = items.reduce((sum, x) => sum + Number(x.saldo_provider || 0), 0);

  const providerColumns: DataTableColumn<Row>[] = [
    {
      id: "provider",
      header: "Provider",
      thClassName: "w-35",
      tdClassName: "whitespace-nowrap",
      render: (d) => (
        <span className="inline-flex items-center gap-2">
          <span className="rounded-md border border-white/10 bg-white/5 px-2 py-1 text-xs font-semibold uppercase tracking-wide">
            {d.provider}
          </span>
          {d.provider === provider ? <span className="text-xs text-muted-foreground">selected</span> : null}
        </span>
      ),
    },
    {
      id: "saldo_internal",
      header: "Saldo Internal",
      thClassName: "w-35",
      tdClassName: "whitespace-nowrap",
      render: (d) => (
        <span className="rounded-md bg-emerald-500/15 px-2 py-1 font-medium text-emerald-300">Rp {fmtID(d.saldo_internal)}</span>
      ),
    },
    {
      id: "saldo_provider",
      header: "Saldo Snapshot",
      thClassName: "w-35",
      tdClassName: "whitespace-nowrap",
      render: (d) => (
        <span className="rounded-md bg-sky-500/15 px-2 py-1 font-medium text-sky-300">Rp {fmtID(d.saldo_provider)}</span>
      ),
    },
    {
      id: "selisih",
      header: "Selisih",
      thClassName: "w-35",
      tdClassName: "whitespace-nowrap",
      render: (d) => (
        <span
          className={[
            "rounded-md px-2 py-1 font-medium",
            d.selisih < 0 ? "bg-rose-500/15 text-rose-300" : "bg-amber-500/15 text-amber-300",
          ].join(" ")}
        >
          Rp {fmtID(d.selisih)}
        </span>
      ),
    },
    {
      id: "snapshot_at",
      header: "Snapshot At",
      thClassName: "w-35",
      tdClassName: "whitespace-nowrap",
      render: (d) => (d.snapshot_at ? new Date(d.snapshot_at).toLocaleString() : "-"),
    },
  ];

  const providerActions: DataTableActions<Row> = {
    header: "Aksi",
    align: "right",
    thClassName: "w-55",
    tdClassName: "whitespace-nowrap",
    render: (d) => (
      <div className="flex flex-nowrap items-center justify-end gap-2">
        <Button
          type="button"
          variant="success"
          className="h-9 rounded-full px-2"
          onClick={() => pickProviderForDeposit(d.provider)}
        >
          <span className="mr-2 inline-flex h-4 w-4 items-center justify-center">
            <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14" />
              <path d="M5 12h14" />
            </svg>
          </span>
          Tambah Saldo
        </Button>
        {!isWalletMode ? (
          <Button
            type="button"
            variant="warning"
            className="h-9 rounded-full px-2"
            onClick={() => pickProviderForAdjust(d.provider)}
          >
            <span className="mr-2 inline-flex h-4 w-4 items-center justify-center">
              <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M3 10a9 9 0 0 1 15.55-6.36L21 6" />
                <path d="M21 14a9 9 0 0 1-15.55 6.36L3 18" />
                <path d="M21 6h-5" />
                <path d="M3 18h5" />
              </svg>
            </span>
            Koreksi Saldo
          </Button>
        ) : null}
        <Button
          type="button"
          variant="info"
          className="h-9 rounded-full px-2"
          onClick={() => openHistoryModal(d.provider)}
          disabled={historyLoading}
        >
          <span className="inline-flex h-4 w-4 items-center justify-center">
            <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6Z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </span>
          {historyProvider === d.provider && historyLoading ? "Loading..." : "Lihat Riwayat"}
        </Button>
      </div>
    ),
  };

  const historyColumns: DataTableColumn<HistoryRow>[] = [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap", render: (h) => new Date(h.dibuat_pada).toLocaleString() },
    { id: "ref_id", header: "RefID", tdClassName: "whitespace-nowrap", render: (h) => h.ref_id || "-" },
    {
      id: "bank",
      header: "Bank Sumber",
      tdClassName: "whitespace-nowrap",
      render: (h) => h.bank_nama || "-",
    },
    {
      id: "arah",
      header: "Arah",
      tdClassName: "whitespace-nowrap",
      render: (h) => (
        <span
          className={[
            "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold",
            h.arah === "credit" ? "bg-emerald-500/15 text-emerald-300" : "bg-rose-500/15 text-rose-300",
          ].join(" ")}
        >
          {h.arah}
        </span>
      ),
    },
    { id: "jumlah", header: "Jumlah", tdClassName: "whitespace-nowrap", render: (h) => `Rp ${fmtID(h.jumlah)}` },
    { id: "saldo_sebelum", header: "Saldo Sebelum", tdClassName: "whitespace-nowrap", render: (h) => `Rp ${fmtID(h.saldo_sebelum)}` },
    { id: "saldo_sesudah", header: "Saldo Sesudah", tdClassName: "whitespace-nowrap", render: (h) => `Rp ${fmtID(h.saldo_sesudah)}` },
    { id: "alasan", header: "Alasan", tdClassName: "whitespace-nowrap", render: (h) => h.alasan },
    { id: "catatan", header: "Catatan", tdClassName: "whitespace-nowrap", render: (h) => h.catatan || "-" },
    {
      id: "diubah_oleh",
      header: "Diubah Oleh",
      tdClassName: "whitespace-nowrap",
      render: (h) => (h.diubah_oleh_nama && h.diubah_oleh ? `${h.diubah_oleh_nama} (${h.diubah_oleh})` : h.diubah_oleh ? String(h.diubah_oleh) : "-"),
    },
  ];

  function pickProviderForAdjust(p: string) {
    setProvider(p);
    setMode("credit");
    setAmountDigits("");
    setNote("");
    setAdjustOpen(true);
  }

  function pickProviderForDeposit(p: string) {
    setDepositProvider(p);
    setDepositBankID("");
    setDepositAmountDigits("");
    setDepositAdminDigits("");
    setDepositNote("");
    setDepositOpen(true);
  }

  function openHistoryModal(p: string) {
    setHistoryOpen(true);
    void loadHistory(p, 0, historyArah);
  }

  return (
    <div className="space-y-4 p-2">
      <div className="mx-auto w-full">
        <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
          <div>
            <div className="text-lg font-semibold tracking-tight">Provider Wallets</div>
            <div className="text-sm text-muted-foreground">
              Saldo internal vs saldo provider (snapshot callback).
            </div>
          </div>

          <Button
            variant="primary"
            className="h-10 md:w-auto"
            onClick={load}
            disabled={loading}
          >
            {loading ? "Loading..." : "Refresh"}
          </Button>
        </div>
        <div className="grid grid-cols-2 gap-3 border-b border-white/10 mt-2">
            <div className="w-full rounded-2xl border border-emerald-400/25 bg-linear-to-r from-emerald-500/15 to-teal-500/10 px-4 py-3 text-sm text-emerald-100 shadow-[0_14px_30px_-20px_rgba(16,185,129,0.8)]">
              <div className="text-emerald-200/90">Total Saldo Internal (Semua Provider)</div>
              <div className="font-semibold text-emerald-50">Rp {fmtID(totalSaldoInternal)}</div>
            </div>
            <div className="w-full rounded-2xl border border-sky-400/25 bg-linear-to-r from-sky-500/15 to-cyan-500/10 px-4 py-3 text-sm text-sky-100 shadow-[0_14px_30px_-20px_rgba(14,165,233,0.8)]">
              <div className="text-sky-200/90">Total Saldo Snapshot (Semua Provider)</div>
              <div className="font-semibold text-sky-50">Rp {fmtID(totalSaldoProvider)}</div>
            </div>
          </div>

        <DataTable<Row>
          columns={providerColumns}
          rows={items}
          rowKey={(row) => row.provider}
          emptyText="Belum ada data."
          minWidthClassName="min-w-190"
          wrapperClassName="mt-5 overflow-auto rounded-md border border-white/10 bg-linear-to-br from-slate-900/50 via-slate-900/35 to-slate-800/30 shadow-xl"
          showRowNumber={false}
          actions={providerActions}
        />

        <div className="h-8" />
      </div>

      <AppModal
        open={historyOpen}
        onClose={() => {
          if (historyLoading) return;
          setHistoryOpen(false);
        }}
        title={`Riwayat Mutasi Provider${historyProvider ? ` • ${historyProvider.toUpperCase()}` : ""}`}
        subtitle="Riwayat credit dan debit saldo internal provider."
        maxWidthClassName="max-w-7xl"
        footer={
          <div className="flex flex-col gap-3 text-sm md:flex-row md:items-center md:justify-end">
            <div className="mr-auto text-xs text-muted-foreground">
              Halaman {historyCurrentPage} / {historyTotalPages} • Total {historyTotal} data
            </div>
            <div className="flex flex-wrap items-center justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => loadHistory(historyProvider, 0, historyArah)}
                disabled={historyLoading || !historyProvider || historyCurrentPage <= 1}
                aria-label="Halaman awal"
              >
                <ChevronsLeft className="h-4 w-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  loadHistory(
                    historyProvider,
                    Math.max(0, historyOffset - historyLimit),
                    historyArah
                  )
                }
                disabled={historyLoading || historyOffset === 0 || !historyProvider}
                aria-label="Halaman sebelumnya"
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              {historyPageButtons[0] > 1 ? <span className="px-1 text-slate-400">...</span> : null}
              {historyPageButtons.map((p) => (
                <Button
                  key={p}
                  type="button"
                  size="sm"
                  variant={p === historyCurrentPage ? "primary" : "outline"}
                  className={p === historyCurrentPage ? "border-cyan-400 bg-cyan-500/15 text-cyan-100" : ""}
                  onClick={() =>
                    loadHistory(
                      historyProvider,
                      (p - 1) * historyLimit,
                      historyArah
                    )
                  }
                  disabled={historyLoading || !historyProvider}
                >
                  {p}
                </Button>
              ))}
              {historyPageButtons[historyPageButtons.length - 1] < historyTotalPages ? (
                <span className="px-1 text-slate-400">...</span>
              ) : null}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  loadHistory(
                    historyProvider,
                    historyOffset + historyLimit,
                    historyArah
                  )
                }
                disabled={historyLoading || historyOffset + historyLimit >= historyTotal || !historyProvider}
                aria-label="Halaman berikutnya"
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  loadHistory(
                    historyProvider,
                    Math.max(0, (historyTotalPages - 1) * historyLimit),
                    historyArah
                  )
                }
                disabled={historyLoading || !historyProvider || historyCurrentPage >= historyTotalPages}
                aria-label="Halaman akhir"
              >
                <ChevronsRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        }
      >
        <div className="mb-3 flex items-center justify-between gap-2">
          <div className="text-sm font-semibold">
            {historyProvider ? `Provider ${historyProvider.toUpperCase()}` : "Riwayat Provider"}
          </div>
          <select
            className="h-9 rounded-md border border-border/60 bg-background px-2 text-sm"
            value={historyArah}
            onChange={(e) => {
              const nextArah = e.target.value as "" | "credit" | "debit";
              setHistoryArah(nextArah);
              if (historyProvider) {
                void loadHistory(historyProvider, 0, nextArah);
              }
            }}
            disabled={!historyProvider || historyLoading}
          >
            <option value="">Semua Arah</option>
            <option value="credit">Credit</option>
            <option value="debit">Debit</option>
          </select>
        </div>

        <DataTable<HistoryRow>
          columns={historyColumns}
          rows={historyItems}
          rowKey={(row) => row.id}
          emptyText={historyLoading ? "Memuat riwayat..." : "Tidak ada riwayat."}
          minWidthClassName="min-w-220"
          wrapperClassName="overflow-auto rounded-md border border-white/10 bg-black/15"
          showRowNumber={false}
        />
      </AppModal>

      <AppModal
        open={depositOpen}
        onClose={() => {
          if (depositSubmitting) return;
          setDepositOpen(false);
          setDepositBankID("");
          setDepositAmountDigits("");
          setDepositAdminDigits("");
          setDepositNote("");
        }}
        title="Tambah Saldo Provider"
        subtitle={`Provider: ${depositSelectedLabel}`}
        maxWidthClassName="max-w-lg"
        footer={
          <div className="flex items-center justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              className="border-zinc-500 bg-zinc-900 text-zinc-100 hover:bg-zinc-800"
              onClick={() => {
                if (depositSubmitting) return;
                setDepositOpen(false);
                setDepositBankID("");
                setDepositAmountDigits("");
                setDepositAdminDigits("");
                setDepositNote("");
              }}
              disabled={depositSubmitting}
            >
              Batal
            </Button>
            <Button
              type="button"
              variant="success"
              onClick={submitDeposit}
              disabled={depositSubmitting}
            >
              {depositSubmitting ? "Menyimpan..." : "Tambah Saldo"}
            </Button>
          </div>
        }
      >
        <div className="grid gap-3">
          <select
            className="h-11 rounded-md border border-border/60 bg-background px-3 text-sm"
            value={depositBankID}
            onChange={(e) => setDepositBankID(e.target.value)}
            disabled={depositSubmitting}
          >
            <option value="">Pilih bank sumber</option>
            {banks.map((bank) => (
              <option key={bank.id} value={String(bank.id)}>
                {bank.nama} - {bank.atas_nama} - {bank.nomor_rekening} {bank.aktif ? "(Aktif)" : "(Nonaktif)"}
              </option>
            ))}
          </select>

          <div className="rounded-xl border border-white/10 bg-white/5 p-3 text-sm">
            {selectedSourceBank ? (
              <div className="space-y-1">
                <div className="font-medium text-white">{selectedSourceBank.nama}</div>
                <div className="text-white/75">{selectedSourceBank.atas_nama}</div>
                <div className="font-mono text-cyan-200">{selectedSourceBank.nomor_rekening}</div>
                <div className="text-emerald-300">Saldo bank: Rp {fmtID(Number(selectedSourceBank.saldo || 0))}</div>
                <div className={selectedSourceBank.aktif ? "text-sky-300" : "text-amber-300"}>
                  Status: {selectedSourceBank.aktif ? "Aktif" : "Nonaktif"}
                </div>
              </div>
            ) : (
              <div className="text-white/55">Pilih bank aktif sebagai sumber saldo provider.</div>
            )}
          </div>

          <Input
            className="h-11"
            inputMode="numeric"
            value={depositAmountDigits}
            onChange={(e) => setDepositAmountDigits(onlyDigits(e.target.value))}
            placeholder="Nominal tambah saldo"
            disabled={depositSubmitting}
          />
          <div className="text-xs text-muted-foreground">
            {depositAmountDigits ? (
              <>
                Preview: <span className="font-medium text-foreground">{depositAmountPreview}</span>
              </>
            ) : (
              "Preview: -"
            )}
          </div>

          <Input
            className="h-11"
            inputMode="numeric"
            value={depositAdminDigits}
            onChange={(e) => setDepositAdminDigits(onlyDigits(e.target.value))}
            placeholder="Admin"
            disabled={depositSubmitting}
          />
          <div className="text-xs text-muted-foreground">
            {depositAdminDigits ? (
              <>
                Admin: <span className="font-medium text-foreground">{depositAdminPreview}</span>
              </>
            ) : (
              "Admin: -"
            )}
            {depositBankDebit > 0 ? (
              <>
                {" "}• Bank terpotong: <span className="font-medium text-foreground">Rp {fmtID(depositBankDebit)}</span>
              </>
            ) : null}
          </div>

          <Input
            className="h-11"
            value={depositNote}
            onChange={(e) => setDepositNote(e.target.value)}
            placeholder="Catatan transfer bank ke provider"
            disabled={depositSubmitting}
          />
        </div>
      </AppModal>

      {!isWalletMode ? (
        <AppModal
          open={adjustOpen}
          onClose={() => {
            if (adjustSubmitting) return;
            setAdjustOpen(false);
            setMode("credit");
            setAmountDigits("");
            setNote("");
          }}
          title="Koreksi Saldo Internal"
          subtitle={`Provider: ${selectedLabel}`}
          maxWidthClassName="max-w-lg"
          footer={
            <div className="flex items-center justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                className="border-zinc-500 bg-zinc-900 text-zinc-100 hover:bg-zinc-800"
                onClick={() => {
                  if (adjustSubmitting) return;
                  setAdjustOpen(false);
                  setMode("credit");
                  setAmountDigits("");
                  setNote("");
                }}
                disabled={adjustSubmitting}
              >
                Batal
              </Button>
              <Button
                type="button"
                variant={mode === "debit" ? "danger" : "primary"}
                onClick={submitAdjust}
                disabled={adjustSubmitting}
              >
                {adjustSubmitting ? "Menyimpan..." : "Simpan"}
              </Button>
            </div>
          }
        >
          <div className="grid gap-3">
            <select
              className="h-11 rounded-md border border-border/60 bg-background px-3 text-sm"
              value={mode}
              onChange={(e) => setMode(e.target.value as "credit" | "debit")}
              disabled={adjustSubmitting}
            >
              <option value="credit">Tambah (Credit)</option>
              <option value="debit">Kurangi (Debit)</option>
            </select>

            <Input
              className="h-11"
              inputMode="numeric"
              value={amountDigits}
              onChange={(e) => setAmountDigits(onlyDigits(e.target.value))}
              placeholder="Nominal mutasi"
              disabled={adjustSubmitting}
            />
            <div className="text-xs text-muted-foreground">
              {amountDigits ? (
                <>
                  Preview: <span className="font-medium text-foreground">{amountPreview}</span>
                </>
              ) : (
                "Preview: -"
              )}
            </div>

            <Input
              className="h-11"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Catatan (opsional)"
              disabled={adjustSubmitting}
            />
          </div>
        </AppModal>
      ) : null}
    </div>
  );
}
