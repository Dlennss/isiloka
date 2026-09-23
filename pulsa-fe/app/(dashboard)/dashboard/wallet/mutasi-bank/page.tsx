"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Copy, Landmark, Plus, RefreshCcw, Search, Wallet } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type BankRow = {
  id: number;
  nama: string;
  nomor_rekening: string;
  atas_nama: string;
  saldo: number;
  aktif: boolean;
};

type ProviderOption = {
  provider: string;
};

type HistoryRow = {
  id: number;
  bank_id: number;
  bank_nama: string;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan: string;
  saldo_sebelum: number;
  saldo_sesudah: number;
  provider?: string | null;
  member_id?: number | null;
  member_nama?: string | null;
  diubah_oleh?: number | null;
  diubah_oleh_nama?: string | null;
  dibuat_pada: string;
};

const PAGE_SIZE = 20;
const POLL_MS = 10_000;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function rupiahWholeDigits(value: string): string {
  let next = (value || "").trim();
  if (!next) return "";

  const lastComma = next.lastIndexOf(",");
  const lastDot = next.lastIndexOf(".");
  const decimalIndex = Math.max(lastComma, lastDot);
  if (decimalIndex >= 0) {
    const tail = next.slice(decimalIndex + 1).replace(/\D/g, "");
    const hasMixedSeparators = lastComma >= 0 && lastDot >= 0;
    const hasDecimalTail = tail.length > 0 && tail.length <= 2;
    if (hasMixedSeparators || hasDecimalTail) {
      next = next.slice(0, decimalIndex);
    }
  }

  return next.replace(/[^\d]/g, "");
}

function formatRupiahDigits(digits: string): string {
  if (!digits) return "";
  const n = Number(digits);
  if (!Number.isFinite(n)) return "";
  return `Rp ${fmtID(n)}`;
}

function bankLabel(bank: BankRow): string {
  return `${bank.nama} - ${bank.atas_nama} - ${bank.nomor_rekening}`;
}

function isAdminFeeMutation(row: HistoryRow): boolean {
  const text = `${row.alasan} ${row.catatan}`.trim().toUpperCase();
  return text.includes("BANK_TRANSFER_ADMIN_FEE") || text.includes("ADMIN_FEE") || text.includes("ADMIN FEE") || text.includes("BIAYA ADMIN");
}

function mutationHasTarget(row: HistoryRow): boolean {
  return Boolean(row.member_id) || Boolean(row.provider?.trim());
}

