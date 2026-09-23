"use client";

import { useEffect, useMemo, useState } from "react";
import { ArrowRightLeft, RefreshCw, Search } from "lucide-react";
import { AppModal } from "@/components/ui/app-modal";
import { alertConfirm, alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID, moneyInput } from "@/lib/format";

type ProviderWallet = {
  provider: string;
  saldo_internal: number;
};

type LoketBayarSummary = {
  provider: string;
  saldo_internal: number;
  snapshot_saldo_provider?: number;
  snapshot_at?: string;
  max_amount: number;
};

type ProviderAccount = {
  id: number;
  provider: string;
  nama: string;
  bank: string;
  nomor_rekening: string;
  nomor_rekening_digits: string;
  aktif: boolean;
};

type TransferRow = {
  id: number;
  direction: string;
  ledger_type: string;
  ref_id: string;
  provider: string;
  bank_name: string;
  account_no: string;
  account_name: string;
  amount: number;
  admin_fee: number;
  status: string;
  reason: string;
  note: string;
  created_by_name: string;
  created_at: string;
  processed_at?: string;
  callback_at?: string;
  reversed_at?: string;
};

function authHeader(): Record<string, string> {
  const token = localStorage.getItem("auth_token") || "";
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function statusClass(status: string): string {
  const s = String(status || "").toLowerCase();
  if (s === "success") return "rounded-full bg-emerald-500/15 px-2 py-1 text-xs font-medium text-emerald-300";
  if (s === "failed" || s === "create_failed") return "rounded-full bg-red-500/15 px-2 py-1 text-xs font-medium text-red-300";
  if (s === "ready" || s === "requested" || s === "processing") return "rounded-full bg-sky-500/15 px-2 py-1 text-xs font-medium text-sky-300";
  return "rounded-full bg-slate-500/15 px-2 py-1 text-xs font-medium text-slate-300";
}

function displayStatus(status: string): string {
  const s = String(status || "-").replace(/_/g, " ");
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function directionClass(direction: string): string {
  if (direction === "in") return "rounded-full bg-emerald-500/15 px-2 py-1 text-xs font-medium text-emerald-300";
  return "rounded-full bg-rose-500/15 px-2 py-1 text-xs font-medium text-rose-300";
}

function directionLabel(direction: string): string {
  return direction === "in" ? "Masuk" : "Keluar";
}

function ledgerLabel(item: TransferRow): string {
  if (item.ledger_type === "smpay_wede") return "Wede SMPAY";
  if (item.ledger_type === "refund") return "Refund";
  if (item.ledger_type === "transfer_provider") return "Transfer Provider";
  return displayStatus(item.ledger_type || item.direction);
}

export default function LoketBayarProviderTransferPage() {
  const [summary, setSummary] = useState<LoketBayarSummary | null>(null);
  const [providers, setProviders] = useState<ProviderWallet[]>([]);
  const [accounts, setAccounts] = useState<ProviderAccount[]>([]);
  const [items, setItems] = useState<TransferRow[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [loading, setLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [statusFilter, setStatusFilter] = useState("");
  const [providerFilter, setProviderFilter] = useState("");
  const [searchText, setSearchText] = useState("");

  const [open, setOpen] = useState(false);
  const [provider, setProvider] = useState("");
  const [accountID, setAccountID] = useState("");
  const [amountDigits, setAmountDigits] = useState("");
  const [note, setNote] = useState("");
  const [checking, setChecking] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [inquiry, setInquiry] = useState<TransferRow | null>(null);

  const selectedProvider = providers.find((item) => item.provider === provider);
  const selectedAccount = accounts.find((item) => String(item.id) === accountID);
  const amount = Number(amountDigits || "0");
  const totalDebit = amount;

  const columns: DataTableColumn<TransferRow>[] = [
    { id: "time", header: "Waktu", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (item) => new Date(item.created_at).toLocaleString("id-ID") },
    { id: "direction", header: "Tipe", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap", render: (item) => <span className={directionClass(item.direction)}>{directionLabel(item.direction)}</span> },
    { id: "ref", header: "RefID", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap font-mono text-cyan-200", render: (item) => item.ref_id },
    { id: "source", header: "Sumber/Tujuan", thClassName: "whitespace-nowrap", tdClassName: "min-w-72 text-slate-200", render: (item) => {
      if (item.direction === "in") return item.note || item.account_name || ledgerLabel(item);
      return `${item.bank_name} • ${item.account_name || "-"} • ${item.account_no}`;
    } },
    { id: "provider", header: "Provider", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap font-semibold uppercase text-slate-100", render: (item) => item.provider || ledgerLabel(item) },
    { id: "amount", header: "Nominal", thClassName: "whitespace-nowrap text-right", tdClassName: "whitespace-nowrap text-right font-semibold", render: (item) => (
      <span className={item.direction === "in" ? "text-emerald-300" : "text-rose-300"}>
        {item.direction === "in" ? "+" : "-"} Rp {fmtID(item.amount)}
      </span>
    ) },
    { id: "fee", header: "Admin", thClassName: "whitespace-nowrap text-right", tdClassName: "whitespace-nowrap text-right text-slate-300", render: (item) => (item.admin_fee > 0 ? `Rp ${fmtID(item.admin_fee)}` : "-") },
    { id: "status", header: "Status", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap", render: (item) => <span className={statusClass(item.status)}>{displayStatus(item.status)}</span> },
    { id: "actor", header: "User/Member", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (item) => item.created_by_name || "-" },
    { id: "reason", header: "Keterangan", thClassName: "whitespace-nowrap", tdClassName: "min-w-56 text-slate-400", render: (item) => item.reason || item.note || "-" },
  ];

  async function loadSummary() {
    const r = await fetch("/api/admin/loketbayar-transfer/summary", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    if (!r.ok || !j.ok) throw new Error(j.error || "Gagal memuat saldo LoketBayar");
    setSummary(j.summary || null);
  }

  async function loadProviders() {
    const r = await fetch("/api/admin/provider/wallets", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    const rows = Array.isArray(j.data) ? j.data : [];
    setProviders(
      rows
        .map((row: ProviderWallet) => row)
        .filter((row: ProviderWallet) => row.provider.toLowerCase() !== "loketbayar")
        .sort((a: ProviderWallet, b: ProviderWallet) => a.provider.localeCompare(b.provider, "id")),
    );
  }

  async function loadHistory() {
    setHistoryLoading(true);
    try {
      const qs = new URLSearchParams({ limit: "100" });
      if (statusFilter) qs.set("status", statusFilter);
      if (providerFilter) qs.set("provider", providerFilter);
      if (searchText.trim()) qs.set("q", searchText.trim());
      const r = await fetch(`/api/admin/loketbayar-transfer/transfers?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat history LoketBayar");
        setItems([]);
        setTotalItems(0);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
      setTotalItems(Number(j.total || 0));
    } finally {
      setHistoryLoading(false);
    }
  }

  async function loadAccounts(nextProvider: string) {
    if (!nextProvider) {
      setAccounts([]);
      return;
    }
    const qs = new URLSearchParams({ provider: nextProvider, aktif: "1", limit: "500" });
    const r = await fetch(`/api/admin/provider-accounts?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    const rows = Array.isArray(j.items) ? j.items : [];
    setAccounts(rows);
  }

  async function loadAll() {
    setLoading(true);
    try {
      await Promise.all([loadSummary(), loadProviders(), loadHistory()]);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal memuat data LoketBayar");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadAll();
  }, []);

  useEffect(() => {
    void loadAccounts(provider);
    setAccountID("");
    setInquiry(null);
  }, [provider]);

  function openTransfer() {
    const firstProvider = providers[0]?.provider || "";
    setProvider(firstProvider);
    setAccountID("");
    setAmountDigits("");
    setNote("");
    setInquiry(null);
    setOpen(true);
  }

  async function resetHistoryFilters() {
    setStatusFilter("");
    setProviderFilter("");
    setSearchText("");
    setHistoryLoading(true);
    try {
      const r = await fetch("/api/admin/loketbayar-transfer/transfers?limit=100", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal memuat history LoketBayar");
        setItems([]);
        setTotalItems(0);
        return;
      }
      setItems(Array.isArray(j.items) ? j.items : []);
      setTotalItems(Number(j.total || 0));
    } finally {
      setHistoryLoading(false);
    }
  }

  async function submitTransfer() {
    if (checking) return;
    if (!provider) {
      await alertWarning("Provider wajib dipilih.");
      return;
    }
    if (!accountID) {
      await alertWarning("Rekening tujuan wajib dipilih.");
      return;
    }
    if (!Number.isFinite(amount) || amount < 10000 || amount > 50000000) {
      await alertWarning("Nominal wajib Rp 10.000 sampai Rp 50.000.000.");
      return;
    }
    if (summary && summary.saldo_internal < totalDebit) {
      await alertWarning("Saldo LoketBayar tidak cukup.");
      return;
    }

    setChecking(true);
    try {
      const r = await fetch("/api/admin/loketbayar-transfer/transfers/create", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          provider,
          provider_rekening_id: Number(accountID),
          amount,
          note: note.trim(),
        }),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        if (j.item) setInquiry(j.item);
        await alertError(j.error || "Proses transfer gagal");
        return;
      }
      const nextInquiry = j.item || null;
      setInquiry(nextInquiry);
      if (nextInquiry) {
        await confirmTransferNow(nextInquiry);
      }
    } finally {
      setChecking(false);
    }
  }

  async function confirmTransferNow(item: TransferRow) {
    const ok = await alertConfirm({
      title: "Transfer Sekarang?",
      text: [
        `Provider: ${item.provider.toUpperCase()}`,
        `Tujuan: ${item.account_name || "-"} - ${item.bank_name} ${item.account_no}`,
        `Nominal: Rp ${fmtID(item.amount)}`,
      ].join("\n"),
      confirmButtonText: "Transfer Sekarang",
    });
    if (!ok) return;
    await processTransfer(item);
  }

  async function processTransfer(item: TransferRow) {
    if (processing || !item) return;

    setProcessing(true);
    try {
      const r = await fetch(`/api/admin/loketbayar-transfer/transfers/${item.id}/process`, {
        method: "POST",
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        if (j.item) setInquiry(j.item);
        await alertError(j.error || "Transfer LoketBayar gagal");
        return;
      }
      await alertSuccess(`Transfer dibuat. Status: ${displayStatus(j.item?.status || "-")}`);
      setOpen(false);
      setInquiry(null);
      await loadAll();
    } finally {
      setProcessing(false);
    }
  }

  const providerOptions = useMemo(() => providers.map((item) => item.provider), [providers]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-slate-100">Transfer LoketBayar ke Provider</h1>
          <p className="mt-1 text-sm text-slate-400">Transfer saldo internal LoketBayar ke rekening provider lewat TRFBANK.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button type="button" variant="outline" size="sm" onClick={() => void loadAll()} disabled={loading}>
            <RefreshCw className="h-4 w-4" />
            Refresh
          </Button>
          <Button type="button" variant="primary" size="sm" onClick={openTransfer}>
            <ArrowRightLeft className="h-4 w-4" />
            Transfer LoketBayar
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <div className="rounded-lg border border-white/10 bg-slate-900/70 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Saldo Internal LoketBayar</div>
          <div className="mt-2 text-2xl font-semibold text-slate-100">Rp {fmtID(summary?.saldo_internal || 0)}</div>
        </div>
        <div className="rounded-lg border border-white/10 bg-slate-900/70 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Snapshot Provider LoketBayar</div>
          <div className="mt-2 text-2xl font-semibold text-slate-100">Rp {fmtID(summary?.snapshot_saldo_provider || 0)}</div>
          <div className="mt-2 text-xs text-slate-400">
            {summary?.snapshot_at ? new Date(summary.snapshot_at).toLocaleString("id-ID") : "-"}
          </div>
        </div>
        <div className="rounded-lg border border-white/10 bg-slate-900/70 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Batas Sekali Transfer</div>
          <div className="mt-2 text-2xl font-semibold text-slate-100">Rp {fmtID(summary?.max_amount || 50000000)}</div>
        </div>
        <div className="rounded-lg border border-white/10 bg-slate-900/70 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Provider Tujuan</div>
          <div className="mt-2 text-2xl font-semibold text-slate-100">{fmtID(providerOptions.length)}</div>
        </div>
      </div>

      <section className="rounded-lg border border-white/10 bg-slate-900/60">
        <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
          <h2 className="text-sm font-semibold text-slate-100">Transaksi LoketBayar ke Provider</h2>
          <span className="text-xs text-slate-400">{historyLoading ? "Memuat..." : `${items.length} dari ${fmtID(totalItems)} data`}</span>
        </div>
        <div className="grid gap-3 border-b border-white/10 px-4 py-3 md:grid-cols-[170px_190px_1fr_auto_auto] md:items-end">
          <label className="space-y-1">
            <span className="text-xs font-medium text-slate-300">Status</span>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="h-10 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-blue-500"
            >
              <option value="">Semua</option>
              <option value="success">Sukses</option>
              <option value="ready">Ready</option>
              <option value="processing">Processing</option>
              <option value="requested">Requested</option>
              <option value="failed">Gagal</option>
              <option value="create_failed">Create gagal</option>
            </select>
          </label>
          <label className="space-y-1">
            <span className="text-xs font-medium text-slate-300">Provider</span>
            <select
              value={providerFilter}
              onChange={(e) => setProviderFilter(e.target.value)}
              className="h-10 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-blue-500"
            >
              <option value="">Semua</option>
              {providerOptions.map((item) => (
                <option key={item} value={item}>
                  {item.toUpperCase()}
                </option>
              ))}
            </select>
          </label>
          <label className="space-y-1">
            <span className="text-xs font-medium text-slate-300">Cari transaksi</span>
            <Input
              value={searchText}
              placeholder="RefID, nominal, bank, user, keterangan"
              onChange={(e) => setSearchText(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") void loadHistory();
              }}
            />
          </label>
          <Button type="button" variant="info" size="sm" onClick={() => void loadHistory()} disabled={historyLoading}>
            <Search className="h-4 w-4" />
            Cari
          </Button>
          <Button type="button" variant="outline" size="sm" onClick={() => void resetHistoryFilters()} disabled={historyLoading}>
            Reset
          </Button>
        </div>
        <DataTable
          columns={columns}
          rows={items}
          rowKey={(item) => `${item.direction}-${item.id}-${item.ref_id}`}
          loading={historyLoading}
          minWidthClassName="min-w-[1500px]"
          emptyText="Belum ada transaksi LoketBayar."
        />
      </section>

      <AppModal open={open} onClose={() => setOpen(false)} title="Transfer LoketBayar" subtitle="Transfer saldo LoketBayar ke provider." maxWidthClassName="max-w-2xl">
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-2">
              <span className="text-xs font-medium text-slate-300">Provider</span>
              <select
                value={provider}
                onChange={(e) => setProvider(e.target.value)}
                className="h-10 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-blue-500"
              >
                <option value="">Pilih provider</option>
                {providerOptions.map((item) => (
                  <option key={item} value={item}>
                    {item.toUpperCase()}
                  </option>
                ))}
              </select>
            </label>
            <label className="space-y-2">
              <span className="text-xs font-medium text-slate-300">Rekening Tujuan</span>
              <select
                value={accountID}
                onChange={(e) => {
                  setAccountID(e.target.value);
                  setInquiry(null);
                }}
                className="h-10 w-full rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100 outline-none focus:border-blue-500"
              >
                <option value="">Pilih rekening</option>
                {accounts.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.bank} - {item.nama} - {item.nomor_rekening}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-2">
              <span className="text-xs font-medium text-slate-300">Nominal</span>
              <Input
                value={amountDigits ? `Rp ${fmtID(amountDigits)}` : ""}
                inputMode="numeric"
                placeholder="Rp 0"
                onChange={(e) => {
                  setAmountDigits(moneyInput(e.target.value).digits);
                  setInquiry(null);
                }}
              />
            </label>
            <label className="space-y-2">
              <span className="text-xs font-medium text-slate-300">Catatan</span>
              <Input value={note} placeholder="Opsional" onChange={(e) => setNote(e.target.value)} />
            </label>
          </div>

          <div className="rounded-md border border-white/10 bg-slate-950/60 p-3 text-sm text-slate-300">
            <div className="flex justify-between gap-4">
              <span>Saldo LoketBayar</span>
              <strong className="text-slate-100">Rp {fmtID(summary?.saldo_internal || 0)}</strong>
            </div>
            <div className="mt-1 flex justify-between gap-4">
              <span>Total potong LoketBayar</span>
              <strong className="text-slate-100">Rp {fmtID(totalDebit || 0)}</strong>
            </div>
            <div className="mt-1 flex justify-between gap-4">
              <span>Saldo internal provider</span>
              <strong className="text-slate-100">Rp {fmtID(selectedProvider?.saldo_internal || 0)}</strong>
            </div>
          </div>

          {selectedAccount ? (
            <div className="rounded-md border border-white/10 bg-slate-950/50 p-3 text-sm text-slate-300">
              Tujuan database: <strong className="text-slate-100">{selectedAccount.bank}</strong> {selectedAccount.nama} {selectedAccount.nomor_rekening}
            </div>
          ) : null}

          {inquiry ? (
            <div className="rounded-md border border-emerald-400/20 bg-emerald-500/10 p-3 text-sm text-emerald-100">
              Tujuan LoketBayar: {inquiry.account_name} • {inquiry.bank_name} {inquiry.account_no}
            </div>
          ) : null}

          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="outline" size="sm" onClick={() => setOpen(false)}>
              Batal
            </Button>
            <Button
              type="button"
              variant="primary"
              size="sm"
              onClick={() => {
                if (inquiry) {
                  void confirmTransferNow(inquiry);
                  return;
                }
                void submitTransfer();
              }}
              disabled={checking || processing}
            >
              {processing ? "Transfer..." : checking ? "Memproses..." : inquiry ? "Transfer Sekarang" : "Proses"}
            </Button>
          </div>
        </div>
      </AppModal>
    </div>
  );
}
