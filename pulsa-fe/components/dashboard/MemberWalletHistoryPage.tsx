"use client";

import { useEffect, useMemo, useState } from "react";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Filter, RotateCcw, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type MemberOption = {
  id: number;
  nama?: string;
};

type RawMemberItem = {
  id?: number | string;
  nama?: string;
};

type MutasiRow = {
  id: number;
  member_id: number;
  member_nama?: string;
  bank_id?: number;
  bank_nama?: string;
  bank_nomor_rekening?: string;
  bank_atas_nama?: string;
  ref_id: string;
  arah: string;
  jumlah: number;
  alasan: string;
  catatan?: string;
  saldo_sebelum?: number;
  saldo_sesudah?: number;
  diubah_oleh?: number;
  diubah_oleh_nama?: string;
  dibuat_pada: string;
};

const PAGE_SIZE = 10;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function visiblePages(current: number, total: number) {
  const pages = new Set<number>([1, total]);
  for (let i = current - 2; i <= current + 2; i += 1) {
    if (i >= 1 && i <= total) pages.add(i);
  }
  return Array.from(pages).sort((a, b) => a - b);
}

function toLocalDateInputValue(date: Date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export default function MemberWalletHistoryPage() {
  const now = useMemo(() => new Date(), []);
  const defaultFrom = useMemo(() => toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)), [now]);
  const defaultTo = useMemo(() => toLocalDateInputValue(now), [now]);

  const [rows, setRows] = useState<MutasiRow[]>([]);
  const [members, setMembers] = useState<MemberOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [showFiltersMobile, setShowFiltersMobile] = useState(false);

  const [memberID, setMemberID] = useState("");
  const [arah, setArah] = useState("credit");
  const [refID, setRefID] = useState("");
  const [from, setFrom] = useState(defaultFrom);
  const [to, setTo] = useState(defaultTo);
  const [offset, setOffset] = useState(0);

  const [appliedMemberID, setAppliedMemberID] = useState("");
  const [appliedArah, setAppliedArah] = useState("credit");
  const [appliedRefID, setAppliedRefID] = useState("");
  const [appliedFrom, setAppliedFrom] = useState(defaultFrom);
  const [appliedTo, setAppliedTo] = useState(defaultTo);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    let cancelled = false;
    async function loadMembers() {
      try {
        const r = await fetch(`/api/admin/members?role=member&limit=500`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        const items: RawMemberItem[] = Array.isArray(j?.items) ? j.items : [];
        const next = items
          .filter((x) => Number(x?.id || 0) > 0)
          .map((x) => ({ id: Number(x.id), nama: String(x.nama || `Member #${x.id}`) }));
        if (!cancelled) setMembers(next);
      } catch {
        if (!cancelled) setMembers([]);
      }
    }
    void loadMembers();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError("");
      try {
        const qs = new URLSearchParams();
        qs.set("limit", String(PAGE_SIZE));
        qs.set("offset", String(offset));
        if (appliedMemberID) qs.set("member_id", appliedMemberID);
        if (appliedArah) qs.set("arah", appliedArah);
        if (appliedRefID.trim()) qs.set("ref_id", appliedRefID.trim());
        if (appliedFrom) qs.set("from", appliedFrom);
        if (appliedTo) qs.set("to", appliedTo);
        const r = await fetch(`/api/admin/wallets/members/deposit-history?${qs.toString()}`, {
          headers: authHeader(),
          cache: "no-store",
        });
        const j = await r.json().catch(() => ({}));
        if (!r.ok) {
          throw new Error(typeof j?.error === "string" && j.error ? j.error : "gagal memuat histori deposit member");
        }
        setRows(Array.isArray(j?.items) ? j.items : []);
        setTotal(Number(j?.total || 0));
      } catch (err) {
        setRows([]);
        setTotal(0);
        setError(err instanceof Error ? err.message : "gagal memuat histori deposit member");
      } finally {
        setLoading(false);
      }
    }
    void load();
  }, [offset, appliedMemberID, appliedArah, appliedRefID, appliedFrom, appliedTo]);

  const columns: DataTableColumn<MutasiRow>[] = [
    { id: "waktu", header: "Waktu", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID") },
    { id: "member", header: "Member", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-100", render: (x) => x.member_nama ? `${x.member_nama} (#${x.member_id})` : `#${x.member_id}` },
    {
      id: "bank",
      header: "Bank",
      thClassName: "whitespace-nowrap",
      tdClassName: "text-slate-200",
      render: (x) => (
        <div className="min-w-[220px]">
          <div className="font-medium text-slate-100">{x.bank_nama || "-"}</div>
          <div className="text-xs text-slate-300">{x.bank_atas_nama || "-"}</div>
          <div className="font-mono text-xs text-cyan-200">{x.bank_nomor_rekening || "-"}</div>
        </div>
      ),
    },
    { id: "arah", header: "Arah", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap uppercase text-slate-200", render: (x) => x.arah },
    { id: "jumlah", header: "Jumlah", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-cyan-200 font-medium", render: (x) => `Rp ${fmtID(x.jumlah)}` },
    { id: "alasan", header: "Alasan", thClassName: "whitespace-nowrap", tdClassName: "text-slate-200", render: (x) => x.alasan },
    { id: "catatan", header: "Catatan", thClassName: "whitespace-nowrap", tdClassName: "text-slate-300", render: (x) => x.catatan || "-" },
    { id: "saldo_sebelum", header: "Saldo Sebelum", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => `Rp ${fmtID(Number(x.saldo_sebelum || 0))}` },
    { id: "saldo_sesudah", header: "Saldo Sesudah", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => `Rp ${fmtID(Number(x.saldo_sesudah || 0))}` },
    { id: "ref_id", header: "Ref ID", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.ref_id || "-" },
    { id: "actor", header: "Diubah Oleh", thClassName: "whitespace-nowrap", tdClassName: "whitespace-nowrap text-slate-300", render: (x) => x.diubah_oleh_nama ? `${x.diubah_oleh_nama} (#${x.diubah_oleh || "-"})` : "-" },
  ];

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const pageButtons = visiblePages(currentPage, totalPages);

  return (
    <div className="space-y-4 p-2">
      <div className="rounded-3xl border border-white/10 bg-[linear-gradient(135deg,rgba(15,23,42,.95),rgba(30,41,59,.82))] px-5 py-5 shadow-[0_18px_60px_rgba(2,6,23,.28)]">
        <div className="inline-flex items-center gap-2 rounded-full border border-cyan-400/20 bg-cyan-400/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-200">
          Wallet Audit
        </div>
        <h1 className="mt-3 text-2xl font-semibold tracking-tight text-white">Riwayat Deposit Member</h1>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-300">Lihat mutasi credit hasil approve deposit member tanpa harus membuka member satu per satu.</p>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-950/70 p-4 shadow-[0_18px_60px_rgba(2,6,23,.2)] backdrop-blur">
        <div className="mb-3 flex items-center justify-between gap-3 lg:hidden">
          <div className="text-sm font-medium text-slate-200">Filter Mutasi</div>
          <Button type="button" variant="outline" className="h-10 rounded-2xl border-white/10 bg-slate-900/70 text-slate-200" onClick={() => setShowFiltersMobile((v) => !v)}>
            <Filter className="mr-2 h-4 w-4" />
            {showFiltersMobile ? "Tutup Filter" : "Buka Filter"}
          </Button>
        </div>

        <div className={`${showFiltersMobile ? "grid" : "hidden"} gap-3 lg:grid lg:grid-cols-10`}>
          <div className="space-y-2 lg:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Member</label>
            <select value={memberID} onChange={(e) => setMemberID(e.target.value)} className="h-11 w-full rounded-2xl border border-white/10 bg-slate-900/80 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
              <option value="">Semua member</option>
              {members.map((m) => (
                <option key={m.id} value={String(m.id)}>
                  {m.nama} (#{m.id})
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2 lg:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Arah</label>
            <select value={arah} onChange={(e) => setArah(e.target.value)} className="h-11 w-full rounded-2xl border border-white/10 bg-slate-900/80 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50">
              <option value="">Semua</option>
              <option value="credit">Credit</option>
              <option value="debit">Debit</option>
            </select>
          </div>
          <div className="space-y-2 lg:col-span-2">
            <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Ref ID</label>
            <Input value={refID} onChange={(e) => setRefID(e.target.value)} placeholder="Cari ref id" className="h-11 rounded-2xl border-white/10 bg-slate-900/80 text-slate-100" />
          </div>
          <div className="grid grid-cols-2 gap-3 lg:col-span-4">
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Dari</label>
              <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="h-11 rounded-2xl border-white/10 bg-slate-900/80 text-slate-100" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">Sampai</label>
              <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="h-11 rounded-2xl border-white/10 bg-slate-900/80 text-slate-100" />
            </div>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap justify-end gap-2">
          <Button
            variant="primary"
            className="h-11 rounded-2xl"
            onClick={() => {
              setOffset(0);
              setAppliedMemberID(memberID);
              setAppliedArah(arah);
              setAppliedRefID(refID);
              setAppliedFrom(from);
              setAppliedTo(to);
              setShowFiltersMobile(false);
            }}
            disabled={loading}
          >
            <Search className="mr-2 h-4 w-4" />
            Cari
          </Button>
          <Button
            variant="outline"
            className="h-11 rounded-2xl border-white/10 bg-slate-900/70 text-slate-200"
            onClick={() => {
              setMemberID("");
              setArah("credit");
              setRefID("");
              setFrom(defaultFrom);
              setTo(defaultTo);
              setOffset(0);
              setAppliedMemberID("");
              setAppliedArah("credit");
              setAppliedRefID("");
              setAppliedFrom(defaultFrom);
              setAppliedTo(defaultTo);
              setShowFiltersMobile(false);
            }}
            disabled={loading}
          >
            <RotateCcw className="mr-2 h-4 w-4" />
            Reset
          </Button>
        </div>
      </div>

      {error ? (
        <div className="rounded-2xl border border-rose-400/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {error}
        </div>
      ) : null}

      <DataTable columns={columns} rows={rows} rowKey={(x) => x.id} rowNumberStart={offset + 1} minWidthClassName="min-w-[1380px]" emptyText="Belum ada riwayat deposit member." loading={loading} />

      {rows.length ? (
        <div className="mt-3 flex flex-wrap items-center justify-end gap-2 rounded-md border border-white/15 bg-linear-to-r from-slate-900/80 via-slate-900/65 to-cyan-950/25 p-2.5 shadow-[0_16px_36px_-28px_rgba(6,182,212,0.8)]">
          <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => setOffset(0)} disabled={loading || currentPage <= 1}><ChevronsLeft className="h-4 w-4" /></Button>
          <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => setOffset((v) => Math.max(0, v - PAGE_SIZE))} disabled={loading || currentPage <= 1}><ChevronLeft className="h-4 w-4" /></Button>
          {pageButtons.map((page) => (
            <Button key={page} type="button" variant={page === currentPage ? "primary" : "outline"} className={page === currentPage ? "h-9 min-w-9 rounded-xl px-3" : "h-9 min-w-9 rounded-xl border-white/10 bg-slate-900/70 px-3 text-slate-200"} onClick={() => setOffset((page - 1) * PAGE_SIZE)} disabled={loading}>
              {page}
            </Button>
          ))}
          <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => setOffset((v) => v + PAGE_SIZE)} disabled={loading || currentPage >= totalPages}><ChevronRight className="h-4 w-4" /></Button>
          <Button type="button" variant="outline" className="h-9 w-9 rounded-xl border-white/10 bg-slate-900/70 p-0 text-slate-200" onClick={() => setOffset((totalPages - 1) * PAGE_SIZE)} disabled={loading || currentPage >= totalPages}><ChevronsRight className="h-4 w-4" /></Button>
        </div>
      ) : null}
    </div>
  );
}
