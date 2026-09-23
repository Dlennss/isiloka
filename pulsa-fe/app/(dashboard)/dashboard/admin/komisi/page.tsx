"use client";

import { useEffect, useMemo, useState } from "react";
import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DateField } from "@/components/ui/date-field";
import { alertError, alertSuccess } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

type Scope = "all" | "retail" | "h2h";

type DailyRow = {
  scope: string;
  day: string;
  month_key: string;
  transaction_count: number;
  transaction_amount: number;
  provider_payment_amount: number;
  margin_amount: number;
  commission_amount: number;
  transaction_expense_amount: number;
  member_deposit_amount: number;
  provider_deposit_amount: number;
  deposit_gap_amount: number;
  profit_amount: number;
};

type MonthGroup = {
  monthKey: string;
  rows: DailyRow[];
  totalTransactions: number;
  totalAmount: number;
  totalProviderPayment: number;
  totalMargin: number;
  totalCommission: number;
  totalTransactionExpense: number;
  totalMemberDeposit: number;
  totalProviderDeposit: number;
  totalDepositGap: number;
  totalProfit: number;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtIDR(v: number) {
  return `Rp ${fmtID(Number(v || 0))}`;
}

function scopeLabel(scope: string) {
  switch ((scope || "").toLowerCase()) {
    case "retail":
      return "Retail";
    case "h2h":
      return "H2H";
    default:
      return "Gabungan";
  }
}

export default function AdminBusinessReportPage() {
  const [scope, setScope] = useState<Scope>("all");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [dailyRows, setDailyRows] = useState<DailyRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshingCache, setRefreshingCache] = useState(false);
  const [err, setErr] = useState("");

  async function load(currentScope: Scope, currentFrom = from, currentTo = to) {
    setLoading(true);
    setErr("");
    try {
      const qs = new URLSearchParams({ scope: currentScope, months: "3" });
      if (currentFrom) qs.set("from", currentFrom);
      if (currentTo) qs.set("to", currentTo);
      const res = await fetch(`/api/admin/reports/daily?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal memuat laporan harian.");
      }
      setDailyRows(Array.isArray(json.items) ? json.items : []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat laporan bisnis.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(scope);
  }, [scope]);

  const groupedMonths = useMemo<MonthGroup[]>(() => {
    const bucket = new Map<string, DailyRow[]>();
    for (const row of dailyRows) {
      const key = row.month_key || "unknown";
      const arr = bucket.get(key) || [];
      arr.push(row);
      bucket.set(key, arr);
    }
    return Array.from(bucket.entries())
      .sort((a, b) => b[0].localeCompare(a[0]))
      .map(([monthKey, rows]) => {
        const ordered = [...rows].sort((a, b) => b.day.localeCompare(a.day));
        return {
          monthKey,
          rows: ordered,
          totalTransactions: ordered.reduce((sum, row) => sum + Number(row.transaction_count || 0), 0),
          totalAmount: ordered.reduce((sum, row) => sum + Number(row.transaction_amount || 0), 0),
          totalProviderPayment: ordered.reduce((sum, row) => sum + Number(row.provider_payment_amount || 0), 0),
          totalMargin: ordered.reduce((sum, row) => sum + Number(row.margin_amount || 0), 0),
          totalCommission: ordered.reduce((sum, row) => sum + Number(row.commission_amount || 0), 0),
          totalTransactionExpense: ordered.reduce((sum, row) => sum + Number(row.transaction_expense_amount || 0), 0),
          totalMemberDeposit: ordered.reduce((sum, row) => sum + Number(row.member_deposit_amount || 0), 0),
          totalProviderDeposit: ordered.reduce((sum, row) => sum + Number(row.provider_deposit_amount || 0), 0),
          totalDepositGap: ordered.reduce((sum, row) => sum + Number(row.deposit_gap_amount || 0), 0),
          totalProfit: ordered.reduce((sum, row) => sum + Number(row.profit_amount || 0), 0),
        };
      });
  }, [dailyRows]);

  const summary = useMemo(() => {
    return {
      totalTransactions: dailyRows.reduce((sum, item) => sum + Number(item.transaction_count || 0), 0),
      totalAmount: dailyRows.reduce((sum, item) => sum + Number(item.transaction_amount || 0), 0),
      totalProviderPayment: dailyRows.reduce((sum, item) => sum + Number(item.provider_payment_amount || 0), 0),
      totalMargin: dailyRows.reduce((sum, item) => sum + Number(item.margin_amount || 0), 0),
      totalPaidCommission: dailyRows.reduce((sum, item) => sum + Number(item.commission_amount || 0), 0),
      totalTransactionExpense: dailyRows.reduce((sum, item) => sum + Number(item.transaction_expense_amount || 0), 0),
      totalMemberDeposit: dailyRows.reduce((sum, item) => sum + Number(item.member_deposit_amount || 0), 0),
      totalProviderDeposit: dailyRows.reduce((sum, item) => sum + Number(item.provider_deposit_amount || 0), 0),
      totalDepositGap: dailyRows.reduce((sum, item) => sum + Number(item.deposit_gap_amount || 0), 0),
      totalProfit: dailyRows.reduce((sum, item) => sum + Number(item.profit_amount || 0), 0),
    };
  }, [dailyRows]);

  function exportDailyCsv() {
    const rows = [
      [
        "scope",
        "day",
        "month_key",
        "transaction_count",
        "transaction_amount",
        "provider_payment_amount",
        "margin_amount",
        "commission_amount",
        "transaction_expense_amount",
        "member_deposit_amount",
        "provider_deposit_amount",
        "deposit_gap_amount",
        "profit_amount",
      ],
      ...dailyRows.map((item) => [
        item.scope,
        item.day,
        item.month_key,
        String(item.transaction_count),
        String(item.transaction_amount),
        String(item.provider_payment_amount),
        String(item.margin_amount),
        String(item.commission_amount),
        String(item.transaction_expense_amount || 0),
        String(item.member_deposit_amount),
        String(item.provider_deposit_amount),
        String(item.deposit_gap_amount),
        String(item.profit_amount),
      ]),
    ];
    downloadCsv(`daily-business-${scope}.csv`, rows);
  }

  async function refreshThreeMonthCache() {
    setRefreshingCache(true);
    setErr("");
    try {
      const res = await fetch("/api/admin/reports/daily/refresh-cache?months=3", {
        method: "POST",
        headers: authHeader(),
        cache: "no-store",
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal update data 3 bulan.");
      }
      await load(scope, from, to);
      await alertSuccess(`Data 3 bulan berhasil diupdate. Baris cache: ${fmtID(Number(json?.item?.refreshed_rows || 0))}.`);
    } catch (e) {
      await alertError(e instanceof Error ? e.message : "Gagal update data 3 bulan.");
    } finally {
      setRefreshingCache(false);
    }
  }

  return (
    <div className="-m-2 min-h-screen bg-[#f0f8f8] p-3 text-[#071b14] sm:p-5 lg:p-6">
      <div className="mx-auto max-w-7xl space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-2xl font-black tracking-tight text-[#071b14]">Laporan Bisnis</div>
          <div className="mt-1 max-w-3xl text-sm font-semibold leading-6 text-[#315847]">
            Rekap harian 3 bulan terakhir: transaksi, bayar provider, margin, komisi, pengeluaran transaksi, deposit member, dan deposit provider berbasis mutasi bank.
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {(["all", "retail", "h2h"] as Scope[]).map((item) => (
            <Button
              key={item}
              variant={scope === item ? "primary" : "outline"}
              className="h-10"
              onClick={() => setScope(item)}
              disabled={loading}
            >
              {scopeLabel(item)}
            </Button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-2">
        <DateField label="Dari" className="h-10 w-44" value={from} onChange={(e) => setFrom(e.target.value)} />
        <DateField label="Sampai" className="h-10 w-44" value={to} onChange={(e) => setTo(e.target.value)} />
        <Button className="h-10" onClick={() => void load(scope, from, to)} disabled={loading}>
          {loading ? "Memuat..." : "Terapkan Filter"}
        </Button>
        <Button
          variant="outline"
          className="h-10"
          onClick={() => {
            setFrom("");
            setTo("");
            void load(scope, "", "");
          }}
          disabled={loading}
        >
          Reset
        </Button>
        <Button variant="outline" className="h-10" onClick={exportDailyCsv} disabled={loading || dailyRows.length === 0}>
          Export Harian CSV
        </Button>
        <Button variant="outline" className="h-10" onClick={() => void refreshThreeMonthCache()} disabled={loading || refreshingCache}>
          <RefreshCw className={`mr-2 h-4 w-4 ${refreshingCache ? "animate-spin" : ""}`} />
          {refreshingCache ? "Updating..." : "Update Data 3 Bulan"}
        </Button>
      </div>

      {err ? (
        <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-bold text-rose-700">
          {err}
        </div>
      ) : null}

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {[
          { label: "Jumlah Trx", value: fmtID(summary.totalTransactions) },
          { label: "Nominal Trx", value: fmtIDR(summary.totalAmount) },
          { label: "Total Bayar Provider", value: fmtIDR(summary.totalProviderPayment) },
          { label: "Margin", value: fmtIDR(summary.totalMargin) },
          { label: "Komisi Dibayar", value: fmtIDR(summary.totalPaidCommission) },
          { label: "Pengeluaran Transaksi", value: fmtIDR(summary.totalTransactionExpense) },
          { label: "Deposit Member", value: fmtIDR(summary.totalMemberDeposit) },
          { label: "Deposit Provider Bank", value: fmtIDR(summary.totalProviderDeposit) },
          { label: "Selisih Deposit", value: fmtIDR(summary.totalDepositGap) },
          { label: "Keuntungan Bersih", value: fmtIDR(summary.totalProfit) },
        ].map((item) => (
          <div
            key={item.label}
            className="rounded-[22px] border border-emerald-200 bg-[linear-gradient(135deg,#ffffff_0%,#f1fff8_100%)] p-4 shadow-[0_14px_30px_rgba(6,78,59,0.08)]"
          >
            <div className="text-xs font-black uppercase tracking-[0.12em] text-[#13464b]">{item.label}</div>
            <div className="mt-2 text-lg font-black text-[#071b14]">{item.value}</div>
          </div>
        ))}
      </section>

      <section className="space-y-4">
        <div className="rounded-[24px] border border-emerald-200 bg-white p-4 shadow-[0_14px_30px_rgba(6,78,59,0.08)]">
          <div className="text-base font-black text-[#071b14]">Rekap Harian 3 Bulan</div>
          <div className="mt-1 text-sm font-semibold text-[#315847]">
            Data harian dikelompokkan per bulan, dengan total setiap item pada bagian bawah bulan.
          </div>
        </div>

        {loading ? (
          <div className="rounded-[24px] border border-emerald-200 bg-white px-4 py-6 text-sm font-semibold text-[#315847]">
            Memuat laporan harian...
          </div>
        ) : null}

        {!loading && groupedMonths.length === 0 ? (
          <div className="rounded-[24px] border border-emerald-200 bg-white px-4 py-6 text-sm font-semibold text-[#315847]">
            Belum ada data transaksi pada periode ini.
          </div>
        ) : null}

        {!loading
          ? groupedMonths.map((group) => (
              <div
                key={group.monthKey}
                className="overflow-hidden rounded-[24px] border border-emerald-200 bg-white shadow-[0_14px_30px_rgba(6,78,59,0.08)]"
              >
                <div className="border-b border-emerald-100 bg-[linear-gradient(135deg,#064e3b_0%,#047857_70%,#55c72f_140%)] px-4 py-3">
                  <div className="text-sm font-black uppercase tracking-wide text-white">{group.monthKey}</div>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-[1480px] w-full text-sm">
                    <thead className="bg-[#f1fff8] text-left">
                      <tr>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Tanggal</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Jumlah Trx</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Nominal Trx</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Total Bayar Provider</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Margin</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Komisi Dibayar</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Pengeluaran Transaksi</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Deposit Member</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Deposit Provider Bank</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Selisih Deposit</th>
                        <th className="px-4 py-3 text-xs font-black uppercase tracking-wide text-[#13464b]">Keuntungan Bersih</th>
                      </tr>
                    </thead>
                    <tbody>
                      {group.rows.map((row) => (
                        <tr key={`${group.monthKey}-${row.day}`} className="border-t border-emerald-100 odd:bg-white even:bg-[#fbfffd]">
                          <td className="px-4 py-3 font-semibold text-[#315847]">{new Date(row.day).toLocaleDateString("id-ID")}</td>
                          <td className="px-4 py-3 font-semibold text-[#315847]">{fmtID(Number(row.transaction_count || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#071b14]">{fmtIDR(Number(row.transaction_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#13464b]">{fmtIDR(Number(row.provider_payment_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#087e8b]">{fmtIDR(Number(row.margin_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#3f6212]">{fmtIDR(Number(row.commission_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-rose-700">{fmtIDR(Number(row.transaction_expense_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#155e75]">{fmtIDR(Number(row.member_deposit_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#13464b]">{fmtIDR(Number(row.provider_deposit_amount || 0))}</td>
                          <td className="px-4 py-3 font-bold text-[#315847]">{fmtIDR(Number(row.deposit_gap_amount || 0))}</td>
                          <td className="px-4 py-3 font-black text-[#087e8b]">{fmtIDR(Number(row.profit_amount || 0))}</td>
                        </tr>
                      ))}
                      <tr className="border-t-2 border-[#13464b] bg-[#e8fff4]">
                        <td className="px-4 py-3 font-black text-[#13464b]">Total {group.monthKey}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtID(group.totalTransactions)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalAmount)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalProviderPayment)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalMargin)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalCommission)}</td>
                        <td className="px-4 py-3 font-black text-rose-700">{fmtIDR(group.totalTransactionExpense)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalMemberDeposit)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalProviderDeposit)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalDepositGap)}</td>
                        <td className="px-4 py-3 font-black text-[#13464b]">{fmtIDR(group.totalProfit)}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            ))
          : null}
      </section>
      </div>
    </div>
  );
}

function escapeCsvCell(value: string) {
  if (/[",\n]/.test(value)) {
    return `"${value.replace(/"/g, "\"\"")}"`;
  }
  return value;
}

function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows.map((row) => row.map((cell) => escapeCsvCell(cell)).join(",")).join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const href = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = href;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(href);
}