export default function WalletBankMutationPage() {
  const [banks, setBanks] = useState<BankRow[]>([]);
  const [providers, setProviders] = useState<ProviderOption[]>([]);
  const [selectedBankID, setSelectedBankID] = useState("");
  const [rows, setRows] = useState<HistoryRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [bankLoading, setBankLoading] = useState(false);

  const [arah, setArah] = useState("");
  const [query, setQuery] = useState("");
  const [appliedArah, setAppliedArah] = useState("");
  const [appliedQuery, setAppliedQuery] = useState("");

  const [manualOpen, setManualOpen] = useState(false);
  const [manualDigits, setManualDigits] = useState("");
  const [manualSender, setManualSender] = useState("");
  const [manualNote, setManualNote] = useState("");
  const [manualSaving, setManualSaving] = useState(false);

  const [topupOpen, setTopupOpen] = useState(false);
  const [topupProvider, setTopupProvider] = useState("");
  const [topupDigits, setTopupDigits] = useState("");
  const [topupAdminDigits, setTopupAdminDigits] = useState("");
  const [topupNote, setTopupNote] = useState("");
  const [topupSaving, setTopupSaving] = useState(false);

  const [assignOpen, setAssignOpen] = useState(false);
  const [assignRow, setAssignRow] = useState<HistoryRow | null>(null);
  const [assignProvider, setAssignProvider] = useState("");
  const [assignProviderSearch, setAssignProviderSearch] = useState("");
  const [assignNote, setAssignNote] = useState("");
  const [assignSaving, setAssignSaving] = useState(false);

  const selectedBank = useMemo(() => banks.find((bank) => String(bank.id) === selectedBankID) || null, [banks, selectedBankID]);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const topupTotalDebit = Number(topupDigits || "0") + Number(topupAdminDigits || "0");
  const filteredAssignProviders = useMemo(() => {
    const q = assignProviderSearch.trim().toLowerCase();
    if (!q) return providers;
    return providers.filter((item) => item.provider.toLowerCase().includes(q));
  }, [assignProviderSearch, providers]);

  const loadBanks = useCallback(async (options: { showLoading?: boolean; silent?: boolean } = {}) => {
    const showLoading = options.showLoading ?? true;
    if (showLoading) setBankLoading(true);
    try {
      const r = await fetch("/api/admin/master/bank", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat rekening");
      setBanks(Array.isArray(j.items) ? j.items : []);
    } catch (err) {
      if (!options.silent) {
        await alertError(err instanceof Error ? err.message : "Gagal memuat rekening");
      }
    } finally {
      if (showLoading) setBankLoading(false);
    }
  }, []);

  const loadProviders = useCallback(async () => {
    try {
      const r = await fetch("/api/admin/provider/wallets", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const data = Array.isArray(j?.data) ? j.data : [];
      const next = data.map((item: { provider?: string }) => ({ provider: String(item.provider || "") })).filter((item: ProviderOption) => item.provider);
      setProviders(next);
      setTopupProvider((prev) => prev || next[0]?.provider || "");
    } catch {
      setProviders([]);
    }
  }, []);

  const loadHistory = useCallback(async (
    bankID = selectedBankID,
    nextOffset = offset,
    nextArah = appliedArah,
    nextQuery = appliedQuery,
    options: { showLoading?: boolean; silent?: boolean } = {},
  ) => {
    if (!bankID) {
      setRows([]);
      setTotal(0);
      setOffset(0);
      return;
    }

    const showLoading = options.showLoading ?? true;
    if (showLoading) setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("bank_id", bankID);
      qs.set("limit", String(PAGE_SIZE));
      qs.set("prioritize_unassigned", "1");
      qs.set("offset", String(nextOffset));
      if (nextArah) qs.set("arah", nextArah);
      if (nextQuery.trim()) qs.set("q", nextQuery.trim());
      const r = await fetch(`/api/admin/bank/history?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat mutasi bank");
      setRows(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
      setOffset(nextOffset);
    } catch (err) {
      if (!options.silent) {
        await alertError(err instanceof Error ? err.message : "Gagal memuat mutasi bank");
      }
    } finally {
      if (showLoading) setLoading(false);
    }
  }, [appliedArah, appliedQuery, offset, selectedBankID]);

  useEffect(() => {
    void loadBanks();
    void loadProviders();
  }, [loadBanks, loadProviders]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      if (document.visibilityState !== "visible") return;
      if (manualOpen || manualSaving || topupOpen || topupSaving || assignOpen || assignSaving) return;

      void loadBanks({ showLoading: false, silent: true });
      if (selectedBankID) {
        void loadHistory(selectedBankID, offset, appliedArah, appliedQuery, { showLoading: false, silent: true });
      }
    }, POLL_MS);

    return () => window.clearInterval(timer);
  }, [appliedArah, appliedQuery, assignOpen, assignSaving, loadBanks, loadHistory, manualOpen, manualSaving, offset, selectedBankID, topupOpen, topupSaving]);

  async function copyRefID(refID: string) {
    if (!refID) return;
    try {
      await navigator.clipboard.writeText(refID);
      await alertSuccess("RefID mutasi berhasil disalin.");
    } catch {
      await alertError("Gagal menyalin RefID mutasi.");
    }
  }

  async function submitManualMutation() {
    if (!selectedBank) return;
    const amount = Number(manualDigits || "0");
    if (amount <= 0) {
      await alertWarning("Nominal mutasi wajib lebih dari 0.");
      return;
    }
    if (!manualSender.trim()) {
      await alertWarning("Pengirim wajib diisi.");
      return;
    }

    setManualSaving(true);
    try {
      const r = await fetch("/api/admin/bank/manual-mutasi", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: selectedBank.id,
          amount,
          sender: manualSender.trim(),
          note: manualNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal menambahkan mutasi");

      await alertSuccess(`RefID: ${j.ref_id || "-"} • saldo rekening: Rp ${fmtID(Number(j.saldo || 0))}`);
      setManualOpen(false);
      setManualDigits("");
      setManualSender("");
      setManualNote("");
      setAppliedArah("credit");
      setAppliedQuery("");
      setArah("credit");
      setQuery("");
      await loadBanks();
      await loadHistory(String(selectedBank.id), 0, "credit", "");
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal menambahkan mutasi");
    } finally {
      setManualSaving(false);
    }
  }

  async function submitProviderTopup() {
    if (!selectedBank) return;
    const amount = Number(topupDigits || "0");
    const adminFee = Number(topupAdminDigits || "0");
    if (!topupProvider) {
      await alertWarning("Provider wajib dipilih.");
      return;
    }
    if (amount <= 0) {
      await alertWarning("Nominal top up provider wajib lebih dari 0.");
      return;
    }
    if (!Number.isFinite(adminFee) || adminFee < 0) {
      await alertWarning("Admin bank tidak valid.");
      return;
    }

    setTopupSaving(true);
    try {
      const r = await fetch("/api/admin/provider/wallets/deposit", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: selectedBank.id,
          provider: topupProvider,
          amount,
          admin_fee: adminFee,
          note: topupNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal top up provider");

      await alertSuccess(
        `RefID: ${j.refid || "-"} • saldo rekening: Rp ${fmtID(Number(j.bank_saldo || 0))} • saldo provider: Rp ${fmtID(Number(j.saldo_internal || 0))}`
      );
      setTopupOpen(false);
      setTopupDigits("");
      setTopupAdminDigits("");
      setTopupNote("");
      await loadBanks();
      await loadHistory(String(selectedBank.id), 0, appliedArah, appliedQuery);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal top up provider");
    } finally {
      setTopupSaving(false);
    }
  }

  function openAssignProvider(row: HistoryRow) {
    setAssignRow(row);
    setAssignProviderSearch("");
    setAssignProvider((prev) => prev || providers[0]?.provider || "");
    setAssignNote("");
    setAssignOpen(true);
  }

  async function submitAssignProvider() {
    if (!selectedBank || !assignRow) return;
    if (!assignProvider) {
      await alertWarning("Provider wajib dipilih.");
      return;
    }

    setAssignSaving(true);
    try {
      const r = await fetch("/api/admin/bank/provider-credit-from-mutasi", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          mutasi_bank_id: assignRow.id,
          provider: assignProvider,
          note: assignNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal menambahkan saldo provider");

      await alertSuccess(
        `RefID: ${j.ref_id || assignRow.ref_id || "-"} • provider ${String(j.provider || assignProvider).toUpperCase()} • saldo provider: Rp ${fmtID(Number(j.saldo_internal || 0))}`
      );
      setAssignOpen(false);
      setAssignRow(null);
      setAssignProviderSearch("");
      setAssignNote("");
      await loadBanks();
      await loadHistory(String(selectedBank.id), offset, appliedArah, appliedQuery);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal menambahkan saldo provider");
    } finally {
      setAssignSaving(false);
    }
  }

  const columns: DataTableColumn<HistoryRow>[] = [
    { id: "time", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-300", render: (row) => new Date(row.dibuat_pada).toLocaleString("id-ID") },
    { id: "ref", header: "RefID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (row) => row.ref_id || "-" },
    {
      id: "arah",
      header: "Arah",
      tdClassName: "whitespace-nowrap",
      render: (row) => (
        <span className={row.arah.toLowerCase() === "credit" ? "text-emerald-300" : "text-rose-300"}>
          {row.arah}
        </span>
      ),
    },
    { id: "amount", header: "Nominal", tdClassName: "whitespace-nowrap text-slate-100", render: (row) => `Rp ${fmtID(Number(row.jumlah || 0))}` },
    { id: "reason", header: "Alasan", tdClassName: "whitespace-nowrap text-slate-300", render: (row) => row.alasan || "-" },
    {
      id: "target",
      header: "Target",
      tdClassName: "text-slate-300",
      render: (row) =>
        isAdminFeeMutation(row)
          ? "Biaya admin"
          : row.member_id
          ? `${row.member_nama || "Member"} (#${row.member_id})`
          : row.provider
            ? row.alasan === "BANK_TRANSFER_TO_PROVIDER"
              ? `Provider ${row.provider}`
              : `Target ${row.provider}`
            : "-",
    },
    { id: "note", header: "Catatan", tdClassName: "text-slate-300", render: (row) => row.catatan || "-" },
  ];

  const actions: DataTableActions<HistoryRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (row) => {
      const hasTarget = mutationHasTarget(row);
      const canCopyRef = row.arah.toLowerCase() === "credit" && Boolean(row.ref_id) && !hasTarget;
      const canAssignProvider = row.arah.toLowerCase() === "debit" && !hasTarget && !isAdminFeeMutation(row);
      if (!canCopyRef && !canAssignProvider) return <span className="text-slate-500">-</span>;
      return (
        <div className="flex justify-end gap-2">
          {canAssignProvider ? (
            <Button type="button" variant="info" size="sm" className="h-8" onClick={() => openAssignProvider(row)}>
              <Wallet className="mr-1.5 h-3.5 w-3.5" />
              Provider
            </Button>
          ) : null}
          {canCopyRef ? (
            <Button type="button" variant="outline" size="sm" className="h-8" onClick={() => void copyRefID(row.ref_id)}>
              <Copy className="mr-1.5 h-3.5 w-3.5" />
              Copy RefID
            </Button>
          ) : null}
        </div>
      );
    },
  };

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-cyan-950/25 p-5 shadow-[0_16px_40px_-24px_rgba(34,211,238,0.28)]">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="text-lg font-semibold text-white">Mutasi Bank</div>
            <div className="mt-1 text-sm text-white/60">Pilih rekening sebelum melihat dan menambahkan mutasi.</div>
          </div>
          <div className="grid h-12 w-12 place-items-center rounded-2xl border border-cyan-400/20 bg-cyan-400/10 text-cyan-300">
            <Landmark className="h-5 w-5" />
          </div>
        </div>

        <div className="mt-4 grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto]">
          <select
            className="h-11 rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100"
            value={selectedBankID}
            onChange={(e) => {
              const next = e.target.value;
              setSelectedBankID(next);
              setArah("");
              setQuery("");
              setAppliedArah("");
              setAppliedQuery("");
              void loadHistory(next, 0, "", "");
            }}
            disabled={bankLoading}
          >
            <option value="">Pilih rekening bank</option>
            {banks.map((bank) => (
              <option key={bank.id} value={String(bank.id)}>
                {bankLabel(bank)} | saldo Rp {fmtID(Number(bank.saldo || 0))}
              </option>
            ))}
          </select>
          <Button type="button" variant="outline" className="h-11" onClick={() => void loadBanks()} disabled={bankLoading}>
            <RefreshCcw className="mr-2 h-4 w-4" />
            Refresh Rekening
          </Button>
        </div>
      </div>

      {selectedBank ? (
        <>
          <div className="grid gap-3 md:grid-cols-[1fr_auto]">
            <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4">
              <div className="text-sm text-slate-400">Rekening dipilih</div>
              <div className="mt-1 font-semibold text-white">{selectedBank.nama}</div>
              <div className="mt-1 text-sm text-slate-300">{selectedBank.atas_nama}</div>
              <div className="mt-1 font-mono text-sm text-cyan-200">{selectedBank.nomor_rekening}</div>
              <div className="mt-3 inline-flex rounded-md bg-emerald-500/15 px-2 py-1 text-sm font-medium text-emerald-300">
                Rp {fmtID(Number(selectedBank.saldo || 0))}
              </div>
            </div>
            <div className="flex flex-col justify-end gap-2 rounded-2xl border border-white/10 bg-slate-950/70 p-4 sm:flex-row md:flex-col">
              <Button type="button" variant="success" className="h-10" onClick={() => setManualOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Tambah Mutasi
              </Button>
              <Button type="button" variant="info" className="h-10" onClick={() => setTopupOpen(true)}>
                <Wallet className="mr-2 h-4 w-4" />
                Top Up Provider
              </Button>
            </div>
          </div>

          <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
            <div className="grid gap-3 lg:grid-cols-[160px_minmax(0,1fr)_auto_auto]">
              <select className="h-11 rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100" value={arah} onChange={(e) => setArah(e.target.value)}>
                <option value="">Semua Arah</option>
                <option value="credit">Credit</option>
                <option value="debit">Debit</option>
              </select>
              <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Cari ref id, nominal, alasan, catatan, member, tujuan, actor" className="h-11" />
              <Button
                type="button"
                variant="primary"
                onClick={() => {
                  setAppliedArah(arah);
                  setAppliedQuery(query);
                  void loadHistory(selectedBankID, 0, arah, query);
                }}
                disabled={loading}
              >
                <Search className="mr-2 h-4 w-4" />
                Cari
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setArah("");
                  setQuery("");
                  setAppliedArah("");
                  setAppliedQuery("");
                  void loadHistory(selectedBankID, 0, "", "");
                }}
                disabled={loading}
              >
                Reset
              </Button>
            </div>

            <DataTable<HistoryRow>
              columns={columns}
              rows={rows}
              actions={actions}
              rowKey={(row) => row.id}
              emptyText="Belum ada mutasi bank untuk rekening ini."
              minWidthClassName="min-w-220"
              showRowNumber={false}
              wrapperClassName="mt-4 overflow-auto rounded-md border border-white/10 bg-slate-950/35"
              loading={loading}
              pagination={{
                page: currentPage,
                totalPages,
                onPrev: () => void loadHistory(selectedBankID, Math.max(0, offset - PAGE_SIZE), appliedArah, appliedQuery),
                onNext: () => void loadHistory(selectedBankID, offset + PAGE_SIZE, appliedArah, appliedQuery),
                onPageChange: (page) => void loadHistory(selectedBankID, Math.max(0, (page - 1) * PAGE_SIZE), appliedArah, appliedQuery),
                disablePrev: loading || currentPage <= 1,
                disableNext: loading || currentPage >= totalPages,
              }}
            />
          </div>
        </>
      ) : (
        <div className="rounded-2xl border border-white/10 bg-slate-950/60 p-8 text-center text-sm text-slate-300">
          Pilih rekening bank terlebih dahulu.
        </div>
      )}

      <AppModal
        open={manualOpen}
        onClose={() => {
          if (manualSaving) return;
          setManualOpen(false);
        }}
        title={`Tambah Mutasi${selectedBank ? `: ${selectedBank.nama}` : ""}`}
        subtitle={selectedBank ? `${selectedBank.atas_nama} - ${selectedBank.nomor_rekening}` : "Pilih rekening lebih dulu."}
      >
        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Nominal</label>
            <Input value={manualDigits} onChange={(e) => setManualDigits(rupiahWholeDigits(e.target.value))} placeholder="Masukkan nominal" inputMode="numeric" disabled={manualSaving} />
            {manualDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(manualDigits)}</div> : null}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Pengirim</label>
            <Input value={manualSender} onChange={(e) => setManualSender(e.target.value)} placeholder="Nama pengirim mutasi" disabled={manualSaving} />
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input value={manualNote} onChange={(e) => setManualNote(e.target.value)} placeholder="Catatan opsional" disabled={manualSaving} />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setManualOpen(false)} disabled={manualSaving}>Batal</Button>
            <Button variant="success" onClick={() => void submitManualMutation()} disabled={manualSaving || !selectedBank}>
              {manualSaving ? "Menyimpan..." : "Simpan Mutasi"}
            </Button>
          </div>
        </div>
      </AppModal>

      <AppModal
        open={topupOpen}
        onClose={() => {
          if (topupSaving) return;
          setTopupOpen(false);
        }}
        title={`Top Up Provider${selectedBank ? `: ${selectedBank.nama}` : ""}`}
        subtitle={selectedBank ? `${selectedBank.atas_nama} - ${selectedBank.nomor_rekening}` : "Pilih rekening lebih dulu."}
      >
        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Provider</label>
            <select className="h-11 rounded-md border border-input bg-background px-3 text-sm" value={topupProvider} onChange={(e) => setTopupProvider(e.target.value)} disabled={topupSaving}>
              <option value="">Pilih provider</option>
              {providers.map((item) => (
                <option key={item.provider} value={item.provider}>
                  {item.provider.toUpperCase()}
                </option>
              ))}
            </select>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Nominal</label>
            <Input value={topupDigits} onChange={(e) => setTopupDigits(rupiahWholeDigits(e.target.value))} placeholder="Masukkan nominal top up" inputMode="numeric" disabled={topupSaving} />
            {topupDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(topupDigits)}</div> : null}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Admin Bank</label>
            <Input value={topupAdminDigits} onChange={(e) => setTopupAdminDigits(rupiahWholeDigits(e.target.value))} placeholder="Masukkan admin bank" inputMode="numeric" disabled={topupSaving} />
            {topupAdminDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(topupAdminDigits)}</div> : null}
          </div>
          <div className="rounded-md border border-white/10 bg-slate-950/45 p-3 text-sm text-slate-300">
            Total rekening terpotong: <span className="font-medium text-slate-100">Rp {fmtID(topupTotalDebit)}</span>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input value={topupNote} onChange={(e) => setTopupNote(e.target.value)} placeholder="Catatan transfer provider" disabled={topupSaving} />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setTopupOpen(false)} disabled={topupSaving}>Batal</Button>
            <Button variant="success" onClick={() => void submitProviderTopup()} disabled={topupSaving || !selectedBank}>
              {topupSaving ? "Memproses..." : "Proses Top Up"}
            </Button>
          </div>
        </div>
      </AppModal>

      <AppModal
        open={assignOpen}
        onClose={() => {
          if (assignSaving) return;
          setAssignOpen(false);
          setAssignRow(null);
        }}
        title="Tambah Saldo Provider"
        subtitle={assignRow ? `${assignRow.ref_id || "-"} - Rp ${fmtID(Number(assignRow.jumlah || 0))}` : "Pilih provider untuk mutasi debit ini."}
      >
        <div className="grid gap-3">
          {assignRow ? (
            <div className="rounded-md border border-white/10 bg-slate-950/45 p-3 text-sm text-slate-300">
              <div className="grid gap-1">
                <div>
                  Bank: <span className="font-medium text-slate-100">{assignRow.bank_nama}</span>
                </div>
                <div>
                  Nominal keluar: <span className="font-medium text-rose-200">Rp {fmtID(Number(assignRow.jumlah || 0))}</span>
                </div>
                <div className="break-words">
                  Catatan: <span className="text-slate-200">{assignRow.catatan || "-"}</span>
                </div>
              </div>
            </div>
          ) : null}
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Cari Provider</label>
            <Input value={assignProviderSearch} onChange={(e) => setAssignProviderSearch(e.target.value)} placeholder="Cari nama provider" disabled={assignSaving} />
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Provider Tujuan</label>
            <select className="h-11 rounded-md border border-input bg-background px-3 text-sm" value={assignProvider} onChange={(e) => setAssignProvider(e.target.value)} disabled={assignSaving}>
              <option value="">Pilih provider</option>
              {filteredAssignProviders.map((item) => (
                <option key={item.provider} value={item.provider}>
                  {item.provider.toUpperCase()}
                </option>
              ))}
            </select>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input value={assignNote} onChange={(e) => setAssignNote(e.target.value)} placeholder="Catatan opsional" disabled={assignSaving} />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => {
                setAssignOpen(false);
                setAssignRow(null);
              }}
              disabled={assignSaving}
            >
              Batal
            </Button>
            <Button variant="success" onClick={() => void submitAssignProvider()} disabled={assignSaving || !assignRow || !assignProvider}>
              {assignSaving ? "Memproses..." : "Tambah Saldo Provider"}
            </Button>
          </div>
        </div>
      </AppModal>
    </div>
  );
}
