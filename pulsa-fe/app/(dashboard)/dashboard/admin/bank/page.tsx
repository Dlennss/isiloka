"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { Copy, Eye, Landmark, Pencil, Plus, Repeat, Wallet } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

type BankRow = {
  id: number;
  nama: string;
  nomor_rekening: string;
  atas_nama: string;
  saldo: number;
  aktif: boolean;
  dibuat_pada?: string;
  diubah_pada?: string;
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
  target_ref_id?: string | null;
  member_id?: number | null;
  member_nama?: string | null;
  diubah_oleh?: number | null;
  diubah_oleh_nama?: string | null;
  dibuat_pada: string;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function onlyDigits(s: string): string {
  return (s || "").replace(/[^\d]/g, "");
}

function formatRupiahDigits(digits: string): string {
  const n = Number(digits || "0");
  if (!digits || !Number.isFinite(n)) return "";
  return `Rp ${fmtID(n)}`;
}

const BCA_OPERATIONAL_ACCOUNT = "3432738881";
const DEFAULT_INTERNAL_TRANSFER_FEE = "6500";

function normalizeBankSortText(value: string): string {
  return (value || "").trim().toUpperCase().replace(/\s+/g, " ");
}

function isPtPulsaAccount(item: BankRow): boolean {
  const accountName = normalizeBankSortText(item.atas_nama);
  return accountName === "PT PULSA" || accountName.startsWith("PT PULSA ");
}

function isSystemQrisBank(item: BankRow): boolean {
  const bankName = normalizeBankSortText(item.nama);
  const accountName = normalizeBankSortText(item.atas_nama);
  return bankName === "SYSTEM QRIS" || accountName === "SYSTEM QRIS";
}

function isBCAOperationalBank(item: BankRow): boolean {
  return onlyDigits(item.nomor_rekening) === BCA_OPERATIONAL_ACCOUNT;
}

function canTransferToBCAOperational(item: BankRow): boolean {
  return Boolean(item.aktif) && !isBCAOperationalBank(item);
}

function compareBankDisplayOrder(a: BankRow, b: BankRow): number {
  const qrisOrder = Number(isSystemQrisBank(a)) - Number(isSystemQrisBank(b));
  if (qrisOrder !== 0) return qrisOrder;

  const ptPulsaOrder = Number(isPtPulsaAccount(b)) - Number(isPtPulsaAccount(a));
  if (ptPulsaOrder !== 0) return ptPulsaOrder;

  const activeOrder = Number(b.aktif) - Number(a.aktif);
  if (activeOrder !== 0) return activeOrder;

  const bankNameOrder = normalizeBankSortText(a.nama).localeCompare(normalizeBankSortText(b.nama), "id");
  if (bankNameOrder !== 0) return bankNameOrder;

  const accountNameOrder = normalizeBankSortText(a.atas_nama).localeCompare(normalizeBankSortText(b.atas_nama), "id");
  if (accountNameOrder !== 0) return accountNameOrder;

  return a.id - b.id;
}

export default function AdminBankPage() {
  const pathname = usePathname();
  const isWalletMode = pathname.startsWith("/dashboard/wallet");
  const isOperatorMode = pathname.startsWith("/dashboard/operator");
  const [items, setItems] = useState<BankRow[]>([]);
  const [providers, setProviders] = useState<ProviderOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);

  const [openForm, setOpenForm] = useState(false);
  const [editing, setEditing] = useState<BankRow | null>(null);
  const [nama, setNama] = useState("");
  const [nomorRekening, setNomorRekening] = useState("");
  const [atasNama, setAtasNama] = useState("");
  const [saldoDigits, setSaldoDigits] = useState("");
  const [aktif, setAktif] = useState(true);
  const [saving, setSaving] = useState(false);

  const [adjustOpen, setAdjustOpen] = useState(false);
  const [adjustBank, setAdjustBank] = useState<BankRow | null>(null);
  const [adjustDirection, setAdjustDirection] = useState<"credit" | "debit">("credit");
  const [adjustDigits, setAdjustDigits] = useState("");
  const [adjustNote, setAdjustNote] = useState("");
  const [adjustSaving, setAdjustSaving] = useState(false);

  const [transferOpen, setTransferOpen] = useState(false);
  const [transferBank, setTransferBank] = useState<BankRow | null>(null);
  const [transferTujuan, setTransferTujuan] = useState("");
  const [transferDigits, setTransferDigits] = useState("");
  const [transferAdminFeeDigits, setTransferAdminFeeDigits] = useState("");
  const [transferNote, setTransferNote] = useState("");
  const [transferSaving, setTransferSaving] = useState(false);

  const [manualOpen, setManualOpen] = useState(false);
  const [manualBank, setManualBank] = useState<BankRow | null>(null);
  const [manualDigits, setManualDigits] = useState("");
  const [manualSender, setManualSender] = useState("");
  const [manualNote, setManualNote] = useState("");
  const [manualSaving, setManualSaving] = useState(false);

  const [providerTopupOpen, setProviderTopupOpen] = useState(false);
  const [providerTopupBank, setProviderTopupBank] = useState<BankRow | null>(null);
  const [providerTopupProvider, setProviderTopupProvider] = useState("");
  const [providerTopupDigits, setProviderTopupDigits] = useState("");
  const [providerTopupAdminDigits, setProviderTopupAdminDigits] = useState("");
  const [providerTopupNote, setProviderTopupNote] = useState("");
  const [providerTopupSaving, setProviderTopupSaving] = useState(false);

  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyBankID, setHistoryBankID] = useState<number>(0);
  const [historyItems, setHistoryItems] = useState<HistoryRow[]>([]);
  const [historyArah, setHistoryArah] = useState("");
  const [historyQuery, setHistoryQuery] = useState("");
  const [historyAppliedArah, setHistoryAppliedArah] = useState("");
  const [historyAppliedQuery, setHistoryAppliedQuery] = useState("");
  const [historyLimit] = useState(10);
  const [historyOffset, setHistoryOffset] = useState(0);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [actionMenu, setActionMenu] = useState<{ key: string; top: number; left: number; item: BankRow } | null>(null);

  async function loadBanks() {
    setLoading(true);
    try {
      const banksRes = await fetch("/api/admin/master/bank", { headers: authHeader(), cache: "no-store" });
      const banksJson = await banksRes.json().catch(() => ({}));

      if (!banksRes.ok || !banksJson.ok) {
        await alertError(banksJson.error || "Gagal memuat bank");
        setItems([]);
      } else {
        setItems(Array.isArray(banksJson.items) ? banksJson.items : []);
      }
    } finally {
      setLoading(false);
    }
  }

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/provider/wallets", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const rows = Array.isArray(j?.data) ? j.data : [];
      const next = rows.map((item: { provider?: string }) => ({ provider: String(item.provider || "") })).filter((item: ProviderOption) => item.provider);
      setProviders(next);
      if (!providerTopupProvider && next[0]?.provider) {
        setProviderTopupProvider(next[0].provider);
      }
    } catch {
      setProviders([]);
    }
  }

  async function loadHistory(bankID: number, offset: number = 0, nextArah: string = historyAppliedArah, nextQuery: string = historyAppliedQuery) {
    if (bankID <= 0) {
      setHistoryItems([]);
      setHistoryTotal(0);
      setHistoryOffset(0);
      setHistoryBankID(0);
      return;
    }
    setHistoryLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("bank_id", String(bankID));
      if (nextArah) qs.set("arah", nextArah);
      if (nextQuery.trim()) qs.set("q", nextQuery.trim());
      qs.set("limit", String(historyLimit));
      qs.set("offset", String(offset));

      const r = await fetch(`/api/admin/bank/history?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat histori bank");
        return;
      }

      setHistoryBankID(bankID);
      setHistoryOpen(true);
      setHistoryItems(Array.isArray(j.items) ? j.items : []);
      setHistoryTotal(Number(j.total || 0));
      setHistoryOffset(offset);
    } finally {
      setHistoryLoading(false);
    }
  }

  useEffect(() => {
    void loadBanks();
    void loadProviders();
  }, []);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-action-dropdown]")) return;
      setActionMenu(null);
    }

    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") setActionMenu(null);
    }

    document.addEventListener("mousedown", handleOutsideClick);
    document.addEventListener("keydown", handleEsc);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleEsc);
    };
  }, []);

  const totalSaldoBank = useMemo(() => items.reduce((sum, item) => sum + Number(item.saldo || 0), 0), [items]);
  const sortedItems = useMemo(() => [...items].sort(compareBankDisplayOrder), [items]);
  const historyTotalPages = Math.max(1, Math.ceil(historyTotal / historyLimit));
  const historyCurrentPage = Math.floor(historyOffset / historyLimit) + 1;
  const selectedHistoryBank = items.find((item) => item.id === historyBankID) || null;
  const providerTopupAmount = Number(providerTopupDigits || "0");
  const providerTopupAdminFee = Number(providerTopupAdminDigits || "0");
  const providerTopupBankDebit = providerTopupAmount + providerTopupAdminFee;

  function openActionMenu(item: BankRow, element: HTMLElement) {
    const rect = element.getBoundingClientRect();
    const menuWidth = 208;
    const viewportPadding = 12;
    const preferredLeft = rect.right - menuWidth;
    const left = Math.max(viewportPadding, Math.min(preferredLeft, window.innerWidth - menuWidth - viewportPadding));
    const top = Math.min(rect.bottom + 8, window.innerHeight - 320);
    setActionMenu({ key: String(item.id), top, left, item });
  }

  async function copyBankInfo(item: BankRow) {
    const text = [
      `Bank: ${item.nama || "-"}`,
      `Nama Rekening: ${item.atas_nama || "-"}`,
      `Nomor Rekening: ${item.nomor_rekening || "-"}`,
    ].join("\n");

    try {
      await navigator.clipboard.writeText(text);
      await alertSuccess("Data rekening bank berhasil disalin.");
    } catch {
      await alertError("Gagal menyalin data rekening bank.");
    }
  }

  async function copyRefID(refID: string) {
    if (!refID) return;
    try {
      await navigator.clipboard.writeText(refID);
      await alertSuccess("RefID mutasi berhasil disalin.");
    } catch {
      await alertError("Gagal menyalin RefID mutasi.");
    }
  }

  function openManualMutation(item: BankRow) {
    setManualBank(item);
    setManualDigits("");
    setManualSender("");
    setManualNote("");
    setManualOpen(true);
  }

  function openProviderTopup(item: BankRow) {
    if (!providers.length) void loadProviders();
    setProviderTopupBank(item);
    setProviderTopupProvider(providerTopupProvider || providers[0]?.provider || "");
    setProviderTopupDigits("");
    setProviderTopupAdminDigits("");
    setProviderTopupNote("");
    setProviderTopupOpen(true);
  }

  function openCreate() {
    setEditing(null);
    setNama("");
    setNomorRekening("");
    setAtasNama("");
    setSaldoDigits("");
    setAktif(true);
    setOpenForm(true);
  }

  function openEdit(item: BankRow) {
    setEditing(item);
    setNama(item.nama || "");
    setNomorRekening(item.nomor_rekening || "");
    setAtasNama(item.atas_nama || "");
    setSaldoDigits(String(item.saldo || 0));
    setAktif(Boolean(item.aktif));
    setOpenForm(true);
  }

  async function toggleBankActive(item: BankRow) {
    const nextAktif = !item.aktif;
    const confirmed = await alertConfirm({
      title: nextAktif ? "Aktifkan Bank" : "Nonaktifkan Bank",
      text: `Bank ${item.nama} akan ${nextAktif ? "diaktifkan" : "dinonaktifkan"}. Lanjutkan?`,
      confirmButtonText: nextAktif ? "Aktifkan" : "Nonaktifkan",
    });
    if (!confirmed) return;

    const r = await fetch(`/api/admin/master/bank/${item.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", ...authHeader() },
      body: JSON.stringify({
        nama: item.nama,
        nomor_rekening: item.nomor_rekening,
        atas_nama: item.atas_nama,
        saldo: item.saldo,
        aktif: nextAktif,
      }),
    });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) {
      await alertError(j.error || `Gagal ${nextAktif ? "mengaktifkan" : "menonaktifkan"} bank`);
      return;
    }

    await alertSuccess(`Bank berhasil ${nextAktif ? "diaktifkan" : "dinonaktifkan"}.`);
    await loadBanks();
  }

  async function submitForm() {
    const saldo = Number(saldoDigits || "0");
    const payload = {
      nama: nama.trim(),
      nomor_rekening: nomorRekening.trim(),
      atas_nama: atasNama.trim(),
      saldo,
      aktif,
    };

    if (!payload.nama) {
      await alertWarning("Jenis bank wajib diisi.");
      return;
    }
    if (!payload.atas_nama) {
      await alertWarning("Nama rekening wajib diisi.");
      return;
    }
    if (!payload.nomor_rekening) {
      await alertWarning("Nomor rekening wajib diisi.");
      return;
    }
    if (!Number.isFinite(saldo) || saldo < 0) {
      await alertWarning("Saldo harus 0 atau lebih.");
      return;
    }

    setSaving(true);
    try {
      const isEdit = Boolean(editing?.id);
      const url = isEdit ? `/api/admin/master/bank/${editing?.id}` : "/api/admin/master/bank";
      const method = isEdit ? "PUT" : "POST";

      const r = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(payload),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menyimpan data bank");
        return;
      }

      await alertSuccess(isEdit ? "Bank berhasil diperbarui." : "Bank berhasil ditambahkan.");
      setOpenForm(false);
      await loadBanks();
    } finally {
      setSaving(false);
    }
  }

  async function submitAdjust() {
    if (!adjustBank) return;
    const amount = Number(adjustDigits || "0");
    if (amount <= 0) {
      await alertWarning("Nominal koreksi wajib lebih dari 0.");
      return;
    }
    if (!adjustNote.trim()) {
      await alertWarning("Catatan koreksi wajib diisi.");
      return;
    }

    setAdjustSaving(true);
    try {
      const r = await fetch("/api/admin/bank/adjust", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: adjustBank.id,
          amount,
          direction: adjustDirection,
          note: adjustNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal koreksi saldo bank");
        return;
      }

      await alertSuccess(`RefID: ${j.ref_id || "-"} • saldo bank: Rp ${fmtID(Number(j.saldo || 0))}`);
      setAdjustOpen(false);
      setAdjustDigits("");
      setAdjustNote("");
      await loadBanks();
      await loadHistory(adjustBank.id, 0);
    } finally {
      setAdjustSaving(false);
    }
  }

  async function submitManualMutation() {
    if (!manualBank) return;
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
          bank_id: manualBank.id,
          amount,
          sender: manualSender.trim(),
          note: manualNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal menambahkan mutasi rekening");
        return;
      }

      await alertSuccess(`RefID: ${j.ref_id || "-"} • saldo rekening: Rp ${fmtID(Number(j.saldo || 0))}`);
      setManualOpen(false);
      setManualDigits("");
      setManualSender("");
      setManualNote("");
      await loadBanks();
      setHistoryAppliedArah("credit");
      setHistoryAppliedQuery("");
      await loadHistory(manualBank.id, 0, "credit", "");
    } finally {
      setManualSaving(false);
    }
  }

  async function submitProviderTopup() {
    if (!providerTopupBank) return;
    const amount = Number(providerTopupDigits || "0");
    const adminFee = Number(providerTopupAdminDigits || "0");
    if (!providerTopupProvider) {
      await alertWarning("Provider wajib dipilih.");
      return;
    }
    if (amount <= 0) {
      await alertWarning("Nominal top up provider wajib lebih dari 0.");
      return;
    }
    if (adminFee < 0 || !Number.isFinite(adminFee)) {
      await alertWarning("Admin bank tidak valid.");
      return;
    }

    setProviderTopupSaving(true);
    try {
      const r = await fetch("/api/admin/provider/wallets/deposit", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          bank_id: providerTopupBank.id,
          provider: providerTopupProvider,
          amount,
          admin_fee: adminFee,
          note: providerTopupNote.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal top up provider");
        return;
      }

      await alertSuccess(
        `RefID: ${j.refid || "-"} • saldo rekening: Rp ${fmtID(Number(j.bank_saldo || 0))} • saldo provider: Rp ${fmtID(Number(j.saldo_internal || 0))}`
      );
      setProviderTopupOpen(false);
      setProviderTopupDigits("");
      setProviderTopupAdminDigits("");
      setProviderTopupNote("");
      await loadBanks();
      await loadHistory(providerTopupBank.id, 0);
    } finally {
      setProviderTopupSaving(false);
    }
  }

  async function submitTransfer() {
    if (!transferBank) return;
    const amount = Number(transferDigits || "0");
    const adminFee = Number(transferAdminFeeDigits || "0");
    if (amount <= 0) {
      await alertWarning("Nominal transfer wajib lebih dari 0.");
      return;
    }
    if (isWalletMode && !canTransferToBCAOperational(transferBank)) {
      await alertWarning("Bank sumber tidak bisa BCA OPERASIONAL.");
      return;
    }
    if (adminFee < 0 || !Number.isFinite(adminFee)) {
      await alertWarning("Biaya admin tidak valid.");
      return;
    }
    if (!isWalletMode && !transferTujuan.trim()) {
      await alertWarning("Tujuan transfer wajib diisi.");
      return;
    }
    if (!transferNote.trim()) {
      await alertWarning("Catatan transfer wajib diisi.");
      return;
    }

    setTransferSaving(true);
    try {
      const r = await fetch(isWalletMode ? "/api/admin/bank/internal-transfer" : "/api/admin/bank/transfer", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify(
          isWalletMode
            ? {
                bank_id: transferBank.id,
                amount,
                admin_fee: adminFee,
                note: transferNote.trim(),
              }
            : {
                bank_id: transferBank.id,
                tujuan: transferTujuan.trim(),
                amount,
                note: transferNote.trim(),
              }
        ),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal transfer saldo bank");
        return;
      }

      if (isWalletMode) {
        await alertSuccess(
          `RefID: ${j.ref_id || "-"} • BCA OPERASIONAL: Rp ${fmtID(Number(j.destination_bank_saldo || 0))} • saldo sumber: Rp ${fmtID(Number(j.source_bank_saldo || 0))}`
        );
      } else {
        await alertSuccess(
          `RefID: ${j.ref_id || "-"} • tujuan: ${j.tujuan || "-"} • saldo bank: Rp ${fmtID(Number(j.saldo || 0))}`
        );
      }
      setTransferOpen(false);
      setTransferDigits("");
      setTransferAdminFeeDigits("");
      setTransferNote("");
      setTransferTujuan("");
      await loadBanks();
      await loadHistory(transferBank.id, 0);
    } finally {
      setTransferSaving(false);
    }
  }

  const bankColumns: DataTableColumn<BankRow>[] = [
    {
      id: "jenis_bank",
      header: "Jenis Bank",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (item) => item.nama || "-",
    },
    {
      id: "nama_rekening",
      header: "Nama Rekening",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (item) => item.atas_nama || "-",
    },
    {
      id: "rekening",
      header: "Nomor Rekening",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (item) => item.nomor_rekening || "-",
    },
  ];
  if (!isOperatorMode) {
    bankColumns.push({
      id: "saldo",
      header: "Saldo",
      tdClassName: "whitespace-nowrap overflow-visible",
      render: (item) => (
        <span className="rounded-md border border-emerald-300 bg-emerald-100 px-2 py-1 font-bold text-emerald-900">
          Rp {fmtID(Number(item.saldo || 0))}
        </span>
      ),
    });
    bankColumns.push({
      id: "aktif",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (item) =>
        item.aktif ? (
          <span className="rounded-full border border-emerald-400 bg-emerald-100 px-2 py-0.5 text-xs font-bold text-emerald-900">
            Aktif
          </span>
        ) : (
          <span className="rounded-full border border-rose-400 bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-900">
            Nonaktif
          </span>
        ),
    });
  }

  const bankActions: DataTableActions<BankRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (item) => {
      const actionKey = String(item.id);
      const isOpen = actionMenu?.key === actionKey;

      return (
        <div className="inline-flex" data-action-dropdown>
          <Button
            type="button"
            variant="info"
            size="sm"
            className="h-8"
            onClick={(e) => {
              const target = e.currentTarget as HTMLElement;
              if (isOpen) {
                setActionMenu(null);
                return;
              }
              openActionMenu(item, target);
            }}
            aria-label="Aksi"
          >
            <Eye className="h-4 w-4" />
            Aksi
          </Button>
        </div>
      );
    },
  };


  const historyColumns: DataTableColumn<HistoryRow>[] = [
    { id: "dibuat", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-300", render: (item) => new Date(item.dibuat_pada).toLocaleString("id-ID") },
    { id: "ref_id", header: "RefID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (item) => item.ref_id },
    {
      id: "arah",
      header: "Arah",
      tdClassName: "whitespace-nowrap",
      render: (item) => (
        <span className={item.arah.toLowerCase() === "credit" ? "font-semibold text-emerald-800" : "font-semibold text-rose-800"}>
          {item.arah}
        </span>
      ),
    },
    { id: "jumlah", header: "Jumlah", tdClassName: "whitespace-nowrap text-slate-100", render: (item) => `Rp ${fmtID(Number(item.jumlah || 0))}` },
    { id: "alasan", header: "Alasan", tdClassName: "whitespace-nowrap text-slate-300", render: (item) => item.alasan },
    {
      id: "target",
      header: "Target",
      tdClassName: "text-slate-300",
      render: (item) => {
        const targetRefID = item.target_ref_id?.trim() || "";
        if (targetRefID) {
          return <span className="font-mono text-xs text-slate-200">{targetRefID}</span>;
        }
        if (item.member_id && item.ref_id) {
          return <span className="font-mono text-xs text-slate-200">{item.ref_id}</span>;
        }
        if (item.provider) {
          if (item.alasan === "BANK_TRANSFER_OUT") return `Tujuan ${item.provider}`;
          if (item.alasan === "BANK_TRANSFER_TO_PROVIDER") return `Provider ${item.provider}`;
          return `Target ${item.provider}`;
        }
        return "-";
      },
    },
    {
      id: "actor",
      header: "Dilakukan Oleh",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (item) => (item.diubah_oleh_nama && item.diubah_oleh ? `${item.diubah_oleh_nama} (#${item.diubah_oleh})` : item.diubah_oleh ? `#${item.diubah_oleh}` : "-"),
    },
    {
      id: "catatan",
      header: "Catatan",
      tdClassName: "text-slate-300",
      render: (item) => item.catatan || "-",
    },
  ];

  const historyActions: DataTableActions<HistoryRow> | undefined = isWalletMode
    ? {
        header: "Aksi",
        align: "right",
        tdClassName: "whitespace-nowrap",
        render: (item) => {
          const canCopyRef = item.arah.toLowerCase() === "credit" && Boolean(item.ref_id) && !item.member_id && !item.provider;
          if (!canCopyRef) return <span className="text-slate-500">-</span>;
          return (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="h-8"
              onClick={() => void copyRefID(item.ref_id)}
            >
              <Copy className="mr-1.5 h-3.5 w-3.5" />
              Copy RefID
            </Button>
          );
        },
      }
    : undefined;

  return (
    <div className="space-y-4 p-2">
      <div className="grid gap-3 md:grid-cols-[1.4fr_1fr]">
        <div className="rounded-2xl border border-emerald-900 bg-[#13464b] p-5 shadow-[0_16px_40px_-24px_rgba(6,78,59,0.35)]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <div className="text-lg font-semibold text-white">Bank Admin</div>
              <div className="mt-1 text-sm font-medium text-emerald-100">
                {isWalletMode
                  ? "Lihat data bank operasional, ubah status aktif/nonaktif, dan transfer saldo ke BCA OPERASIONAL."
                  : isOperatorMode
                    ? "Operator transaksi hanya bisa melihat data rekening bank dan menyalinnya untuk kebutuhan operasional."
                  : "Kelola master bank, koreksi saldo bank, dan transfer keluar dari saldo bank."}
              </div>
            </div>
            <div className="grid h-12 w-12 place-items-center rounded-2xl border border-white/25 bg-white/10 text-lime-300">
              <Landmark className="h-5 w-5" />
            </div>
          </div>

          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            <div className="rounded-2xl border border-white/25 bg-white/10 px-4 py-4">
              <div className="text-sm font-semibold text-emerald-100">{isOperatorMode ? "Data Rekening" : "Total Saldo Bank"}</div>
              <div className="mt-1 text-2xl font-black text-white">{isOperatorMode ? "Terbatas" : `Rp ${fmtID(totalSaldoBank)}`}</div>
            </div>
            <div className="rounded-2xl border border-lime-200 bg-lime-300 px-4 py-4">
              <div className="text-sm font-semibold text-emerald-950">Jumlah Bank</div>
              <div className="mt-1 text-2xl font-black text-emerald-950">{items.length}</div>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border border-emerald-200 bg-white p-5 shadow-[0_16px_40px_-24px_rgba(6,78,59,0.18)]">
          <div className="text-base font-bold text-emerald-950">{isWalletMode ? "Akses Operator Wallet" : isOperatorMode ? "Akses Operator Transaksi" : "Aksi Cepat"}</div>
          <div className="mt-1 text-sm font-medium text-emerald-900/75">
            {isWalletMode
              ? "Operator wallet bisa transfer saldo dari bank internal lain ke BCA OPERASIONAL 3432738881."
              : isOperatorMode
                ? "Operator transaksi hanya bisa melihat data bank yang terbatas dan menyalinnya."
                : "Tambahkan bank baru atau pilih bank yang ada untuk operasi saldo."}
          </div>
          {!isWalletMode && !isOperatorMode ? (
            <div className="mt-4">
              <Button
                type="button"
                variant="primary"
                className="h-11 rounded-full px-4"
                onClick={openCreate}
              >
                <Plus className="mr-2 h-4 w-4" />
                Tambah Bank
              </Button>
            </div>
          ) : null}
        </div>
      </div>

      <div className="rounded-2xl border border-white/10 bg-card/80 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
        <div className="border-b border-white/10 bg-linear-to-r from-white/10 via-white/5 to-transparent px-4 py-3">
          <div className="text-lg font-semibold">Daftar Bank</div>
          <div className="text-xs font-medium text-emerald-900/70">
            {isWalletMode
              ? "Gunakan aksi untuk aktif/nonaktifkan bank, transfer ke BCA OPERASIONAL, atau melihat histori."
              : isOperatorMode
                ? "Operator transaksi hanya bisa melihat jenis bank, nama rekening, nomor rekening, dan menyalin data."
                : "Gunakan aksi untuk edit, koreksi saldo, transfer keluar, atau melihat histori."}
          </div>
        </div>
        <div className="px-4 pb-4 pt-4">
          <DataTable<BankRow>
            columns={bankColumns}
            rows={sortedItems}
            actions={bankActions}
            rowKey={(item) => String(item.id)}
            emptyText={loading ? "Memuat..." : "Belum ada bank."}
            minWidthClassName="min-w-200"
            showRowNumber={false}
            wrapperClassName="overflow-x-auto overflow-y-visible rounded-md border border-white/10 bg-slate-950/35"
          />
        </div>
      </div>

      {actionMenu ? (
        <div
          className="fixed inset-0 z-40"
          onMouseDown={() => setActionMenu(null)}
        >
          <div
            className="fixed z-50 w-52 rounded-xl border border-white/12 bg-slate-900/95 p-1.5 shadow-2xl backdrop-blur"
            style={{ top: actionMenu.top, left: actionMenu.left }}
            onMouseDown={(e) => e.stopPropagation()}
            data-action-dropdown
          >
            {isOperatorMode ? (
              <button
                type="button"
                className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
                onClick={() => {
                  const item = actionMenu.item;
                  setActionMenu(null);
                  void copyBankInfo(item);
                }}
              >
                <Copy className="h-3.5 w-3.5" />
                Copy
              </button>
            ) : null}

            {!isOperatorMode ? (
              <>
                {isWalletMode ? (
                  <>
                    <button
                      type="button"
                      className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition ${
                        actionMenu.item.aktif
                          ? "text-rose-800 hover:bg-rose-50"
                          : "text-emerald-800 hover:bg-emerald-50"
                      }`}
                      onClick={() => {
                        const item = actionMenu.item;
                        setActionMenu(null);
                        void toggleBankActive(item);
                      }}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      {actionMenu.item.aktif ? "Nonaktifkan" : "Aktifkan"}
                    </button>
                    {canTransferToBCAOperational(actionMenu.item) ? (
                      <button
                        type="button"
                        className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50"
                        onClick={() => {
                          const item = actionMenu.item;
                          setActionMenu(null);
                          setTransferBank(item);
                          setTransferTujuan(BCA_OPERATIONAL_ACCOUNT);
                          setTransferDigits("");
                          setTransferAdminFeeDigits(DEFAULT_INTERNAL_TRANSFER_FEE);
                          setTransferNote(`Transfer ke BCA OPERASIONAL ${BCA_OPERATIONAL_ACCOUNT}`);
                          setTransferOpen(true);
                        }}
                      >
                        <Repeat className="h-3.5 w-3.5" />
                        Transfer BCA OP
                      </button>
                    ) : null}
                    <button
                      type="button"
                      className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
                      onClick={() => {
                        const item = actionMenu.item;
                        setActionMenu(null);
                        void loadHistory(item.id, 0);
                      }}
                    >
                      <Eye className="h-3.5 w-3.5" />
                      Histori
                    </button>
                  </>
                ) : (
                  <button
                    type="button"
                    className="flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
                    onClick={() => {
                      const item = actionMenu.item;
                      setActionMenu(null);
                      openEdit(item);
                    }}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    Edit
                  </button>
                )}
                {!isWalletMode ? (
                  <>
                    <button
                      type="button"
                      className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm font-semibold text-amber-800 transition hover:bg-amber-50"
                      onClick={() => {
                        const item = actionMenu.item;
                        setActionMenu(null);
                        setAdjustBank(item);
                        setAdjustDirection("credit");
                        setAdjustDigits("");
                        setAdjustNote("");
                        setAdjustOpen(true);
                      }}
                    >
                      <Wallet className="h-3.5 w-3.5" />
                      Koreksi
                    </button>
                    <button
                      type="button"
                      className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50"
                      onClick={() => {
                        const item = actionMenu.item;
                        setActionMenu(null);
                        setTransferBank(item);
                        setTransferTujuan("");
                        setTransferDigits("");
                        setTransferAdminFeeDigits("");
                        setTransferNote("");
                        setTransferOpen(true);
                      }}
                    >
                      <Repeat className="h-3.5 w-3.5" />
                      Transfer
                    </button>
                    <button
                      type="button"
                      className="mt-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm text-slate-100 transition hover:bg-white/10"
                      onClick={() => {
                        const item = actionMenu.item;
                        setActionMenu(null);
                        void loadHistory(item.id, 0);
                      }}
                    >
                      <Eye className="h-3.5 w-3.5" />
                      Histori
                    </button>
                  </>
                ) : null}
              </>
            ) : null}
          </div>
        </div>
      ) : null}

      {!isOperatorMode ? (
      <AppModal
        open={historyOpen}
        onClose={() => {
          setHistoryOpen(false);
          setHistoryBankID(0);
          setHistoryItems([]);
          setHistoryTotal(0);
          setHistoryOffset(0);
          setHistoryArah("");
          setHistoryQuery("");
          setHistoryAppliedArah("");
          setHistoryAppliedQuery("");
        }}
        title={`Histori Mutasi Bank${selectedHistoryBank ? `: ${selectedHistoryBank.nama}` : ""}`}
        subtitle="Lihat riwayat mutasi saldo bank, filter arah transaksi, dan telusuri referensi mutasinya."
        maxWidthClassName="max-w-6xl"
      >
        <div className="grid gap-4">
          <div className="grid gap-3 lg:grid-cols-[180px_minmax(0,1fr)_auto_auto_auto_auto]">
            <select
              className="h-11 rounded-md border border-white/10 bg-slate-900/80 px-3 text-sm text-slate-100"
              value={historyArah}
              onChange={(e) => setHistoryArah(e.target.value)}
            >
              <option value="">Semua Arah</option>
              <option value="credit">Credit</option>
              <option value="debit">Debit</option>
            </select>
            <Input
              value={historyQuery}
              onChange={(e) => setHistoryQuery(e.target.value)}
              placeholder="Cari ref id, alasan, catatan, member, tujuan, actor"
              className="h-11"
            />
            <Button
              type="button"
              variant="primary"
              onClick={() => {
                setHistoryAppliedArah(historyArah);
                setHistoryAppliedQuery(historyQuery);
                void loadHistory(historyBankID, 0, historyArah, historyQuery);
              }}
              disabled={!historyBankID}
            >
              Cari
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setHistoryArah("");
                setHistoryQuery("");
                setHistoryAppliedArah("");
                setHistoryAppliedQuery("");
                void loadHistory(historyBankID, 0, "", "");
              }}
              disabled={!historyBankID}
            >
              Reset
            </Button>
          </div>

          <DataTable<HistoryRow>
            columns={historyColumns}
            rows={historyItems}
            actions={historyActions}
            rowKey={(item) => String(item.id)}
            emptyText="Belum ada histori bank."
            minWidthClassName="min-w-220"
            showRowNumber={false}
            wrapperClassName="overflow-auto rounded-md border border-white/10 bg-slate-950/35"
            loading={historyLoading}
            pagination={{
              page: historyCurrentPage,
              totalPages: historyTotalPages,
              onPrev: () => void loadHistory(historyBankID, Math.max(0, historyOffset - historyLimit)),
              onNext: () => void loadHistory(historyBankID, historyOffset + historyLimit),
              onPageChange: (page) => void loadHistory(historyBankID, Math.max(0, (page - 1) * historyLimit)),
              disablePrev: historyLoading || historyCurrentPage <= 1,
              disableNext: historyLoading || historyCurrentPage >= historyTotalPages,
            }}
          />
        </div>
      </AppModal>
      ) : null}

      {isWalletMode ? (
      <AppModal
        open={manualOpen}
        onClose={() => {
          if (manualSaving) return;
          setManualOpen(false);
        }}
        title={`Tambah Mutasi${manualBank ? `: ${manualBank.nama}` : ""}`}
        subtitle={manualBank ? `${manualBank.atas_nama} - ${manualBank.nomor_rekening}` : "Input mutasi uang masuk rekening."}
      >
        <div className="grid gap-3">
          <div className="rounded-xl border border-white/10 bg-white/5 p-3 text-sm">
            <div className="font-medium text-white">{manualBank?.nama || "-"}</div>
            <div className="font-medium text-emerald-950">{manualBank?.atas_nama || "-"}</div>
            <div className="font-mono text-emerald-700">{manualBank?.nomor_rekening || "-"}</div>
            <div className="font-semibold text-emerald-700">Saldo: Rp {fmtID(Number(manualBank?.saldo || 0))}</div>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Nominal</label>
            <Input
              value={manualDigits}
              onChange={(e) => setManualDigits(onlyDigits(e.target.value))}
              placeholder="Masukkan nominal"
              inputMode="numeric"
              disabled={manualSaving}
            />
            {manualDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(manualDigits)}</div> : null}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Pengirim</label>
            <Input
              value={manualSender}
              onChange={(e) => setManualSender(e.target.value)}
              placeholder="Nama pengirim mutasi"
              disabled={manualSaving}
            />
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input
              value={manualNote}
              onChange={(e) => setManualNote(e.target.value)}
              placeholder="Catatan opsional"
              disabled={manualSaving}
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setManualOpen(false)} disabled={manualSaving}>Batal</Button>
            <Button variant="success" onClick={() => void submitManualMutation()} disabled={manualSaving}>
              {manualSaving ? "Menyimpan..." : "Simpan Mutasi"}
            </Button>
          </div>
        </div>
      </AppModal>
      ) : null}

      {isWalletMode ? (
      <AppModal
        open={providerTopupOpen}
        onClose={() => {
          if (providerTopupSaving) return;
          setProviderTopupOpen(false);
        }}
        title={`Top Up Provider${providerTopupBank ? `: ${providerTopupBank.nama}` : ""}`}
        subtitle={providerTopupBank ? `${providerTopupBank.atas_nama} - ${providerTopupBank.nomor_rekening}` : "Saldo provider akan bertambah sesuai nominal top up."}
      >
        <div className="grid gap-3">
          <div className="rounded-xl border border-white/10 bg-white/5 p-3 text-sm">
            <div className="font-medium text-white">{providerTopupBank?.nama || "-"}</div>
            <div className="font-medium text-emerald-950">{providerTopupBank?.atas_nama || "-"}</div>
            <div className="font-mono text-emerald-700">{providerTopupBank?.nomor_rekening || "-"}</div>
            <div className="font-semibold text-emerald-700">Saldo rekening: Rp {fmtID(Number(providerTopupBank?.saldo || 0))}</div>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Provider</label>
            <select
              className="h-11 rounded-md border border-input bg-background px-3 text-sm"
              value={providerTopupProvider}
              onChange={(e) => setProviderTopupProvider(e.target.value)}
              disabled={providerTopupSaving}
            >
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
            <Input
              value={providerTopupDigits}
              onChange={(e) => setProviderTopupDigits(onlyDigits(e.target.value))}
              placeholder="Masukkan nominal top up"
              inputMode="numeric"
              disabled={providerTopupSaving}
            />
            {providerTopupDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(providerTopupDigits)}</div> : null}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Admin Bank</label>
            <Input
              value={providerTopupAdminDigits}
              onChange={(e) => setProviderTopupAdminDigits(onlyDigits(e.target.value))}
              placeholder="Masukkan admin bank"
              inputMode="numeric"
              disabled={providerTopupSaving}
            />
            {providerTopupAdminDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(providerTopupAdminDigits)}</div> : null}
          </div>
          <div className="rounded-md border border-white/10 bg-slate-950/45 p-3 text-sm text-slate-300">
            Total rekening terpotong: <span className="font-medium text-slate-100">Rp {fmtID(providerTopupBankDebit)}</span>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input
              value={providerTopupNote}
              onChange={(e) => setProviderTopupNote(e.target.value)}
              placeholder="Catatan transfer provider"
              disabled={providerTopupSaving}
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setProviderTopupOpen(false)} disabled={providerTopupSaving}>Batal</Button>
            <Button variant="success" onClick={() => void submitProviderTopup()} disabled={providerTopupSaving}>
              {providerTopupSaving ? "Memproses..." : "Proses Top Up"}
            </Button>
          </div>
        </div>
      </AppModal>
      ) : null}

      {!isWalletMode && !isOperatorMode ? (
      <AppModal
        open={openForm}
        onClose={() => setOpenForm(false)}
        title={editing ? "Edit Bank" : "Tambah Bank"}
        subtitle={isWalletMode ? "Ubah status aktif atau nonaktif bank." : "Atur informasi bank yang digunakan untuk operasional admin."}
      >
        <div className="grid gap-3">
          {isWalletMode ? (
            <>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Jenis Bank</label>
                <Input value={nama} readOnly />
              </div>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Nama Rekening</label>
                <Input value={atasNama} readOnly />
              </div>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Nomor Rekening</label>
                <Input value={nomorRekening} readOnly />
              </div>
            </>
          ) : (
            <>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Jenis Bank</label>
                <Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="Contoh: BCA" />
              </div>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Nama Rekening</label>
                <Input value={atasNama} onChange={(e) => setAtasNama(e.target.value)} placeholder="Contoh: PT Isiloka" />
              </div>
              <div className="grid gap-1.5">
                <label className="text-sm font-medium">Nomor Rekening</label>
                <Input value={nomorRekening} onChange={(e) => setNomorRekening(e.target.value)} placeholder="Contoh: 1234567890" />
              </div>
              {!editing ? (
                <div className="grid gap-1.5">
                  <label className="text-sm font-medium">Saldo Awal</label>
                  <Input
                    value={saldoDigits}
                    onChange={(e) => setSaldoDigits(onlyDigits(e.target.value))}
                    placeholder="Masukkan saldo awal"
                    inputMode="numeric"
                  />
                  {saldoDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(saldoDigits)}</div> : null}
                </div>
              ) : (
                <div className="grid gap-1.5">
                  <label className="text-sm font-medium">Saldo Saat Ini</label>
                  <Input value={`Rp ${fmtID(Number(editing.saldo || 0))}`} readOnly />
                </div>
              )}
            </>
          )}
          <label className="inline-flex items-center gap-2 text-sm">
            <input type="checkbox" checked={aktif} onChange={(e) => setAktif(e.target.checked)} />
            Aktif
          </label>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setOpenForm(false)}>Batal</Button>
            <Button onClick={() => void submitForm()} disabled={saving}>{saving ? "Menyimpan..." : "Simpan"}</Button>
          </div>
        </div>
      </AppModal>
      ) : null}

      {!isWalletMode && !isOperatorMode ? (
      <AppModal
        open={adjustOpen}
        onClose={() => setAdjustOpen(false)}
        title={`Koreksi Saldo Bank${adjustBank ? `: ${adjustBank.nama}` : ""}`}
        subtitle="Gunakan credit untuk menambah saldo bank dan debit untuk mengurangi saldo bank."
      >
        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Arah</label>
            <select
              className="h-11 rounded-md border border-input bg-background px-3 text-sm"
              value={adjustDirection}
              onChange={(e) => setAdjustDirection(e.target.value as "credit" | "debit")}
            >
              <option value="credit">Credit</option>
              <option value="debit">Debit</option>
            </select>
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Nominal</label>
            <Input
              value={adjustDigits}
              onChange={(e) => setAdjustDigits(onlyDigits(e.target.value))}
              placeholder="Masukkan nominal"
              inputMode="numeric"
            />
            {adjustDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(adjustDigits)}</div> : null}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input value={adjustNote} onChange={(e) => setAdjustNote(e.target.value)} placeholder="Alasan koreksi saldo" />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setAdjustOpen(false)}>Batal</Button>
            <Button onClick={() => void submitAdjust()} disabled={adjustSaving}>{adjustSaving ? "Menyimpan..." : "Simpan"}</Button>
          </div>
        </div>
      </AppModal>
      ) : null}

      {!isOperatorMode ? (
      <AppModal
        open={transferOpen}
        onClose={() => setTransferOpen(false)}
        title={isWalletMode ? `Transfer ke BCA OPERASIONAL${transferBank ? `: ${transferBank.nama}` : ""}` : `Transfer Saldo Bank${transferBank ? `: ${transferBank.nama}` : ""}`}
        subtitle={isWalletMode ? "Tujuan dikunci ke BCA OPERASIONAL 3432738881. Nominal transfer masuk ke BCA, biaya admin hanya mengurangi bank sumber." : "Transfer keluar akan mengurangi saldo bank untuk kebutuhan seperti gaji, pembelian, atau pembagian deviden."}
      >
        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Tujuan Transfer</label>
            {isWalletMode ? (
              <Input value={`BCA OPERASIONAL - ${BCA_OPERATIONAL_ACCOUNT}`} readOnly />
            ) : (
              <Input
                value={transferTujuan}
                onChange={(e) => setTransferTujuan(e.target.value)}
                placeholder="Contoh: bayar gaji, beli alat kantor, bagi deviden"
              />
            )}
          </div>
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Nominal</label>
            <Input
              value={transferDigits}
              onChange={(e) => setTransferDigits(onlyDigits(e.target.value))}
              placeholder="Masukkan nominal"
              inputMode="numeric"
            />
            {transferDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(transferDigits)}</div> : null}
          </div>
          {isWalletMode ? (
            <div className="grid gap-1.5">
              <label className="text-sm font-medium">Biaya Admin</label>
              <Input
                value={transferAdminFeeDigits}
                onChange={(e) => setTransferAdminFeeDigits(onlyDigits(e.target.value))}
                placeholder="Masukkan biaya admin"
                inputMode="numeric"
              />
              {transferAdminFeeDigits ? <div className="text-xs text-muted-foreground">{formatRupiahDigits(transferAdminFeeDigits)}</div> : null}
            </div>
          ) : null}
          <div className="grid gap-1.5">
            <label className="text-sm font-medium">Catatan</label>
            <Input value={transferNote} onChange={(e) => setTransferNote(e.target.value)} placeholder={isWalletMode ? "Catatan transfer ke BCA OPERASIONAL" : "Catatan internal transfer keluar"} />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={() => setTransferOpen(false)}>Batal</Button>
            <Button onClick={() => void submitTransfer()} disabled={transferSaving}>{transferSaving ? "Menyimpan..." : "Transfer"}</Button>
          </div>
        </div>
      </AppModal>
      ) : null}
    </div>
  );
}
