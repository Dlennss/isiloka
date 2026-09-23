"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DateField } from "@/components/ui/date-field";
import { authHeader, downloadXlsx, fmtDateTime, fmtIDR, todayJakarta } from "@/components/dashboard/auditorHelpers";

type DetailRow = {
  scope: string;
  ref_id: string;
  occurred_at: string;
  status: string;
  member_nama?: string | null;
  member_email?: string | null;
  product_code: string;
  product_name: string;
  destination: string;
  provider?: string | null;
  harga_beli: number;
  harga_jual: number;
  komisi: number;
  margin: number;
  provider_ref?: string | null;
  status_note?: string | null;
};

type SummaryRow = {
  transaction_count: number;
  sales_amount: number;
  provider_amount: number;
  commission_amount: number;
  margin_amount: number;
};

export function AuditorTradingDetailClient() {
  const today = todayJakarta();
  const pageSize = 100;
  const [scope, setScope] = useState("all");
  const [from, setFrom] = useState(today);
  const [to, setTo] = useState(today);
  const [refID, setRefID] = useState("");
  const [items, setItems] = useState<DetailRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [page, setPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [totals, setTotals] = useState({ count: 0, hargaBeli: 0, hargaJual: 0, komisi: 0, margin: 0 });

  async function load(nextPage = 1) {
    setLoading(true);
    try {
      const qs = new URLSearchParams({ scope, from, to, limit: String(pageSize), offset: String((nextPage - 1) * pageSize) });
      if (refID) qs.set("ref_id", refID);
      const summaryQs = new URLSearchParams({ scope, from, to, period: "day" });
      const [r, summaryR] = await Promise.all([
        fetch(`/api/admin/auditor/trading/details?${qs.toString()}`, { headers: authHeader(), cache: "no-store" }),
        fetch(`/api/admin/auditor/trading/summary?${summaryQs.toString()}`, { headers: authHeader(), cache: "no-store" }),
      ]);
      const j = await r.json().catch(() => ({}));
      const summaryJ = await summaryR.json().catch(() => ({}));
      if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal memuat transaksi jual beli");
      if (!summaryR.ok || !summaryJ?.ok) throw new Error(summaryJ?.error || "Gagal memuat total transaksi jual beli");
      const rows = Array.isArray(j.items) ? j.items : [];
      const summaryItems = Array.isArray(summaryJ.items) ? (summaryJ.items as SummaryRow[]) : [];
      const aggregate = summaryItems.reduce(
        (acc, row) => {
          acc.count += Number(row.transaction_count || 0);
          acc.hargaBeli += Number(row.provider_amount || 0);
          acc.hargaJual += Number(row.sales_amount || 0);
          acc.komisi += Number(row.commission_amount || 0);
          acc.margin += Number(row.margin_amount || 0);
          return acc;
        },
        { count: 0, hargaBeli: 0, hargaJual: 0, komisi: 0, margin: 0 }
      );
      setItems(rows);
      setTotalItems(Number(j.total || 0));
      setTotals(aggregate);
      setPage(nextPage);
    } finally {
      setLoading(false);
    }
  }

  async function handleDownloadAll() {
    setDownloading(true);
    try {
      const allRows: DetailRow[] = [];
      let offset = 0;
      const batchSize = 500;
      while (true) {
        const qs = new URLSearchParams({ scope, from, to, limit: String(batchSize), offset: String(offset) });
        if (refID) qs.set("ref_id", refID);
        const r = await fetch(`/api/admin/auditor/trading/details?${qs.toString()}`, { headers: authHeader(), cache: "no-store" });
        const j = await r.json().catch(() => ({}));
        if (!r.ok || !j?.ok) throw new Error(j?.error || "Gagal mengunduh transaksi jual beli");
        const rows = Array.isArray(j.items) ? (j.items as DetailRow[]) : [];
        allRows.push(...rows);
        if (rows.length < batchSize) break;
        offset += batchSize;
      }
      await downloadXlsx(
        `auditor-jual-beli-${from}-${to}.xlsx`,
        allRows.map((row) => ({
          ref_id: row.ref_id,
          waktu: row.occurred_at,
          scope: row.scope,
          member: row.member_nama || row.member_email || "",
          kode_produk: row.product_code,
          nama_produk: row.product_name,
          tujuan: row.destination,
          provider: row.provider || "",
          harga_beli: row.harga_beli,
          harga_jual: row.harga_jual,
          komisi: row.komisi,
          margin: row.margin,
          provider_ref: row.provider_ref || "",
          status: row.status,
        })),
        "JualBeli"
      );
    } finally {
      setDownloading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Transaksi Jual Beli</div>
        <div className="text-sm text-muted-foreground">Harga beli provider, harga jual ke member, komisi, dan margin per ref id.</div>
      </div>
      <div className="flex flex-wrap items-end gap-2">
        <select className="h-10 rounded-md border border-white/15 bg-slate-950 px-3" value={scope} onChange={(e) => setScope(e.target.value)}>
          <option value="all">Gabungan</option>
          <option value="retail">Retail</option>
          <option value="h2h">H2H</option>
        </select>
        <DateField label="Dari" className="h-10 w-44" value={from} onChange={(e) => setFrom(e.target.value)} />
        <DateField label="Sampai" className="h-10 w-44" value={to} onChange={(e) => setTo(e.target.value)} />
        <Input className="h-10 w-64" value={refID} onChange={(e) => setRefID(e.target.value)} placeholder="Cari ref id / invoice" />
        <Button className="h-10" onClick={() => void load(1)} disabled={loading}>{loading ? "Memuat..." : "Terapkan"}</Button>
        <Button
          variant="outline"
          className="h-10"
          disabled={totalItems === 0 || downloading}
          onClick={() => void handleDownloadAll()}
        >
          {downloading ? "Mengunduh..." : "Download XLSX"}
        </Button>
      </div>

      <div className="overflow-x-auto rounded-md border border-white/15 bg-slate-950/50">
        <table className="min-w-full text-sm">
          <thead className="text-left text-slate-400">
            <tr className="border-b border-white/10">
              <th className="px-3 py-2">Ref ID</th>
              <th className="px-3 py-2">Waktu</th>
              <th className="px-3 py-2">Produk</th>
              <th className="px-3 py-2">Member</th>
              <th className="px-3 py-2">Provider</th>
              <th className="px-3 py-2 text-right">Harga Beli</th>
              <th className="px-3 py-2 text-right">Harga Jual</th>
              <th className="px-3 py-2 text-right">Komisi</th>
              <th className="px-3 py-2 text-right">Margin</th>
            </tr>
          </thead>
          <tbody>
            {items.map((row) => (
              <tr key={`${row.scope}-${row.ref_id}-${row.provider_ref || ""}`} className="border-b border-white/5 align-top">
                <td className="px-3 py-2 font-mono text-xs">{row.ref_id}<div className="mt-1 text-[11px] text-slate-500">{row.status}</div></td>
                <td className="px-3 py-2">{fmtDateTime(row.occurred_at)}</td>
                <td className="px-3 py-2">{row.product_code}<div className="mt-1 text-xs text-slate-500">{row.product_name} • {row.destination}</div></td>
                <td className="px-3 py-2">{row.member_nama || "-"}<div className="mt-1 text-xs text-slate-500">{row.member_email || "-"}</div></td>
                <td className="px-3 py-2">{row.provider || "-"}</td>
                <td className="px-3 py-2 text-right">{fmtIDR(row.harga_beli)}</td>
                <td className="px-3 py-2 text-right">{fmtIDR(row.harga_jual)}</td>
                <td className="px-3 py-2 text-right">{fmtIDR(row.komisi)}</td>
                <td className="px-3 py-2 text-right">{fmtIDR(row.margin)}</td>
              </tr>
            ))}
            {items.length === 0 ? <tr><td className="px-3 py-6 text-center text-slate-400" colSpan={9}>Tidak ada data.</td></tr> : null}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-white/15 bg-slate-950/50 px-4 py-3 text-sm">
        <div className="text-slate-300">
          Menampilkan <span className="font-semibold text-white">{items.length}</span> data dari total <span className="font-semibold text-white">{totalItems}</span> transaksi sukses.
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" className="h-9" disabled={loading || page <= 1} onClick={() => void load(page - 1)}>
            Sebelumnya
          </Button>
          <div className="min-w-24 text-center text-slate-300">Halaman {page}</div>
          <Button variant="outline" className="h-9" disabled={loading || page * pageSize >= totalItems} onClick={() => void load(page + 1)}>
            Berikutnya
          </Button>
        </div>
      </div>

      <div className="rounded-md border border-white/15 bg-slate-950/50 p-4">
        <div className="mb-3 text-sm font-semibold text-slate-200">Total</div>
        <div className="grid gap-3 md:grid-cols-5">
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Jumlah Transaksi</div>
            <div className="mt-1 text-base font-semibold text-white">{totals.count}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Harga Beli</div>
            <div className="mt-1 text-base font-semibold text-white">{fmtIDR(totals.hargaBeli)}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Harga Jual</div>
            <div className="mt-1 text-base font-semibold text-white">{fmtIDR(totals.hargaJual)}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Komisi</div>
            <div className="mt-1 text-base font-semibold text-white">{fmtIDR(totals.komisi)}</div>
          </div>
          <div className="rounded-md border border-white/10 bg-slate-900/70 p-3">
            <div className="text-xs uppercase tracking-wide text-slate-400">Total Margin</div>
            <div className="mt-1 text-base font-semibold text-emerald-300">{fmtIDR(totals.margin)}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
