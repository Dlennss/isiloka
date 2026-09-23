"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { DateField } from "@/components/ui/date-field";
import { authHeader, downloadXlsx, fmtDateOnly, fmtIDR, todayJakarta } from "@/components/dashboard/auditorHelpers";
import { fmtID } from "@/lib/format";

type SummaryRow = {
  scope: string;
  period_key: string;
  transaction_count: number;
  sales_amount: number;
  provider_amount: number;
  commission_amount: number;
  margin_amount: number;
};

type MonthGroup = {
  monthKey: string;
  rows: SummaryRow[];
  totalTransactions: number;
  totalSales: number;
  totalProvider: number;
  totalCommission: number;
  totalMargin: number;
};

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

function threeMonthsAgoDate() {
  const now = new Date(`${todayJakarta()}T12:00:00+07:00`);
  now.setMonth(now.getMonth() - 2);
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function monthBucket(periodKey: string, period: string) {
  if (!periodKey) return "unknown";
  if (period === "day") return periodKey.slice(0, 7);
  return periodKey.slice(0, 7);
}

export function AuditorTradingSummaryClient() {
  const today = todayJakarta();
  const [scope, setScope] = useState("all");
  const period = "day";
  const [from, setFrom] = useState(threeMonthsAgoDate());
  const [to, setTo] = useState(today);
  const [items, setItems] = useState<SummaryRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");

  async function load(nextFrom = from, nextTo = to, nextScope = scope) {
    setLoading(true);
    setErr("");
    try {
      const qs = new URLSearchParams({ period, scope: nextScope });
      if (nextFrom) qs.set("from", nextFrom);
      if (nextTo) qs.set("to", nextTo);
      const r = await fetch(`/api/admin/auditor/trading/summary?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) {
        throw new Error(j?.error || "Gagal memuat transaksi auditor");
      }
      setItems(Array.isArray(j.items) ? j.items : []);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat transaksi auditor");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(threeMonthsAgoDate(), today, scope);
  }, [period]);

  const summary = useMemo(
    () => ({
      trx: items.reduce((sum, row) => sum + Number(row.transaction_count || 0), 0),
      sales: items.reduce((sum, row) => sum + Number(row.sales_amount || 0), 0),
      provider: items.reduce((sum, row) => sum + Number(row.provider_amount || 0), 0),
      commission: items.reduce((sum, row) => sum + Number(row.commission_amount || 0), 0),
      margin: items.reduce((sum, row) => sum + Number(row.margin_amount || 0), 0),
    }),
    [items],
  );

  const groupedMonths = useMemo<MonthGroup[]>(() => {
    const bucket = new Map<string, SummaryRow[]>();
    for (const row of items) {
      const key = monthBucket(row.period_key || "", period);
      const arr = bucket.get(key) || [];
      arr.push(row);
      bucket.set(key, arr);
    }
    return Array.from(bucket.entries())
      .sort((a, b) => b[0].localeCompare(a[0]))
      .map(([monthKey, rows]) => {
        const ordered = [...rows].sort((a, b) => b.period_key.localeCompare(a.period_key));
        return {
          monthKey,
          rows: ordered,
          totalTransactions: ordered.reduce((sum, row) => sum + Number(row.transaction_count || 0), 0),
          totalSales: ordered.reduce((sum, row) => sum + Number(row.sales_amount || 0), 0),
          totalProvider: ordered.reduce((sum, row) => sum + Number(row.provider_amount || 0), 0),
          totalCommission: ordered.reduce((sum, row) => sum + Number(row.commission_amount || 0), 0),
          totalMargin: ordered.reduce((sum, row) => sum + Number(row.margin_amount || 0), 0),
        };
      });
  }, [items, period]);

  async function exportFile() {
    await downloadXlsx(
      `auditor-transaksi-${period}-${from || "all"}-${to || "all"}.xlsx`,
      items.map((row) => ({
        bulan: monthBucket(row.period_key, period),
        periode: row.period_key,
        scope: row.scope,
        transaksi: row.transaction_count,
        harga_jual: row.sales_amount,
        harga_beli_provider: row.provider_amount,
        komisi: row.commission_amount,
        margin: row.margin_amount,
      })),
      "Transaksi",
    );
  }

  const title = "Transaksi Harian";
  const subtitle = "Rekap transaksi harian 3 bulan terakhir, dipisah per bulan dan dilengkapi total per bulan.";

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">{title}</div>
          <div className="text-sm text-muted-foreground">{subtitle}</div>
        </div>

        <div className="flex flex-wrap gap-2">
          {(["all", "retail", "h2h"] as const).map((item) => (
            <Button key={item} variant={scope === item ? "primary" : "outline"} className="h-10" onClick={() => setScope(item)} disabled={loading}>
              {scopeLabel(item)}
            </Button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-2">
        <DateField label="Dari" className="h-10 w-44" value={from} onChange={(e) => setFrom(e.target.value)} />
        <DateField label="Sampai" className="h-10 w-44" value={to} onChange={(e) => setTo(e.target.value)} />
        <Button className="h-10" onClick={() => void load(scope === "all" ? from : from, to, scope)} disabled={loading}>
          {loading ? "Memuat..." : "Terapkan Filter"}
        </Button>
        <Button
          variant="outline"
          className="h-10"
          onClick={() => {
            const nextFrom = threeMonthsAgoDate();
            const nextTo = today;
            setFrom(nextFrom);
            setTo(nextTo);
            void load(nextFrom, nextTo, scope);
          }}
          disabled={loading}
        >
          Reset
        </Button>
        <Button variant="outline" className="h-10" onClick={() => void exportFile()} disabled={items.length === 0}>
          Download XLSX
        </Button>
      </div>

      {err ? <div className="rounded-md border border-rose-400/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">{err}</div> : null}

      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        {[
          ["Jumlah Trx", fmtID(summary.trx)],
          ["Harga Jual", fmtIDR(summary.sales)],
          ["Harga Beli Provider", fmtIDR(summary.provider)],
          ["Komisi", fmtIDR(summary.commission)],
          ["Margin", fmtIDR(summary.margin)],
        ].map(([label, value]) => (
          <div key={label} className="rounded-md border border-white/15 bg-slate-950/50 p-4 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]">
            <div className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</div>
            <div className="mt-2 text-lg font-semibold text-slate-100">{value}</div>
          </div>
        ))}
      </section>

      <section className="space-y-4">
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]">
          <div className="text-base font-semibold text-slate-100">Rekap Harian 3 Bulan</div>
          <div className="mt-1 text-sm text-slate-400">Data dipisah per bulan, dengan total transaksi, harga jual, harga beli provider, komisi, dan margin pada tiap bulan.</div>
        </div>

        {loading ? (
          <div className="rounded-md border border-white/15 bg-slate-950/50 px-4 py-6 text-sm text-slate-300">Memuat laporan transaksi...</div>
        ) : null}

        {!loading && groupedMonths.length === 0 ? (
          <div className="rounded-md border border-white/15 bg-slate-950/50 px-4 py-6 text-sm text-slate-300">Belum ada data transaksi pada periode ini.</div>
        ) : null}

        {!loading
          ? groupedMonths.map((group) => (
              <div key={group.monthKey} className="overflow-hidden rounded-md border border-white/15 bg-slate-950/50 shadow-[0_18px_42px_-26px_rgba(56,189,248,0.45)]">
                <div className="border-b border-white/10 bg-linear-to-r from-cyan-500/15 via-sky-500/10 to-indigo-500/15 px-4 py-3">
                  <div className="text-sm font-semibold uppercase tracking-wide text-slate-100">{group.monthKey}</div>
                </div>

                <div className="overflow-x-auto">
                  <table className="min-w-[980px] w-full text-sm">
                    <thead className="bg-slate-900/70 text-left">
                      <tr>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">{period === "day" ? "Tanggal" : "Periode"}</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Scope</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Jumlah Trx</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Harga Jual</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Harga Beli Provider</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Komisi</th>
                        <th className="px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-300">Margin</th>
                      </tr>
                    </thead>
                    <tbody>
                      {group.rows.map((row, index) => (
                        <tr key={`${group.monthKey}-${row.period_key}-${index}`} className="border-t border-white/10">
                          <td className="px-4 py-3 text-slate-200">{period === "day" ? fmtDateOnly(row.period_key) : row.period_key}</td>
                          <td className="px-4 py-3 text-slate-200">{scopeLabel(row.scope)}</td>
                          <td className="px-4 py-3 text-slate-200">{fmtID(row.transaction_count)}</td>
                          <td className="px-4 py-3 text-slate-100">{fmtIDR(row.sales_amount)}</td>
                          <td className="px-4 py-3 text-fuchsia-200">{fmtIDR(row.provider_amount)}</td>
                          <td className="px-4 py-3 text-amber-200">{fmtIDR(row.commission_amount)}</td>
                          <td className="px-4 py-3 font-semibold text-emerald-300">{fmtIDR(row.margin_amount)}</td>
                        </tr>
                      ))}
                      <tr className="border-t border-cyan-400/20 bg-cyan-500/5">
                        <td className="px-4 py-3 font-semibold text-slate-100">Total {group.monthKey}</td>
                        <td className="px-4 py-3 font-semibold text-slate-100">{scopeLabel(scope)}</td>
                        <td className="px-4 py-3 font-semibold text-slate-100">{fmtID(group.totalTransactions)}</td>
                        <td className="px-4 py-3 font-semibold text-slate-100">{fmtIDR(group.totalSales)}</td>
                        <td className="px-4 py-3 font-semibold text-fuchsia-200">{fmtIDR(group.totalProvider)}</td>
                        <td className="px-4 py-3 font-semibold text-amber-200">{fmtIDR(group.totalCommission)}</td>
                        <td className="px-4 py-3 font-semibold text-emerald-300">{fmtIDR(group.totalMargin)}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            ))
          : null}
      </section>
    </div>
  );
}
