"use client";

import { useEffect, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";

type AuditRow = {
  transaksi_member_id: number;
  ref_id: string;
  status_member: string;
  transaksi_provider_id: number;
  provider: string;
  status_provider?: string;
  member_created: string;
  provider_created: string;
};

type ProviderItem = {
  id: number;
  nama?: string;
  aktif?: boolean;
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

function toLocalDateInputValue(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

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

export default function AdminStatusMismatchPage() {
  const [rows, setRows] = useState<AuditRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [totalPages, setTotalPages] = useState(1);

  const [providers, setProviders] = useState<ProviderItem[]>([]);
  const [provider, setProvider] = useState("");
  const [statusMember, setStatusMember] = useState("");
  const [mismatchType, setMismatchType] = useState("");
  const [from, setFrom] = useState(() => {
    const now = new Date();
    return toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1));
  });
  const [to, setTo] = useState(() => toLocalDateInputValue(new Date()));
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);
  const activeFilterCount = [provider, statusMember, mismatchType, from, to].filter((v) => v.trim()).length;

  const page = Math.floor(offset / limit) + 1;

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

  async function load(nextOffset = offset) {
    setLoading(true);
    try {
      const qs = new URLSearchParams();
      qs.set("limit", String(limit));
      qs.set("offset", String(nextOffset));
      if (provider.trim()) qs.set("provider", provider.trim());
      if (statusMember.trim()) qs.set("status_member", statusMember.trim());
      if (mismatchType.trim()) qs.set("mismatch_type", mismatchType.trim());
      if (from.trim()) qs.set("from", from.trim());
      if (to.trim()) qs.set("to", to.trim());

      const r = await fetch(`/api/admin/audit/transaksi/status-mismatch?${qs.toString()}`, {
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
    void loadProviders();
    void load(0);
     
  }, []);

  const columns: DataTableColumn<AuditRow>[] = [
    {
      id: "memberCreated",
      header: "Waktu Member",
      tdClassName: "whitespace-nowrap text-slate-100",
      render: (x) => fmtDate(x.member_created),
    },
    {
      id: "providerCreated",
      header: "Waktu Provider",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => fmtDate(x.provider_created),
    },
    {
      id: "ref",
      header: "Ref ID",
      tdClassName: "whitespace-nowrap font-mono text-xs text-slate-300",
      render: (x) => x.ref_id,
    },
    {
      id: "provider",
      header: "Provider",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.provider,
    },
    {
      id: "statusMember",
      header: "Status Member",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.status_member,
    },
    {
      id: "statusProvider",
      header: "Status Provider",
      tdClassName: "whitespace-nowrap text-slate-200",
      render: (x) => x.status_provider || "-",
    },
    {
      id: "memberID",
      header: "Member Trx ID",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => String(x.transaksi_member_id),
    },
    {
      id: "providerID",
      header: "Provider Trx ID",
      tdClassName: "whitespace-nowrap text-slate-300",
      render: (x) => String(x.transaksi_provider_id),
    },
  ];

  return (
    <div className="space-y-4 p-2">
      <div>
        <div className="text-lg font-semibold tracking-tight">Audit Status Mismatch</div>
        <div className="text-sm text-muted-foreground">Perbandingan status transaksi member vs provider terbaru.</div>
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
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-6">
          <select
            value={provider}
            onChange={(e) => setProvider(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          >
            <option value="">semua provider</option>
            {providers.map((p) => (
              <option key={p.id} value={(p.nama || "").trim().toLowerCase()}>
                {p.nama || `provider-${p.id}`}
              </option>
            ))}
          </select>

          <select
            value={statusMember}
            onChange={(e) => setStatusMember(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          >
            <option value="">semua status member</option>
            <option value="success">success</option>
            <option value="pending">pending</option>
            <option value="failed">failed</option>
          </select>

          <select
            value={mismatchType}
            onChange={(e) => setMismatchType(e.target.value)}
            className="h-10 rounded-xl border border-white/15 bg-slate-950/55 px-3 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
          >
            <option value="">semua mismatch</option>
            <option value="member_success_provider_not20">member success / provider bukan 20</option>
            <option value="member_not_success_provider_20">member non-success / provider 20</option>
          </select>

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
                setProvider("");
                setStatusMember("");
                setMismatchType("");
                const now = new Date();
                setFrom(toLocalDateInputValue(new Date(now.getFullYear(), now.getMonth(), 1)));
                setTo(toLocalDateInputValue(now));
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

      <div className="text-sm text-slate-300">Total mismatch: {total}</div>

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(x) => `${x.transaksi_member_id}-${x.transaksi_provider_id}`}
        emptyText={loading ? "Memuat..." : "Tidak ada mismatch."}
        minWidthClassName="min-w-280"
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
