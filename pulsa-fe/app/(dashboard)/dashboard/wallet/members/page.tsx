"use client";

import { useEffect, useMemo, useState } from "react";
import { ArrowDownCircle, ArrowUpCircle, ReceiptText, RefreshCw, Wallet } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AppModal } from "@/components/ui/app-modal";
import { DataTable, type DataTableActions, type DataTableColumn } from "@/components/ui/data-table";
import { alertError, alertSuccess, alertWarning } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

type MemberRow = {
  id: number;
  email: string;
  nama: string;
  role: string;
  aktif: boolean;
  saldo: number;
  dibuat_pada: string;
};

type MutasiRow = {
  id: number;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string | null;
  saldo_sebelum?: number | null;
  saldo_sesudah?: number | null;
  dibuat_pada: string;
};

type TrxRow = {
  id: number;
  ref_id: string;
  perintah: string;
  kode_produk: string;
  tujuan: string;
  qty: number;
  status: string;
  keterangan?: string | null;
  biaya_perkiraan: number;
  biaya_aktual: number;
  dibuat_pada: string;
};

const PAGE_SIZE = 10;
const HISTORY_PAGE_SIZE = 10;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function digitsOnly(s: string): string {
  return String(s || "").replace(/\D+/g, "");
}

function formatDateTime(v: string): string {
  try {
    return new Date(v).toLocaleString("id-ID");
  } catch {
    return v;
  }
}

function statusTone(v: string) {
  const s = String(v || "").toLowerCase();
  if (s === "success" || s === "sukses") return "bg-emerald-500/15 text-emerald-300 ring-1 ring-emerald-400/30";
  if (s === "failed" || s === "gagal" || s === "error") return "bg-rose-500/15 text-rose-300 ring-1 ring-rose-400/30";
  if (s === "pending" || s === "proses") return "bg-amber-500/15 text-amber-300 ring-1 ring-amber-400/30";
  return "bg-white/10 text-slate-200 ring-1 ring-white/20";
}

