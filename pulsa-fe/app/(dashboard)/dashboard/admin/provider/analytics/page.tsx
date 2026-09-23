"use client";

import { useEffect, useMemo, useState } from "react";
import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { alertError, alertSuccess } from "@/components/ui/alerts";
import { fmtID } from "@/lib/format";

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

function isoDateInJakarta(date: Date): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);
  const year = parts.find((part) => part.type === "year")?.value || "0000";
  const month = parts.find((part) => part.type === "month")?.value || "01";
  const day = parts.find((part) => part.type === "day")?.value || "01";
  return `${year}-${month}-${day}`;
}

type SummaryRow = {
  provider: string;
  total: number;
  success: number;
  failed: number;
  sum_qty: number;
  sum_harga: number;
  success_nominal: number;
  deposit_amount: number;
};

type PeriodRow = {
  provider: string;
  period_start: string;
  success_count: number;
  success_nominal: number;
  deposit_amount: number;
};

type AnalyticsResponse = {
  ok?: boolean;
  error?: string;
  items?: SummaryRow[];
  daily_items?: PeriodRow[];
  monthly_items?: PeriodRow[];
};

function todayISO(): string {
  return isoDateInJakarta(new Date());
}

function threeMonthsAgoISO(): string {
  const d = new Date();
  d.setMonth(d.getMonth() - 2, 1);
  return isoDateInJakarta(d);
}

