"use client";

import { useEffect, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { fmtID } from "@/lib/format";

type AnomalyRow = {
  id: number;
  dibuat_pada: string;
  provider: string;
  ref_id?: string;
  kode_respon?: string;
  pesan?: string;
  harga?: number;
  tujuan?: string;
  qty?: number;
  payload_hash?: string;
  is_duplicate: boolean;
  is_suspected_fraud: boolean;
  fraud_reason?: string;
  raw_query?: string;
  raw_body?: string;
};

type ProviderItem = {
  id: number;
  nama?: string;
  aktif?: boolean;
};

function authHeader(): Record<string, string> {
  const t = localStorage.getItem("auth_token") || "";
  if (!t) return {};
  return { Authorization: `Bearer ${t}` };
}

const PAGE_SIZE = 10;

export default function AdminProviderAnomalyPage() {
  const [rows, setRows] = useState<AnomalyRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasNext, setHasNext] = useState(false);

  const [provider, setProvider] = useState("");
  const [status, setStatus] = useState("");
  const [refID, setRefID] = useState("");
  const [q, setQ] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const activeFilterCount = [provider, status, refID, q, from, to].filter((v) => v.trim()).length;

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      if (provider) qs.set("provider", provider);
      if (status) qs.set("status", status);
      if (refID) qs.set("ref_id", refID);
      if (q) qs.set("q", q);
      if (from) qs.set("from", from);
      if (to) qs.set("to", to);
      qs.set("limit", String(PAGE_SIZE + 1));
      qs.set("offset", String(nextOffset));

      const r = await fetch(`/api/admin/provider/anomalies?${qs.toString()}`, {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const all: AnomalyRow[] = Array.isArray(j.items) ? j.items : [];
      setHasNext(all.length > PAGE_SIZE);
      setRows(all.slice(0, PAGE_SIZE));
    } finally {
      setLoading(false);
    }
  }

  async function loadProviders() {
    try {
      const r = await fetch("/api/admin/master/provider", {
        headers: authHeader(),
        cache: "no-store",
      });
      const j = await r.json().catch(() => ({}));
      const items: ProviderItem[] = Array.isArray(j?.items) ? j.items : [];
      setProviders(items.filter((x) => x?.aktif !== false));
    } catch {
      setProviders([]);
    }
  }

  useEffect(() => {
    void loadProviders();
    void load(0);
     
  }, []);

  useEffect(() => {
    void load(offset);
     
  }, [offset]);

  const columns: DataTableColumn<AnomalyRow>[] = [
    {
      id: "waktu",
      header: "Waktu",
      tdClassName: "whitespace-nowrap text-slate-100",
      render: (x) => new Date(x.dibuat_pada).toLocaleString("id-ID"),
    },
    {
      id: "provider",
      header: "Provider",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.provider,
    },
    {
      id: "refid",
      header: "RefID",
      tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
      render: (x) => x.ref_id || "-",
    },
    {
      id: "status",
      header: "Status",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => {
        if (x.is_suspected_fraud) return "suspected_fraud";
        if (x.is_duplicate) return "duplicate";
        return "anomali";
      },
    },
    {
      id: "rc",
      header: "RC",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.kode_respon || "-",
    },
    {
      id: "tujuan",
      header: "Tujuan",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.tujuan || "-",
    },
    {
      id: "qty",
      header: "Qty",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => fmtID(x.qty || 0),
    },
    {
      id: "harga",
      header: "Harga",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => `Rp ${fmtID(x.harga || 0)}`,
    },
    {
      id: "fraudReason",
      header: "Fraud Reason",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => x.fraud_reason || "-",
    },
    {
      id: "pesan",
      header: "Pesan",
      tdClassName: "max-w-96 wrap-break-word text-slate-300",
      render: (x) => x.pesan || "-",
    },
  ];

  return (
    <div className="space-y-4 p-2">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="text-lg font-semibold tracking-tight">Transaksi Anomali Provider</div>
          <div className="text-sm text-muted-foreground">Callback provider yang tidak match / terdeteksi anomali.</div>
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
              <span className="rounded-full bg-cyan-500/20 px-2 py-0.5 text-xs font-semibold text-cyan-200">{activeFilterCount}</span>
            ) : null}
          </Button>
        </div>

        <div
          className={`${mobileFilterOpen ? "block" : "hidden"} w-full rounded-2xl border border-white/10 bg-linear-to-br from-slate-900/85 via-slate-900/65 to-slate-800/45 p-3 shadow-[0_22px_48px_-34px_rgba(56,189,248,0.75)] md:block`}
        >
          <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-slate-300">
            <SlidersHorizontal className="h-3.5 w-3.5 text-cyan-300" />
            Filter Anomali
          </div>
          <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7">
            <select
              className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
            >
              <option value="">provider</option>
              {providers.map((p) => (
                <option key={p.id} value={(p.nama || "").trim().toLowerCase()}>
                  {p.nama || `provider-${p.id}`}
                </option>
              ))}
            </select>
            <select
              className="h-10 w-full rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
            >
              <option value="">status</option>
              <option value="duplicate">duplicate</option>
              <option value="fraud">fraud</option>
            </select>
            <Input
              className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40"
              value={refID}
              onChange={(e) => setRefID(e.target.value)}
              placeholder="ref_id"
            />
            <Input
              className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="search pesan/tujuan"
            />
            <Input
              className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
            />
            <Input
              className="h-10 w-full rounded-xl border-white/15 bg-slate-950/55 text-slate-100 placeholder:text-slate-500 focus-visible:ring-cyan-400/40 scheme-dark [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:opacity-90 [&::-webkit-calendar-picker-indicator]:invert"
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
            />
            <div className="grid grid-cols-2 gap-2">
              <Button
                variant="primary"
                className="h-10 w-full"
                onClick={() => {
                  setOffset(0);
                  void load(0);
                  setMobileFilterOpen(false);
                }}
                disabled={loading}
              >
                {loading ? "Loading..." : "Filter"}
              </Button>
              <Button
                variant="outline"
                className="h-10 w-full border-white/20 bg-slate-900/40 text-slate-100 hover:bg-slate-800/50"
                onClick={() => {
                  setProvider("");
                  setStatus("");
                  setRefID("");
                  setQ("");
                  setFrom("");
                  setTo("");
                  setOffset(0);
                  void load(0);
                  setMobileFilterOpen(false);
                }}
                disabled={loading}
              >
                Reset
              </Button>
            </div>
          </div>
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(x) => x.id}
        rowNumberStart={offset + 1}
        minWidthClassName="min-w-[1650px]"
        emptyText="Tidak ada data anomali."
        pagination={{
          page: Math.floor(offset / PAGE_SIZE) + 1,
          totalPages: hasNext ? Math.floor(offset / PAGE_SIZE) + 2 : Math.floor(offset / PAGE_SIZE) + 1,
          onPrev: () => setOffset((v) => Math.max(0, v - PAGE_SIZE)),
          onNext: () => setOffset((v) => v + PAGE_SIZE),
          onPageChange: (nextPage) => setOffset((nextPage - 1) * PAGE_SIZE),
          disablePrev: loading || offset === 0,
          disableNext: loading || !hasNext,
        }}
      />
    </div>
  );
}