export default function WalletMembersPage() {
  const [items, setItems] = useState<MemberRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [totalSaldo, setTotalSaldo] = useState(0);

  const [adjustOpen, setAdjustOpen] = useState(false);
  const [adjustMember, setAdjustMember] = useState<MemberRow | null>(null);
  const [direction, setDirection] = useState<"credit" | "debit">("credit");
  const [amountDigits, setAmountDigits] = useState("10000");
  const [note, setNote] = useState("");
  const [savingAdjust, setSavingAdjust] = useState(false);

  const [selectedMember, setSelectedMember] = useState<MemberRow | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyTab, setHistoryTab] = useState<"mutasi" | "transaksi">("mutasi");
  const [historyLoading, setHistoryLoading] = useState(false);
  const [mutasiRows, setMutasiRows] = useState<MutasiRow[]>([]);
  const [trxRows, setTrxRows] = useState<TrxRow[]>([]);
  const [historyOffset, setHistoryOffset] = useState(0);
  const [historyHasNext, setHistoryHasNext] = useState(false);
  const [historyTotal, setHistoryTotal] = useState(0);

  async function loadMembers(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("role", "member");
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));
      if (search.trim()) qs.set("search", search.trim());

      const r = await fetch(`/api/admin/members?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const all: MemberRow[] = Array.isArray(j.items) ? j.items : [];
      setItems(all.slice(0, PAGE_SIZE));
      setHasNext(all.length > PAGE_SIZE);
      setTotalSaldo(Number(j.total_saldo || 0));
    } finally {
      setLoading(false);
    }
  }

  async function loadHistory(memberID: number, tab = historyTab, nextOffset = historyOffset) {
    setHistoryLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("member_id", String(memberID));
      qs.set("limit", String(HISTORY_PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));

      const path = tab === "mutasi" ? "/api/admin/history/mutasi" : "/api/admin/history/transaksi";
      const r = await fetch(`${path}?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || `Gagal load ${tab}`);
        return;
      }

      const list = Array.isArray(j.items) ? j.items : [];
      setHistoryHasNext(list.length > HISTORY_PAGE_SIZE);
      setHistoryTotal(Number(j.total || 0));
      if (tab === "mutasi") {
        setMutasiRows(list.slice(0, HISTORY_PAGE_SIZE));
      } else {
        setTrxRows(list.slice(0, HISTORY_PAGE_SIZE));
      }
    } finally {
      setHistoryLoading(false);
    }
  }

  useEffect(() => {
    void loadMembers(offset);
     
  }, [offset]);

  useEffect(() => {
    if (!selectedMember) return;
    void loadHistory(selectedMember.id, historyTab, historyOffset);
     
  }, [selectedMember?.id, historyTab, historyOffset]);

  async function submitAdjust() {
    if (!adjustMember) return;
    const amount = Number(amountDigits || "0");
    if (!Number.isFinite(amount) || amount <= 0) {
      await alertWarning("Nominal tidak valid.");
      return;
    }
    if (!note.trim()) {
      await alertWarning("Catatan wajib diisi.");
      return;
    }

    setSavingAdjust(true);
    try {
      const r = await fetch("/api/admin/wallet/adjust", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeader() },
        body: JSON.stringify({
          member_id: adjustMember.id,
          amount,
          direction,
          reason: `wallet manual ${direction}`,
          note: note.trim(),
        }),
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        await alertError(j.error || "Gagal koreksi saldo.");
        return;
      }
      setAdjustOpen(false);
      setAdjustMember(null);
      setDirection("credit");
      setAmountDigits("10000");
      setNote("");
      await Promise.all([
        loadMembers(offset),
        selectedMember ? loadHistory(selectedMember.id, historyTab, 0) : Promise.resolve(),
      ]);
      setHistoryOffset(0);
      await alertSuccess(`Koreksi saldo berhasil. ref_id=${j.ref_id || "-"}`);
    } finally {
      setSavingAdjust(false);
    }
  }

  const memberColumns: DataTableColumn<MemberRow>[] = [
    {
      id: "nama",
      header: "Member",
      render: (m) => (
        <div className="min-w-0">
          <div className="font-semibold text-slate-100">{m.nama}</div>
          <div className="text-xs text-slate-400">{m.email}</div>
        </div>
      ),
    },
    {
      id: "saldo",
      header: "Saldo",
      tdClassName: "whitespace-nowrap font-semibold text-emerald-200",
      render: (m) => `Rp ${fmtID(m.saldo)}`,
    },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap",
      render: (m) => (
        <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${m.aktif ? "bg-emerald-500/15 text-emerald-300" : "bg-rose-500/15 text-rose-300"}`}>
          {m.aktif ? "Aktif" : "Nonaktif"}
        </span>
      ),
    },
    {
      id: "dibuat",
      header: "Dibuat",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (m) => formatDateTime(m.dibuat_pada),
    },
  ];

  const memberActions: DataTableActions<MemberRow> = {
    header: "Aksi",
    align: "right",
    tdClassName: "whitespace-nowrap",
    render: (m) => {
      const isMutasiActive = selectedMember?.id === m.id && historyTab === "mutasi";
      const isTrxActive = selectedMember?.id === m.id && historyTab === "transaksi";

      return (
        <div className="flex justify-end gap-2">
          <div className="inline-flex rounded-2xl border border-white/10 bg-white/[0.04] p-1 shadow-[0_10px_24px_-18px_rgba(15,23,42,0.9)]">
            <button
              type="button"
              className={`inline-flex h-9 items-center gap-2 rounded-xl px-3 text-sm font-medium transition ${
                isMutasiActive ? "bg-emerald-500/18 text-emerald-200" : "text-slate-200 hover:bg-white/8"
              }`}
              onClick={() => {
                setSelectedMember(m);
                setHistoryOpen(true);
                setHistoryTab("mutasi");
                setHistoryOffset(0);
              }}
            >
              <Wallet className="h-4 w-4" />
              Mutasi
            </button>
            <button
              type="button"
              className={`inline-flex h-9 items-center gap-2 rounded-xl px-3 text-sm font-medium transition ${
                isTrxActive ? "bg-indigo-500/18 text-indigo-200" : "text-slate-200 hover:bg-white/8"
              }`}
              onClick={() => {
                setSelectedMember(m);
                setHistoryOpen(true);
                setHistoryTab("transaksi");
                setHistoryOffset(0);
              }}
            >
              <ReceiptText className="h-4 w-4" />
              Transaksi
            </button>
          </div>
          <Button
            type="button"
            variant="warning"
            className="h-11 rounded-2xl px-4"
            onClick={() => {
              setAdjustMember(m);
              setDirection("credit");
              setAmountDigits("10000");
              setNote("");
              setAdjustOpen(true);
            }}
          >
            <ArrowUpCircle className="mr-2 h-4 w-4" />
            Koreksi Saldo
          </Button>
        </div>
      );
    },
  };

  const mutasiColumns: DataTableColumn<MutasiRow>[] = [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (m) => formatDateTime(m.dibuat_pada) },
    { id: "arah", header: "Arah", tdClassName: "whitespace-nowrap", render: (m) => <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${m.arah === "credit" ? "bg-emerald-500/15 text-emerald-300" : "bg-rose-500/15 text-rose-300"}`}>{m.arah}</span> },
    { id: "jumlah", header: "Jumlah", tdClassName: "whitespace-nowrap font-semibold text-cyan-200", render: (m) => `Rp ${fmtID(m.jumlah)}` },
    { id: "alasan", header: "Alasan", tdClassName: "whitespace-nowrap text-slate-200", render: (m) => m.alasan },
    { id: "catatan", header: "Catatan", tdClassName: "text-slate-300", render: (m) => m.catatan || "-" },
    { id: "saldo", header: "Saldo", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (m) => `${m.saldo_sebelum ?? "-"} → ${m.saldo_sesudah ?? "-"}` },
    { id: "ref", header: "Ref", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (m) => m.ref_id || "-" },
  ];

  const trxColumns: DataTableColumn<TrxRow>[] = [
    { id: "waktu", header: "Waktu", tdClassName: "whitespace-nowrap text-slate-100", render: (t) => formatDateTime(t.dibuat_pada) },
    { id: "ref", header: "Ref", tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300", render: (t) => t.ref_id },
    { id: "produk", header: "Produk", tdClassName: "whitespace-nowrap text-slate-200", render: (t) => t.kode_produk },
    { id: "tujuan", header: "Tujuan", tdClassName: "whitespace-nowrap text-slate-200", render: (t) => t.tujuan },
    { id: "qty", header: "Qty", tdClassName: "whitespace-nowrap text-slate-200", render: (t) => t.qty },
    {
      id: "status",
      header: "Status",
      render: (t) => <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${statusTone(t.status)}`}>{t.status}</span>,
    },
    {
      id: "biaya",
      header: "Biaya",
      tdClassName: "whitespace-nowrap font-semibold text-indigo-200",
      render: (t) => fmtID(t.biaya_aktual || t.biaya_perkiraan),
    },
    { id: "ket", header: "Keterangan", tdClassName: "text-slate-300", render: (t) => t.keterangan || "-" },
  ];

  const historyRows = historyTab === "mutasi" ? mutasiRows : trxRows;
  const historyColumns = historyTab === "mutasi" ? (mutasiColumns as DataTableColumn<MutasiRow | TrxRow>[]) : (trxColumns as DataTableColumn<MutasiRow | TrxRow>[]);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const selectedSubtitle = useMemo(() => {
    if (!selectedMember) return "";
    return `${selectedMember.nama} (${selectedMember.email}) • saldo ${fmtID(selectedMember.saldo)}`;
  }, [selectedMember]);

  return (
    <div className="space-y-5 p-2">
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <div className="text-2xl font-semibold text-white">Saldo Member</div>
          <div className="mt-1 text-sm text-white/65">Operasional wallet: koreksi saldo dan lihat riwayat member tanpa aksi transaksi.</div>
        </div>
        <div className="w-full rounded-2xl border border-emerald-400/25 bg-linear-to-r from-emerald-500/15 to-teal-500/10 px-4 py-3 text-sm text-emerald-100 shadow-[0_14px_30px_-20px_rgba(16,185,129,0.8)] md:w-auto">
          <div className="text-emerald-200/90">Total Saldo Member</div>
          <div className="font-semibold text-emerald-50">Rp {fmtID(totalSaldo)}</div>
        </div>
      </div>

      <div className="rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/70 to-cyan-950/25 p-4 shadow-[0_16px_40px_-24px_rgba(34,211,238,0.28)]">
        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama atau email member"
            className="h-11"
          />
          <div className="flex gap-2">
            <Button
              type="button"
              className="h-11 bg-linear-to-r from-sky-500 to-cyan-500 text-white hover:opacity-90"
              onClick={() => {
                setOffset(0);
                void loadMembers(0);
              }}
              disabled={loading}
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Cari
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-11"
              onClick={() => {
                setSearch("");
                setOffset(0);
                void loadMembers(0);
              }}
              disabled={loading}
            >
              Reset
            </Button>
          </div>
        </div>

        <DataTable
          columns={memberColumns}
          rows={items}
          rowKey={(m) => m.id}
          rowNumberStart={offset + 1}
          minWidthClassName="min-w-[940px]"
          emptyText="Tidak ada member."
          actions={memberActions}
          pagination={{
            page: currentPage,
            totalPages: Math.max(1, currentPage + (hasNext ? 1 : 0)),
            onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
            onNext: () => setOffset((v) => v + PAGE_SIZE),
            onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE),
            disablePrev: loading || offset === 0,
            disableNext: loading || !hasNext,
          }}
        />
      </div>

      <AppModal
        open={historyOpen}
        onClose={() => {
          if (historyLoading) return;
          setHistoryOpen(false);
        }}
        title={selectedMember ? `Riwayat Member • ${selectedMember.nama}` : "Riwayat Member"}
        subtitle={selectedMember ? selectedSubtitle : "Mutasi dan transaksi member terpilih."}
        maxWidthClassName="max-w-6xl"
        footer={
          <>
            <Button variant="outline" onClick={() => setHistoryOpen(false)} disabled={historyLoading}>
              Tutup
            </Button>
          </>
        }
      >
        <div className="space-y-4 md:min-h-[68vh]">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="inline-flex rounded-2xl border border-white/10 bg-white/[0.04] p-1 shadow-[0_10px_24px_-18px_rgba(15,23,42,0.9)]">
              <button
                type="button"
                className={`inline-flex h-10 items-center gap-2 rounded-xl px-4 text-sm font-medium transition ${
                  historyTab === "mutasi" ? "bg-emerald-500/18 text-emerald-200" : "text-slate-200 hover:bg-white/8"
                }`}
                onClick={() => {
                  setHistoryTab("mutasi");
                  setHistoryOffset(0);
                }}
              >
                <Wallet className="h-4 w-4" />
                Mutasi
              </button>
              <button
                type="button"
                className={`inline-flex h-10 items-center gap-2 rounded-xl px-4 text-sm font-medium transition ${
                  historyTab === "transaksi" ? "bg-indigo-500/18 text-indigo-200" : "text-slate-200 hover:bg-white/8"
                }`}
                onClick={() => {
                  setHistoryTab("transaksi");
                  setHistoryOffset(0);
                }}
              >
                <ReceiptText className="h-4 w-4" />
                Transaksi
              </button>
            </div>
            <div className="rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-xs text-slate-300">
              {historyLoading ? "Memuat..." : `${historyTotal} data`}
            </div>
          </div>

          <DataTable
            columns={historyColumns}
            rows={historyRows as (MutasiRow | TrxRow)[]}
            rowKey={(row) => row.id}
            rowNumberStart={historyOffset + 1}
            minWidthClassName={historyTab === "mutasi" ? "min-w-[980px]" : "min-w-[1120px]"}
            emptyText={`Belum ada ${historyTab}.`}
            pagination={{
              page: Math.floor(historyOffset / HISTORY_PAGE_SIZE) + 1,
              totalPages: Math.max(1, Math.ceil(historyTotal / HISTORY_PAGE_SIZE)),
              onPrev: () => setHistoryOffset((v) => Math.max(0, v - HISTORY_PAGE_SIZE)),
              onNext: () => setHistoryOffset((v) => v + HISTORY_PAGE_SIZE),
              disablePrev: historyLoading || historyOffset === 0,
              disableNext: historyLoading || !historyHasNext,
            }}
          />
        </div>
      </AppModal>

      <AppModal
        open={adjustOpen}
        onClose={() => {
          if (savingAdjust) return;
          setAdjustOpen(false);
          setAdjustMember(null);
        }}
        title={adjustMember ? `Koreksi Saldo • ${adjustMember.nama}` : "Koreksi Saldo"}
        subtitle={adjustMember ? `${adjustMember.email} • saldo ${fmtID(adjustMember.saldo)}` : undefined}
        maxWidthClassName="max-w-2xl"
        footer={
          <>
            <Button variant="outline" onClick={() => setAdjustOpen(false)} disabled={savingAdjust}>
              Batal
            </Button>
            <Button variant="success" onClick={() => void submitAdjust()} disabled={savingAdjust}>
              {savingAdjust ? "Menyimpan..." : "Simpan"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-2">
            <button
              type="button"
              className={`rounded-2xl border px-4 py-3 text-left transition ${
                direction === "credit" ? "border-emerald-400/50 bg-emerald-500/12 text-emerald-100" : "border-white/10 bg-white/[0.03] text-slate-200"
              }`}
              onClick={() => setDirection("credit")}
            >
              <div className="flex items-center gap-2 font-semibold">
                <ArrowUpCircle className="h-4 w-4" />
                Tambah Saldo
              </div>
              <div className="mt-1 text-xs text-white/60">Menambah saldo member.</div>
            </button>
            <button
              type="button"
              className={`rounded-2xl border px-4 py-3 text-left transition ${
                direction === "debit" ? "border-rose-400/50 bg-rose-500/12 text-rose-100" : "border-white/10 bg-white/[0.03] text-slate-200"
              }`}
              onClick={() => setDirection("debit")}
            >
              <div className="flex items-center gap-2 font-semibold">
                <ArrowDownCircle className="h-4 w-4" />
                Kurangi Saldo
              </div>
              <div className="mt-1 text-xs text-white/60">Mengurangi saldo member.</div>
            </button>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-xs text-slate-300">Nominal</label>
              <Input
                value={amountDigits ? fmtID(amountDigits) : ""}
                onChange={(e) => setAmountDigits(digitsOnly(e.target.value))}
                inputMode="numeric"
                placeholder="10.000"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-300">Preview</label>
              <div className="flex h-10 items-center rounded-xl border border-white/10 bg-white/[0.03] px-3 text-sm text-slate-200">
                Rp {fmtID(amountDigits || "0")}
              </div>
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs text-slate-300">Catatan</label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              rows={3}
              className="w-full rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-300/40 focus:ring-2 focus:ring-cyan-400/10"
              placeholder="Wajib diisi untuk audit koreksi saldo"
            />
          </div>

          <div className="rounded-xl border border-amber-400/20 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
            Koreksi saldo akan tercatat di mutasi member beserta actor yang melakukan perubahan.
          </div>
        </div>
      </AppModal>
    </div>
  );
}