function fmtMonthLabel(value: string): string {
  if (!value) return "-";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleDateString("id-ID", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
}

function fmtDateTimeWIB(value: string): string {
  if (!value) return "-";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return `${d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
    timeZone: "Asia/Jakarta",
  })} WIB`;
}

export default function ProviderAnalyticsPage() {
  const [items, setItems] = useState<SummaryRow[]>([]);
  const [dailyItems, setDailyItems] = useState<PeriodRow[]>([]);
  const [monthlyItems, setMonthlyItems] = useState<PeriodRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshingCache, setRefreshingCache] = useState(false);
  const [err, setErr] = useState("");
  const [from, setFrom] = useState(threeMonthsAgoISO());
  const [to, setTo] = useState(todayISO());
  const [provider, setProvider] = useState("");

  async function load() {
    setLoading(true);
    setErr("");
    try {
      const qs = new URLSearchParams();
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);

      const r = await fetch(`/api/admin/provider/analytics?${qs.toString()}`, {
        headers: authHeader(),
      });
      const j: AnalyticsResponse = await r.json().catch(() => ({}));
      if (!r.ok || j.ok === false) {
        throw new Error(j.error || "Gagal memuat analytics provider.");
      }
      setItems(Array.isArray(j.items) ? j.items : []);
      setDailyItems(Array.isArray(j.daily_items) ? j.daily_items : []);
      setMonthlyItems(Array.isArray(j.monthly_items) ? j.monthly_items : []);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal memuat analytics provider.";
      setErr(message);
    } finally {
      setLoading(false);
    }
  }

  async function refreshThreeMonthCache() {
    setRefreshingCache(true);
    setErr("");
    try {
      const res = await fetch("/api/admin/provider/analytics/refresh-cache?months=3", {
        method: "POST",
        headers: authHeader(),
        cache: "no-store",
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok || !json?.ok) {
        throw new Error(json?.error || "Gagal update data 3 bulan.");
      }
      await load();
      await alertSuccess(`Data provider 3 bulan berhasil diupdate. Baris cache: ${fmtID(Number(json?.item?.refreshed_rows || 0))}.`);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal update data 3 bulan.";
      setErr(message);
      await alertError(message);
    } finally {
      setRefreshingCache(false);
    }
  }

  useEffect(() => {
    load();
     
  }, []);

  const providers = Array.from(
    new Set(
      [...items, ...dailyItems, ...monthlyItems]
        .map((x) => String(x.provider || "").trim())
        .filter(Boolean),
    ),
  ).sort((a, b) => a.localeCompare(b));

  useEffect(() => {
    if (!providers.length) return;
    if (provider && providers.includes(provider)) return;
    if (providers.includes("yuscom")) {
      setProvider("yuscom");
      return;
    }
    setProvider(providers[0]);
  }, [provider, providers]);

  const activeProvider = provider || (providers.includes("yuscom") ? "yuscom" : providers[0] || "");
  const activeSummary = items.find((x) => x.provider === activeProvider) || null;
  const filteredDaily = dailyItems.filter((x) => x.provider === activeProvider);
  const filteredMonthly = monthlyItems.filter((x) => x.provider === activeProvider);

  const totalSuccessCount = Number(activeSummary?.success || 0);
  const totalSuccessNominal = Number(activeSummary?.success_nominal || 0);
  const totalDepositAmount = Number(activeSummary?.deposit_amount || 0);
  const totalFailedCount = Number(activeSummary?.failed || 0);
  const totalAllTrx = Number(activeSummary?.total || 0);

  const monthlyCards = useMemo(
    () =>
      [...filteredMonthly]
        .sort((a, b) => String(b.period_start).localeCompare(String(a.period_start)))
        .map((row) => ({
          key: `${row.provider}-${row.period_start}`,
          monthKey: String(row.period_start || "").slice(0, 7),
          label: fmtMonthLabel(row.period_start),
          successCount: Number(row.success_count || 0),
          successNominal: Number(row.success_nominal || 0),
          depositAmount: Number(row.deposit_amount || 0),
        })),
    [filteredMonthly],
  );

  const dailyColumns: DataTableColumn<PeriodRow>[] = [
    {
      id: "period_start",
      header: "Waktu",
      tdClassName: "whitespace-nowrap text-slate-100 font-semibold",
      render: (x) => fmtDateTimeWIB(x.period_start),
    },
    {
      id: "success_count",
      header: "Trx Berhasil",
      tdClassName: "whitespace-nowrap text-emerald-300",
      render: (x) => fmtID(Number(x.success_count || 0)),
    },
    {
      id: "success_nominal",
      header: "Nominal Berhasil",
      tdClassName: "whitespace-nowrap text-cyan-200",
      render: (x) => `Rp ${fmtID(Number(x.success_nominal || 0))}`,
    },
    {
      id: "deposit_amount",
      header: "Deposit Provider",
      tdClassName: "whitespace-nowrap text-amber-200",
      render: (x) => `Rp ${fmtID(Number(x.deposit_amount || 0))}`,
    },
  ];

  return (
    <div className="space-y-5 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Provider Analytics</div>
          <div className="text-sm text-muted-foreground">
            Ringkasan transaksi berhasil dan deposit provider per bulan.
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <Input className="h-10 w-40" value={from} onChange={(e) => setFrom(e.target.value)} placeholder="from yyyy-mm-dd" />
          <Input className="h-10 w-40" value={to} onChange={(e) => setTo(e.target.value)} placeholder="to yyyy-mm-dd" />
          <Button variant="primary" className="h-10" onClick={load} disabled={loading}>
            {loading ? "Loading..." : "Filter"}
          </Button>
          <Button variant="outline" className="h-10" onClick={() => void refreshThreeMonthCache()} disabled={loading || refreshingCache}>
            <RefreshCw className={`mr-2 h-4 w-4 ${refreshingCache ? "animate-spin" : ""}`} />
            {refreshingCache ? "Updating..." : "Update Data 3 Bulan"}
          </Button>
        </div>
      </div>

      {err ? (
        <div className="rounded-md border border-rose-400/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {err}
        </div>
      ) : null}

      <div className="overflow-x-auto">
        <div className="flex min-w-max gap-2 rounded-xl border border-white/10 bg-slate-950/40 p-2">
          {providers.map((name) => {
            const active = name === activeProvider;
            return (
              <button
                key={name}
                type="button"
                onClick={() => setProvider(name)}
                className={[
                  "rounded-lg px-4 py-2 text-sm font-medium capitalize transition",
                  active
                    ? "bg-cyan-500/20 text-cyan-100 ring-1 ring-cyan-400/40"
                    : "bg-white/5 text-slate-300 hover:bg-white/10 hover:text-slate-100",
                ].join(" ")}
              >
                {name}
              </button>
            );
          })}
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-5">
        <div className="rounded-md border border-cyan-400/20 bg-cyan-500/10 p-4">
          <div className="text-xs uppercase tracking-wide text-cyan-100/70">Provider Aktif</div>
          <div className="mt-2 text-2xl font-semibold text-cyan-100 capitalize">{activeProvider || "-"}</div>
        </div>
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Trx Berhasil</div>
          <div className="mt-2 text-2xl font-semibold text-emerald-300">{fmtID(totalSuccessCount)}</div>
        </div>
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Nominal Berhasil</div>
          <div className="mt-2 text-2xl font-semibold text-cyan-200">Rp {fmtID(totalSuccessNominal)}</div>
        </div>
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Deposit Provider</div>
          <div className="mt-2 text-2xl font-semibold text-amber-200">Rp {fmtID(totalDepositAmount)}</div>
        </div>
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Trx Gagal</div>
          <div className="mt-2 text-2xl font-semibold text-rose-300">{fmtID(totalFailedCount)}</div>
        </div>
        <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
          <div className="text-xs uppercase tracking-wide text-slate-400">Total Trx</div>
          <div className="mt-2 text-2xl font-semibold text-slate-100">{fmtID(totalAllTrx)}</div>
        </div>
      </div>

      <section className="space-y-3">
        <div>
          <div className="text-base font-semibold text-slate-100">Harian Per Bulan</div>
          <div className="text-sm text-slate-400">Data harian dipisah per bulan, dan setiap bulan punya total sendiri.</div>
        </div>
        {monthlyCards.map((month) => {
          const monthRows = filteredDaily
            .filter((row) => String(row.period_start || "").slice(0, 7) === month.monthKey)
            .sort((a, b) => String(b.period_start).localeCompare(String(a.period_start)));
          return (
            <div key={month.key} className="space-y-2 rounded-xl border border-white/10 bg-slate-950/40 p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="text-base font-semibold text-slate-100">{month.label}</div>
                  <div className="text-sm text-slate-400">Rincian harian dan total untuk bulan ini.</div>
                </div>
                <div className="grid gap-2 sm:grid-cols-3">
                  <div className="rounded-lg border border-white/10 bg-white/5 px-3 py-2">
                    <div className="text-[11px] uppercase tracking-wide text-slate-500">Total Trx Berhasil</div>
                    <div className="mt-1 text-base font-semibold text-emerald-300">{fmtID(month.successCount)}</div>
                  </div>
                  <div className="rounded-lg border border-white/10 bg-white/5 px-3 py-2">
                    <div className="text-[11px] uppercase tracking-wide text-slate-500">Total Nominal</div>
                    <div className="mt-1 text-base font-semibold text-cyan-200">Rp {fmtID(month.successNominal)}</div>
                  </div>
                  <div className="rounded-lg border border-white/10 bg-white/5 px-3 py-2">
                    <div className="text-[11px] uppercase tracking-wide text-slate-500">Total Deposit</div>
                    <div className="mt-1 text-base font-semibold text-amber-200">Rp {fmtID(month.depositAmount)}</div>
                  </div>
                </div>
              </div>
              <DataTable
                columns={dailyColumns}
                rows={monthRows}
                rowKey={(x, i) => `${x.provider}-${x.period_start}-${i}`}
                minWidthClassName="min-w-[860px]"
                emptyText="Tidak ada data harian pada bulan ini."
                loading={loading}
              />
            </div>
          );
        })}
        {!monthlyCards.length ? (
          <div className="rounded-xl border border-dashed border-white/10 bg-slate-950/30 p-5 text-sm text-slate-400">
            Tidak ada data bulanan untuk provider ini.
          </div>
        ) : null}
      </section>
    </div>
  );
}
