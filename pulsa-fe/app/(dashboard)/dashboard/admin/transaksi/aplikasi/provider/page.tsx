"use client";

import { useEffect, useMemo, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type ProviderTrxRow = {
  id: number;
  app_order_id: number;
  invoice_id: string;
  provider: string;
  ref_id: string;
  kode_provider: string;
  produk_nama_snapshot: string;
  dest: string;
  harga_provider: number;
  status: string;
  kode_respon?: string | null;
  pesan?: string | null;
  sn?: string | null;
  dibuat_pada?: string | null;
  diubah_pada?: string | null;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

const PAGE_SIZE = 10;

function money(v: number) {
  return `Rp ${fmtID(v || 0)}`;
}

function statusTone(status: string) {
  const s = (status || "").toLowerCase();
  if (s === "success") return "border border-emerald-400 bg-emerald-100 text-emerald-900";
  if (s === "pending") return "border border-amber-400 bg-amber-100 text-amber-900";
  if (s === "failed") return "border border-rose-400 bg-rose-100 text-rose-900";
  return "border border-slate-400 bg-slate-100 text-slate-800";
}

export default function AdminAppProviderTrxPage() {
  const [items, setItems] = useState<ProviderTrxRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);

  const [draftInvoice, setDraftInvoice] = useState("");
  const [draftRefID, setDraftRefID] = useState("");
  const [draftStatus, setDraftStatus] = useState("");
  const [draftDateFrom, setDraftDateFrom] = useState("");
  const [draftDateTo, setDraftDateTo] = useState("");

  const [invoiceID, setInvoiceID] = useState("");
  const [refID, setRefID] = useState("");
  const [status, setStatus] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const activeFilterCount = [draftInvoice, draftRefID, draftStatus, draftDateFrom, draftDateTo].filter((v) => v.trim()).length;

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (invoiceID) qs.set("invoice_id", invoiceID);
      if (refID) qs.set("ref_id", refID);
      if (status) qs.set("status", status);
      if (dateFrom) qs.set("date_from", dateFrom);
      if (dateTo) qs.set("date_to", dateTo);
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/app/provider-trx?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const all: ProviderTrxRow[] = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > PAGE_SIZE);
      setItems(all.slice(0, PAGE_SIZE));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(offset);
     
  }, [offset]);

  useEffect(() => {
    void load(0);
     
  }, [invoiceID, refID, status, dateFrom, dateTo]);

  function applyFilters() {
    setOffset(0);
    setInvoiceID(draftInvoice.trim());
    setRefID(draftRefID.trim());
    setStatus(draftStatus);
    setDateFrom(draftDateFrom);
    setDateTo(draftDateTo);
  }

  function resetFilters() {
    setDraftInvoice("");
    setDraftRefID("");
    setDraftStatus("");
    setDraftDateFrom("");
    setDraftDateTo("");
    setOffset(0);
    setInvoiceID("");
    setRefID("");
    setStatus("");
    setDateFrom("");
    setDateTo("");
  }

  const columns = useMemo<DataTableColumn<ProviderTrxRow>[]>(
    () => [
      {
        id: "dibuat_pada",
        header: "Waktu",
        tdClassName: "whitespace-nowrap text-slate-100",
        render: (x) => (x.dibuat_pada ? new Date(x.dibuat_pada).toLocaleString("id-ID") : "-"),
      },
      {
        id: "invoice_id",
        header: "Invoice",
        tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
        render: (x) => x.invoice_id,
      },
      {
        id: "kode_provider",
        header: "Kode Provider",
        tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
        render: (x) => x.kode_provider || "-",
      },
      {
        id: "dest",
        header: "Tujuan",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => x.dest,
      },
      {
        id: "harga_provider",
        header: "Harga Provider",
        tdClassName: "whitespace-nowrap text-slate-200",
        render: (x) => money(x.harga_provider),
      },
      {
        id: "status",
        header: "Status",
        tdClassName: "whitespace-nowrap",
        render: (x) => (
          <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-semibold capitalize ${statusTone(x.status)}`}>
            <span className="h-1.5 w-1.5 rounded-full bg-current opacity-80" />
            {x.status}
          </span>
        ),
      },
    ],
    [],
  );

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const shouldShowPagination = offset > 0 || hasNext;

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Provider Aplikasi</div>
          <div className="text-sm text-muted-foreground">Monitoring transaksi provider untuk flow app commerce.</div>
        </div>

        <div className="w-full md:hidden">
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
              <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700">{activeFilterCount}</span>
            ) : null}
          </Button>
        </div>
      </div>

      <div
        className={`${mobileFilterOpen ? "block" : "hidden"} rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-slate-800/45 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
      >
        <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-slate-300 md:hidden">
          <SlidersHorizontal className="h-3.5 w-3.5 text-emerald-600" />
          Filter Provider
        </div>
        <div className="grid grid-cols-2 gap-2 xl:grid-cols-5">
          <div className="col-span-2 grid grid-cols-2 gap-2 xl:col-span-2">
            <div className="space-y-1">
              <Input value={draftInvoice} onChange={(e) => setDraftInvoice(e.target.value)} placeholder="Cari invoice" className="col-span-2 h-10 border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 xl:col-span-1" />
            </div>
            <div className="space-y-1">
              <Input value={draftRefID} onChange={(e) => setDraftRefID(e.target.value)} placeholder="Cari ref id" className="col-span-2 h-10 border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 xl:col-span-1" />
            </div>
          </div>
          <select value={draftStatus} onChange={(e) => setDraftStatus(e.target.value)} className="col-span-2 h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50 xl:col-span-1">
            <option value="">semua status</option>
            <option value="pending">pending</option>
            <option value="success">success</option>
            <option value="failed">failed</option>
          </select>
          <div className="col-span-2 grid grid-cols-2 gap-2 xl:col-span-2">
            <div className="space-y-1">
              <div className="text-[11px] font-medium text-slate-400 md:hidden">Dari</div>
              <Input type="date" value={draftDateFrom} onChange={(e) => setDraftDateFrom(e.target.value)} className="h-10 border-white/15 bg-slate-950/55 text-slate-100" />
            </div>
            <div className="space-y-1">
              <div className="text-[11px] font-medium text-slate-400 md:hidden">Sampai</div>
              <Input type="date" value={draftDateTo} onChange={(e) => setDraftDateTo(e.target.value)} className="h-10 border-white/15 bg-slate-950/55 text-slate-100" />
            </div>
          </div>
        </div>
        <div className="mt-2 grid grid-cols-2 gap-2">
          <Button
            onClick={() => {
              applyFilters();
              setMobileFilterOpen(false);
            }}
            className="h-10 bg-emerald-700 text-white hover:bg-emerald-600"
          >
            Terapkan
          </Button>
          <Button
            onClick={() => {
              resetFilters();
              setMobileFilterOpen(false);
            }}
            variant="outline"
            className="h-10 border-white/15 bg-slate-950/55 text-slate-100 hover:bg-slate-800/50"
          >
            Reset
          </Button>
        </div>
      </div>

      {loading ? (
        <div className="rounded-md border border-white/15 bg-slate-950/50 px-4 py-10 text-center text-sm text-slate-400">
          Memuat provider aplikasi...
        </div>
      ) : (
        <DataTable<ProviderTrxRow>
          columns={columns}
          rows={items}
          rowKey={(row) => row.id}
          emptyText="Tidak ada data provider aplikasi."
          pagination={
            shouldShowPagination
              ? {
                  page: currentPage,
                  totalPages: hasNext ? currentPage + 1 : currentPage,
                  onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
                  onNext: () => setOffset((v) => v + PAGE_SIZE),
                  onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE),
                  disablePrev: loading || offset === 0,
                  disableNext: loading || !hasNext,
                }
              : undefined
          }
        />
      )}
    </div>
  );
}
