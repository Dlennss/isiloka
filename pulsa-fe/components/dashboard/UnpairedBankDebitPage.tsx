"use client";

import { useEffect, useMemo, useState } from "react";
import { RefreshCcw, Search, Wallet } from "lucide-react";
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

type DebitRow = {
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
  pengirim?: string | null;
  penerima?: string | null;
  waktu_mutasi_bank?: string | null;
  dibuat_pada: string;
};

const PAGE_SIZE = 25;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function toDateInput(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function firstDayThisMonth() {
  const now = new Date();
  return toDateInput(new Date(now.getFullYear(), now.getMonth(), 1));
}

function today() {
  return toDateInput(new Date());
}

function displayTime(row: DebitRow) {
  const value = row.waktu_mutasi_bank || row.dibuat_pada;
  return value ? new Date(value).toLocaleString("id-ID") : "-";
}

function bankLabel(bank: BankRow): string {
  return `${bank.nama} - ${bank.atas_nama} - ${bank.nomor_rekening}`;
}

export default function UnpairedBankDebitPage() {
  const [banks, setBanks] = useState<BankRow[]>([]);
  const [providers, setProviders] = useState<ProviderOption[]>([]);
  const [rows, setRows] = useState<DebitRow[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);

  const [bankID, setBankID] = useState("");
  const [from, setFrom] = useState(firstDayThisMonth());
  const [to, setTo] = useState(today());
  const [query, setQuery] = useState("");

  const [assignOpen, setAssignOpen] = useState(false);
  const [assignRow, setAssignRow] = useState<DebitRow | null>(null);
  const [assignProvider, setAssignProvider] = useState("");
  const [assignProviderSearch, setAssignProviderSearch] = useState("");
  const [assignNote, setAssignNote] = useState("");
  const [assignSaving, setAssignSaving] = useState(false);

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const filteredAssignProviders = useMemo(() => {
    const q = assignProviderSearch.trim().toLowerCase();
    if (!q) return providers;
    return providers.filter((item) => item.provider.toLowerCase().includes(q));
  }, [assignProviderSearch, providers]);

  async function loadBanks() {
    try {
      const r = await fetch("/api/admin/master/bank", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat rekening bank.");
      setBanks(Array.isArray(j.items) ? j.items : []);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal memuat rekening bank.");
    }
  }

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/provider/wallets", { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      const data = Array.isArray(j?.data) ? j.data : [];
      const next = data.map((item: { provider?: string }) => ({ provider: String(item.provider || "") })).filter((item: ProviderOption) => item.provider);
      setProviders(next);
      setAssignProvider((prev) => prev || next[0]?.provider || "");
    } catch {
      setProviders([]);
    }
  }

  async function load(nextOffset = offset, overrides?: Partial<{ bankID: string; from: string; to: string; query: string }>) {
    setLoading(true);
    try {
      const bankIDValue = overrides?.bankID ?? bankID;
      const fromValue = overrides?.from ?? from;
      const toValue = overrides?.to ?? to;
      const queryValue = overrides?.query ?? query;
      const qs = new URLSearchParams();
      qs.set("limit", String(PAGE_SIZE));
      qs.set("offset", String(nextOffset));
      if (bankIDValue) qs.set("bank_id", bankIDValue);
      if (fromValue) qs.set("from", fromValue);
      if (toValue) qs.set("to", toValue);
      if (queryValue.trim()) qs.set("q", queryValue.trim());

      const r = await fetch(`/api/admin/bank/unpaired-debits?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat debit tanpa pasangan.");
      setRows(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
      setOffset(nextOffset);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal memuat debit tanpa pasangan.");
    } finally {
      setLoading(false);
    }
  }

  function openAssignProvider(row: DebitRow) {
    setAssignRow(row);
    setAssignProviderSearch("");
    setAssignProvider((prev) => prev || providers[0]?.provider || "");
    setAssignNote("");
    setAssignOpen(true);
  }

  async function submitAssignProvider() {
    if (!assignRow) return;
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
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal menambahkan saldo provider.");

      await alertSuccess(
        `RefID: ${j.ref_id || assignRow.ref_id || "-"} - provider ${String(j.provider || assignProvider).toUpperCase()} - saldo provider: Rp ${fmtID(Number(j.saldo_internal || 0))}`
      );
      setAssignOpen(false);
      setAssignRow(null);
      await load(offset);
    } catch (err) {
      await alertError(err instanceof Error ? err.message : "Gagal menambahkan saldo provider.");
    } finally {
      setAssignSaving(false);
    }
  }

  useEffect(() => {
    void loadBanks();
    void loadProviders();
  }, []);

  useEffect(() => {
    void load(0);
  }, []);

  const columns: DataTableColumn<DebitRow>[] = [
    { id: "time", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-300", render: displayTime },
    { id: "bank", header: "Bank", tdClassName: "whitespace-nowrap text-slate-200", render: (row) => row.bank_nama || "-" },
    { id: "ref", header: "RefID", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (row) => row.ref_id || "-" },
    { id: "amount", header: "Nominal", tdClassName: "whitespace-nowrap font-semibold text-rose-200", render: (row) => `Rp ${fmtID(Number(row.jumlah || 0))}` },
    { id: "reason", header: "Alasan", tdClassName: "whitespace-nowrap text-slate-300", render: (row) => row.alasan || "-" },
    { id: "target", header: "Target Tercatat", tdClassName: "text-slate-300", render: (row) => row.provider?.trim() || "-" },
    { id: "receiver", header: "Penerima", tdClassName: "text-slate-300", render: (row) => row.penerima || "-" },
    { id: "note", header: "Catatan", tdClassName: "min-w-120 text-slate-300", render: (row) => row.catatan || "-" },
  ];

  const actions: DataTableActions<DebitRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (row) => (
      <Button type="button" variant="info" size="sm" className="h-8" onClick={() => openAssignProvider(row)}>
        <Wallet className="mr-1.5 h-3.5 w-3.5" />
        Provider
      </Button>
    ),
  };

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-rose-950/25 p-5 shadow-[0_16px_40px_-24px_rgba(248,113,113,0.28)]">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-white">Debit Tanpa Pasangan</h1>
            <div className="mt-1 text-sm text-white/60">Mutasi debit yang belum punya pasangan provider atau rekening internal.</div>
          </div>
          <Button type="button" variant="outline" className="h-10" onClick={() => void load(0)} disabled={loading}>
            <RefreshCcw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      <div className="rounded-2xl border border-white/10 bg-card/80 p-4 shadow-[0_16px_40px_-24px_rgba(0,0,0,0.8)]">
        <div className="grid gap-3 lg:grid-cols-[minmax(180px,260px)_150px_150px_minmax(0,1fr)_auto_auto]">
          <select className="h-11 rounded-md border border-white/10 bg-slate-950 px-3 text-sm text-slate-100" value={bankID} onChange={(e) => setBankID(e.target.value)}>
            <option value="">Semua rekening</option>
            {banks.map((bank) => (
              <option key={bank.id} value={String(bank.id)}>
                {bankLabel(bank)}
              </option>
            ))}
          </select>
          <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="h-11" />
          <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="h-11" />
          <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Cari ref id, nominal, bank, penerima, catatan" className="h-11" />
          <Button type="button" variant="primary" onClick={() => void load(0)} disabled={loading}>
            <Search className="mr-2 h-4 w-4" />
            Cari
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setBankID("");
              setFrom(firstDayThisMonth());
              setTo(today());
              setQuery("");
              void load(0, { bankID: "", from: firstDayThisMonth(), to: today(), query: "" });
            }}
            disabled={loading}
          >
            Reset
          </Button>
        </div>

        <div className="mt-4 rounded-md border border-white/10 bg-slate-950/35 p-3 text-sm text-slate-300">
          Total: <span className="font-semibold text-white">{fmtID(total)}</span> debit
        </div>

        <DataTable<DebitRow>
          columns={columns}
          rows={rows}
          actions={actions}
          rowKey={(row) => row.id}
          emptyText="Tidak ada debit tanpa pasangan untuk filter ini."
          minWidthClassName="min-w-260"
          showRowNumber={false}
          wrapperClassName="mt-4 overflow-auto rounded-md border border-white/10 bg-slate-950/35"
          loading={loading}
          pagination={{
            page: currentPage,
            totalPages,
            onPrev: () => void load(Math.max(0, offset - PAGE_SIZE)),
            onNext: () => void load(offset + PAGE_SIZE),
            onPageChange: (page) => void load(Math.max(0, (page - 1) * PAGE_SIZE)),
            disablePrev: loading || currentPage <= 1,
            disableNext: loading || currentPage >= totalPages,
          }}
        />
      </div>

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
