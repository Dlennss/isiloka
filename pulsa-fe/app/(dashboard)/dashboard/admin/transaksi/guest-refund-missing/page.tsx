"use client";

import { useEffect, useState } from "react";
import { Search, SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type AuditRow = {
  app_order_id: number;
  invoice_id: string;
  buyer_type: string;
  status: string;
  member_id?: number;
  guest_nama?: string;
  guest_email?: string;
  guest_phone?: string;
  harga_final: number;
  dibuat_pada: string;
  diubah_pada: string;
  provider_trx_id?: number;
  provider_info?: string;
};

type ApiResponse = {
  ok?: boolean;
  items?: AuditRow[];
  total?: number;
  limit?: number;
  offset?: number;
  page?: number;
  total_pages?: number;
  error?: string;
};

const DEFAULT_LIMIT = 10;

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  return t ? { Authorization: `Bearer ${t}` } : {};
}

function fmtDate(s: string) {
  try {
    return new Date(s).toLocaleString("id-ID");
  } catch {
    return s;
  }
}

function fmtMoney(n: number) {
  return new Intl.NumberFormat("id-ID").format(Number(n || 0));
}

export default function AdminGuestRefundMissingPage() {
  const [rows, setRows] = useState<AuditRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [totalPages, setTotalPages] = useState(1);

  const [invoiceID, setInvoiceID] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const activeFilterCount = [invoiceID, from, to].filter((v) => v.trim()).length;

  const page = Math.floor(offset / limit) + 1;

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(limit));
      qs.set("offset", String(nextOffset));
      if (invoiceID.trim()) qs.set("invoice_id", invoiceID.trim());
      if (from.trim()) qs.set("from", from.trim());
      if (to.trim()) qs.set("to", to.trim());

      const r = await fetch(`/api/admin/audit/app-order/guest-refund-missing?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j: ApiResponse = await r.json().catch(() => ({}));
      if (!r.ok || !j.ok) {
        setRows([]);
        setTotal(0);
        setTotalPages(1);
        return;
      }

      setRows(Array.isArray(j.items) ? j.items : []);
      setTotal(Number(j.total || 0));
      setTotalPages(Math.max(1, Number(j.total_pages || 1)));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(0);
     
  }, []);

  const columns: DataTableColumn<AuditRow>[] = [
    {
      id: "dibuat",
      header: "Dibuat",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => fmtDate(x.dibuat_pada),
    },
    {
      id: "invoice",
      header: "Invoice",
      tdClassName: "whitespace-nowrap font-mono text-xs text-slate-200",
      render: (x) => x.invoice_id,
    },
    {
      id: "appOrderID",
      header: "App Order ID",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => String(x.app_order_id),
    },
    {
      id: "guest",
      header: "Guest",
      tdClassName: "text-slate-200",
      render: (x) => (
        <div className="space-y-0.5">
          <div>{x.guest_nama || "-"}</div>
          <div className="text-xs text-slate-400">{x.guest_email || "-"}</div>
          <div className="text-xs text-slate-400">{x.guest_phone || "-"}</div>
        </div>
      ),
    },
    {
      id: "harga",
      header: "Harga Final",
      tdClassName: "whitespace-nowrap text-slate-100",
      render: (x) => `Rp ${fmtMoney(x.harga_final)}`,
    },
    {
      id: "provider",
      header: "Provider Trx",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => x.provider_info || "-",
    },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap text-rose-300",
      render: (x) => x.status,
    },
  ];

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Audit Guest Refund Missing</div>
        <div className="text-sm text-muted-foreground">Order guest gagal yang belum punya ticket refund claim.</div>
      </div>

      <div className="md:hidden">
        <Button
          type="button"
          variant="outline"
          className="h-10 w-full justify-between border-white/15 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
          onClick={() => setMobileFilterOpen((v) => !v)}
        >
          <span className="inline-flex items-center gap-2">
            <SlidersHorizontal className="h-4 w-4" />
            {mobileFilterOpen ? "Tutup Filter" : "Buka Filter"}
          </span>
          {activeFilterCount > 0 ? (
            <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span>
          ) : null}
        </Button>
      </div>

      <div
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/12 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-cyan-950/25 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-5">
          <div className="relative lg:col-span-2">
            <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              value={invoiceID}
              onChange={(e) => setInvoiceID(e.target.value)}
              placeholder="Filter invoice_id"
              className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 pr-3 pl-9 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
            />
          </div>

          <input
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
          />

          <input
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
          />

          <div className="grid grid-cols-2 gap-2">
            <Button
              type="button"
              variant="primary"
              className="h-10"
              disabled={loading}
              onClick={() => {
                const nextOffset = 0;
                setOffset(nextOffset);
                void load(nextOffset);
                setMobileFilterOpen(false);
              }}
            >
              {loading ? "Memuat..." : "Terapkan"}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={loading}
              onClick={() => {
                setInvoiceID("");
                setFrom("");
                setTo("");
                const nextOffset = 0;
                setOffset(nextOffset);
                void load(nextOffset);
                setMobileFilterOpen(false);
              }}
            >
              Reset
            </Button>
          </div>
        </div>
      </div>

      <div className="text-sm text-slate-300">Total missing refund ticket: {total}</div>

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(x) => `${x.app_order_id}-${x.invoice_id}`}
        emptyText={loading ? "Memuat..." : "Tidak ada data."}
        minWidthClassName="min-w-260"
        showRowNumber
        rowNumberStart={offset + 1}
        pagination={{
          page,
          totalPages,
          onPrev: () => {
            const nextOffset = Math.max(0, offset - limit);
            setOffset(nextOffset);
            void load(nextOffset);
          },
          onNext: () => {
            const nextOffset = offset + limit;
            setOffset(nextOffset);
            void load(nextOffset);
          },
          disablePrev: loading || offset <= 0,
          disableNext: loading || page >= totalPages,
        }}
      />
    </div>
  );
}
