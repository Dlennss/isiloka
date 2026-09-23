"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { DateField } from "@/components/ui/date-field";
import { authHeader, downloadXlsx, fmtDateTime, fmtIDR, todayJakarta } from "@/components/dashboard/auditorHelpers";
import { cn } from "@/lib/utils";

type BankRow = { id: number; nama: string };
type FinanceRow = {
  id: number;
  bank_id: number;
  bank_nama: string;
  ref_id: string;
  arah: string;
  amount: number;
  fee: number;
  total_amount: number;
  saldo_bank: number;
  reason: string;
  counterparty?: string | null;
  note: string;
  provider?: string | null;
  member_nama?: string | null;
  occurred_at: string;
};

export function AuditorFinanceClient() {
  const today = todayJakarta();
  const [banks, setBanks] = useState<BankRow[]>([]);
  const [bankID, setBankID] = useState("");
  const [from, setFrom] = useState(today);
  const [to, setTo] = useState(today);
  const [items, setItems] = useState<FinanceRow[]>([]);
  const [loading, setLoading] = useState(false);
  const totals = items.reduce(
    (acc, row) => {
      const amount = Number(row.amount || 0);
      const fee = Number(row.fee || 0);
      if (String(row.arah || "").toLowerCase() === "debit") acc.debit += amount;
      if (String(row.arah || "").toLowerCase() === "credit") acc.kredit += amount;
      acc.fee += fee;
      return acc;
    },
    { debit: 0, kredit: 0, fee: 0 }
  );

  async function loadBanks() {
    const r = await fetch("/api/admin/master/bank", { headers: authHeader(), cache: "no-store" });
    const j = await r.json().catch(() => ({}));
    if (r.ok && j?.ok) {
      const rows = Array.isArray(j.items) ? j.items : [];
      setBanks(rows);
      setBankID((current) => {
        if (current) return current;
        const bca = rows.find((bank: BankRow) => String(bank?.nama || "").trim().toLowerCase() === "bca");
        if (bca) return String(bca.id);
        return rows.length > 0 ? String(rows[0].id) : "";
      });
    }
  }

  async function load() {
    setLoading(true);
    try {
      if (!bankID) {
        setItems([]);
        return;
      }
      const qs = new URLSearchParams({ from, to, limit: "500" });
      qs.set("bank_id", bankID);
      const r = await fetch(`/api/admin/auditor/finance?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat transaksi keuangan");
      setItems(Array.isArray(j.items) ? j.items : []);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadBanks();
  }, []);

  useEffect(() => {
    if (!bankID) return;
    void load();
  }, [bankID]);

  const headerCell = "whitespace-nowrap px-4 py-3";
  const bodyCell = "whitespace-nowrap px-4 py-3 align-middle";

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Transaksi Keuangan</div>
        <div className="text-sm text-muted-foreground">Semua arus kas bank: deposit member, topup provider, withdraw, dan pengeluaran internal.</div>
      </div>
      <div className="flex flex-wrap items-end gap-2">
        <select className="h-10 rounded-md border border-white/15 bg-slate-950 px-3" value={bankID} onChange={(e) => setBankID(e.target.value)}>
          {banks.map((bank) => <option key={bank.id} value={bank.id}>{bank.nama}</option>)}
        </select>
        <DateField label="Dari" className="h-10 w-44" value={from} onChange={(e) => setFrom(e.target.value)} />
        <DateField label="Sampai" className="h-10 w-44" value={to} onChange={(e) => setTo(e.target.value)} />
        <Button className="h-10" onClick={() => void load()} disabled={loading}>{loading ? "Memuat..." : "Terapkan"}</Button>
        <Button variant="outline" className="h-10" disabled={items.length === 0} onClick={() => void downloadXlsx(`auditor-keuangan-${from}-${to}.xlsx`, items, "Keuangan")}>Download XLSX</Button>
      </div>

      <div className="overflow-x-auto rounded-md border border-white/15 bg-slate-950/50">
        <table className="min-w-[1460px] text-sm">
          <thead className="text-left text-slate-400">
            <tr className="border-b border-white/10">
              <th className={cn(headerCell, "min-w-[220px]")}>Ref</th>
              <th className={cn(headerCell, "min-w-[170px]")}>Waktu</th>
              <th className={cn(headerCell, "min-w-[220px]")}>Pihak</th>
              <th className={cn(headerCell, "min-w-[420px]")}>Keterangan</th>
              <th className={cn(headerCell, "min-w-[130px] text-right")}>Debit</th>
              <th className={cn(headerCell, "min-w-[130px] text-right")}>Kredit</th>
              <th className={cn(headerCell, "min-w-[110px] text-right")}>Fee</th>
              <th className={cn(headerCell, "min-w-[150px] text-right")}>Saldo Bank</th>
            </tr>
          </thead>
          <tbody>
            {items.map((row) => (
              <tr key={row.id} className="border-b border-white/5">
                <td className={cn(bodyCell, "font-mono text-xs")}>
                  <div className="max-w-[220px] truncate">{row.ref_id}</div>
                </td>
                <td className={bodyCell}>{fmtDateTime(row.occurred_at)}</td>
                <td className={bodyCell}>
                  <div className="max-w-[220px] truncate">{row.counterparty || row.provider || row.member_nama || "-"}</div>
                </td>
                <td className={bodyCell}>
                  <div className="max-w-[420px] truncate">
                    {row.reason}{row.note ? ` • ${row.note}` : ""}
                  </div>
                </td>
                <td className={cn(bodyCell, "text-right")}>{row.arah.toLowerCase() === "debit" ? fmtIDR(row.amount) : "-"}</td>
                <td className={cn(bodyCell, "text-right")}>{row.arah.toLowerCase() === "credit" ? fmtIDR(row.amount) : "-"}</td>
                <td className={cn(bodyCell, "text-right")}>{fmtIDR(row.fee)}</td>
                <td className={cn(bodyCell, "text-right")}>{fmtIDR(row.saldo_bank)}</td>
              </tr>
            ))}
            {items.length === 0 ? <tr><td className="px-3 py-6 text-center text-slate-400" colSpan={8}>Tidak ada data.</td></tr> : null}
          </tbody>
        </table>
      </div>

      <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
        <div className="grid gap-3 md:grid-cols-3">
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Debit</div>
            <div className="mt-1 text-base font-semibold text-rose-300">{fmtIDR(totals.debit)}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Kredit</div>
            <div className="mt-1 text-base font-semibold text-emerald-300">{fmtIDR(totals.kredit)}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Fee</div>
            <div className="mt-1 text-base font-semibold text-white">{fmtIDR(totals.fee)}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
