"use client";

import { useEffect, useMemo, useState } from "react";
import { RefreshCw, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
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

function todayISO(): string {
  return isoDateInJakarta(new Date());
}

function fmtDate(value: string): string {
  if (!value) return "-";
  const raw = value.slice(0, 10);
  const [year, month, day] = raw.split("-");
  if (!year || !month || !day) return value;
  return `${day}/${month}/${year}`;
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
    hour12: false,
    timeZone: "Asia/Jakarta",
  })} WIB`;
}

type ProductSuccessRow = {
  period_start: string;
  internal_sku: string;
  product_name: string;
  group_name: string;
  success_count: number;
  total_qty: number;
  total_qty_provider: number;
  total_provider_price: number;
  total_member_price: number;
  total_margin: number;
  provider_count: number;
  providers: string;
  first_success_at: string;
  last_success_at: string;
};

type ProductSuccessSummary = {
  group_count: number;
  unique_sku_count: number;
  success_count: number;
  total_qty: number;
  total_qty_provider: number;
  total_provider_price: number;
  total_member_price: number;
  total_margin: number;
};

type ProductSuccessResponse = {
  items?: ProductSuccessRow[];
  total?: number;
  summary?: Partial<ProductSuccessSummary>;
  error?: string;
};

const PAGE_SIZE = 50;

const emptySummary: ProductSuccessSummary = {
  group_count: 0,
  unique_sku_count: 0,
  success_count: 0,
  total_qty: 0,
  total_qty_provider: 0,
  total_provider_price: 0,
  total_member_price: 0,
  total_margin: 0,
};

export default function DailySuccessfulProductsPage() {
  const [items, setItems] = useState<ProductSuccessRow[]>([]);
  const [summary, setSummary] = useState<ProductSuccessSummary>(emptySummary);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [from, setFrom] = useState(todayISO());
  const [to, setTo] = useState(todayISO());
  const [q, setQ] = useState("");
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);

  async function load(nextOffset = offset) {
    setLoading(true);
    setError("");
    try {
      const qs = new URLSearchParams();
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);
      const query = q.trim();
      if (query) qs.set("q", query);
      qs.set("limit", String(PAGE_SIZE));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/provider/products/daily-success?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ProductSuccessResponse = await r.json().catch(() => ({}));
      if (!r.ok) {
        throw new Error(j.error || "Gagal memuat data produk sukses.");
      }
      const nextSummary = { ...emptySummary, ...(j.summary || {}) };
      setItems(Array.isArray(j.items) ? j.items : []);
      setSummary(nextSummary);
      setTotal(Number(j.total || nextSummary.group_count || 0));
      setOffset(nextOffset);
    } catch (err) {
      setItems([]);
      setSummary(emptySummary);
      setTotal(0);
      setError(err instanceof Error ? err.message : "Gagal memuat data produk sukses.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load(0);
     
  }, []);

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const totalMargin = Number(summary.total_margin || 0);

  const cards = useMemo(
    () => [
      { label: "Grup Harian", value: fmtID(summary.group_count), tone: "text-cyan-100" },
      { label: "SKU Internal", value: fmtID(summary.unique_sku_count), tone: "text-sky-200" },
      { label: "Trx Sukses", value: fmtID(summary.success_count), tone: "text-emerald-300" },
      { label: "Nominal", value: `Rp ${fmtID(summary.total_qty)}`, tone: "text-cyan-200" },
      { label: "Margin", value: `Rp ${fmtID(totalMargin)}`, tone: totalMargin >= 0 ? "text-emerald-300" : "text-rose-300" },
    ],
    [summary, totalMargin],
  );

  const columns: DataTableColumn<ProductSuccessRow>[] = [
    {
      id: "period",
      header: "Tanggal",
      tdClassName: "whitespace-nowrap text-slate-100 font-semibold",
      render: (x) => fmtDate(x.period_start),
    },
    {
      id: "sku",
      header: "SKU Internal",
      tdClassName: "min-w-48",
      render: (x) => (
        <div>
          <div className="font-semibold text-cyan-100">{x.internal_sku || "-"}</div>
          <div className="mt-1 text-xs text-slate-400">{x.product_name || "-"}</div>
        </div>
      ),
    },
    {
      id: "group",
      header: "Grup",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => x.group_name || "-",
    },
    {
      id: "success_count",
      header: "Sukses",
      tdClassName: "whitespace-nowrap text-emerald-300 text-right",
      render: (x) => fmtID(x.success_count),
    },
    {
      id: "total_qty",
      header: "Nominal",
      tdClassName: "whitespace-nowrap text-cyan-200 text-right",
      render: (x) => `Rp ${fmtID(x.total_qty)}`,
    },
    {
      id: "provider_price",
      header: "Biaya Provider",
      tdClassName: "whitespace-nowrap text-amber-200 text-right",
      render: (x) => `Rp ${fmtID(x.total_provider_price)}`,
    },
    {
      id: "member_price",
      header: "Harga Member",
      tdClassName: "whitespace-nowrap text-slate-100 text-right",
      render: (x) => `Rp ${fmtID(x.total_member_price)}`,
    },
    {
      id: "margin",
      header: "Margin",
      tdClassName: "whitespace-nowrap text-right",
      render: (x) => {
        const margin = Number(x.total_margin || 0);
        return <span className={margin >= 0 ? "text-emerald-300" : "text-rose-300"}>Rp {fmtID(margin)}</span>;
      },
    },
    {
      id: "providers",
      header: "Provider",
      tdClassName: "min-w-72 text-slate-300",
      render: (x) => (
        <div>
          <div className="text-slate-100">{fmtID(x.provider_count)} provider</div>
          <div className="mt-1 text-xs text-slate-400">{x.providers || "-"}</div>
        </div>
      ),
    },
    {
      id: "last_success_at",
      header: "Sukses Terakhir",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => fmtDateTimeWIB(x.last_success_at),
    },
  ];

  return (
    <div className="space-y-5 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Produk Sukses Harian</div>
          <div className="text-sm text-muted-foreground">
            Transaksi sukses dikelompokkan berdasarkan SKU internal.
          </div>
        </div>

        <div className="grid w-full gap-2 sm:w-auto sm:grid-cols-[140px_140px_minmax(180px,260px)_auto_auto]">
          <Input type="date" className="h-10" value={from} onChange={(e) => setFrom(e.target.value)} placeholder="from yyyy-mm-dd" />
          <Input type="date" className="h-10" value={to} onChange={(e) => setTo(e.target.value)} placeholder="to yyyy-mm-dd" />
          <Input className="h-10" value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari SKU / nama / grup" />
          <Button
            variant="primary"
            className="h-10"
            onClick={() => load(0)}
            disabled={loading}
          >
            <Search className="h-4 w-4" />
            Filter
          </Button>
          <Button variant="outline" className="h-10" onClick={() => load(offset)} disabled={loading} aria-label="Muat ulang">
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-5">
        {cards.map((card) => (
          <div key={card.label} className="rounded-md border border-white/15 bg-slate-950/50 p-4">
            <div className="text-xs uppercase tracking-wide text-slate-400">{card.label}</div>
            <div className={`mt-2 text-2xl font-semibold ${card.tone}`}>{card.value}</div>
          </div>
        ))}
      </div>

      {error ? (
        <div className="rounded-md border border-rose-400/25 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {error}
        </div>
      ) : null}

      <DataTable
        columns={columns}
        rows={items}
        rowKey={(x, i) => `${x.period_start}-${x.internal_sku}-${i}`}
        minWidthClassName="min-w-[1180px]"
        emptyText="Tidak ada data produk sukses."
        loading={loading}
        pagination={{
          page,
          totalPages,
          disablePrev: page <= 1,
          disableNext: page >= totalPages,
          onPrev: () => load(Math.max(0, offset - PAGE_SIZE)),
          onNext: () => load(offset + PAGE_SIZE),
          onPageChange: (targetPage) => load(Math.max(0, (targetPage - 1) * PAGE_SIZE)),
        }}
      />
    </div>
  );
}
