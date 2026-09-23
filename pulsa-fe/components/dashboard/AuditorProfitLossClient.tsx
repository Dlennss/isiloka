"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { MonthField } from "@/components/ui/date-field";
import { authHeader, currentMonthJakarta, downloadXlsx, fmtIDR } from "@/components/dashboard/auditorHelpers";

type Line = { code: string; label: string; amount: number };
type Opening = { account_code: string; account_name: string; account_group: string; amount: number };
type Report = {
  month: string;
  generated_at_wib: string;
  revenue_lines: Line[];
  cost_lines: Line[];
  expense_lines: Line[];
  other_income_lines: Line[];
  other_expense_lines: Line[];
  asset_lines: Line[];
  liability_lines: Line[];
  equity_lines: Line[];
  net_profit: number;
  current_year_profit: number;
  total_asset: number;
  total_liability: number;
  total_equity: number;
  total_liability_equity: number;
  opening_balances: Opening[];
};

function Section({ title, items }: { title: string; items: Line[] }) {
  return (
    <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
      <div className="mb-2 text-sm font-semibold uppercase tracking-wide text-slate-300">{title}</div>
      <div className="space-y-2">
        {items.map((item) => (
          <div key={item.code} className="flex items-center justify-between gap-4 text-sm">
            <div className="text-slate-300">{item.label}</div>
            <div className="font-medium text-slate-100">{fmtIDR(item.amount)}</div>
          </div>
        ))}
        {items.length === 0 ? <div className="text-sm text-slate-500">Belum ada data.</div> : null}
      </div>
    </div>
  );
}

export function AuditorProfitLossClient() {
  const [month, setMonth] = useState(currentMonthJakarta());
  const [item, setItem] = useState<Report | null>(null);
  const [loading, setLoading] = useState(false);

  async function load(nextMonth = month) {
    setLoading(true);
    try {
      const qs = new URLSearchParams({ month: nextMonth });
      const r = await fetch(`/api/admin/auditor/profit-loss?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
      const j = await r.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat rugi laba");
      setItem(j.item || null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function exportFile() {
    if (!item) return;
    const rows = [
      ...item.revenue_lines.map((v) => ({ kelompok: "Revenue", akun: v.label, amount: v.amount })),
      ...item.cost_lines.map((v) => ({ kelompok: "Cost", akun: v.label, amount: v.amount })),
      ...item.expense_lines.map((v) => ({ kelompok: "Expense", akun: v.label, amount: v.amount })),
      ...item.other_income_lines.map((v) => ({ kelompok: "Other Income", akun: v.label, amount: v.amount })),
      ...item.other_expense_lines.map((v) => ({ kelompok: "Other Expense", akun: v.label, amount: v.amount })),
      { kelompok: "Summary", akun: "Net Profit", amount: item.net_profit },
      { kelompok: "Summary", akun: "Total Asset", amount: item.total_asset },
      { kelompok: "Summary", akun: "Total Liability", amount: item.total_liability },
      { kelompok: "Summary", akun: "Total Equity", amount: item.total_equity },
    ];
    await downloadXlsx(`auditor-rugi-laba-${item.month}.xlsx`, rows, "RugiLaba");
  }

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <div className="text-lg font-semibold tracking-tight">Rugi Laba</div>
          <div className="text-sm text-muted-foreground">Laporan bulanan WIB berbasis transaksi, komisi, pengeluaran internal, saldo member, dan saldo provider.</div>
        </div>
        <div className="flex items-end gap-2">
          <MonthField label="Bulan" className="h-10 w-44" value={month} onChange={(e) => setMonth(e.target.value)} />
          <Button className="h-10" onClick={() => void load()} disabled={loading}>{loading ? "Memuat..." : "Terapkan"}</Button>
          <Button variant="outline" className="h-10" disabled={!item} onClick={() => void exportFile()}>Download XLSX</Button>
        </div>
      </div>

      {item ? (
        <>
          <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {[
              ["Net Profit", fmtIDR(item.net_profit)],
              ["Profit Tahun Berjalan", fmtIDR(item.current_year_profit)],
              ["Total Asset", fmtIDR(item.total_asset)],
              ["Total Liab + Equity", fmtIDR(item.total_liability_equity)],
            ].map(([label, value]) => (
              <div key={label} className="rounded-md border border-white/15 bg-slate-950/50 p-4">
                <div className="text-xs uppercase tracking-wide text-slate-400">{label}</div>
                <div className="mt-2 text-lg font-semibold">{value}</div>
              </div>
            ))}
          </section>

          <section className="grid gap-4 xl:grid-cols-2">
            <Section title="Revenue" items={item.revenue_lines} />
            <Section title="Cost" items={item.cost_lines} />
            <Section title="Expense" items={item.expense_lines} />
            <Section title="Other Income" items={item.other_income_lines} />
            <Section title="Other Expense" items={item.other_expense_lines} />
            <Section title="Asset" items={item.asset_lines} />
            <Section title="Liability" items={item.liability_lines} />
            <Section title="Equity" items={item.equity_lines} />
          </section>
        </>
      ) : null}
    </div>
  );
}
